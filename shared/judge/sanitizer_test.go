package judge

import (
	"testing"
)

func TestSanitizeOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty",
			input:    "",
			expected: "",
		},
		{
			name:     "no piston path",
			input:    "ReferenceError: c is not defined\n    at solution.js:2:14",
			expected: "ReferenceError: c is not defined\n    at solution.js:2:14",
		},
		{
			name:     "with piston path",
			input:    "ReferenceError: c is not defined\n    at /piston/jobs/1a2b3c4d5e/solution.js:2:14",
			expected: "ReferenceError: c is not defined\n    at solution.js:2:14",
		},
		{
			name:     "multiple piston paths",
			input:    "Error in /piston/jobs/foo/bar.js\nAt /piston/jobs/foo/baz.js",
			expected: "Error in bar.js\nAt baz.js",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeOutput(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeOutput() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name      string
		stderr    string
		isCompile bool
		runSignal *string
		expected  string
	}{
		{
			name:      "compile error explicit",
			stderr:    "anything",
			isCompile: true,
			runSignal: nil,
			expected:  "compile_error",
		},
		{
			name:      "timeout signal",
			stderr:    "",
			isCompile: false,
			runSignal: strPtr("SIGKILL"),
			expected:  "timeout",
		},
		{
			name:      "runtime error signal",
			stderr:    "",
			isCompile: false,
			runSignal: strPtr("SIGSEGV"),
			expected:  "runtime_error",
		},
		{
			name:      "syntax error in execution",
			stderr:    "SyntaxError: Unexpected token",
			isCompile: false,
			runSignal: nil,
			expected:  "compile_error",
		},
		{
			name:      "reference error",
			stderr:    "ReferenceError: c is not defined",
			isCompile: false,
			runSignal: nil,
			expected:  "runtime_error",
		},
		{
			name:      "type error",
			stderr:    "TypeError: Cannot read properties of undefined",
			isCompile: false,
			runSignal: nil,
			expected:  "runtime_error",
		},
		{
			name:      "out of memory",
			stderr:    "FATAL ERROR: Ineffective mark-compacts near heap limit Allocation failed - JavaScript heap out of memory",
			isCompile: false,
			runSignal: nil,
			expected:  "out_of_memory",
		},
		{
			name:      "internal error",
			stderr:    "some internal error occurred",
			isCompile: false,
			runSignal: nil,
			expected:  "internal_error",
		},
		{
			name:      "unknown error",
			stderr:    "something went wrong",
			isCompile: false,
			runSignal: nil,
			expected:  "runtime_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ClassifyError(tt.stderr, tt.isCompile, tt.runSignal)
			if result != tt.expected {
				t.Errorf("ClassifyError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}
