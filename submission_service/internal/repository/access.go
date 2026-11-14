package repository

import (
	"context"
	"innogen-backend/shared/catalog"
)

// ReaderRole uses current database state, never a caller-supplied role header.
func (r *SubmissionRepository) ReaderRole(ctx context.Context, uid string) (bool, error) {
	var active, admin bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1 AND is_active),EXISTS(SELECT 1 FROM users WHERE id=$1 AND is_active AND role='admin')`, uid).Scan(&active, &admin)
	if err != nil {
		return false, err
	}
	if !active {
		return false, catalog.ErrAccountInactive
	}
	return admin, nil
}

func (r *SubmissionRepository) VisibleProblem(ctx context.Context, uid, pid string) (bool, error) {
	var visible bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM problems p JOIN levels l ON l.id=p.level_id JOIN packages pk ON pk.id=l.package_id WHERE p.id=$2 AND ((p.status='published' AND pk.status='published') OR EXISTS(SELECT 1 FROM submissions s WHERE s.user_id=$1 AND s.problem_id=p.id)))`, uid, pid).Scan(&visible)
	return visible, err
}
