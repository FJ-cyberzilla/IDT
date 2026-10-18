package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Color Palette
	ColorPrimary   = lipgloss.Color("#00d7ff") // Cyan
	ColorSecondary = lipgloss.Color("#af87ff") // Purple
	ColorSuccess   = lipgloss.Color("#00ffaf") // Emerald
	ColorWarning   = lipgloss.Color("#ffaf00") // Amber
	ColorError     = lipgloss.Color("#ff5f5f") // Crimson
	ColorMuted     = lipgloss.Color("#585858") // Gray
	ColorBg        = lipgloss.Color("#1c1c1c") // Dark Gray

	// Base Box & Margins
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2).
			Margin(1, 1).
			Background(ColorBg)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(ColorSecondary).
			PaddingBottom(1).
			MarginBottom(1)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(ColorPrimary).
			Padding(0, 1)

	StatusSuccessStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorSuccess).
				SetString("✔ MATCH")

	StatusWarningStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorWarning).
				SetString("⚠ PARTIAL")

	StatusErrorStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorError).
				SetString("✖ MISMATCH")

	ProtocolStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true).
			MarginTop(1)

	MenuItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			PaddingLeft(2)

	MenuSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorPrimary).
				Background(lipgloss.Color("#262626")).
				PaddingLeft(2).
				BorderStyle(lipgloss.NormalBorder()).
				BorderLeft(true).
				BorderForeground(ColorPrimary)

	// Phase / Dynamic Status Badge Styles
	PhaseMenuStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00ffff")).Background(lipgloss.Color("#005f5f")).Padding(0, 1)
	PhaseInitStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ffff00")).Background(lipgloss.Color("#5f5f00")).Padding(0, 1)
	PhaseNetworkStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#af87ff")).Background(lipgloss.Color("#4f0080")).Padding(0, 1)
	PhaseReadyStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00ffaf")).Background(lipgloss.Color("#005f3f")).Padding(0, 1)
	PhaseRunningStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00d7ff")).Background(lipgloss.Color("#005f87")).Padding(0, 1)
	PhaseResultsStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00ff87")).Background(lipgloss.Color("#005f00")).Padding(0, 1)
	PhaseServersStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#87d7ff")).Background(lipgloss.Color("#005faf")).Padding(0, 1)
	PhaseCheckStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#87ffaf")).Background(lipgloss.Color("#00875f")).Padding(0, 1)
	PhaseTuneStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ffaf00")).Background(lipgloss.Color("#875f00")).Padding(0, 1)
	PhaseErrorStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ffffff")).Background(ColorError).Padding(0, 1)

	RedAttentionStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorError).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(ColorError).
				Padding(1, 2).
				Margin(1, 0)
)
