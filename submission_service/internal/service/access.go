package service

import "context"

func (s *SubmissionService) ReaderRole(ctx context.Context, uid string) (bool, error) {
	return s.repo.ReaderRole(ctx, uid)
}
func (s *SubmissionService) VisibleProblem(ctx context.Context, uid, pid string) (bool, error) {
	return s.repo.VisibleProblem(ctx, uid, pid)
}
