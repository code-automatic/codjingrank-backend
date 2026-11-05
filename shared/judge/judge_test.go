package judge

import (
	"testing"

	"innogen-backend/shared/constants"
)

func TestEvaluateJSONNormalization(t *testing.T) {
	tests := []struct {
		name           string
		expectedOutput string
		runStdout      string
		expectedStatus string
	}{
		{
			name:           "exact match",
			expectedOutput: "hello",
			runStdout:      "hello",
			expectedStatus: constants.StatusAccepted,
		},
		{
			name:           "json array match with spaces",
			expectedOutput: "[0,1]",
			runStdout:      "[0, 1]",
			expectedStatus: constants.StatusAccepted,
		},
		{
			name:           "json integer match",
			expectedOutput: "3",
			runStdout:      " 3 ",
			expectedStatus: constants.StatusAccepted,
		},
		{
			name:           "json boolean match",
			expectedOutput: "true",
			runStdout:      "true",
			expectedStatus: constants.StatusAccepted,
		},
		{
			name:           "json mismatch",
			expectedOutput: "[0,1]",
			runStdout:      "[1,0]",
			expectedStatus: constants.StatusWrongAnswer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Evaluate(tt.expectedOutput, "", tt.runStdout, "", 0, nil, 0.1, 1000)
			if res.Status != tt.expectedStatus {
				t.Errorf("expected %s, got %s", tt.expectedStatus, res.Status)
			}
		})
	}
}
