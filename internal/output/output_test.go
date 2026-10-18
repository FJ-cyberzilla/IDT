package output

import (
	"testing"

	"idt/internal/compare"
	"idt/internal/config"
	"idt/internal/runner"
)

func TestColor(t *testing.T) {
	if c := color(10); c != green {
		t.Error("expected green for <30ms")
	}
	if c := color(50); c != yellow {
		t.Error("expected yellow for <80ms")
	}
	if c := color(100); c != red {
		t.Error("expected red for >=80ms")
	}
}

func TestRenderSmoke(t *testing.T) {
	// Smoke test Render doesn't panic
	results := []runner.Result{}
	verdicts := map[string]compare.Verdict{}
	opts := config.DefaultOptions()
	Render(results, verdicts, opts)
}
