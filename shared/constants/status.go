// Package constants defines shared status string values used across services.
package constants

// Submission and run statuses.
const (
	StatusAccepted          = "Accepted"
	StatusWrongAnswer       = "Wrong Answer"
	StatusCompilationError  = "Compilation Error"
	StatusRuntimeError      = "Runtime Error"
	StatusTimeLimitExceeded = "Time Limit Exceeded"
	StatusInternalError     = "Internal Error"
	StatusPending           = "Pending"
	StatusRunning           = "Running"
)
