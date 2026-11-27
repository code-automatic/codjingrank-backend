// Package audit records administrative and moderation actions.
package audit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AuditEntry is one immutable audit trail record.
// ActorID and ActorEmail identify the administrator or moderator that acted.
type AuditEntry struct {
	ActorID      string
	ActorEmail   string
	Action       string
	ResourceType string
	ResourceID   string
	Detail       any
}

// Write inserts a single audit log row. A nil pool, missing action, or invalid
// detail value is returned as an error; the caller decides whether it is fatal.
func Write(ctx context.Context, pool *pgxpool.Pool, entry AuditEntry) error {
	if pool == nil {
		return errors.New("audit: nil database pool")
	}
	if entry.Action == "" {
		return errors.New("audit: action is required")
	}
	var detail any
	if entry.Detail != nil {
		encoded, err := json.Marshal(entry.Detail)
		if err != nil {
			return fmt.Errorf("audit: encode detail: %w", err)
		}
		detail = encoded
	}
	_, err := pool.Exec(ctx,
		`INSERT INTO audit_logs (actor_id, actor_email, action, resource_type, resource_id, detail)
		 VALUES (NULLIF($1, '')::uuid, NULLIF($2, ''), $3, NULLIF($4, ''), NULLIF($5, ''), $6::jsonb)`,
		entry.ActorID, entry.ActorEmail, entry.Action, entry.ResourceType, entry.ResourceID, detail,
	)
	if err != nil {
		return fmt.Errorf("audit: insert: %w", err)
	}
	return nil
}
