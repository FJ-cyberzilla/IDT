package ui

import (
	"context"
	"net"
	"strings"
	"time"

	"idt/internal/config"
	"idt/internal/enterprise"
	"idt/internal/runner"
	"idt/internal/timediag"

	tea "github.com/charmbracelet/bubbletea"
)

func mostCommonIPs(samples [][]string) []net.IP {
	counts := map[string]int{}
	raw := map[string][]net.IP{}
	for _, ips := range samples {
		key := strings.Join(ips, ",")
		counts[key]++
		if _, ok := raw[key]; !ok {
			parsed := make([]net.IP, 0, len(ips))
			for _, s := range ips {
				if ip := net.ParseIP(s); ip != nil {
					parsed = append(parsed, ip)
				}
			}
			raw[key] = parsed
		}
	}
	var best string
	bestN := -1
	for k, n := range counts {
		if n > bestN {
			bestN, best = n, k
		}
	}
	return raw[best]
}

func runColdStartCmd(opts config.Options, dialers []config.Dialer) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		res := runner.Run(ctx, dialers, opts)
		return coldStartCompleteMsg{results: res}
	}
}

func runHotlineCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()
		cfg := enterprise.ProtocolTestConfig{
			Protocol:   enterprise.ProtoVoIP,
			TargetName: "Enterprise SIP Media Relay Hotline",
			Condition:  enterprise.ConditionCongested,
			Iterations: 15,
		}
		reps := enterprise.RunProtocolTests(ctx, []enterprise.ProtocolTestConfig{cfg})
		var rep enterprise.ProtocolTestResult
		if len(reps) > 0 {
			rep = reps[0]
		}
		return hotlineCompleteMsg{report: rep}
	}
}

func runDNSSECCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		res := enterprise.AuditDNSSEC(ctx, "cloudflare-dns.com")
		return dnssecCompleteMsg{result: res}
	}
}

func runTimezoneCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		reps := enterprise.AnalyzeTimezones(ctx)
		return timezoneCompleteMsg{reports: reps}
	}
}

func runCorruptionCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		res := enterprise.RunDNSCorruptionDetection(ctx, enterprise.DefaultCorruptionConfig)
		return corruptionCompleteMsg{result: res}
	}
}

type timeDiagCompleteMsg struct {
	report timediag.WorkflowReport
}

func runTimeDiagCmd(opts config.Options, dialers []config.Dialer) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		engine := timediag.NewNetworkWorkflowEngine(opts, dialers)
		report, _ := engine.Run(ctx)
		return timeDiagCompleteMsg{report: report}
	}
}
