package service

import (
	"context"
	"errors"

	"innogen-backend/run_service/internal/dto"
	"innogen-backend/run_service/internal/repository"
	"innogen-backend/shared/judge"
	"innogen-backend/shared/languageutil"
	"innogen-backend/shared/models"
	"innogen-backend/shared/piston"
)

// Sentinel errors for the service layer.
var (
	ErrInvalidInput     = errors.New("invalid input")
	ErrProblemNotFound  = errors.New("problem not found")
	ErrLanguageNotFound = errors.New("language not found")
)

// RunService handles the code execution business logic.
type RunService struct {
	repo         *repository.Repository
	pistonClient *piston.Client
}

// New creates a new RunService with the given dependencies.
func New(repo *repository.Repository, pistonClient *piston.Client) *RunService {
	return &RunService{repo: repo, pistonClient: pistonClient}
}

// Run executes the user's code against all sample test cases for the given problem.
func (s *RunService) Run(ctx context.Context, req dto.RunRequest) (*dto.RunResponse, error) {
	if _, msg := req.Validate(); msg != "" {
		return nil, ErrInvalidInput
	}

	// Verify problem exists and get details
	problem, err := s.repo.GetProblemByID(ctx, req.ProblemID)
	if err != nil {
		return nil, err
	}
	if problem == nil {
		return nil, ErrProblemNotFound
	}

	// Load language config
	lang, err := s.repo.GetLanguageByID(ctx, req.LanguageID)
	if err != nil {
		return nil, err
	}
	if lang == nil {
		return nil, ErrLanguageNotFound
	}

	// Load sample test cases
	testCases, err := s.repo.GetSampleTestCases(ctx, req.ProblemID)
	if err != nil {
		return nil, err
	}

	// Determine the file name for code submission
	fileName := languageutil.DetermineFileName(lang)

	// Execute each test case
	results := make([]dto.TestResult, 0, len(testCases))
	passed := 0
	compilationFailed := false
	internalError := false

	for _, tc := range testCases {
		if compilationFailed {
			// All remaining tests get Compilation Error
			results = append(results, dto.TestResult{
				TestCaseID: tc.ID,
				Status:     dto.StatusCompilationError,
				InputData:  inputData(tc),
			})
			continue
		}

		stdin := ""
		if tc.InputData != nil {
			stdin = *tc.InputData
		}

		// Combine code if function mode
		codeToRun := req.Code
		if problem.ExecutionMode == "function" && problem.DriverCode != nil {
			codeToRun = req.Code + "\n\n" + *problem.DriverCode
		}

		// Call Piston
		pistonResp, err := s.pistonClient.Execute(ctx, lang.PistonAlias, lang.PistonVersion, fileName, codeToRun, stdin, 0)
		if err != nil {
			internalError = true
			msg := err.Error()
			results = append(results, dto.TestResult{
				TestCaseID:   tc.ID,
				Status:       dto.StatusInternalError,
				InputData:    inputData(tc),
				ErrorMessage: &msg,
			})
			continue
		}

		// Prepare judge inputs
		var compileStderr string
		if pistonResp.Compile != nil && pistonResp.Compile.Code != 0 {
			compileStderr = pistonResp.Compile.Stderr
		}

		var runStdout, runStderr string
		var runCode int
		var runSignal *string
		var runTime float64
		if pistonResp.Run != nil {
			runStdout = pistonResp.Run.Stdout
			runStderr = pistonResp.Run.Stderr
			runCode = pistonResp.Run.Code
			runSignal = pistonResp.Run.Signal
			runTime = pistonResp.Run.CPUTime
		}

		// Judge the result
		jr := judge.Evaluate(tc.ExpectedOutput, compileStderr, runStdout, runStderr, runCode, runSignal, runTime, 0)

		tr := dto.TestResult{
			TestCaseID:     tc.ID,
			Status:         jr.Status,
			InputData:      inputData(tc),
			ExpectedOutput: tc.ExpectedOutput,
			ActualOutput:   jr.ActualOutput,
			RuntimeMs:      jr.RuntimeMs,
			ErrorMessage:   jr.ErrorMessage,
			ErrorType:      jr.ErrorType,
			Stderr:         jr.Stderr,
		}

		if jr.Status == dto.StatusCompilationError {
			compilationFailed = true
		}

		if jr.Status == dto.StatusAccepted {
			passed++
		}

		results = append(results, tr)
	}

	// Determine overall status
	overallStatus := computeOverallStatus(results, passed, len(testCases), internalError)

	return &dto.RunResponse{
		Status:  overallStatus,
		Passed:  passed,
		Total:   len(testCases),
		Results: results,
	}, nil
}

// inputData returns the input data string from a test case, or empty string if nil.
func inputData(tc models.TestCase) string {
	if tc.InputData != nil {
		return *tc.InputData
	}
	return ""
}

// computeOverallStatus determines the aggregated status from all test results.
func computeOverallStatus(results []dto.TestResult, passed, total int, internalError bool) string {
	if total == 0 {
		return dto.StatusAccepted
	}

	// Check for error statuses in priority order
	hasCompilationError := false
	hasRuntimeError := false
	hasInternalError := internalError

	for _, r := range results {
		switch r.Status {
		case dto.StatusCompilationError:
			hasCompilationError = true
		case dto.StatusRuntimeError:
			hasRuntimeError = true
		case dto.StatusInternalError:
			hasInternalError = true
		}
	}

	if hasCompilationError {
		return dto.StatusCompilationError
	}
	if hasInternalError {
		return dto.StatusInternalError
	}
	if hasRuntimeError {
		return dto.StatusRuntimeError
	}
	if passed == total {
		return dto.StatusAccepted
	}
	return dto.StatusWrongAnswer
}
