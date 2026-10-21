package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// View renders the current state of the Bubble Tea UI model.
func (m Model) View() string {
	var sb strings.Builder

	width := m.width
	if width <= 0 {
		width = 95
	}

	header := HeaderStyle.Render("🌐 IDT — ENTERPRISE SECURE DIALER COMMAND CENTER")
	sb.WriteString(header + "\n\n")

	switch m.state {
	case stateMenu:
		sb.WriteString(RenderPhaseBadge("MENU", "MENU") + "\n\n")
		sb.WriteString(RenderMenuChoices(m.choices, m.cursor, width))

	case stateWarmup:
		elapsed := time.Since(m.startTime).Round(time.Second)
		timerStr := formatTimerWithETA(elapsed, 15*time.Second)
		sb.WriteString(RenderPhaseBadge("WARMUP", "WARMUP") + "\n\n")
		sb.WriteString(RenderWarmupProgress(m.warmupStep, m.spinner.View(), m.connType, width, elapsed))
		sb.WriteString(Dim("\n  Progress: " + timerStr))

	case stateRunning:
		elapsed := time.Since(m.startTime).Round(time.Second)
		timerStr := formatTimerWithETA(elapsed, 45*time.Second)
		sb.WriteString(RenderPhaseBadge("RUNNING", "RUNNING") + "\n\n")
		sb.WriteString(fmt.Sprintf("\n  %s %s %s\n\n", m.spinner.View(), lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(m.loadingText), Dim(timerStr)))
		sb.WriteString(HelpStyle.Render("SOC Command Center • ESC: Cancel Operation  │  q: Disconnect & Quit"))

	case stateResults, stateDashboard:
		sb.WriteString(RenderPhaseBadge("RESULTS", "RESULTS") + "\n\n")
		sb.WriteString(RenderTuneResultsView(m.sortedResults, m.verdicts, m.opts, width))

	case stateServers:
		sb.WriteString(RenderPhaseBadge("SERVERS", "SERVERS") + "\n\n")
		sb.WriteString(TitleStyle.Render(" ⚙ CONFIGURED SERVERS & TOPOLOGY ") + "\n")
		sb.WriteString(Dim(strings.Repeat("─", width-6)) + "\n\n")
		for _, d := range m.dialers {
			addr := d.Address
			if width < 60 && len(addr) > 28 {
				addr = addr[:25] + "..."
			}
			sb.WriteString(fmt.Sprintf(" • [%s] %s (%s)\n", ProtocolStyle.Render(d.Protocol), d.Name, addr))
		}
		sb.WriteString("\n" + HelpStyle.Render("ENTER/ESC: Return to Main Command Center"))

	case stateEnterpriseMenu:
		sb.WriteString(RenderPhaseBadge("ENTERPRISE", "ENTER") + "\n\n")
		sb.WriteString(RenderEnterpriseMenu(m.enterpriseChoices, m.enterpriseCursor, width))

	case stateMatrixRunning, stateProtocolRunning, stateReplayRunning:
		elapsed := time.Since(m.startTime).Round(time.Second)
		timerStr := formatTimerWithETA(elapsed, 30*time.Second)
		sb.WriteString(RenderPhaseBadge("ENTERPRISE", "ENTER") + "\n\n")
		sb.WriteString(fmt.Sprintf("\n  %s %s %s\n\n", m.spinner.View(), lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(m.loadingText), Dim(timerStr)))
		sb.WriteString(HelpStyle.Render("Enterprise Subsystem • ESC: Cancel Operation  │  q: Quit"))

	case stateMatrixResults:
		sb.WriteString(RenderPhaseBadge("ENTERPRISE", "ENTER") + "\n\n")
		sb.WriteString(RenderMatrixResultsView(m.matrixResults, width))

	case stateProtocolResults:
		sb.WriteString(RenderPhaseBadge("ENTERPRISE", "ENTER") + "\n\n")
		sb.WriteString(RenderProtocolResultsView(m.protocolResults, width))

	case stateReplayResults:
		sb.WriteString(RenderPhaseBadge("ENTERPRISE", "ENTER") + "\n\n")
		sb.WriteString(RenderReplayResultsView(m.replayResult, width))

	case stateColdStartRunning, stateHotlineRunning, stateDNSSECRunning, stateTimezoneRunning:
		elapsed := time.Since(m.startTime).Round(time.Second)
		timerStr := formatTimerWithETA(elapsed, 20*time.Second)
		sb.WriteString(RenderPhaseBadge("ENTERPRISE", "ENTER") + "\n\n")
		sb.WriteString(fmt.Sprintf("\n  %s %s %s\n\n", m.spinner.View(), lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(m.loadingText), Dim(timerStr)))
		sb.WriteString(HelpStyle.Render("Subsystem Execution • ESC: Cancel Operation  │  q: Quit"))

	case stateColdStartResults:
		sb.WriteString(RenderPhaseBadge("COLD START", "COLD") + "\n\n")
		sb.WriteString(RenderColdStartResultsView(m.coldStartResults, width))

	case stateHotlineResults:
		sb.WriteString(RenderPhaseBadge("HOTLINE", "HOTLINE") + "\n\n")
		sb.WriteString(RenderHotlineResultsView(m.hotlineReport, width))

	case stateDNSSECResults:
		sb.WriteString(RenderPhaseBadge("DNSSEC", "DNSSEC") + "\n\n")
		sb.WriteString(RenderDNSSECResultsView(m.dnssecResult, width))

	case stateTimezoneResults:
		sb.WriteString(RenderPhaseBadge("TIMEZONE", "TIMEZONE") + "\n\n")
		sb.WriteString(RenderTimezoneResultsView(m.timezoneReports, width))

	case stateAbout:
		sb.WriteString(RenderPhaseBadge("ABOUT", "ABOUT") + "\n\n")
		sb.WriteString(RenderAboutView(width))

	case stateCheckRunning:
		elapsed := time.Since(m.startTime).Round(time.Second)
		timerStr := formatTimerWithETA(elapsed, 10*time.Second)
		sb.WriteString(RenderPhaseBadge("CHECK", "CHECK") + "\n\n")
		sb.WriteString(fmt.Sprintf("\n  %s %s %s\n\n", m.spinner.View(), lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(m.loadingText), Dim(timerStr)))
		sb.WriteString(HelpStyle.Render("Probing connectivity anchors... ESC: Cancel"))

	case stateCheckResults:
		sb.WriteString(RenderPhaseBadge("CHECK", "CHECK") + "\n\n")
		sb.WriteString(RenderCheckResultsView(m.checkResults, width))

	case stateTuneRunning:
		elapsed := time.Since(m.startTime).Round(time.Second)
		timerStr := formatTimerWithETA(elapsed, 25*time.Second)
		sb.WriteString(RenderPhaseBadge("TUNE", "TUNE") + "\n\n")
		sb.WriteString(fmt.Sprintf("\n  %s %s %s\n\n", m.spinner.View(), lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(m.loadingText), Dim(timerStr)))
		sb.WriteString(HelpStyle.Render("Computing latency matrix... ESC: Cancel"))

	case stateTuneResults:
		sb.WriteString(RenderPhaseBadge("TUNE", "TUNE") + "\n\n")
		sb.WriteString(RenderTuneResultsView(m.sortedResults, m.verdicts, m.opts, width))

	case stateOfflineError:
		sb.WriteString(RenderPhaseBadge("ERROR", "ERROR") + "\n\n")
		sb.WriteString(RenderErrorCard(m.offlineReason, width))
	}

	boxWidth := width - 2
	if boxWidth < 50 {
		boxWidth = 50
	}
	return lipgloss.NewStyle().Width(boxWidth).Padding(1, 1).Render(sb.String())
}
