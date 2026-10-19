package ui

import (
	"fmt"
	"strings"
	"time"

	"idt/internal/compare"
	"idt/internal/config"
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
	case "ERROR":
		return PhaseErrorStyle.Render(" 🚨 ATTENTION: NO INTERNET CONNECTION / STALLED NETWORK ")
	default:
		return PhaseMenuStyle.Render(fmt.Sprintf(" 🔵 STATUS: %s ", stateName))
	}
}

// RenderWarmupProgress renders the real-time animated warm-up checklist with checkmarks, active spinner, and pending states.
func RenderWarmupProgress(step warmupStep, spinnerView string, connType string) string {
	var sb strings.Builder

	// Step 1: Init
	if step > stepInit {
		sb.WriteString("  ✅ Initializing IDT Integrated Command Center & Configuration\n")
	} else if step == stepInit {
		sb.WriteString(fmt.Sprintf("  %s %s\n", spinnerView, lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("Initializing IDT Integrated Command Center...")))
	} else {
		sb.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("  ⏳ Initializing IDT Integrated Command Center...") + "\n")
	}

	// Step 2: Network
	if step > stepNetwork {
		ct := connType
		if ct == "" {
			ct = "Normal ISP Connection"
		}
		sb.WriteString(fmt.Sprintf("  ✅ Network Route & VPN Detection [%s]\n", ct))
	} else if step == stepNetwork {
		sb.WriteString(fmt.Sprintf("  %s %s\n", spinnerView, lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary).Render("Detecting network route & VPN status...")))
	} else {
		sb.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("  ⏳ Network Route & VPN Detection...") + "\n")
	}

	// Step 3: Preflight
	if step > stepPreflight {
		sb.WriteString("  ✅ Public Resolver Reachability Check (1.1.1.1 / 8.8.8.8)\n")
	} else if step == stepPreflight {
		sb.WriteString(fmt.Sprintf("  %s %s\n", spinnerView, lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render("Testing pre-flight public resolver reachability...")))
	} else {
		sb.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("  ⏳ Public Resolver Reachability Check...") + "\n")
	}

	// Step 4: Ready
	if step > stepReady {
		sb.WriteString("  ✅ Secure Resolvers & Connection Pools Prepared\n")
	} else if step == stepReady {
		sb.WriteString(fmt.Sprintf("  %s %s\n", spinnerView, lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render("Preparing secure connection pools...")))
	} else {
		sb.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("  ⏳ Secure Resolvers & Connection Pools Prepared...") + "\n")
	}

	sb.WriteString("\n" + HelpStyle.Render("Integrated Dashboard • Real-time Warm-up Progress & System Verification..."))
	return sb.String()
}

// RenderMenuChoices renders the interactive menu choices with cursor highlighting.
func RenderMenuChoices(choices []string, cursor int) string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(" SELECT COMMAND / VIEW ") + "\n\n")
	for i, choice := range choices {
		if i == cursor {
			sb.WriteString(MenuSelectedStyle.Render(choice) + "\n")
		} else {
			sb.WriteString(MenuItemStyle.Render(choice) + "\n")
		}
	}
	sb.WriteString(HelpStyle.Render("\n[↑/↓] Navigate • [Enter] Select • [q] Quit"))
	return sb.String()
}

// RenderErrorCard renders the red attention error card for self-aware offline / stalled states.
func RenderErrorCard(reason string) string {
	errMsg := fmt.Sprintf("CRITICAL ERROR: %s\n\n❌ Self-awareness warmup check FAILED: The application cannot reach public resolvers (1.1.1.1 / 8.8.8.8).\nCheck your Wi-Fi, mobile data, airplane mode, or active VPN connection.", reason)
	return RedAttentionStyle.Render(errMsg) + "\n\n" + HelpStyle.Render("[r] Retry Warmup • [Enter] Return to Menu • [q] Quit")
}

// RenderCheckView renders real-world connectivity diagnostics telemetry.
func RenderCheckView(opts config.Options, dialers []config.Dialer) string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(" 🔍 REAL-WORLD CONNECTIVITY & HEALTH CHECK ") + "\n\n")
	sb.WriteString(fmt.Sprintf("Target Domain: %s\n", opts.Target))
	sb.WriteString(fmt.Sprintf("Timeout Config: %v\n", opts.Timeout))
	sb.WriteString(fmt.Sprintf("Jitter Window: %v - %v\n\n", opts.JitterMin, opts.JitterMax))
	sb.WriteString("Healthy Check Endpoints:\n")
	for i, d := range dialers {
		if i >= 5 {
			break
		}
		sb.WriteString(fmt.Sprintf("  ✅ [%s] %s (%s)\n", ProtocolStyle.Render(d.Protocol), d.Name, d.Address))
	}
	sb.WriteString(HelpStyle.Render("\n[Enter] Return to Menu"))
	return sb.String()
}

// RenderTuneView renders real-world tuning analysis parameters.
func RenderTuneView(opts config.Options) string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(" 📈 RTT DISTRIBUTION & TUNING ANALYSIS ") + "\n\n")
	sb.WriteString("Analysis Parameters:\n")
	sb.WriteString(fmt.Sprintf(" • Max Dials Limit: %d\n", opts.MaxDials))
	sb.WriteString(fmt.Sprintf(" • Dials per endpoint: %d\n", opts.Dials))
	sb.WriteString(fmt.Sprintf(" • Target Verification Domain: %s\n\n", opts.Target))
	sb.WriteString("Recommendation: Enable DoH/DoT encapsulation for optimal privacy and bypass of local network injection.\n")
	sb.WriteString(HelpStyle.Render("\n[Enter] Return to Menu"))
	return sb.String()
}

// RenderCheckResultsView renders live connectivity check results with color spectrum and comprehensive details.
func RenderCheckResultsView(results []runner.Result) string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(" 🔍 REAL-WORLD CONNECTIVITY & HEALTH CHECK ") + "\n\n")
	sb.WriteString(Cyan(fmt.Sprintf("Health check — example.com — 5 dials — sequential (Color Spectrum: %s | %s | %s)\n\n",
		Green("Optimal"), Yellow("Warning"), Red("Unreachable"))))

	var (
		dohOK, dotOK, anySuccess bool
		stdLocal                 bool
	)

	for _, r := range results {
		status := Red("❌ no answer")
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
			status = Red("❌ unreachable / filtered")
		}

		// Comprehensive detail: which is what and where
		protocolTag := Cyan("[" + r.Dialer.Protocol + "]")
		sb.WriteString(fmt.Sprintf("  %-12s %-10s %-45s  %s\n",
			Bold(r.Dialer.Name), protocolTag, Dim(r.Dialer.Address), status))
	}

	sb.WriteString("\n" + Bold("SETTINGS & NETWORK RESOLUTION ANALYSIS:") + "\n")
	switch {
	case !anySuccess:
		sb.WriteString("  " + Red("❌ No dialer reachable. Name resolution is broken.") + "\n")
		sb.WriteString("     Guidance: Check airplane mode, Wi-Fi/cellular data, or VPN state.\n")
	case !dohOK:
		sb.WriteString("  " + Red("❌ Encrypted DoH dialers are unreachable.") + "\n")
		sb.WriteString("     Guidance: Firewall blocking port 443 or captive portal active.\n")
	case !dotOK:
		sb.WriteString("  " + Yellow("⚠ Settings Mismatch / Network Restriction: DoH works, but DoT (port 853) is blocked.") + "\n")
		sb.WriteString("     Resolution: Android Private DNS (hostname mode) will NOT resolve here.\n")
		sb.WriteString("     Fix: Switch network or use DoH-based upstreams in your app settings.\n")
	default:
		sb.WriteString("  " + Green("✅ All encrypted protocols (DoH + DoT) are fully reachable and optimal.") + "\n")
		sb.WriteString("     Resolution: Your current settings (Private DNS / Termux upstreams) are fully supported.\n")
	}

	if stdLocal {
		sb.WriteString("  " + Yellow("⚠ Local Interception Notice: Standard port 53 is answering locally (VPN/ISP intercept).") + "\n")
		sb.WriteString("     Resolution: Upstream DNS settings differ from local ISP defaults; encryption bypasses this.\n")
	}

	sb.WriteString(HelpStyle.Render("\n[Enter] Return to Menu • [q] Quit"))
	return sb.String()
}

// RenderTuneResultsView renders live tuning analysis results.
func RenderTuneResultsView(sorted []runner.Result, verdicts map[string]compare.Verdict, opts config.Options) string {
	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(" 📈 RTT DISTRIBUTION & TUNING ANALYSIS ") + "\n\n")
	rec := tune.GetRecommendationString(sorted, verdicts, opts)
	sb.WriteString(rec)
	sb.WriteString(HelpStyle.Render("\n[Enter] Return to Menu • [q] Quit"))
	return sb.String()
}
