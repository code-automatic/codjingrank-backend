package judge

import (
	"testing"

	"innogen-backend/run_service/internal/dto"
	"innogen-backend/shared/judge"
)

func intPtr(v int) *int       { return &v }
func strPtr(v string) *string { return &v }

func TestEvaluate_Accepted(t *testing.T) {
	result := judge.Evaluate("hello\n", "", "hello\n", "", 0, nil, 0.5, 0)

	if result.Status != dto.StatusAccepted {
		t.Errorf("expected Status=%q, got %q", dto.StatusAccepted, result.Status)
	}
	if result.ActualOutput != "hello" {
		t.Errorf("expected ActualOutput=%q, got %q", "hello", result.ActualOutput)
	}
	if result.RuntimeMs == nil {
		t.Fatal("expected RuntimeMs to be non-nil")
	}
	if *result.RuntimeMs != 500 {
		t.Errorf("expected RuntimeMs=500, got %d", *result.RuntimeMs)
	}
	if result.ErrorMessage != nil {
		t.Errorf("expected ErrorMessage to be nil, got %v", *result.ErrorMessage)
	}
}

func TestEvaluate_WrongAnswer(t *testing.T) {
	result := judge.Evaluate("expected output", "", "actual output", "", 0, nil, 0.123, 0)

	if result.Status != dto.StatusWrongAnswer {
		t.Errorf("expected Status=%q, got %q", dto.StatusWrongAnswer, result.Status)
	}
	if result.ActualOutput != "actual output" {
		t.Errorf("expected ActualOutput=%q, got %q", "actual output", result.ActualOutput)
	}
	if result.RuntimeMs == nil {
		t.Fatal("expected RuntimeMs to be non-nil")
	}
	if *result.RuntimeMs != 123 {
		t.Errorf("expected RuntimeMs=123, got %d", *result.RuntimeMs)
	}
	if result.ErrorMessage != nil {
		t.Errorf("expected ErrorMessage to be nil, got %v", *result.ErrorMessage)
	}
}

func TestEvaluate_CompilationError(t *testing.T) {
	result := judge.Evaluate("expected", "some compile error", "run stdout", "run stderr", 0, nil, 0.5, 0)

	if result.Status != dto.StatusCompilationError {
		t.Errorf("expected Status=%q, got %q", dto.StatusCompilationError, result.Status)
	}
	if result.ErrorMessage == nil {
		t.Fatal("expected ErrorMessage to be non-nil")
	}
	if *result.ErrorMessage != "some compile error" {
		t.Errorf("expected ErrorMessage=%q, got %q", "some compile error", *result.ErrorMessage)
	}
	if result.ActualOutput != "" {
		t.Errorf("expected ActualOutput to be empty, got %q", result.ActualOutput)
	}
	if result.RuntimeMs != nil {
		t.Errorf("expected RuntimeMs to be nil, got %d", *result.RuntimeMs)
	}
}

func TestEvaluate_CompilationError_EmptyStderr(t *testing.T) {
	result := judge.Evaluate("expected", "   ", "run stdout", "run stderr", 0, nil, 0.5, 0)

	if result.Status != dto.StatusCompilationError {
		t.Errorf("expected Status=%q, got %q", dto.StatusCompilationError, result.Status)
	}
	if result.ErrorMessage == nil {
		t.Fatal("expected ErrorMessage to be non-nil")
	}
	if *result.ErrorMessage != "Compilation error" {
		t.Errorf("expected default ErrorMessage=%q, got %q", "Compilation error", *result.ErrorMessage)
	}
}

func TestEvaluate_RuntimeError_ExitCode(t *testing.T) {
	result := judge.Evaluate("expected", "", "some stdout", "segmentation fault", 1, nil, 0.5, 0)

	if result.Status != dto.StatusRuntimeError {
		t.Errorf("expected Status=%q, got %q", dto.StatusRuntimeError, result.Status)
	}
	if result.ErrorMessage == nil {
		t.Fatal("expected ErrorMessage to be non-nil")
	}
	if *result.ErrorMessage != "segmentation fault" {
		t.Errorf("expected ErrorMessage=%q, got %q", "segmentation fault", *result.ErrorMessage)
	}
	if result.ActualOutput != "some stdout" {
		t.Errorf("expected ActualOutput=%q, got %q", "some stdout", result.ActualOutput)
	}
	if result.RuntimeMs != nil {
		t.Errorf("expected RuntimeMs to be nil, got %d", *result.RuntimeMs)
	}
}

func TestEvaluate_RuntimeError_ExitCode_EmptyStderr(t *testing.T) {
	result := judge.Evaluate("expected", "", "some stdout", "", 1, nil, 0.5, 0)

	if result.Status != dto.StatusRuntimeError {
		t.Errorf("expected Status=%q, got %q", dto.StatusRuntimeError, result.Status)
	}
	if result.ErrorMessage == nil {
		t.Fatal("expected ErrorMessage to be non-nil")
	}
	if *result.ErrorMessage != "Runtime error" {
		t.Errorf("expected default ErrorMessage=%q, got %q", "Runtime error", *result.ErrorMessage)
	}
	if result.ActualOutput != "some stdout" {
		t.Errorf("expected ActualOutput=%q, got %q", "some stdout", result.ActualOutput)
	}
}

func TestEvaluate_RuntimeError_Signal(t *testing.T) {
	signal := "SIGKILL"
	result := judge.Evaluate("expected", "", "partial stdout", "", 0, &signal, 0.5, 0)

	if result.Status != dto.StatusRuntimeError {
		t.Errorf("expected Status=%q, got %q", dto.StatusRuntimeError, result.Status)
	}
	if result.ErrorMessage == nil {
		t.Fatal("expected ErrorMessage to be non-nil")
	}
	if *result.ErrorMessage != "Killed by signal: SIGKILL" {
		t.Errorf("expected ErrorMessage=%q, got %q", "Killed by signal: SIGKILL", *result.ErrorMessage)
	}
	if result.ActualOutput != "partial stdout" {
		t.Errorf("expected ActualOutput=%q, got %q", "partial stdout", result.ActualOutput)
	}
	if result.RuntimeMs != nil {
		t.Errorf("expected RuntimeMs to be nil, got %d", *result.RuntimeMs)
	}
}

func TestEvaluate_RuntimeError_Signal_WinsOverExitCode(t *testing.T) {
	signal := "SIGTERM"
	result := judge.Evaluate("expected", "", "stdout", "stderr", 1, &signal, 0.5, 0)

	if result.Status != dto.StatusRuntimeError {
		t.Errorf("expected Status=%q, got %q", dto.StatusRuntimeError, result.Status)
	}
	if result.ErrorMessage == nil {
		t.Fatal("expected ErrorMessage to be non-nil")
	}
	if *result.ErrorMessage != "Killed by signal: SIGTERM" {
		t.Errorf("expected ErrorMessage=%q, got %q", "Killed by signal: SIGTERM", *result.ErrorMessage)
	}
}

func TestEvaluate_OutputTrimming(t *testing.T) {
	result := judge.Evaluate("hello", "", "  hello  \n\n  ", "", 0, nil, 0.5, 0)

	if result.Status != dto.StatusAccepted {
		t.Errorf("expected Status=%q, got %q", dto.StatusAccepted, result.Status)
	}
	if result.ActualOutput != "hello" {
		t.Errorf("expected trimmed ActualOutput=%q, got %q", "hello", result.ActualOutput)
	}
}

func TestEvaluate_OutputTrimming_ExpectedTrailingSpaces(t *testing.T) {
	result := judge.Evaluate("  world  \n", "", "world", "", 0, nil, 0.5, 0)

	if result.Status != dto.StatusAccepted {
		t.Errorf("expected Status=%q, got %q", dto.StatusAccepted, result.Status)
	}
}

func TestEvaluate_RuntimeMs_Positive(t *testing.T) {
	tests := []struct {
		name     string
		runTime  float64
		expected int
	}{
		{"zero point five seconds", 0.5, 500},
		{"one second", 1.0, 1000},
		{"one hundred twenty three ms", 0.123, 123},
		{"round up", 0.1236, 124},
		{"round down", 0.1234, 123},
		{"very small", 0.001, 1},
		{"large value", 60.0, 60000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := judge.Evaluate("a", "", "a", "", 0, nil, tt.runTime, 0)
			if result.RuntimeMs == nil {
				t.Fatal("expected RuntimeMs to be non-nil")
			}
			if *result.RuntimeMs != tt.expected {
				t.Errorf("expected RuntimeMs=%d, got %d", tt.expected, *result.RuntimeMs)
			}
		})
	}
}

func TestEvaluate_RuntimeMs_ZeroOrNegative(t *testing.T) {
	tests := []struct {
		name    string
		runTime float64
	}{
		{"zero", 0},
		{"negative", -1.5},
		{"negative small", -0.001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := judge.Evaluate("a", "", "a", "", 0, nil, tt.runTime, 0)
			if result.RuntimeMs != nil {
				t.Errorf("expected RuntimeMs to be nil, got %d", *result.RuntimeMs)
			}
		})
	}
}

func TestEvaluate_CompilationError_PrecedesAll(t *testing.T) {
	signal := "SIGINT"
	result := judge.Evaluate("expected", "compile failed", "stdout", "stderr", 1, &signal, 0.5, 0)

	if result.Status != dto.StatusCompilationError {
		t.Errorf("expected Status=%q, got %q", dto.StatusCompilationError, result.Status)
	}
	if result.ErrorMessage == nil {
		t.Fatal("expected ErrorMessage to be non-nil")
	}
	if *result.ErrorMessage != "compile failed" {
		t.Errorf("expected ErrorMessage=%q, got %q", "compile failed", *result.ErrorMessage)
	}
}

func TestEvaluate_EmptyOutput_Accepted(t *testing.T) {
	result := judge.Evaluate("", "", "", "", 0, nil, 0, 0)

	if result.Status != dto.StatusAccepted {
		t.Errorf("expected Status=%q, got %q", dto.StatusAccepted, result.Status)
	}
	if result.ActualOutput != "" {
		t.Errorf("expected ActualOutput to be empty, got %q", result.ActualOutput)
	}
	if result.RuntimeMs != nil {
		t.Errorf("expected RuntimeMs to be nil, got %d", *result.RuntimeMs)
	}
}

func TestEvaluate_NilSignal(t *testing.T) {
	// nil signal with exit code 0 and matching output = Accepted
	result := judge.Evaluate("output", "", "output", "", 0, nil, 0.5, 0)

	if result.Status != dto.StatusAccepted {
		t.Errorf("expected Status=%q, got %q", dto.StatusAccepted, result.Status)
	}
}

func TestEvaluate_EmptySignalString(t *testing.T) {
	// empty string signal with exit code 0 and matching output = Accepted
	empty := ""
	result := judge.Evaluate("output", "", "output", "", 0, &empty, 0.5, 0)

	if result.Status != dto.StatusAccepted {
		t.Errorf("expected Status=%q, got %q", dto.StatusAccepted, result.Status)
	}
}
