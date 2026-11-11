package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
)

type Page[T any] struct {
	Items []T `json:"items"`
	Page  int `json:"page"`
	Size  int `json:"size"`
	Total int `json:"total"`
}
type Language struct {
	ID             string `json:"id"`
	Slug           string `json:"slug"`
	Name           string `json:"name"`
	RuntimeVersion string `json:"runtimeVersion"`
	FileExtension  string `json:"fileExtension"`
}
type Tag struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}
type Problem struct {
	ID             string  `json:"id"`
	Slug           string  `json:"slug"`
	Title          string  `json:"title"`
	PackageID      string  `json:"packageId"`
	PackageSlug    string  `json:"packageSlug"`
	LevelID        string  `json:"levelId"`
	Difficulty     string  `json:"difficulty"`
	Tags           []Tag   `json:"tags"`
	Solved         bool    `json:"solved"`
	Unlocked       bool    `json:"unlocked"`
	AcceptanceRate float64 `json:"acceptanceRate"`
}
type Sample struct {
	ID             string  `json:"id"`
	InputData      *string `json:"inputData"`
	ExpectedOutput string  `json:"expectedOutput"`
	DisplayOrder   int     `json:"displayOrder"`
}
type Detail struct {
	Problem          Problem    `json:"problem"`
	StatementMD      string     `json:"statementMd"`
	ExecutionMode    string     `json:"executionMode"`
	FunctionName     *string    `json:"functionName"`
	TimeLimitMS      int        `json:"timeLimitMs"`
	MemoryLimitMB    int        `json:"memoryLimitMb"`
	OutputLimitBytes int        `json:"outputLimitBytes"`
	Samples          []Sample   `json:"sampleTestCases"`
	Languages        []Language `json:"languages"`
}
type Template struct {
	ProblemID   string `json:"problemId"`
	LanguageID  string `json:"languageId"`
	StarterCode string `json:"starterCode"`
	FileName    string `json:"fileName"`
}

func (r *Repository) Languages(ctx context.Context, problemID string) ([]Language, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,slug,name,piston_version,file_extension FROM languages WHERE is_active AND ($1='' OR id IN(SELECT language_id FROM problem_languages WHERE problem_id=NULLIF($1,'')::uuid)) ORDER BY slug,id`, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Language{}
	for rows.Next() {
		var l Language
		if err = rows.Scan(&l.ID, &l.Slug, &l.Name, &l.RuntimeVersion, &l.FileExtension); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}
func (r *Repository) Tags(ctx context.Context) ([]Tag, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,slug,name FROM tags ORDER BY slug,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Tag{}
	for rows.Next() {
		var t Tag
		if err = rows.Scan(&t.ID, &t.Slug, &t.Name); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

const packageColumns = `p.id,p.slug,p.name,p.description,p.display_order,(SELECT count(*) FROM problems pr JOIN levels l ON l.id=pr.level_id WHERE l.package_id=p.id AND pr.status='published')`

func (r *Repository) Packages(ctx context.Context, page, size int, q string) (Page[Package], error) {
	out := Page[Package]{Items: []Package{}, Page: page, Size: size}
	const where = ` FROM packages p WHERE p.status='published' AND ($1='' OR p.name ILIKE '%'||$1||'%' OR p.slug ILIKE '%'||$1||'%')`
	if err := r.pool.QueryRow(ctx, `SELECT count(*)`+where, q).Scan(&out.Total); err != nil {
		return out, err
	}
	rows, err := r.pool.Query(ctx, `SELECT `+packageColumns+where+` ORDER BY p.display_order,p.id LIMIT $2 OFFSET $3`, q, size, (page-1)*size)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var p Package
		if err = rows.Scan(&p.ID, &p.Slug, &p.Name, &p.Description, &p.DisplayOrder, &p.TotalProblems); err != nil {
			return out, err
		}
		out.Items = append(out.Items, p)
	}
	return out, rows.Err()
}
func (r *Repository) PackageBySlug(ctx context.Context, slug string) (*Package, error) {
	var p Package
	err := r.pool.QueryRow(ctx, `SELECT `+packageColumns+` FROM packages p WHERE slug=$1 AND status='published'`, slug).Scan(&p.ID, &p.Slug, &p.Name, &p.Description, &p.DisplayOrder, &p.TotalProblems)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &p, err
}

// Every previous published problem must be solved; checking only the previous level would skip gaps.
const unlockedSQL = `NOT EXISTS(SELECT 1 FROM levels prev JOIN problems required ON required.level_id=prev.id AND required.status='published' WHERE prev.package_id=l.package_id AND prev.level_no<l.level_no AND NOT EXISTS(SELECT 1 FROM user_problem_progress pp WHERE pp.problem_id=required.id AND pp.user_id=$1::uuid))`
const problemColumns = `p.id,p.slug,p.title,pk.id,pk.slug,l.id,l.difficulty,
 COALESCE((SELECT jsonb_agg(jsonb_build_object('id',t.id,'slug',t.slug,'name',t.name) ORDER BY t.slug,t.id) FROM problem_tags pt JOIN tags t ON t.id=pt.tag_id WHERE pt.problem_id=p.id),'[]'::jsonb),
 EXISTS(SELECT 1 FROM user_problem_progress pp WHERE pp.problem_id=p.id AND pp.user_id=$1::uuid),` + unlockedSQL + `,COALESCE((SELECT 100.0*count(*) FILTER(WHERE s.status='Accepted')/NULLIF(count(*),0) FROM submissions s WHERE s.problem_id=p.id AND s.status NOT IN('Pending','Running','Internal Error')),0)`
const problemFrom = ` FROM problems p JOIN levels l ON l.id=p.level_id JOIN packages pk ON pk.id=l.package_id WHERE p.status='published' AND pk.status='published'`

type Filter struct {
	Package, LevelID, Difficulty, Tag, Query string
	Solved                                   *bool
	Page, Size                               int
}

func scanProblem(row pgx.Row) (Problem, error) {
	var p Problem
	var tags []byte
	err := row.Scan(&p.ID, &p.Slug, &p.Title, &p.PackageID, &p.PackageSlug, &p.LevelID, &p.Difficulty, &tags, &p.Solved, &p.Unlocked, &p.AcceptanceRate)
	if err == nil {
		err = json.Unmarshal(tags, &p.Tags)
	}
	return p, err
}
func (r *Repository) Problems(ctx context.Context, userID string, f Filter) (Page[Problem], error) {
	out := Page[Problem]{Items: []Problem{}, Page: f.Page, Size: f.Size}
	const filters = ` AND ($2='' OR pk.slug=$2) AND ($3='' OR l.id=NULLIF($3,'')::uuid) AND ($4='' OR l.difficulty::text=$4)
 AND ($5='' OR EXISTS(SELECT 1 FROM problem_tags pt JOIN tags t ON t.id=pt.tag_id WHERE pt.problem_id=p.id AND t.slug=$5))
 AND ($6='' OR p.title ILIKE '%'||$6||'%' OR p.slug ILIKE '%'||$6||'%')
 AND ($7::boolean IS NULL OR EXISTS(SELECT 1 FROM user_problem_progress pp WHERE pp.problem_id=p.id AND pp.user_id=$1::uuid)=$7)`
	args := []any{userID, f.Package, f.LevelID, f.Difficulty, f.Tag, f.Query, f.Solved}
	if err := r.pool.QueryRow(ctx, `SELECT count(*)`+problemFrom+filters, args...).Scan(&out.Total); err != nil {
		return out, err
	}
	args = append(args, f.Size, (f.Page-1)*f.Size)
	rows, err := r.pool.Query(ctx, `SELECT `+problemColumns+problemFrom+filters+` ORDER BY pk.display_order,pk.id,l.level_no,p.created_at,p.id LIMIT $8 OFFSET $9`, args...)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		p, err := scanProblem(rows)
		if err != nil {
			return out, err
		}
		out.Items = append(out.Items, p)
	}
	return out, rows.Err()
}
func (r *Repository) ProblemBySlug(ctx context.Context, slug, userID string) (*Detail, error) {
	p, err := scanProblem(r.pool.QueryRow(ctx, `SELECT `+problemColumns+problemFrom+` AND p.slug=$2`, userID, slug))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	d := &Detail{Problem: p}
	if !p.Unlocked {
		return d, nil
	}
	err = r.pool.QueryRow(ctx, `SELECT statement_md,execution_mode,function_name,time_limit_ms,memory_limit_mb,output_limit_bytes FROM problems WHERE id=$1`, p.ID).Scan(&d.StatementMD, &d.ExecutionMode, &d.FunctionName, &d.TimeLimitMS, &d.MemoryLimitMB, &d.OutputLimitBytes)
	if err != nil {
		return nil, err
	}
	d.Samples, err = r.Samples(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	d.Languages, err = r.Languages(ctx, p.ID)
	return d, err
}
func (r *Repository) Samples(ctx context.Context, problemID string) ([]Sample, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,input_data,expected_output,display_order FROM problem_samples WHERE problem_id=$1 ORDER BY display_order,id`, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Sample{}
	for rows.Next() {
		var s Sample
		if err = rows.Scan(&s.ID, &s.InputData, &s.ExpectedOutput, &s.DisplayOrder); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
func (r *Repository) Template(ctx context.Context, problemID, languageID string) (*Template, error) {
	var t Template
	err := r.pool.QueryRow(ctx, `SELECT ct.problem_id,ct.language_id,ct.starter_code,ct.file_name FROM code_templates ct JOIN languages l ON l.id=ct.language_id AND l.is_active JOIN problem_languages pl ON pl.problem_id=ct.problem_id AND pl.language_id=ct.language_id WHERE ct.problem_id=$1 AND ct.language_id=$2`, problemID, languageID).Scan(&t.ProblemID, &t.LanguageID, &t.StarterCode, &t.FileName)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &t, err
}
