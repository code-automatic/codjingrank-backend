package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

// adminUserJSON deliberately excludes credentials and refresh-token state.
const adminUserJSON = `jsonb_build_object('id',id,'email',email,'username',username,'fullName',full_name,'avatarUrl',avatar_url,'bio',bio,'role',role,'isActive',is_active,'createdAt',created_at,'updatedAt',updated_at)`

// FindUsers returns a page of users ordered by creation time descending.
// q matches email or username, role filters an exact role, and isActive is
// optional. All values are bound as query parameters.
func (r *AdminRepository) FindUsers(ctx context.Context, page, size int, q, role string, isActive *bool) ([]AdminUser, int, error) {
	where := ` WHERE ($1 = '' OR email ILIKE '%' || $1 || '%' ESCAPE '\' OR username ILIKE '%' || $1 || '%' ESCAPE '\')
		AND ($2 = '' OR role::text = $2)
		AND ($3::boolean IS NULL OR is_active = $3)`
	args := []any{escapeLike(q), role, isActive}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM users`+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("admin.FindUsers count: %w", err)
	}

	args = append(args, size, (page-1)*size)
	var raw json.RawMessage
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(jsonb_agg(v ORDER BY c DESC, i DESC), '[]')
		 FROM (SELECT `+adminUserJSON+` v, created_at c, id i FROM users`+where+` ORDER BY created_at DESC, id DESC LIMIT $4 OFFSET $5) x`,
		args...,
	).Scan(&raw)
	if err != nil {
		return nil, 0, fmt.Errorf("admin.FindUsers: %w", err)
	}
	items := make([]AdminUser, 0)
	if err = json.Unmarshal(raw, &items); err != nil {
		return nil, 0, fmt.Errorf("admin.FindUsers decode: %w", err)
	}
	return items, total, nil
}

// FindUserByID returns an admin-safe user projection, or nil when not found.
func (r *AdminRepository) FindUserByID(ctx context.Context, id string) (*AdminUser, error) {
	var raw json.RawMessage
	err := r.pool.QueryRow(ctx, `SELECT `+adminUserJSON+` FROM users WHERE id=$1`, id).Scan(&raw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("admin.FindUserByID: %w", err)
	}
	var user AdminUser
	if err = json.Unmarshal(raw, &user); err != nil {
		return nil, fmt.Errorf("admin.FindUserByID decode: %w", err)
	}
	return &user, nil
}

// UpdateUser applies only the fields present in patch and returns the updated
// user, or nil when the user does not exist. Deactivating a user revokes their
// refresh tokens, and demoting or deactivating the last active admin is
// rejected with ErrLastAdminProtected.
func (r *AdminRepository) UpdateUser(ctx context.Context, id string, patch UpdateUserRequest) (*AdminUser, error) {
	sets := make([]string, 0, 3)
	args := make([]any, 0, 4)
	if patch.Role != nil {
		args = append(args, *patch.Role)
		sets = append(sets, fmt.Sprintf("role=$%d::user_role", len(args)))
	}
	if patch.IsActive != nil {
		args = append(args, *patch.IsActive)
		sets = append(sets, fmt.Sprintf("is_active=$%d", len(args)))
	}
	if patch.FullName != nil {
		args = append(args, nullIfBlank(*patch.FullName))
		sets = append(sets, fmt.Sprintf("full_name=$%d", len(args)))
	}
	if len(sets) == 0 {
		return r.FindUserByID(ctx, id)
	}
	args = append(args, id)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("admin.UpdateUser begin: %w", err)
	}
	defer tx.Rollback(ctx)

	var currentRole string
	var currentActive bool
	if err = tx.QueryRow(ctx, `SELECT role::text, is_active FROM users WHERE id=$1 FOR UPDATE`, id).Scan(&currentRole, &currentActive); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("admin.UpdateUser lock: %w", err)
	}
	nextRole := currentRole
	if patch.Role != nil {
		nextRole = *patch.Role
	}
	nextActive := currentActive
	if patch.IsActive != nil {
		nextActive = *patch.IsActive
	}
	if currentRole == "admin" && currentActive && (nextRole != "admin" || !nextActive) {
		// Lock every active admin so concurrent demotions cannot both pass the
		// "last admin" check and leave the platform without an administrator.
		adminRows, qerr := tx.Query(ctx, `SELECT id::text FROM users WHERE role='admin' AND is_active FOR UPDATE`)
		if qerr != nil {
			return nil, fmt.Errorf("admin.UpdateUser admin lock: %w", qerr)
		}
		activeAdmins := 0
		for adminRows.Next() {
			activeAdmins++
		}
		adminRows.Close()
		if qerr = adminRows.Err(); qerr != nil {
			return nil, fmt.Errorf("admin.UpdateUser admin count: %w", qerr)
		}
		if activeAdmins <= 1 {
			return nil, ErrLastAdminProtected
		}
	}

	var raw json.RawMessage
	err = tx.QueryRow(ctx,
		`UPDATE users SET `+strings.Join(sets, ", ")+fmt.Sprintf(` WHERE id=$%d RETURNING `, len(args))+adminUserJSON,
		args...,
	).Scan(&raw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("admin.UpdateUser: %w", err)
	}
	if currentActive && !nextActive {
		if _, err = tx.Exec(ctx, `UPDATE refresh_tokens SET revoked_at=CURRENT_TIMESTAMP WHERE user_id=$1 AND revoked_at IS NULL`, id); err != nil {
			return nil, fmt.Errorf("admin.UpdateUser revoke refresh tokens: %w", err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("admin.UpdateUser commit: %w", err)
	}
	var user AdminUser
	if err = json.Unmarshal(raw, &user); err != nil {
		return nil, fmt.Errorf("admin.UpdateUser decode: %w", err)
	}
	return &user, nil
}

// ErrLastAdminProtected is returned when a patch would leave the platform
// without any active administrator.
var ErrLastAdminProtected = errors.New("last active admin protected")

// escapeLike makes user input a literal ILIKE pattern by escaping the
// wildcard characters. Callers must use ESCAPE '\' in the query.
func escapeLike(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "%", `\%`)
	value = strings.ReplaceAll(value, "_", `\_`)
	return value
}

func nullIfBlank(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
