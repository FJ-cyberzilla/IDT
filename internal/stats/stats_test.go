package stats

import (
	"testing"
	"time"
)

func TestComputeEmpty(t *testing.T) {
	s := Compute(nil)
	if s.Count != 0 {
		t.Errorf("expected count 0, got %d", s.Count)
	}
}

func TestComputeBasic(t *testing.T) {
	samples := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		300 * time.Millisecond,
	}
	s := Compute(samples)
	if s.Count != 3 {
		t.Errorf("expected count 3, got %d", s.Count)
	}
	if s.Min != 100*time.Millisecond {
		t.Errorf("expected min 100ms, got %v", s.Min)
	}
	if s.Max != 300*time.Millisecond {
		t.Errorf("expected max 300ms, got %v", s.Max)
	}
	if s.Median != 200*time.Millisecond {
		t.Errorf("expected median 200ms, got %v", s.Median)
	}
}
