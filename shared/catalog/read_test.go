package catalog

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestCatalogReadsPublishedSafeProjection(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := New(pool)
	langs, err := repo.Languages(ctx, "")
	if err != nil || len(langs) == 0 {
		t.Fatalf("seed languages: %v %v", langs, err)
	}
	raw, _ := json.Marshal(langs)
	for _, field := range []string{"pistonAlias", "pistonVersion", "defaultFileName"} {
		if strings.Contains(string(raw), field) {
			t.Fatalf("private runtime leaked %s", raw)
		}
	}
	pkg, err := repo.CreatePackage(ctx, "catalog-read", "Catalog read", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Exec(ctx, "DELETE FROM levels WHERE package_id=$1", pkg.ID)
		pool.Exec(ctx, "DELETE FROM packages WHERE id=$1", pkg.ID)
	})
	page, err := repo.Packages(ctx, 1, 20, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range page.Items {
		if p.ID == pkg.ID {
			t.Fatal("draft package exposed")
		}
	}
	if _, err = pool.Exec(ctx, "UPDATE packages SET status='published',published_at=now() WHERE id=$1", pkg.ID); err != nil {
		t.Fatal(err)
	}
	got, err := repo.PackageBySlug(ctx, pkg.Slug)
	if err != nil || got == nil || got.ID != pkg.ID {
		t.Fatalf("package not found: %+v %v", got, err)
	}
	page, err = repo.Packages(ctx, 1, 20, "catalog-read")
	if err != nil || page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("search page: %+v %v", page, err)
	}
	page, err = repo.Packages(ctx, 2, 20, "catalog-read")
	if err != nil || page.Total != 1 || len(page.Items) != 0 {
		t.Fatalf("empty later page lost total: %+v %v", page, err)
	}
}

func TestProblemProjectionAndTemplate(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := New(pool)
	pkg, err := repo.CreatePackage(ctx, "projection-test", "Projection", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var problemID, languageID string
	t.Cleanup(func() {
		for _, q := range []string{
			`UPDATE problems SET status='draft',published_at=NULL WHERE level_id IN(SELECT id FROM levels WHERE package_id=$1)`,
			`DELETE FROM problem_test_cases WHERE problem_id IN(SELECT id FROM problems WHERE level_id IN(SELECT id FROM levels WHERE package_id=$1))`,
			`DELETE FROM problems WHERE level_id IN(SELECT id FROM levels WHERE package_id=$1)`,
			`DELETE FROM levels WHERE package_id=$1`, `DELETE FROM packages WHERE id=$1`,
		} {
			if _, e := pool.Exec(ctx, q, pkg.ID); e != nil {
				t.Errorf("cleanup: %v", e)
			}
		}
	})
	err = pool.QueryRow(ctx, `INSERT INTO problems(level_id,slug,title,statement_md) SELECT id,'projection-problem','Projection','Public statement' FROM levels WHERE package_id=$1 AND level_no=1 RETURNING id`, pkg.ID).Scan(&problemID)
	if err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `SELECT id FROM languages WHERE is_active ORDER BY slug LIMIT 1`).Scan(&languageID); err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{
		`INSERT INTO problem_languages(problem_id,language_id) VALUES($1,$2)`,
		`INSERT INTO code_templates(problem_id,language_id,starter_code,wrapper_code,file_name) VALUES($1,$2,'public starter','PRIVATE WRAPPER','main.txt')`,
	} {
		if _, err = pool.Exec(ctx, q, problemID, languageID); err != nil {
			t.Fatal(err)
		}
	}
	for _, q := range []string{
		`INSERT INTO problem_samples(problem_id,expected_output) VALUES($1,'PUBLIC SAMPLE')`,
		`INSERT INTO problem_test_cases(problem_id,expected_output) VALUES($1,'HIDDEN OUTPUT')`,
		`UPDATE problems SET status='published',published_at=now() WHERE id=$1`,
	} {
		if _, err = pool.Exec(ctx, q, problemID); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = pool.Exec(ctx, `UPDATE packages SET status='published',published_at=now() WHERE id=$1`, pkg.ID); err != nil {
		t.Fatal(err)
	}
	uid := "00000000-0000-0000-0000-000000000000"
	page, err := repo.Problems(ctx, uid, Filter{Package: pkg.Slug, Page: 1, Size: 20})
	if err != nil || page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("problem page: %+v %v", page, err)
	}
	detail, err := repo.ProblemBySlug(ctx, "projection-problem", uid)
	if err != nil || detail == nil {
		t.Fatalf("detail: %+v %v", detail, err)
	}
	if !detail.Problem.Unlocked || detail.StatementMD != "Public statement" || len(detail.Samples) != 1 || len(detail.Languages) != 1 {
		t.Fatalf("incomplete detail: %+v", detail)
	}
	template, err := repo.Template(ctx, problemID, languageID)
	if err != nil || template == nil {
		t.Fatalf("template: %+v %v", template, err)
	}
	if template.StarterCode != "public starter" || template.FileName != "main.txt" {
		t.Fatalf("template: %+v", template)
	}
	raw, _ := json.Marshal([]any{page, detail, template})
	for _, secret := range []string{"PRIVATE WRAPPER", "HIDDEN OUTPUT", "wrapperCode", "pistonAlias"} {
		if strings.Contains(string(raw), secret) {
			t.Fatalf("leaked %s", secret)
		}
	}
	solved := true
	page, err = repo.Problems(ctx, uid, Filter{Package: pkg.Slug, Solved: &solved, Page: 1, Size: 20})
	if err != nil || page.Total != 0 {
		t.Fatalf("solved filter: %+v %v", page, err)
	}
	if _, err = pool.Exec(ctx, `UPDATE packages SET status='archived' WHERE id=$1`, pkg.ID); err != nil {
		t.Fatal(err)
	}
	detail, err = repo.ProblemBySlug(ctx, "projection-problem", uid)
	if err != nil || detail != nil {
		t.Fatalf("archived package exposed: %+v %v", detail, err)
	}
}
