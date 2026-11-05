package judge

import (
	"regexp"
	"strings"
)

// pistonJobPathRegex matches /piston/jobs/<id>/
var pistonJobPathRegex = regexp.MustCompile(`(?m)/piston/jobs/[^/]+/`)

// SanitizeOutput removes internal Piston paths from the output.
func SanitizeOutput(output string) string {
	if output == "" {
		return ""
	}
	// Replace /piston/jobs/<id>/ with nothing, so /piston/jobs/<id>/solution.js becomes solution.js
	return pistonJobPathRegex.ReplaceAllString(output, "")
}

// ClassifyError classifies the error type based on the stderr and run status.
// Types: compile_error, runtime_error, timeout, out_of_memory, internal_error
func ClassifyError(stderr string, isCompile bool, runSignal *string) string {
	if isCompile {
		return "compile_error"
	}
	if runSignal != nil {
		sig := *runSignal
		if sig == "SIGKILL" || sig == "SIGXCPU" {
			return "timeout"
		}
		if sig == "SIGSEGV" || sig == "SIGABRT" {
			return "runtime_error"
		}
	}

	lowerStderr := strings.ToLower(stderr)
	if strings.Contains(lowerStderr, "syntaxerror") {
		// Usually if it fails to parse during execution (e.g. JS eval or similar)
		return "compile_error"
	}
	if strings.Contains(lowerStderr, "out of memory") || strings.Contains(lowerStderr, "heap out of memory") {
		return "out_of_memory"
	}
	if strings.Contains(lowerStderr, "referenceerror") || strings.Contains(lowerStderr, "typeerror") {
		return "runtime_error"
	}
	if strings.Contains(lowerStderr, "internal error") {
		return "internal_error"
	}

	return "runtime_error" // Default fallback for execution errors
}
