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

// RenderPhaseBadge returns a color-coded status badge for the current dashboard phase.
func RenderPhaseBadge(stateName string, phase string) string {
	switch phase {
	case "MENU":
		return PhaseMenuStyle.Render(" 🟢 STATUS: MAIN MENU - AWAITING ACTION ")
	case "WARMUP":
		return PhaseInitStyle.Render(" 🟡 STATUS: WARMUP - REAL-TIME SYSTEM VERIFICATION ")
	case "RUNNING":
		return PhaseRunningStyle.Render(" 🔵 STATUS: BENCHMARK - EXECUTING DIALERS ")
	case "RESULTS":
		return PhaseResultsStyle.Render(" ✅ STATUS: COMPLETED - TELEMETRY READY ")
	case "SERVERS":
		return PhaseServersStyle.Render(" ⚙ STATUS: CONFIGURATION - ACTIVE SERVERS ")
	case "CHECK":
		return PhaseCheckStyle.Render(" 🔍 STATUS: CONNECTIVITY - DIAGNOSTICS ")
	case "TUNE":
		return PhaseTuneStyle.Render(" 📈 STATUS: TUNING - ANALYSIS & RECOMMENDATIONS ")
	case "ABOUT":
		return PhaseAboutStyle.Render(" ℹ️ STATUS: ABOUT US - CAPABILITIES & PRIVACY ")
	case "ERROR":
		return PhaseErrorStyle.Render(" 🚨 ATTENTION: NO INTERNET CONNECTION / STALLED NETWORK ")
	default:
		return PhaseMenuStyle.Render(fmt.Sprintf(" 🔵 STATUS: %s ", stateName))
	}
}

// RenderWarmupProgress renders the real-time animated warm-up checklist with checkmarks, active spinner, and pending states.
func RenderWarmupProgress(step warmupStep, spinnerView string, connType string, width int) string {
	var sb strings.Builder

	// Step 1: Init
	if step > stepInit {
		sb.WriteString("  ✅ Initializing IDT Command Center & Configuration\n")
	} else if step == stepInit {
		sb.WriteString(fmt.Sprintf("  %s %s\n", spinnerView, lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("Initializing IDT Command Center...")))
	} else {
		sb.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("  ⏳ Initializing IDT Command Center...") + "\n")
	}

	// Step 2: Network
	if step > stepNetwork {
		ct := connType
		if ct == "" {
			ct = "Normal ISP Connection"
		}
		sb.WriteString(fmt.Sprintf("  ✅ Route & VPN [%s]\n", ct))
	} else if step == stepNetwork {
		sb.WriteString(fmt.Sprintf("  %s %s\n", spinnerView, lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render("Detecting network route & VPN...")))
	} else {
		sb.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("  ⏳ Route & VPN Detection...") + "\n")
	}

	// Step 3: Preflight
	if step > stepPreflight {
		sb.WriteString("  ✅ Public Resolver Check (1.1.1.1 / 8.8.8.8)\n")
	} else if step == stepPreflight {
		sb.WriteString(fmt.Sprintf("  %s %s\n", spinnerView, lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render("Testing public resolver reachability...")))
	} else {
		sb.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("  ⏳ Public Resolver Check...") + "\n")
	}

	// Step 4: Ready
	if step > stepReady {
		sb.WriteString("  ✅ Secure Resolvers & Pools Prepared\n")
	} else if step == stepReady {
		sb.WriteString(fmt.Sprintf("  %s %s\n", spinnerView, lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render("Preparing connection pools...")))
	} else {
		sb.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("  ⏳ Secure Resolvers & Pools Ready...") + "\n")
	}

	sb.WriteString("\n" + HelpStyle.Render("[ESC] Cancel • [q] Quit"))
	return sb.String()
}

// RenderMenuChoices renders the interactive menu choices with cursor highlighting.
func RenderMenuChoices(choices []string, cursor int, width int) string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(" SELECT COMMAND / VIEW ") + "\n\n")
	for i, choice := range choices {
		if i == cursor {
			sb.WriteString(MenuSelectedStyle.Render(choice) + "\n")
		} else {
			sb.WriteString(MenuItemStyle.Render(choice) + "\n")
		}
	}
	sb.WriteString(HelpStyle.Render("\n[↑/↓] Navigate • [Enter] Select • [ESC] Menu • [q] Quit"))
	return sb.String()
}

// RenderErrorCard renders the red attention error card for self-aware offline / stalled states.
func RenderErrorCard(reason string, width int) string {
	errMsg := fmt.Sprintf("CRITICAL ERROR: %s\n\n❌ Warmup check FAILED: Cannot reach public resolvers (1.1.1.1 / 8.8.8.8).\nCheck Wi-Fi, mobile data, airplane mode, or VPN.", reason)
	return RedAttentionStyle.Render(errMsg) + "\n\n" + HelpStyle.Render("[r] Retry • [ESC/Enter] Menu • [q] Quit")
}

// RenderCheckResultsView renders live connectivity check results with color spectrum and responsive adaptation.
func RenderCheckResultsView(results []runner.Result, width int) string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(" 🔍 CONNECTIVITY & HEALTH CHECK ") + "\n\n")
	sb.WriteString(Cyan(fmt.Sprintf("Health check (Color Spectrum: %s | %s | %s)\n\n",
		Green("Optimal"), Yellow("Warning"), Red("Unreachable"))))

	var (
		dohOK, dotOK, anySuccess bool
		stdLocal                 bool
	)

	addrLen := 40
	if width > 0 && width < 70 {
		addrLen = width - 28
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

		protocolTag := Cyan("[" + r.Dialer.Protocol + "]")
		sb.WriteString(fmt.Sprintf("  %-10s %-8s %-*s  %s\n",
			Bold(r.Dialer.Name), protocolTag, addrLen, Dim(addr), status))
	}

	sb.WriteString("\n" + Bold("RESOLUTION ANALYSIS:") + "\n")
	switch {
	case !anySuccess:
		sb.WriteString("  " + Red("❌ No dialer reachable. Check network / VPN.") + "\n")
	case !dohOK:
		sb.WriteString("  " + Red("❌ Encrypted DoH unreachable (port 443 blocked).") + "\n")
	case !dotOK:
		sb.WriteString("  " + Yellow("⚠ DoT (port 853) blocked. Android Private DNS might fail.") + "\n")
	default:
		sb.WriteString("  " + Green("✅ Encrypted protocols (DoH + DoT) optimal.") + "\n")
	}

	if stdLocal {
		sb.WriteString("  " + Yellow("⚠ Local port 53 interception detected.") + "\n")
	}

	sb.WriteString(HelpStyle.Render("\n[ESC/Enter] Return to Menu • [q] Quit"))
	return sb.String()
}

// RenderTuneResultsView renders live tuning analysis results.
func RenderTuneResultsView(sorted []runner.Result, verdicts map[string]compare.Verdict, opts config.Options, width int) string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(" 📈 RTT TUNING ANALYSIS ") + "\n\n")
	rec := tune.GetRecommendationString(sorted, verdicts, opts)
	sb.WriteString(rec)
	sb.WriteString(HelpStyle.Render("\n[ESC/Enter] Return to Menu • [q] Quit"))
	return sb.String()
}

// RenderAboutView renders About Us, capabilities, usage, branding, GitHub, contact email cyberzilla.systsms@gmail.com, and privacy/anti-copying legal notices with polished color and bullets.
func RenderAboutView(width int) string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(" ℹ️ ABOUT US & APP CAPABILITIES ") + "\n\n")

	sb.WriteString(Bold(Cyan("• BRANDING & ORGANIZATION:")) + "\n")
	sb.WriteString(fmt.Sprintf("  %s\n\n", Yellow("FJ™ Cybertronic Systems")))

	sb.WriteString(Bold(Cyan("• CORE APP CAPABILITIES & USAGE:")) + "\n")
	sb.WriteString("  - High-performance network testing & DNS diagnostics command center.\n")
	sb.WriteString("  - Multi-protocol connection benchmarking (DoH, DoT, Standard UDP/TCP).\n")
	sb.WriteString("  - Mobile-optimized TUI interface designed for Termux and compact terminals.\n\n")

	sb.WriteString(Bold(Cyan("• CONTACT & REPOSITORY:")) + "\n")
	sb.WriteString("  - GitHub: https://github.com/cyberzilla-systems/idt\n")
	sb.WriteString("  - Support / Inquiries: cyberzilla.systsms@gmail.com\n\n")

	sb.WriteString(Bold(Red("• PRIVACY POLICY & LEGAL NOTICE:")) + "\n")
	sb.WriteString("  - Anti-Copying / IP Protection: Unauthorized copying, distribution, or reverse engineering is strictly illegal and protected by law.\n")
	sb.WriteString("  - Illegal Usage Prohibition: Any malicious, illegal, or unauthorized network probing is strictly prohibited. Violators will face legal prosecution.\n")

	sb.WriteString(HelpStyle.Render("\n[ESC/Enter] Return to Menu • [q] Quit"))
	return sb.String()
}

// RenderEnterpriseMenu renders the advanced enterprise submenu choices with cursor highlighting.
func RenderEnterpriseMenu(choices []string, cursor int, width int) string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(" 🏢 ADVANCED ENTERPRISE & PROTOCOL SIMULATION ") + "\n\n")
	for i, choice := range choices {
		if i == cursor {
			sb.WriteString(MenuSelectedStyle.Render(choice) + "\n")
		} else {
			sb.WriteString(MenuItemStyle.Render(choice) + "\n")
		}
	}
	sb.WriteString(HelpStyle.Render("\n[↑/↓] Navigate • [Enter] Select • [ESC] Main Menu • [q] Quit"))
	return sb.String()
}

// RenderMatrixResultsView renders remote office and data center performance matrix results.
func RenderMatrixResultsView(results []enterprise.MatrixResult, width int) string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(" 🌐 REMOTE OFFICE & DATA CENTER MATRIX ") + "\n\n")
	sb.WriteString(Cyan("Validating latency, jitter, loss & SLA compliance across remote sites.\n\n"))

	for _, r := range results {
		statusColor := Green(r.SLAStatus)
		if r.SLAStatus == "WARNING" {
			statusColor = Yellow(r.SLAStatus)
		} else if r.SLAStatus == "VIOLATED" {
			statusColor = Red(r.SLAStatus)
		}

		pairName := fmt.Sprintf("%s ➔ %s", r.Source.Region, r.Destination.Region)
		sb.WriteString(fmt.Sprintf("  %-18s | Latency: %-7v | Loss: %4.1f%% | Throughput: %4.0fMbps | %s\n",
			Bold(pairName), r.Latency, r.Loss, r.Throughput, statusColor))
	}

	sb.WriteString(HelpStyle.Render("\n[ESC/Enter] Return to Submenu • [q] Quit"))
	return sb.String()
}

// RenderProtocolResultsView renders multi-protocol testing results under dynamic conditions.
func RenderProtocolResultsView(results []enterprise.ProtocolTestResult, width int) string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(" 📡 MULTI-PROTOCOL DYNAMIC SIMULATION ") + "\n\n")
	sb.WriteString(Cyan("Testing HTTP/VoIP/SaaS/Custom apps across dynamic conditions.\n\n"))

	for _, res := range results {
		statusColor := Green(res.Status)
		if res.Status == "DEGRADED" {
			statusColor = Yellow(res.Status)
		} else if res.Status == "CRITICAL" {
			statusColor = Red(res.Status)
		}

		sb.WriteString(fmt.Sprintf("  %-22s [%s]\n", Bold(string(res.Config.Protocol)), Cyan(string(res.Config.Condition))))
		sb.WriteString(fmt.Sprintf("    Target: %s\n", Dim(res.Config.TargetName)))
		sb.WriteString(fmt.Sprintf("    Success: %5.1f%% | RTT: %-7v | Score: %3d/100 | Status: %s\n",
			res.SuccessRate, res.AvgLatency, res.Score, statusColor))
		sb.WriteString(fmt.Sprintf("    Details: %s\n\n", Dim(res.Details)))
	}

	sb.WriteString(HelpStyle.Render("\n[ESC/Enter] Return to Submenu • [q] Quit"))
	return sb.String()
}

// RenderReplayResultsView renders live traffic capture & lab replay simulation results.
func RenderReplayResultsView(res enterprise.ReplayResult, width int) string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(" 📼 LIVE TRAFFIC CAPTURE & LAB REPLAY ") + "\n\n")
	sb.WriteString(Cyan("Replaying production traffic in controlled lab setting before rollout.\n\n"))

	sb.WriteString(fmt.Sprintf("  • Total Captured Packets: %d\n", res.TotalCaptured))
	sb.WriteString(fmt.Sprintf("  • Successfully Replayed:  %d\n", res.TotalReplayed))
	sb.WriteString(fmt.Sprintf("  • Dropped / Anomalies:    %d\n", res.DroppedPackets))
	sb.WriteString(fmt.Sprintf("  • Replay Duration:        %v\n", res.ReplayDuration))
	sb.WriteString(fmt.Sprintf("  • Average Replayed RTT:   %v\n", res.AvgReplayedRTT))

	anomalyStatus := Green("NO ANOMALIES DETECTED")
	if res.AnomalyDetected {
		anomalyStatus = Red("ANOMALIES / LATENCY SPIKES DETECTED")
	}
	sb.WriteString(fmt.Sprintf("  • Rollout Readiness:      %s\n\n", anomalyStatus))

	sb.WriteString(Bold("REPLAY TRANSACTION LOG:") + "\n")
	for _, logLine := range res.LogSummary {
		sb.WriteString("  " + logLine + "\n")
	}

	sb.WriteString(HelpStyle.Render("\n[ESC/Enter] Return to Submenu • [q] Quit"))
	return sb.String()
}


