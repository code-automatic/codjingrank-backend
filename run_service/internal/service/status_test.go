package service

import (
	"innogen-backend/run_service/internal/dto"
	"testing"
)

func TestOverallStatusPreservesTimeout(t *testing.T) {
	for _, statuses := range [][]string{
		{dto.StatusTimeLimitExceeded},
		{dto.StatusAccepted, dto.StatusTimeLimitExceeded},
		{dto.StatusWrongAnswer, dto.StatusTimeLimitExceeded},
	} {
		var results []dto.TestResult
		passed := 0
		for _, status := range statuses {
			results = append(results, dto.TestResult{Status: status})
			if status == dto.StatusAccepted {
				passed++
			}
		}
		if got := computeOverallStatus(results, passed, len(results), false); got != dto.StatusTimeLimitExceeded {
			t.Errorf("statuses=%v: got %q", statuses, got)
		}
	}
}
