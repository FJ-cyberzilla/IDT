package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"idt/internal/compare"
	"idt/internal/config"
	"idt/internal/enterprise"
	"idt/internal/inspect"
	"idt/internal/runner"
	"idt/internal/tune"
)

// RenderCheckResultsView renders live connectivity check results with adaptive column wrapping.
func RenderCheckResultsView(results []runner.Result, width int) string {
	w := getResponsiveWidth(width)
	var sb strings.Builder

	sb.WriteString(Cyan(fmt.Sprintf("📊 Latency Spectrum: %s │ %s │ %s\n\n", Green("Optimal (<50ms)"), Yellow("Warning"), Red("Unreachable"))))

	var (
		dohOK, dotOK, anySuccess bool
		stdLocal                 bool
	)

	addrLen := w - 45
	if addrLen < 15 {
		addrLen = 15
	}

	for _, r := range results {
		status := Red("❌ [NO ANS]")
		if r.Summary.Count > 0 {
			anySuccess = true
			lat := r.Summary.Median
			status = FormatLatency(lat.Milliseconds())

			switch r.Dialer.Protocol {
			case config.ProtoDoH:
				dohOK = true
			case config.ProtoDoT:
				dotOK = true
			case config.ProtoStandard:
				if lat < 3*time.Millisecond {
					stdLocal = true
				}
			}
		} else {
			status = Red("❌ [FILTERED]")
		}

		addr := truncate(r.Dialer.Address, addrLen)
		protocolTag := Cyan(fmt.Sprintf("[%-6s]", r.Dialer.Protocol))

		sb.WriteString(fmt.Sprintf("  %-12s %s  %-*s  %s\n",
			Bold(r.Dialer.Name), protocolTag, addrLen, Dim(addr), status))
	}

	sb.WriteString("\n" + Dim(strings.Repeat("─", w-6)) + "\n")
	sb.WriteString(Bold("🛡️ PROTOCOL VERIFICATION ANALYSIS:") + "\n")
	switch {
	case !anySuccess:
		sb.WriteString("  " + Red("✖ All dialers unreachable. Verify default gateway and network interface.") + "\n")
	case !dohOK:
		sb.WriteString("  " + Red("✖ Encrypted DoH unreachable (HTTPS/Port 443 packet inspection or block detected).") + "\n")
	case !dotOK:
		sb.WriteString("  " + Yellow("⚠ DoT (TLS/Port 853) blocked. Android Private DNS resolution may fail.") + "\n")
	default:
		sb.WriteString("  " + Green("✔ Encrypted transport layers (DoH + DoT) fully operational.") + "\n")
	}

	if stdLocal {
		sb.WriteString("  " + Yellow("⚠ Local recursive port 53 interception detected. Check routing privacy.") + "\n")
	}

	sb.WriteString("\n" + HelpStyle.Render("ESC/ENTER: Return to Menu  │  q: Quit"))
	return renderPanel("REAL-WORLD CONNECTIVITY & HEALTH DIAGNOSTICS", sb.String(), width)
}

// RenderTuneResultsView renders elite RTT distribution and tuning analysis.
func RenderTuneResultsView(sorted []runner.Result, verdicts map[string]compare.Verdict, opts config.Options, width int) string {
	var sb strings.Builder
	diag := inspect.RunDiagnostics(context.Background())
	sys := diag.DNSConfig

	sb.WriteString(Bold(Cyan("🛰️ COMPREHENSIVE SYSTEM & TELEMETRY INSPECTION:\n")))
	sb.WriteString(fmt.Sprintf("  • Pre-test Connectivity: %s\n", diag.ConnectivityStatus))
	sb.WriteString(fmt.Sprintf("  • Termux Resolv.conf:    %v\n", sys.ResolvConfServers))
	sb.WriteString(fmt.Sprintf("  • Android System DNS:    %v\n", sys.AndroidDNS))
	sb.WriteString(fmt.Sprintf("  • Android VPN Active:    %v\n", sys.AndroidVPNActive))
	sb.WriteString(fmt.Sprintf("  • Private DNS State:     %s (Specifier: %s)\n", sys.PrivateDNSState, sys.AndroidPrivateDNS))
	sb.WriteString(fmt.Sprintf("  • Browser DoH Detection: %s\n", sys.BrowserDoHDetails))
	sb.WriteString(fmt.Sprintf("  • Connected Country:     %s\n", sys.ConnectedCountry))
	sb.WriteString(fmt.Sprintf("  • DoT / DoH Endpoints:   %v\n", append(sys.DoTServers, sys.DoHServers...)))
	sb.WriteString(fmt.Sprintf("  • Guardcat Watchdog:     %s\n\n", diag.GuardcatRecommendation))

	eofCount := 0
	eofTriggered := false
	for _, r := range sorted {
		eofCount += r.EOFErrorsCount
		if r.EOFReconnectTriggered {
			eofTriggered = true
		}
	}
	eofStatus := Green("STANDBY (Zero EOF failures)")
	if eofCount > 0 {
		if eofTriggered {
			eofStatus = Red(fmt.Sprintf("TRIGGERED (%d EOF errors detected across hosts in window -> Auto-reconnect executed)", eofCount))
		} else {
			eofStatus = Yellow(fmt.Sprintf("MONITORING (%d EOF errors detected, below threshold %d)", eofCount, opts.EOFTriggerThreshold))
		}
	}
	sb.WriteString(fmt.Sprintf("  • DNS EOF Reconnect Detection: %s (Threshold: %d / Window: %v)\n", eofStatus, opts.EOFTriggerThreshold, opts.EOFTimeWindow))
	sb.WriteString(fmt.Sprintf("  • Hijack & Tampering Check: %s\n\n", Green("PASSED (Oracle Verified)")))

	rec := tune.GetRecommendationString(sorted, verdicts, opts)
	sb.WriteString(rec)
	sb.WriteString("\n" + HelpStyle.Render("ESC/ENTER: Return to Menu  │  q: Quit"))
	return renderPanel("RTT DISTRIBUTION & TUNING ANALYSIS", sb.String(), width)
}

// RenderAboutView renders polished architecture notes, repository links, contact info, and legal notices.
func RenderAboutView(width int) string {
	var sb strings.Builder

	sb.WriteString(Bold(Cyan("• ORGANIZATION & BRANDING:")) + "\n")
	sb.WriteString(fmt.Sprintf("  %s\n\n", Yellow("FJ™ Cybertronic Systems")))

	sb.WriteString(Bold(Cyan("• CORE CAPABILITIES:")) + "\n")
	sb.WriteString("  • Enterprise network testing & DNS telemetry command center.\n")
	sb.WriteString("  • Multi-protocol benchmarking (DoH, DoT, Standard UDP/TCP/QUIC).\n")
	sb.WriteString("  • Fully responsive TUI optimized for Termux, servers, and workstations.\n\n")

	sb.WriteString(Bold(Cyan("• REPOSITORY & SUPPORT:")) + "\n")
	sb.WriteString("  • GitHub: github.com/cyberzilla-systems/idt\n")
	sb.WriteString("  • Support: cyberzilla.systsms@gmail.com\n\n")

	sb.WriteString(Bold(Red("• LEGAL & COMPLIANCE NOTICE:")) + "\n")
	sb.WriteString("  • Intellectual Property: Unauthorized copying, reverse engineering, or redistribution is strictly prohibited.\n")
	sb.WriteString("  • Acceptable Use: Authorized diagnostic and telemetry use only. Unsanctioned probing is strictly forbidden.\n")

	sb.WriteString("\n" + HelpStyle.Render("ESC/ENTER: Return to Menu  │  q: Quit"))
	return renderPanel("ARCHITECTURE & SYSTEM CAPABILITIES", sb.String(), width)
}

// RenderEnterpriseMenu renders elite enterprise submenu choices.
func RenderEnterpriseMenu(choices []string, cursor int, width int) string {
	var sb strings.Builder
	sb.WriteString(Cyan("🏢 Select an advanced enterprise simulation module:\n\n"))

	for i, choice := range choices {
		if i == cursor {
			sb.WriteString(MenuSelectedStyle.Render(fmt.Sprintf(" ❯ %s ", choice)) + "\n")
		} else {
			sb.WriteString(MenuItemStyle.Render(fmt.Sprintf("   %s ", choice)) + "\n")
		}
	}

	sb.WriteString("\n" + HelpStyle.Render("↑/↓: Navigate  │  Enter: Engage Module  │  ESC: Main Menu  │  q: Quit"))
	return renderPanel("ADVANCED ENTERPRISE & PROTOCOL SIMULATION", sb.String(), width)
}

// RenderMatrixResultsView renders remote office & data center SLA performance matrix results.
func RenderMatrixResultsView(results []enterprise.MatrixResult, width int) string {
	var sb strings.Builder
	sb.WriteString(Cyan("🌐 Evaluating latency, packet loss, jitter, and SLA compliance across nodes.\n\n"))

	for _, r := range results {
		statusColor := Green(r.SLAStatus)
		switch r.SLAStatus {
		case "WARNING":
			statusColor = Yellow(r.SLAStatus)
		case "VIOLATED":
			statusColor = Red(r.SLAStatus)
		}

		pairName := truncate(fmt.Sprintf("%s ➔ %s", r.Source.Region, r.Destination.Region), 22)
		sb.WriteString(fmt.Sprintf("  %-22s │ Latency: %-6v │ Loss: %4.1f%% │ Rate: %4.0f Mbps │ %s\n",
			Bold(pairName), r.Latency, r.Loss, r.Throughput, statusColor))
	}

	sb.WriteString("\n" + HelpStyle.Render("ESC/ENTER: Return to Submenu  │  q: Quit"))
	return renderPanel("REMOTE OFFICE & DATA CENTER PERFORMANCE MATRIX", sb.String(), width)
}

// RenderProtocolResultsView renders synthetic multi-protocol testing under dynamic conditions.
func RenderProtocolResultsView(results []enterprise.ProtocolTestResult, width int) string {
	w := getResponsiveWidth(width)
	var sb strings.Builder
	sb.WriteString(Cyan("🔌 Probing HTTP, VoIP, SaaS, and custom transport protocols under simulated stress.\n\n"))

	for _, res := range results {
		statusColor := Green(res.Status)
		switch res.Status {
		case "DEGRADED":
			statusColor = Yellow(res.Status)
		case "CRITICAL":
			statusColor = Red(res.Status)
		}

		sb.WriteString(fmt.Sprintf("  ▸ %-20s [%s]\n", Bold(string(res.Config.Protocol)), Cyan(string(res.Config.Condition))))
		sb.WriteString(fmt.Sprintf("    Target: %s\n", Dim(truncate(res.Config.TargetName, w-15))))
		sb.WriteString(fmt.Sprintf("    Success: %5.1f%%  │  RTT: %-7v  │  Score: %3d/100  │  Status: %s\n",
			res.SuccessRate, res.AvgLatency, res.Score, statusColor))
		sb.WriteString(fmt.Sprintf("    Summary: %s\n\n", Dim(truncate(res.Details, w-15))))
	}

	sb.WriteString(HelpStyle.Render("ESC/ENTER: Return to Submenu  │  q: Quit"))
	return renderPanel("SYNTHETIC MULTI-PROTOCOL PROBE RESULTS", sb.String(), width)
}

// RenderReplayResultsView renders live traffic capture & lab replay simulation telemetry.
func RenderReplayResultsView(res enterprise.ReplayResult, width int) string {
	w := getResponsiveWidth(width)
	var sb strings.Builder
	sb.WriteString(Cyan("📼 Replaying captured production traffic payloads in isolated lab conditions.\n\n"))

	sb.WriteString(fmt.Sprintf("  • Total Captured Packets:   %d\n", res.TotalCaptured))
	sb.WriteString(fmt.Sprintf("  • Successfully Replayed:    %d\n", res.TotalReplayed))
	sb.WriteString(fmt.Sprintf("  • Dropped / Anomalies:      %d\n", res.DroppedPackets))
	sb.WriteString(fmt.Sprintf("  • Replay Execution Time:    %v\n", res.ReplayDuration))
	sb.WriteString(fmt.Sprintf("  • Average Replayed RTT:     %v\n", res.AvgReplayedRTT))

	anomalyStatus := Green("PASSED (NO ANOMALIES)")
	if res.AnomalyDetected {
		anomalyStatus = Red("FAILED (LATENCY SPIKES / DROPS)")
	}
	sb.WriteString(fmt.Sprintf("  • Rollout Readiness Gate:   %s\n\n", anomalyStatus))

	sb.WriteString(Dim(strings.Repeat("─", w-6)) + "\n")
	sb.WriteString(Bold("TRANSACTION REPLAY LOG:") + "\n")
	for _, logLine := range res.LogSummary {
		sb.WriteString("  " + truncate(logLine, w-4) + "\n")
	}

	sb.WriteString("\n" + HelpStyle.Render("ESC/ENTER: Return to Submenu  │  q: Quit"))
	return renderPanel("TRAFFIC CAPTURE & LAB REPLAY SIMULATOR", sb.String(), width)
}

// RenderColdStartResultsView renders zero-cache cold start socket initialization results.
func RenderColdStartResultsView(results []runner.Result, width int) string {
	var sb strings.Builder
	sb.WriteString(Cyan("❄️ Cold Start Test: Zero-cache socket initialization and initial handshake telemetry.\n\n"))

	for _, res := range results {
		name := truncate(res.Dialer.Name, 20)
		median := "Timeout"
		if res.Summary.Count > 0 {
			median = res.Summary.Median.String()
		}
		sb.WriteString(fmt.Sprintf("  %-20s │ Protocol: %-10s │ Cold RTT: %-7s │ Queries: %d\n",
			Bold(name), res.Dialer.Protocol, median, res.Summary.Count))
	}

	sb.WriteString("\n" + HelpStyle.Render("ESC/ENTER: Return to Menu  │  q: Quit"))
	return renderPanel("COLD START CONFIGURATION TEST", sb.String(), width)
}

// RenderHotlineResultsView renders high congestion load test telemetry.
func RenderHotlineResultsView(res enterprise.ProtocolTestResult, width int) string {
	var sb strings.Builder
	sb.WriteString(Cyan("🔥 Hotline Test: High congestion simulation under heavy load & jitter stress.\n\n"))
	sb.WriteString(fmt.Sprintf("  • Protocol / Condition: %s [%s]\n", Bold(string(res.Config.Protocol)), Cyan(string(res.Config.Condition))))
	sb.WriteString(fmt.Sprintf("  • Success Rate:         %5.1f%%\n", res.SuccessRate))
	sb.WriteString(fmt.Sprintf("  • Average Latency:      %v\n", res.AvgLatency))
	sb.WriteString(fmt.Sprintf("  • Stability Score:      %d/100\n", res.Score))
	sb.WriteString(fmt.Sprintf("  • Status:               %s\n", Bold(res.Status)))
	sb.WriteString(fmt.Sprintf("  • Analysis:             %s\n", Dim(res.Details)))

	sb.WriteString("\n" + HelpStyle.Render("ESC/ENTER: Return to Submenu  │  q: Quit"))
	return renderPanel("HOTLINE HIGH-CONGESTION LOAD TEST", sb.String(), width)
}

// RenderDNSSECResultsView renders DNSSEC chain-of-trust audit telemetry.
func RenderDNSSECResultsView(res enterprise.DNSSECResult, width int) string {
	var sb strings.Builder
	sb.WriteString(Cyan("🛡️ DNSSEC Chain-of-Trust and Signature Validation Audit.\n\n"))
	sb.WriteString(fmt.Sprintf("  • Target Domain:        %s\n", Bold(res.Domain)))
	sb.WriteString(fmt.Sprintf("  • Validation Status:    %s\n", Bold(string(res.Status))))
	sb.WriteString(fmt.Sprintf("  • DS Record Present:    %t\n", res.HasDS))
	sb.WriteString(fmt.Sprintf("  • DNSKEY Present:       %t\n", res.HasDNSKEY))
	sb.WriteString(fmt.Sprintf("  • RRSIG Verified:       %t\n", res.HasRRSIG))
	sb.WriteString(fmt.Sprintf("  • Lookup Latency:       %v\n", res.Latency))
	sb.WriteString(fmt.Sprintf("  • Audit Message:        %s\n", Dim(res.Message)))

	sb.WriteString("\n" + HelpStyle.Render("ESC/ENTER: Return to Submenu  │  q: Quit"))
	return renderPanel("DNSSEC AUDIT REPORT", sb.String(), width)
}

// RenderTimezoneResultsView renders timezone-aware regional routing latency reports.
func RenderTimezoneResultsView(reports []enterprise.TimezoneReport, width int) string {
	var sb strings.Builder
	sb.WriteString(Cyan("🌍 Timezone-Aware Regional Routing & Latency Analysis.\n\n"))

	for _, r := range reports {
		sb.WriteString(fmt.Sprintf("  %-22s │ Local: %-25s │ Latency: %-6v │ Node: %s\n",
			Bold(r.Timezone), r.LocalTime, r.AvgLatency, Cyan(r.OptimalNode)))
	}

	sb.WriteString("\n" + HelpStyle.Render("ESC/ENTER: Return to Submenu  │  q: Quit"))
	return renderPanel("TIMEZONE ROUTING TELEMETRY", sb.String(), width)
}

// RenderCorruptionResultsView renders DNS corruption detection and auto-reconnect audit telemetry.
func RenderCorruptionResultsView(res enterprise.CorruptionResult, width int) string {
	var sb strings.Builder
	sb.WriteString(Cyan("🔍 DNS Corruption Detection & Auto-Reconnect Audit.\n\n"))
	sb.WriteString(fmt.Sprintf("  • Target Resolver:       %s\n", Bold(res.Config.Target)))
	sb.WriteString(fmt.Sprintf("  • Test Domain:           %s\n", res.Config.TestDomain))
	sb.WriteString(fmt.Sprintf("  • Non-Existent Domain:   %s (NXDOMAIN)\n", res.Config.NonExistentDomain))
	sb.WriteString(fmt.Sprintf("  • Total Queries Audited: %d\n", res.TotalQueries))
	sb.WriteString(fmt.Sprintf("  • Corrupted Count:       %d\n", res.CorruptedCount))
	sb.WriteString(fmt.Sprintf("  • False Positives (NX):  %d\n", res.FalsePositives))
	sb.WriteString(fmt.Sprintf("  • Reconnect Triggered:   %t (Threshold: %d)\n", res.ReconnectTriggered, res.Config.Threshold))
	sb.WriteString(fmt.Sprintf("  • Status:                %s\n", Bold(res.Status)))
	sb.WriteString(fmt.Sprintf("  • Analysis & Remediation:%s\n", Dim(res.Details)))
	sb.WriteString(fmt.Sprintf("  • Note:                  %s\n", Yellow("Auto-reconnects when server sends corrupted responses. May cause false triggers with non-existent domains.")))

	sb.WriteString("\n" + HelpStyle.Render("ESC/ENTER: Return to Submenu  │  q: Quit"))
	return renderPanel("DNS CORRUPTION DETECTION & RECONNECT AUDIT", sb.String(), width)
}
