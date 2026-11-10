package admin

import (
	"encoding/json"
	"innogen-backend/shared/response"
	"net/http"
)

func (h *v1Handler) clone(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Slug  string `json:"slug"`
		Title string `json:"title"`
	}
	if !decode(w, r, &b) {
		return
	}
	if !slugPattern.MatchString(b.Slug) || !validText(b.Title, 200) {
		response.ErrorSimple(w, 400, "Invalid clone fields")
		return
	}
	tx, err := h.pool.Begin(r.Context())
	if h.fail(w, err) {
		return
	}
	defer tx.Rollback(r.Context())
	if _, err = lockProblem(r.Context(), tx, r.PathValue("id")); h.fail(w, err) {
		return
	}
	var id string
	err = tx.QueryRow(r.Context(), `INSERT INTO problems(level_id,slug,title,statement_md,execution_mode,function_name,time_limit_ms,memory_limit_mb,output_limit_bytes) SELECT level_id,$2,$3,statement_md,execution_mode,function_name,time_limit_ms,memory_limit_mb,output_limit_bytes FROM problems WHERE id=$1 RETURNING id`, r.PathValue("id"), b.Slug, b.Title).Scan(&id)
	if h.fail(w, err) {
		return
	}
	for _, q := range []string{
		`INSERT INTO problem_languages(problem_id,language_id) SELECT $2,language_id FROM problem_languages WHERE problem_id=$1`,
		`INSERT INTO problem_tags(problem_id,tag_id) SELECT $2,tag_id FROM problem_tags WHERE problem_id=$1`,
		`INSERT INTO code_templates(problem_id,language_id,starter_code,wrapper_code,file_name) SELECT $2,language_id,starter_code,wrapper_code,file_name FROM code_templates WHERE problem_id=$1`,
		`INSERT INTO problem_samples(problem_id,input_data,expected_output,display_order) SELECT $2,input_data,expected_output,display_order FROM problem_samples WHERE problem_id=$1`,
		`INSERT INTO problem_test_cases(problem_id,input_data,expected_output,display_order,is_active) SELECT $2,input_data,expected_output,display_order,is_active FROM problem_test_cases WHERE problem_id=$1`,
	} {
		_, err = tx.Exec(r.Context(), q, r.PathValue("id"), id)
		if h.fail(w, err) {
			return
		}
	}
	var raw json.RawMessage
	err = tx.QueryRow(r.Context(), `SELECT `+problemJSON+` FROM problems WHERE id=$1`, id).Scan(&raw)
	if h.fail(w, err) || h.fail(w, tx.Commit(r.Context())) {
		return
	}
	response.Success(w, 201, map[string]any{"problem": raw}, "Problem cloned")
}
