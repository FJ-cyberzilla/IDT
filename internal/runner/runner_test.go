package runner

import (
	"context"
	"testing"
	"time"

	"idt/internal/config"
	"idt/internal/stats"
)

func TestNewDialer(t *testing.T) {
	dStandard := config.Dialer{Name: "Test", Protocol: config.ProtoStandard, Address: "1.1.1.1:53"}
	if dl := NewDialer(dStandard, time.Second); dl == nil {
		t.Error("expected non-nil Standard dialer")
	}

	dDoT := config.Dialer{Name: "Test", Protocol: config.ProtoDoT, Address: "one.one.one.one:853"}
	if dl := NewDialer(dDoT, time.Second); dl == nil {
		t.Error("expected non-nil DoT dialer")
	}

	dDoH := config.Dialer{Name: "Test", Protocol: config.ProtoDoH, Address: "https://cloudflare-dns.com/dns-query"}
	if dl := NewDialer(dDoH, time.Second); dl == nil {
		t.Error("expected non-nil DoH dialer")
	}

	dUnknown := config.Dialer{Name: "Test", Protocol: "Unknown", Address: "foo"}
	if dl := NewDialer(dUnknown, time.Second); dl != nil {
		t.Error("expected nil for unknown protocol")
	}
}

func TestSortByMedian(t *testing.T) {
	results := []Result{
		{Dialer: config.Dialer{Name: "Slow"}, Summary: stats.Summary{Count: 3, Median: 100 * time.Millisecond}},
		{Dialer: config.Dialer{Name: "Fast"}, Summary: stats.Summary{Count: 3, Median: 20 * time.Millisecond}},
		{Dialer: config.Dialer{Name: "Dead"}, Summary: stats.Summary{Count: 0}},
	}

	sorted := SortByMedian(results)
	if sorted[0].Dialer.Name != "Fast" {
		t.Errorf("expected Fastest first, got %s", sorted[0].Dialer.Name)
	}
	if sorted[2].Dialer.Name != "Dead" {
		t.Errorf("expected Dead last, got %s", sorted[2].Dialer.Name)
	}
}

func BenchmarkRun(b *testing.B) {
	dialers := make([]config.Dialer, 10)
	for i := range dialers {
		dialers[i] = config.Dialer{
			Name:     "Test",
			Protocol: config.ProtoStandard,
			Address:  "127.0.0.1:1",
		}
	}
	opts := config.Options{
		Dials:   5,
		Timeout: 5 * time.Millisecond,
		Target:  "127.0.0.1:1",
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Run(context.Background(), dialers, opts)
	}
}

