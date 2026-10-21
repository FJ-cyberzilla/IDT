package ui

import (
	"fmt"
	"strings"
	"time"

	"idt/internal/compare"
	"idt/internal/config"
	"idt/internal/enterprise"
	"idt/internal/runner"
	"idt/internal/tune"

	"github.com/charmbracelet/lipgloss"
)

// RenderPhaseBadge returns a high-contrast, elite status badge for the current dashboard phase.
func RenderPhaseBadge(stateName string, phase string) string {
	switch phase {
	case "MENU":
		return PhaseMenuStyle.Render(" 🟢 STATUS: MAIN COMMAND CENTER • READY ")
	case "WARMUP":
		return PhaseInitStyle.Render(" 🟡 STATUS: SYSTEM VERIFICATION & WARMUP ")
	case "RUNNING":
		return PhaseRunningStyle.Render(" 🔵 STATUS: EXECUTING MULTI-PROTOCOL DIALERS ")
	case "RESULTS":
		return PhaseResultsStyle.Render(" ✅ STATUS: TELEMETRY & VERDICTS COMPUTED ")
	case "SERVERS":
		return PhaseServersStyle.Render(" ⚙ STATUS: ACTIVE SERVERS & RESOLVER TOPOLOGY ")
	case "CHECK":
		return PhaseCheckStyle.Render(" 🔍 STATUS: REAL-WORLD CONNECTIVITY DIAGNOSTICS ")
	case "TUNE":
		return PhaseTuneStyle.Render(" 📈 STATUS: RTT DISTRIBUTION & OPTIMIZATION ")
	case "ABOUT":
		return PhaseAboutStyle.Render(" ℹ️ STATUS: ARCHITECTURE & SYSTEM CAPABILITIES ")
	case "ERROR":
		return PhaseErrorStyle.Render(" 🚨 CRITICAL: NETWORK STALLED OR OFFLINE ")
	default:
		return PhaseMenuStyle.Render(fmt.Sprintf(" 🔵 STATUS: %s ", stateName))
	}
}

// RenderWarmupProgress renders a sleek real-time animated warm-up checklist with precise step tracking and elapsed time.
func RenderWarmupProgress(step warmupStep, spinnerView string, connType string, width int, elapsed time.Duration) string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(" SYSTEM INITIALIZATION & PREFLIGHT ") + "\n\n")

	// Helper for step items
	renderStep := func(currentStep warmupStep, targetStep warmupStep, label string, activeText string) {
		if step > targetStep {
			sb.WriteString(fmt.Sprintf("  %s %s\n", Green("✔"), label))
		} else if step == targetStep {
			elapsedStr := fmt.Sprintf("[%02d:%02d]", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
			sb.WriteString(fmt.Sprintf("  %s %s %s\n", spinnerView, lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(activeText), Dim(elapsedStr)))
		} else {
			sb.WriteString(fmt.Sprintf("  %s %s\n", Dim("•"), Dim(label)))
		}
	}

	ct := connType
	if ct == "" {
		ct = "Normal ISP Connection"
	}

	renderStep(step, stepInit, "Initialize Command Center & Config", "Initializing command center & configuration...")
	renderStep(step, stepNetwork, fmt.Sprintf("Network Route & VPN Detection [%s]", ct), "Detecting network topology & active tunnels...")
	renderStep(step, stepPreflight, "Public Resolver Reachability (1.1.1.1 / 8.8.8.8)", "Probing public resolvers...")
	renderStep(step, stepReady, "Secure Connection Pools & Resolvers Prepared", "Allocating secure connection pools...")

	sb.WriteString("\n" + HelpStyle.Render("────────────────────────────────────────\n[ESC] Cancel Operation • [q] Quit"))
	return sb.String()
}

// RenderMenuChoices renders elite interactive menu options with custom selection cursors and subtle borders.
func RenderMenuChoices(choices []string, cursor int, width int) string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(" SELECT COMMAND CENTER OPERATION ") + "\n\n")

	for i, choice := range choices {
		if i == cursor {
			sb.WriteString(MenuSelectedStyle.Render(fmt.Sprintf(" ❯ %s ", choice)) + "\n")
		} else {
			sb.WriteString(MenuItemStyle.Render(fmt.Sprintf("   %s ", choice)) + "\n")
		}
	}

	sb.WriteString("\n" + HelpStyle.Render("[↑/↓] Navigate • [Enter] Execute • [ESC] Menu • [q] Quit"))
	return sb.String()
}

// RenderErrorCard renders a striking attention card for offline or stalled states.
func RenderErrorCard(reason string, width int) string {
	errMsg := fmt.Sprintf("CRITICAL NETWORK FAULT\n\n%s\n\nReason: %s\n\nAction Required: Check Wi-Fi, cellular data, firewall rules, or VPN tunnel status.",
		Red("✖ Preflight connectivity check failed against public resolvers (1.1.1.1 / 8.8.8.8)."), reason)
	return RedAttentionStyle.Render(errMsg) + "\n\n" + HelpStyle.Render("[r] Retry Preflight • [ESC/Enter] Return to Menu • [q] Quit")
}

// RenderCheckResultsView renders live connectivity check results with adaptive column wrapping.
func RenderCheckResultsView(results []runner.Result, width int) string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(" 🔍 REAL-WORLD CONNECTIVITY & HEALTH DIAGNOSTICS ") + "\n\n")
	sb.WriteString(Cyan(fmt.Sprintf("Spectrum: %s | %s | %s\n\n", Green("Optimal (<50ms)"), Yellow("Warning"), Red("Unreachable"))))

	var (
		dohOK, dotOK, anySuccess bool
		stdLocal                 bool
	)

	addrLen := 38
	if width > 0 && width < 75 {
		addrLen = width - 30
		if addrLen < 12 {
			addrLen = 12
		}
	}

	for _, r := range results {
		status := Red("❌ no ans")
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
			status = Red("❌ filtered")
		}

		addr := r.Dialer.Address
		if len(addr) > addrLen {
			addr = addr[:addrLen-3] + "..."
		}

		protocolTag := Cyan(fmt.Sprintf("[%-6s]", r.Dialer.Protocol))
		sb.WriteString(fmt.Sprintf("  %-12s %s  %-*s  %s\n",
			Bold(r.Dialer.Name), protocolTag, addrLen, Dim(addr), status))
	}

	sb.WriteString("\n" + Bold("PROTOCOL VERIFICATION ANALYSIS:") + "\n")
	switch {
	case !anySuccess:
		sb.WriteString("  " + Red("✖ All dialers unreachable. Verify gateway and network interface.") + "\n")
	case !dohOK:
		sb.WriteString("  " + Red("✖ Encrypted DoH unreachable (HTTPS/443 inspection or block detected).") + "\n")
	case !dotOK:
		sb.WriteString("  " + Yellow("⚠ DoT (TLS/853) blocked. Android Private DNS resolution may fail.") + "\n")
	default:
		sb.WriteString("  " + Green("✔ Encrypted transport layers (DoH + DoT) fully operational.") + "\n")
	}

	if stdLocal {
		sb.WriteString("  " + Yellow("⚠ Local recursive port 53 interception detected.") + "\n")
	}

	sb.WriteString(HelpStyle.Render("\n[ESC/Enter] Return to Menu • [q] Quit"))
	return sb.String()
}

// RenderTuneResultsView renders elite RTT distribution and tuning analysis.
func RenderTuneResultsView(sorted []runner.Result, verdicts map[string]compare.Verdict, opts config.Options, width int) string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(" 📈 RTT DISTRIBUTION & TUNING ANALYSIS ") + "\n\n")
	rec := tune.GetRecommendationString(sorted, verdicts, opts)
	sb.WriteString(rec)
	sb.WriteString(HelpStyle.Render("\n[ESC/Enter] Return to Menu • [q] Quit"))
	return sb.String()
}

// RenderAboutView renders polished architecture notes, repository links, contact info, and legal notices.
func RenderAboutView(width int) string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(" ℹ️ ARCHITECTURE & SYSTEM CAPABILITIES ") + "\n\n")

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
	sb.WriteString("  • Acceptable Use: Authorized diagnostic and telemetry use only. Unsanctioned probing is strictly forbidden.")

	sb.WriteString(HelpStyle.Render("\n[ESC/Enter] Return to Menu • [q] Quit"))
	return sb.String()
}

// RenderEnterpriseMenu renders elite enterprise submenu choices.
func RenderEnterpriseMenu(choices []string, cursor int, width int) string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(" 🏢 ADVANCED ENTERPRISE & PROTOCOL SIMULATION ") + "\n\n")
	for i, choice := range choices {
		if i == cursor {
			sb.WriteString(MenuSelectedStyle.Render(fmt.Sprintf(" ❯ %s ", choice)) + "\n")
		} else {
			sb.WriteString(MenuItemStyle.Render(fmt.Sprintf("   %s ", choice)) + "\n")
		}
	}
	sb.WriteString(HelpStyle.Render("\n[↑/↓] Navigate • [Enter] Select • [ESC] Main Menu • [q] Quit"))
	return sb.String()
}

// RenderMatrixResultsView renders remote office & data center SLA performance matrix results.
func RenderMatrixResultsView(results []enterprise.MatrixResult, width int) string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(" 🌐 REMOTE OFFICE & DATA CENTER PERFORMANCE MATRIX ") + "\n\n")
	sb.WriteString(Cyan("Evaluating latency, packet loss, jitter, and SLA compliance across nodes.\n\n"))

	for _, r := range results {
		statusColor := Green(r.SLAStatus)
		switch r.SLAStatus {
		case "WARNING":
			statusColor = Yellow(r.SLAStatus)
		case "VIOLATED":
			statusColor = Red(r.SLAStatus)
		}

		pairName := fmt.Sprintf("%s ➔ %s", r.Source.Region, r.Destination.Region)
		sb.WriteString(fmt.Sprintf("  %-20s │ Latency: %-6v │ Loss: %4.1f%% │ Rate: %4.0f Mbps │ %s\n",
			Bold(pairName), r.Latency, r.Loss, r.Throughput, statusColor))
	}

	sb.WriteString(HelpStyle.Render("\n[ESC/Enter] Return to Submenu • [q] Quit"))
	return sb.String()
}

// RenderProtocolResultsView renders synthetic multi-protocol testing under dynamic conditions.
func RenderProtocolResultsView(results []enterprise.ProtocolTestResult, width int) string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(" 📡 SYNTHETIC MULTI-PROTOCOL PROBE RESULTS ") + "\n\n")
	sb.WriteString(Cyan("Evaluating HTTP, VoIP, SaaS, and custom protocols under simulated network stress.\n\n"))

	for _, res := range results {
		statusColor := Green(res.Status)
		switch res.Status {
		case "DEGRADED":
			statusColor = Yellow(res.Status)
		case "CRITICAL":
			statusColor = Red(res.Status)
		}

		sb.WriteString(fmt.Sprintf("  ▸ %-20s [%s]\n", Bold(string(res.Config.Protocol)), Cyan(string(res.Config.Condition))))
		sb.WriteString(fmt.Sprintf("    Target: %s\n", Dim(res.Config.TargetName)))
		sb.WriteString(fmt.Sprintf("    Success: %5.1f%%  │  RTT: %-7v  │  Score: %3d/100  │  Status: %s\n",
			res.SuccessRate, res.AvgLatency, res.Score, statusColor))
		sb.WriteString(fmt.Sprintf("    Summary: %s\n\n", Dim(res.Details)))
	}

	sb.WriteString(HelpStyle.Render("\n[ESC/Enter] Return to Submenu • [q] Quit"))
	return sb.String()
}

// RenderReplayResultsView renders live traffic capture & lab replay simulation telemetry.
func RenderReplayResultsView(res enterprise.ReplayResult, width int) string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(" 📼 TRAFFIC CAPTURE & LAB REPLAY SIMULATOR ") + "\n\n")
	sb.WriteString(Cyan("Replaying captured production traffic payloads in isolated lab conditions.\n\n"))

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

	sb.WriteString(Bold("TRANSACTION REPLAY LOG:") + "\n")
	for _, logLine := range res.LogSummary {
		sb.WriteString("  " + logLine + "\n")
	}

	sb.WriteString(HelpStyle.Render("\n[ESC/Enter] Return to Submenu • [q] Quit"))
	return sb.String()
}
