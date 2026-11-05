// Package judge provides output comparison and test result evaluation logic.
package judge

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"

	"innogen-backend/shared/constants"
)

// EvaluateResult holds the outcome of evaluating a single Piston execution.
type EvaluateResult struct {
	Status       string
	ActualOutput string
	RuntimeMs    *int
	ErrorMessage *string
	ErrorType    *string
	Stderr       *string
}

// trimOutput removes leading and trailing whitespace from a string.
func trimOutput(s string) string {
	return strings.TrimSpace(s)
}

// Evaluate compares the Piston execution result against the expected output.
//
// Parameters:
//   - expectedOutput: the expected output from the test case
//   - compileStderr: non-empty if compilation failed
//   - runStdout: stdout from the Piston run stage
//   - runStderr: stderr from the Piston run stage
//   - runCode: exit code from the Piston run stage (0 = success)
//   - runSignal: signal that killed the process (nil if no signal)
//   - runTime: execution time in seconds (0 if not measured)
//   - timeLimitMs: the problem's time limit in ms (0 if no explicit limit)
func Evaluate(expectedOutput, compileStderr, runStdout, runStderr string, runCode int, runSignal *string, runTime float64, timeLimitMs int) EvaluateResult {
	compileStderr = SanitizeOutput(compileStderr)
	runStdout = SanitizeOutput(runStdout)
	runStderr = SanitizeOutput(runStderr)

	// Compilation error
	if compileStderr != "" {
		msg := trimOutput(compileStderr)
		if msg == "" {
			msg = "Compilation error"
		}

		// If compile step failed, check if syntax error
		errType := ClassifyError(msg, true, nil)

		return EvaluateResult{
			Status:       constants.StatusCompilationError,
			ErrorMessage: &msg,
			ErrorType:    &errType,
			Stderr:       &msg,
		}
	}

	// Killed by signal — distinguish TLE from RTE based on timeLimitMs
	if runSignal != nil && *runSignal != "" {
		msg := "Killed by signal: " + *runSignal
		status := constants.StatusRuntimeError
		if timeLimitMs > 0 && (*runSignal == "SIGKILL" || *runSignal == "SIGXCPU") {
			status = constants.StatusTimeLimitExceeded
			msg = "Time Limit Exceeded"
		}

		actual := trimOutput(runStdout)
		errType := ClassifyError(msg, false, runSignal)

		return EvaluateResult{
			Status:       status,
			ActualOutput: actual,
			ErrorMessage: &msg,
			ErrorType:    &errType,
		}
	}

	// Runtime error (non-zero exit code)
	if runCode != 0 {
		msg := trimOutput(runStderr)
		if msg == "" {
			msg = "Runtime error"
		}
		actual := trimOutput(runStdout)

		errType := ClassifyError(msg, false, nil)

		// Extract just the first line for the short message if possible
		shortMsg := msg
		if idx := strings.Index(shortMsg, "\n"); idx != -1 {
			shortMsg = shortMsg[:idx]
		}

		return EvaluateResult{
			Status:       constants.StatusRuntimeError,
			ActualOutput: actual,
			ErrorMessage: &shortMsg,
			ErrorType:    &errType,
			Stderr:       &msg,
		}
	}

	// Successful execution — compare outputs
	actual := trimOutput(runStdout)
	expected := trimOutput(expectedOutput)
	rtMs := runtimeToMs(runTime)

	status := constants.StatusWrongAnswer
	if actual == expected {
		status = constants.StatusAccepted
	} else {
		// Attempt JSON normalization
		var actualJSON, expectedJSON any
		if errA := json.Unmarshal([]byte(actual), &actualJSON); errA == nil {
			if errE := json.Unmarshal([]byte(expected), &expectedJSON); errE == nil {
				if reflect.DeepEqual(actualJSON, expectedJSON) {
					status = constants.StatusAccepted
				}
			}
		}
	}

	return EvaluateResult{
		Status:       status,
		ActualOutput: actual,
		RuntimeMs:    rtMs,
	}
}

// runtimeToMs converts a duration in seconds to milliseconds integer.
// If seconds is 0 or negative, returns nil.
func runtimeToMs(seconds float64) *int {
	if seconds <= 0 {
		return nil
	}
	ms := int(math.Round(seconds * 1000))
	return &ms
}
