package catalog

import (
	"context"
	"errors"
)

var (
	ErrProblemNotFound     = errors.New("problem not found")
	ErrLevelLocked         = errors.New("level locked")
	ErrLanguageUnsupported = errors.New("language unsupported")
	ErrAccountInactive     = errors.New("account inactive or missing")
)

// CheckExecution applies the same learner policy to both Run and Submit.
func (r *Repository) CheckExecution(ctx context.Context, userID, problemID, languageID string) error {
	var active bool
	if err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1 AND is_active)`, userID).Scan(&active); err != nil {
		return err
	}
	if !active {
		return ErrAccountInactive
	}
	var exists, unlocked, supported bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1`+problemFrom+` AND p.id=$2),COALESCE((SELECT `+unlockedSQL+problemFrom+` AND p.id=$2),false),EXISTS(SELECT 1 FROM problem_languages pl JOIN languages l ON l.id=pl.language_id AND l.is_active WHERE pl.problem_id=$2 AND pl.language_id=$3)`, userID, problemID, languageID).Scan(&exists, &unlocked, &supported)
	if err != nil {
		return err
	}
	if !exists {
		return ErrProblemNotFound
	}
	if !unlocked {
		return ErrLevelLocked
	}
	if !supported {
		return ErrLanguageUnsupported
	}
	return nil
}
