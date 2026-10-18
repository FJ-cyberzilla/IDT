package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// EventType represents the category of instrumentation event.
type EventType string

const (
	EventWarmup   EventType = "warmup"
	EventTest     EventType = "test"
	EventSnapshot EventType = "snapshot"
)

// Event represents a single JSONL log record.
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Type      EventType `json:"type"`

	// Warmup fields
	CandidateName string        `json:"candidate_name,omitempty"`
	CandidateAddr string        `json:"candidate_addr,omitempty"`
	Outcome       string        `json:"outcome,omitempty"`
	RTT           time.Duration `json:"rtt,omitempty"`
	Error         string        `json:"error,omitempty"`
	DropReason    string        `json:"drop_reason,omitempty"`

	// Test fields
	ResolverName    string   `json:"resolver_name,omitempty"`
	Domain          string   `json:"domain,omitempty"`
	Field           string   `json:"field,omitempty"`
	AnswersGiven    []string `json:"answers_given,omitempty"`
	AnswersExpected []string `json:"answers_expected,omitempty"`

	// Snapshot fields
	CandidatesIn      int            `json:"candidates_in,omitempty"`
	LiveAfterProbe    int            `json:"live_after_probe,omitempty"`
	LiveAfterBaseline int            `json:"live_after_baseline,omitempty"`
	DropCounts        map[string]int `json:"drop_counts,omitempty"`
}

// Logger defines the interface for recording instrumentation events.
type Logger interface {
	LogWarmup(candidateAddr, candidateName, outcome string, rtt time.Duration, errStr, dropReason string)
	LogTest(resolverName, domain, field, outcome string, given, expected []string, rtt time.Duration, errStr string)
	LogSnapshot(in, afterProbe, afterBaseline int, dropCounts map[string]int)
	Close() error
}

// JSONLFileLogger writes events as JSONL to a file in the runs directory.
type JSONLFileLogger struct {
	mu      sync.Mutex
	file    *os.File
	encoder *json.Encoder
}

// NewJSONLFileLogger creates a new JSONL file logger in the specified runs dir.
func NewJSONLFileLogger(runsDir string) (*JSONLFileLogger, error) {
	if runsDir == "" {
		runsDir = "runs"
	}
	if err := os.MkdirAll(runsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create runs dir: %w", err)
	}

	filename := fmt.Sprintf("run_%s.jsonl", time.Now().Format("20060102_150405"))
	path := filepath.Join(runsDir, filename)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open run log file: %w", err)
	}

	return &JSONLFileLogger{
		file:    f,
		encoder: json.NewEncoder(f),
	}, nil
}

func (l *JSONLFileLogger) writeEvent(e Event) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		return
	}
	_ = l.encoder.Encode(e)
}

func (l *JSONLFileLogger) LogWarmup(addr, name, outcome string, rtt time.Duration, errStr, dropReason string) {
	l.writeEvent(Event{
		Timestamp:     time.Now(),
		Type:          EventWarmup,
		CandidateAddr: addr,
		CandidateName: name,
		Outcome:       outcome,
		RTT:           rtt,
		Error:         errStr,
		DropReason:    dropReason,
	})
}

func (l *JSONLFileLogger) LogTest(resolverName, domain, field, outcome string, given, expected []string, rtt time.Duration, errStr string) {
	l.writeEvent(Event{
		Timestamp:       time.Now(),
		Type:            EventTest,
		ResolverName:    resolverName,
		Domain:          domain,
		Field:           field,
		Outcome:         outcome,
		AnswersGiven:    given,
		AnswersExpected: expected,
		RTT:             rtt,
		Error:           errStr,
	})
}

func (l *JSONLFileLogger) LogSnapshot(in, afterProbe, afterBaseline int, dropCounts map[string]int) {
	l.writeEvent(Event{
		Timestamp:         time.Now(),
		Type:              EventSnapshot,
		CandidatesIn:      in,
		LiveAfterProbe:    afterProbe,
		LiveAfterBaseline: afterBaseline,
		DropCounts:        dropCounts,
	})
}

func (l *JSONLFileLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		err := l.file.Close()
		l.file = nil
		return err
	}
	return nil
}

// NoopLogger is a dummy logger for testing or when logging is disabled.
type NoopLogger struct{}

func (NoopLogger) LogWarmup(string, string, string, time.Duration, string, string) {}
func (NoopLogger) LogTest(string, string, string, string, []string, []string, time.Duration, string) {
}
func (NoopLogger) LogSnapshot(int, int, int, map[string]int) {}
func (NoopLogger) Close() error                              { return nil }
