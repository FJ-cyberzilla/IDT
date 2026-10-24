package timediag

import (
	"context"
	"fmt"
	"sync"
	"time"

	"idt/internal/config"
	"idt/internal/dialer"
	"idt/internal/errorhandling"
	"idt/internal/guardcat"
	"idt/internal/helper"
	"idt/internal/runner"
)

// ChronologicalEvent represents a timestamped step in the network test sequence,
// immediately communicating both the chronological aspect and the diagnostic purpose.
type ChronologicalEvent struct {
	Timestamp     time.Time                         `json:"timestamp"`
	SequenceIndex int                               `json:"sequence_index"`
	PhaseName     string                            `json:"phase_name"`
	Target        string                            `json:"target"`
	Protocol      string                            `json:"protocol"`
	Duration      time.Duration                     `json:"duration"`
	Success       bool                              `json:"success"`
	ErrorDetails  errorhandling.NetworkErrorDetails `json:"error_details"`
	DiagnosticMsg string                            `json:"diagnostic_msg"`
}

// WorkflowReport summarizes the complete chronological and diagnostic run.
type WorkflowReport struct {
	StartTime      time.Time            `json:"start_time"`
	EndTime        time.Time            `json:"end_time"`
	TotalDuration  time.Duration        `json:"total_duration"`
	Events         []ChronologicalEvent `json:"events"`
	RunnerResults  []runner.Result      `json:"runner_results"`
	OverallStatus  string               `json:"overall_status"`
	SummaryVerdict string               `json:"summary_verdict"`
	GuardcatReport string               `json:"guardcat_report"`
}

// NetworkWorkflowEngine orchestrates synchronized, chronological network and DNS diagnostic test sequences.
type NetworkWorkflowEngine struct {
	mu       sync.RWMutex
	opts     config.Options
	dialers  []config.Dialer
	events   []ChronologicalEvent
	guardcat *guardcat.Guardcat
}

// NewNetworkWorkflowEngine creates a new NetworkWorkflowEngine instance with Guardcat monitoring.
func NewNetworkWorkflowEngine(opts config.Options, dialers []config.Dialer) *NetworkWorkflowEngine {
	if len(dialers) == 0 {
		dialers = config.NormalizeDialers(config.DefaultDialers())
	}
	return &NetworkWorkflowEngine{
		opts:     opts,
		dialers:  dialers,
		events:   make([]ChronologicalEvent, 0),
		guardcat: guardcat.New(guardcat.DefaultConfig),
	}
}

// Run executes the synchronized chronological network workflow and DNS test sequences
// with panic recovery via helper.SafeRun, retry logic via helper.Retry, structured
// error handling via errorhandling.ResolveError, and Guardcat monitoring & circuit breaking.
func (e *NetworkWorkflowEngine) Run(ctx context.Context) (WorkflowReport, error) {
	startTime := time.Now()

	// 1. Pre-test state: DNS Resolution Readiness & local socket stack verification
	readinessErr := e.guardcat.EnsureReadiness(ctx, "1.1.1.1:53")
	if readinessErr != nil {
		// Log or note pre-test warning but allow workflow to attempt execution
	}

	var events []ChronologicalEvent
	var seqMu sync.Mutex
	var wg sync.WaitGroup

	sem := make(chan struct{}, 4) // Bounded concurrency for diagnostic probe sequencing

	for idx, d := range e.dialers {
		e.guardcat.RegisterEndpoint(d.Name, d.Address, d.Protocol)

		// Circuit breaker check: skip or downgrade if endpoint is marked unhealthy
		if !e.guardcat.IsHealthy(d.Address, d.Protocol) {
			seqMu.Lock()
			events = append(events, ChronologicalEvent{
				Timestamp:     time.Now(),
				SequenceIndex: idx,
				PhaseName:     fmt.Sprintf("Seq-%02d-%s-%s-BYPASSED", idx+1, d.Name, d.Protocol),
				Target:        d.Address,
				Protocol:      d.Protocol,
				Duration:      0,
				Success:       false,
				DiagnosticMsg: "Guardcat Circuit Breaker: Endpoint bypassed due to consecutive failures/EOF threshold.",
			})
			seqMu.Unlock()
			continue
		}

		select {
		case <-ctx.Done():
			return WorkflowReport{}, ctx.Err()
		case sem <- struct{}{}:
		}

		wg.Add(1)
		go func(sequenceNum int, dl config.Dialer) {
			defer func() { <-sem }()
			defer wg.Done()

			stepStart := time.Now()
			var answer dialer.Answer
			var dialErr error

			// Safe execution with panic recovery
			evalErr := helper.SafeRun(func() error {
				// Retry operation with jittered backoff
				retryErr := helper.Retry(ctx, 2, 50*time.Millisecond, 500*time.Millisecond, func(c context.Context) error {
					var innerErr error
					answer, innerErr = runner.Dial(c, dl, e.opts.Target, e.opts.Timeout)
					return innerErr
				})
				dialErr = retryErr
				return nil
			})

			if evalErr != nil {
				dialErr = evalErr
			}

			stepDuration := time.Since(stepStart)
			errDetails := errorhandling.ResolveError(dialErr, dl.Address)

			success := dialErr == nil && len(answer.IPs) > 0
			if success {
				e.guardcat.RecordSuccess(dl.Address, dl.Protocol)
			} else {
				e.guardcat.RecordFailure(dl.Address, dl.Protocol, dialErr)
			}

			diagMsg := "Chronological Sync OK: Transport verified successfully."
			if !success {
				diagMsg = fmt.Sprintf("Diagnostic Alert [%s]: %s", errDetails.Category, errDetails.ActionableHint)
			}

			event := ChronologicalEvent{
				Timestamp:     stepStart,
				SequenceIndex: sequenceNum,
				PhaseName:     fmt.Sprintf("Seq-%02d-%s-%s", sequenceNum+1, dl.Name, dl.Protocol),
				Target:        dl.Address,
				Protocol:      dl.Protocol,
				Duration:      stepDuration,
				Success:       success,
				ErrorDetails:  errDetails,
				DiagnosticMsg: diagMsg,
			}

			seqMu.Lock()
			events = append(events, event)
			seqMu.Unlock()
		}(idx, d)
	}

	wg.Wait()
	endTime := time.Now()

	// Execute core runner results for dashboard comparison matrix
	runnerResults := runner.Run(ctx, e.dialers, e.opts)

	successCount := 0
	for _, ev := range events {
		if ev.Success {
			successCount++
		}
	}

	overallStatus := "✅ Time-Diag Workflow Completed Successfully"
	verdict := "Chronological sequence verified; secure transport channels operational."

	if successCount == 0 {
		overallStatus = "❌ Time-Diag Workflow Failure: All Endpoints Unreachable"
		verdict = "Name resolution and network connectivity are completely blocked."
	} else if successCount < len(e.dialers) {
		overallStatus = "⚠️ Time-Diag Workflow Warning: Partial Reachability"
		verdict = "Some DNS protocols or endpoints experienced degradation or timeout."
	}

	guardcatReport := e.guardcat.RecommendAction()
	if readinessErr != nil {
		guardcatReport = fmt.Sprintf("[PRE-TEST WARNING] Socket readiness probe notice: %v | %s", readinessErr, guardcatReport)
	}

	report := WorkflowReport{
		StartTime:      startTime,
		EndTime:        endTime,
		TotalDuration:  endTime.Sub(startTime),
		Events:         events,
		RunnerResults:  runnerResults,
		OverallStatus:  overallStatus,
		SummaryVerdict: verdict,
		GuardcatReport: guardcatReport,
	}

	e.mu.Lock()
	e.events = events
	e.mu.Unlock()

	return report, nil
}
