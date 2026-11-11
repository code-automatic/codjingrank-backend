// Package catalog implements the fresh UUID catalog and progression read model.
package catalog

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

type Package struct {
	ID            string  `json:"id"`
	Slug          string  `json:"slug"`
	Name          string  `json:"name"`
	Description   *string `json:"description"`
	DisplayOrder  int     `json:"displayOrder"`
	TotalProblems int     `json:"totalProblems"`
}
type LevelProgress struct {
	ID         string `json:"id"`
	PackageID  string `json:"packageId"`
	LevelNo    int    `json:"levelNo"`
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	Difficulty string `json:"difficulty"`
	Total      int    `json:"total"`
	Solved     int    `json:"solved"`
	Unlocked   bool   `json:"unlocked"`
	Completed  bool   `json:"completed"`
}

// CreatePackage creates the package and its fixed three levels in one transaction.
func (r *Repository) CreatePackage(ctx context.Context, slug, name string, description *string, displayOrder int) (*Package, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	p := &Package{}
	err = tx.QueryRow(ctx, `INSERT INTO packages(slug,name,description,display_order) VALUES($1,$2,$3,$4) RETURNING id,slug,name,description,display_order`, slug, name, description, displayOrder).Scan(&p.ID, &p.Slug, &p.Name, &p.Description, &p.DisplayOrder)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO levels(package_id,level_no,slug,name,difficulty,display_order)
 SELECT $1,n,slug,name,difficulty::difficulty_level,n FROM (VALUES(1,'easy','Easy','easy'),(2,'medium','Medium','medium'),(3,'hard','Hard','hard')) v(n,slug,name,difficulty)`, p.ID)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return p, nil
}

// Levels derives unlock from all preceding levels; empty levels never block.
func (r *Repository) Levels(ctx context.Context, packageID, userID string) ([]LevelProgress, error) {
	rows, err := r.pool.Query(ctx, `SELECT l.id,l.package_id,l.level_no,l.slug,l.name,l.difficulty,
 count(p.id),count(upp.id) FROM levels l
 LEFT JOIN problems p ON p.level_id=l.id AND p.status='published'
 LEFT JOIN user_problem_progress upp ON upp.problem_id=p.id AND upp.user_id=$2::uuid
 WHERE l.package_id=$1::uuid GROUP BY l.id ORDER BY l.level_no`, packageID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	levels := []LevelProgress{}
	unlocked := true
	for rows.Next() {
		var l LevelProgress
		if err = rows.Scan(&l.ID, &l.PackageID, &l.LevelNo, &l.Slug, &l.Name, &l.Difficulty, &l.Total, &l.Solved); err != nil {
			return nil, err
		}
		l.Unlocked = unlocked
		l.Completed = l.Solved == l.Total
		unlocked = unlocked && l.Completed
		levels = append(levels, l)
	}
	return levels, rows.Err()
}
