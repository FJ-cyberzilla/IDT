package timediag

import (
	"context"
	"testing"
	"time"

	"idt/internal/config"
)

func TestNetworkWorkflowEngineRun(t *testing.T) {
	opts := config.DefaultOptions()
	opts.Timeout = 1 * time.Second
	opts.Dials = 1

	dialers := []config.Dialer{
		{"Cloudflare", config.ProtoDoH, "https://cloudflare-dns.com/dns-query"},
		{"Google", config.ProtoDoT, "dns.google:853"},
	}

	engine := NewNetworkWorkflowEngine(opts, dialers)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	report, err := engine.Run(ctx)
	if err != nil {
		t.Fatalf("expected no error from engine Run, got %v", err)
	}

	if report.TotalDuration <= 0 {
		t.Errorf("expected positive total duration, got %v", report.TotalDuration)
	}

	if len(report.Events) != len(dialers) {
		t.Errorf("expected %d chronological events, got %d", len(dialers), len(report.Events))
	}

	if report.OverallStatus == "" {
		t.Errorf("expected non-empty overall status")
	}
}
