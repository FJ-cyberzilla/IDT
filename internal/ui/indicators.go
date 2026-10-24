package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	guardcatLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Bold(true)
	preparingStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	readyStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	errorStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	guardcatTileStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).Padding(0, 1)
)

// RenderGuardcatTile renders the permanent Guardcat status tile with live stats and animated pre-test status.
func RenderGuardcatTile(statusPhase string, dotCount int, memoryMB uint64, goroutines int, width int) string {
	w := getResponsiveWidth(width)

	guardcatLabel := guardcatLabelStyle.Render("Guardcat")

	var statusText string
	switch statusPhase {
	case "preparing":
		dots := strings.Repeat(".", (dotCount%3)+1)
		statusText = preparingStyle.Render(fmt.Sprintf("Preparing for curiosity %s", dots))
	case "ready":
		statusText = readyStyle.Render("ready and safe!")
	case "error":
		statusText = errorStyle.Render("monitoring alert: system check warning!")
	default:
		statusText = readyStyle.Render("ready and safe!")
	}

	liveStats := Dim(fmt.Sprintf("│ Mem: %dMB │ Goroutines: %d", memoryMB, goroutines))

	content := fmt.Sprintf("🛡️ %s : %s  %s", guardcatLabel, statusText, liveStats)

	return guardcatTileStyle.Width(w - 2).Render(content)
}

// getResponsiveWidth ensures layouts adapt cleanly to terminal zooming and viewports.
func getResponsiveWidth(width int) int {
	if width <= 0 {
		return 95
	}
	if width < 70 {
		return 70
	}
	if width > 130 {
		return 130
	}
	return width
}

// truncate adapts string lengths to prevent wrapping bugs on narrow terminals.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// renderPanel wraps content inside a sleek command center panel box.
func renderPanel(title string, body string, width int) string {
	w := getResponsiveWidth(width)

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00F0FF")).
		Width(w - 2).
		Padding(1, 2)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FFCC"))

	header := headerStyle.Render("⚡ " + title)
	divider := Dim(strings.Repeat("─", w-6))

	content := fmt.Sprintf("%s\n%s\n\n%s", header, divider, body)
	return panelStyle.Render(content)
}

// RenderPhaseBadge returns a high-contrast, elite command center status badge.
func RenderPhaseBadge(stateName string, phase string) string {
	switch phase {
	case "MENU":
		return PhaseMenuStyle.Render(" 🟢 [SOC STATUS: COMMAND CENTER • IDLE] ")
	case "WARMUP":
		return PhaseInitStyle.Render(" 🟡 [SOC STATUS: PREFLIGHT & ROUTE VERIFICATION] ")
	case "RUNNING":
		return PhaseRunningStyle.Render(" 🔵 [SOC STATUS: EXECUTING MULTI-PROTOCOL PROBES] ")
	case "RESULTS":
		return PhaseResultsStyle.Render(" ✅ [SOC STATUS: TELEMETRY & VERDICTS SECURED] ")
	case "SERVERS":
		return PhaseServersStyle.Render(" ⚙ [SOC STATUS: TOPOLOGY & NODE MONITORING] ")
	case "CHECK":
		return PhaseCheckStyle.Render(" 📡 [SOC STATUS: REAL-WORLD HEALTH DIAGNOSTICS] ")
	case "TUNE":
		return PhaseTuneStyle.Render(" 📈 [SOC STATUS: LATENCY MATRIX & RTT TUNING] ")
	case "ABOUT":
		return PhaseAboutStyle.Render(" 🛡️ [SOC STATUS: SYSTEM ARCHITECTURE & INTEL] ")
	case "ERROR":
		return PhaseErrorStyle.Render(" 🚨 [CRITICAL: NETWORK STALL / TUNNEL FAILURE] ")
	default:
		return PhaseMenuStyle.Render(fmt.Sprintf(" 🔵 [SOC STATUS: %s] ", stateName))
	}
}

// RenderWarmupProgress renders a sleek real-time animated command center warm-up sequence.
func RenderWarmupProgress(step warmupStep, spinnerView string, connType string, width int, elapsed time.Duration) string {
	var sb strings.Builder

	renderStep := func(targetStep warmupStep, label string, activeText string) {
		if step > targetStep {
			sb.WriteString(fmt.Sprintf("  %s %s\n", Green("✔ [COMPLETED]"), label))
		} else if step == targetStep {
			elapsedStr := fmt.Sprintf("[%02d:%02d]", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
			sb.WriteString(fmt.Sprintf("  %s %s %s\n", spinnerView, lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(activeText), Dim(elapsedStr)))
		} else {
			sb.WriteString(fmt.Sprintf("  %s %s\n", Dim("• [PENDING]"), Dim(label)))
		}
	}

	ct := connType
	if ct == "" {
		ct = "Standard Broadband / ISP"
	}

	sb.WriteString(Cyan("🛰️ Initializing telemetric link and validating routing security...\n\n"))
	renderStep(stepInit, "Initialize Command Center & Config Kernel", "Initializing command center & configuration kernel...")
	renderStep(step, fmt.Sprintf("Network Route & Tunnel Detection [%s]", ct), "Detecting network topology & active tunnels...")
	renderStep(stepPreflight, "System & Discovered Recursive Resolver Reachability (Android / Termux / DoH / DoT)", "Probing system & public trust anchors...")
	renderStep(stepReady, "Encrypted Connection Pools & Sockets Allocated", "Allocating secure connection pools...")

	sb.WriteString("\n" + HelpStyle.Render("ESC: Abort Operation  │  q: Disconnect & Quit"))
	return renderPanel("SYSTEM INITIALIZATION & PREFLIGHT", sb.String(), width)
}

// RenderMenuChoices renders elite interactive command center menu options with custom selection cursors.
func RenderMenuChoices(choices []string, cursor int, width int) string {
	var sb strings.Builder
	sb.WriteString(Cyan("🎛️ Select an operational subsystem or diagnostic routine:\n\n"))

	for i, choice := range choices {
		if i == cursor {
			sb.WriteString(MenuSelectedStyle.Render(fmt.Sprintf(" ❯ %s ", choice)) + "\n")
		} else {
			sb.WriteString(MenuItemStyle.Render(fmt.Sprintf("   %s ", choice)) + "\n")
		}
	}

	sb.WriteString("\n" + HelpStyle.Render("↑/↓: Navigate  │  Enter: Engage Subsystem  │  ESC: Menu  │  q: Quit"))
	return renderPanel("COMMAND CENTER OPERATIONS", sb.String(), width)
}

// RenderErrorCard renders a striking telemetry attention card for offline or stalled states.
func RenderErrorCard(reason string, width int) string {
	errMsg := fmt.Sprintf("TELEMETRY LINK INTERRUPTED\n\n%s\n\nReason: %s\n\nStatus: Error handler reset & retry completed. Not reachable this time.",
		Red("✖ Preflight connectivity check failed against system and discovered DNS resolvers (Android/Termux/DoH/DoT)."), reason)

	body := RedAttentionStyle.Render(errMsg) + "\n\n" + HelpStyle.Render("r: Retry Preflight  │  ENTER: Return to Menu  │  q: Quit")
	return renderPanel("CRITICAL NETWORK FAULT", body, width)
}
