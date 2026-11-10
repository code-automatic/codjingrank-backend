package worker

import (
	"context"
	"fmt"
	"log/slog"

	"innogen-backend/shared/constants"
	"innogen-backend/shared/judge"
	"innogen-backend/shared/languageutil"
	"innogen-backend/shared/piston"
	"innogen-backend/submission_service/internal/queue"
	"innogen-backend/submission_service/internal/repository"
)

// Worker processes submission jobs from the Redis queue.
type Worker struct {
	log          *slog.Logger
	repo         *repository.SubmissionRepository
	pistonClient *piston.Client
	queue        *queue.Queue
}

// New creates a new Worker.
func New(log *slog.Logger, repo *repository.SubmissionRepository, pistonClient *piston.Client, q *queue.Queue) *Worker {
	return &Worker{
		log:          log,
		repo:         repo,
		pistonClient: pistonClient,
		queue:        q,
	}
}

// Run starts the worker loop, blocking until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) {
	w.log.Info("worker started")
	w.reconcileStaleSubmissions(ctx)
	for {
		select {
		case <-ctx.Done():
			w.log.Info("worker stopped")
			return
		default:
			if err := w.processOne(ctx); err != nil {
				w.log.Error("worker error", slog.String("error", err.Error()))
			}
		}
	}
}

// processOne dequeues and processes a single job.
func (w *Worker) processOne(ctx context.Context) error {
	submissionID, err := w.queue.Dequeue(ctx)
	if err != nil {
		return fmt.Errorf("dequeue failed: %w", err)
	}
	if submissionID == "" {
		return nil // timeout, no job
	}

	log := w.log.With(slog.String("submissionId", submissionID))
	log.Info("processing submission")

	if err := w.processJob(ctx, submissionID); err != nil {
		log.Error("job processing failed", slog.String("error", err.Error()))
	}
	return nil
}

// processJob handles the full judging flow for a single submission.
func (w *Worker) processJob(ctx context.Context, submissionID string) error {
	// 1. Load submission
	sub, err := w.repo.FindByID(ctx, submissionID)
	if err != nil {
		return fmt.Errorf("find submission: %w", err)
	}
	if sub == nil {
		w.log.Warn("submission not found", slog.String("submissionId", submissionID))
		return nil
	}

	// Duplicate delivery must not execute or overwrite a persisted verdict.
	claimed, err := w.repo.ClaimSubmission(ctx, submissionID)
	if err != nil {
		return fmt.Errorf("claim submission: %w", err)
	}
	if !claimed {
		return nil
	}

	// 3. Load problem
	problem, err := w.repo.GetProblemByID(ctx, sub.ProblemID)
	if err != nil {
		errMsg := "Judging service unavailable"
		_ = w.repo.UpdateSubmissionResult(ctx, submissionID, constants.StatusInternalError, nil, nil, &errMsg, 0, 0)
		return nil
	}
	if problem == nil {
		errMsg := "Problem not found"
		_ = w.repo.UpdateSubmissionResult(ctx, submissionID, constants.StatusInternalError, nil, nil, &errMsg, 0, 0)
		return nil
	}

	// 4. Load language
	lang, err := w.repo.GetLanguageByID(ctx, sub.LanguageID)
	if err != nil {
		errMsg := "Judging service unavailable"
		_ = w.repo.UpdateSubmissionResult(ctx, submissionID, constants.StatusInternalError, nil, nil, &errMsg, 0, 0)
		return nil
	}
	if lang == nil {
		errMsg := "Language not found"
		_ = w.repo.UpdateSubmissionResult(ctx, submissionID, constants.StatusInternalError, nil, nil, &errMsg, 0, 0)
		return nil
	}

	if err = w.repo.LoadTemplate(ctx, problem, sub.LanguageID); err != nil {
		msg := "Judging configuration unavailable"
		return w.repo.UpdateSubmissionResult(ctx, submissionID, constants.StatusInternalError, nil, nil, &msg, 0, 0)
	}
	// 5. Load test cases
	testCases, err := w.repo.GetTestCasesByProblemID(ctx, sub.ProblemID)
	if err != nil {
		errMsg := "Judging service unavailable"
		_ = w.repo.UpdateSubmissionResult(ctx, submissionID, constants.StatusInternalError, nil, nil, &errMsg, 0, 0)
		return nil
	}

	// 6. Determine file name
	fileName := languageutil.DetermineFileName(lang)
	if problem.SolutionFileName != nil {
		fileName = *problem.SolutionFileName
	}

	// 7. Execute and judge each test case
	var (
		passed           int
		maxRuntimeMs     *int
		errorMessage     *string
		overallStatus    = constants.StatusAccepted
		compilationError bool
		internalError    bool
		timeLimitHit     bool
		runtimeError     bool
		wrongAnswer      bool
	)

	for _, tc := range testCases {
		if compilationError {
			continue
		}

		stdin := ""
		if tc.InputData != nil {
			stdin = *tc.InputData
		}

		codeToRun := sub.Code
		if problem.ExecutionMode == "function" && problem.DriverCode != nil {
			codeToRun = sub.Code + "\n\n" + *problem.DriverCode
		}

		pistonResp, err := w.pistonClient.Execute(ctx, lang.PistonAlias, lang.PistonVersion, fileName, codeToRun, stdin, problem.TimeLimitMs)
		if err != nil {
			internalError = true
			msg := err.Error()
			errorMessage = &msg
			break
		}

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

		jr := judge.Evaluate(tc.ExpectedOutput, compileStderr, runStdout, runStderr, runCode, runSignal, runTime, problem.TimeLimitMs)

		switch jr.Status {
		case constants.StatusAccepted:
			passed++
		case constants.StatusCompilationError:
			compilationError = true
			errorMessage = jr.ErrorMessage
		case constants.StatusTimeLimitExceeded:
			timeLimitHit = true
			if errorMessage == nil {
				errorMessage = jr.ErrorMessage
			}
		case constants.StatusRuntimeError:
			runtimeError = true
			if errorMessage == nil {
				errorMessage = jr.ErrorMessage
			}
		case constants.StatusWrongAnswer:
			wrongAnswer = true
		}

		if jr.RuntimeMs != nil {
			if maxRuntimeMs == nil || *jr.RuntimeMs > *maxRuntimeMs {
				maxRuntimeMs = jr.RuntimeMs
			}
		}
	}

	// 8. Determine overall status
	switch {
	case compilationError:
		overallStatus = constants.StatusCompilationError
	case internalError:
		overallStatus = constants.StatusInternalError
	case timeLimitHit:
		overallStatus = constants.StatusTimeLimitExceeded
	case runtimeError:
		overallStatus = constants.StatusRuntimeError
	case wrongAnswer:
		overallStatus = constants.StatusWrongAnswer
	default:
		overallStatus = constants.StatusAccepted
	}

	// Hidden-test/compiler output can include private inputs or wrapper source.
	if errorMessage != nil {
		safe := overallStatus
		errorMessage = &safe
	}
	// 9. Update submission
	if err := w.repo.UpdateSubmissionResult(ctx, submissionID, overallStatus, maxRuntimeMs, nil, errorMessage, passed, len(testCases)); err != nil {
		return fmt.Errorf("update result: %w", err)
	}

	w.log.Info("submission judged",
		slog.String("submissionId", submissionID),
		slog.String("status", overallStatus),
		slog.Int("passed", passed),
		slog.Int("total", len(testCases)),
	)

	return nil
}

// reconcileStaleSubmissions finds Pending or long-Running submissions and re-enqueues them.
// TODO: Implement proper reconciliation. For MVP, this is a placeholder.
func (w *Worker) reconcileStaleSubmissions(ctx context.Context) {
	w.log.Info("stale submission reconciliation placeholder called")
}
