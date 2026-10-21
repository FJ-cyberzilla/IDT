package ui

import "github.com/charmbracelet/lipgloss"

// Elite Cyber-Slate & Neon Palette
var (
        ColorPrimary   = lipgloss.Color("#00F0FF") // Electric Cyan
        ColorSecondary = lipgloss.Color("#7000FF") // Deep Indigo
        ColorSuccess   = lipgloss.Color("#00FF66") // Neon Green
        ColorWarning   = lipgloss.Color("#FFB800") // Amber Warning
        ColorError     = lipgloss.Color("#FF0055") // Cyber Red
        ColorMuted     = lipgloss.Color("#4A5568") // Slate Gray
        ColorText      = lipgloss.Color("#E2E8F0") // Light Off-White
        ColorSubtext   = lipgloss.Color("#A0AEC0") // Muted Text
        ColorSurface   = lipgloss.Color("#1A202C") // Dark Card Surface
        ColorBorder    = lipgloss.Color("#2D3748") // Crisp Border Gray
)

// Global Lipgloss Styles
var (
        BoxStyle = lipgloss.NewStyle().
                Border(lipgloss.RoundedBorder()).
                BorderForeground(ColorBorder).
                Background(ColorSurface).
                Padding(1, 2)

        TitleStyle = lipgloss.NewStyle().
                Foreground(ColorPrimary).
                Bold(true).
                BorderStyle(lipgloss.NormalBorder()).
                BorderBottom(true).
                BorderForeground(ColorBorder).
                PaddingBottom(0).
                MarginBottom(1)

        SubtitleStyle = lipgloss.NewStyle().
                Foreground(ColorSubtext).
                Italic(true)

        MenuItemStyle = lipgloss.NewStyle().
                Foreground(ColorText).
                PaddingLeft(2)

        MenuSelectedStyle = lipgloss.NewStyle().
                Foreground(ColorPrimary).
                Background(lipgloss.Color("#22303F")).
                Bold(true).
                PaddingLeft(2).
                Border(lipgloss.NormalBorder(), false, false, false, true).
                BorderForeground(ColorPrimary)

        BadgeSuccess = lipgloss.NewStyle().
                Foreground(lipgloss.Color("#000000")).
                Background(ColorSuccess).
                Bold(true).
                Padding(0, 1)

        BadgeWarning = lipgloss.NewStyle().
                Foreground(lipgloss.Color("#000000")).
                Background(ColorWarning).
                Bold(true).
                Padding(0, 1)

        BadgeError = lipgloss.NewStyle().
                Foreground(lipgloss.Color("#FFFFFF")).
                Background(ColorError).
                Bold(true).
                Padding(0, 1)

        FooterStyle = lipgloss.NewStyle().
                Foreground(ColorMuted).
                MarginTop(1)

        PhaseMenuStyle = lipgloss.NewStyle().
                Foreground(ColorPrimary).
                Background(ColorSurface).
                Bold(true).
                Padding(0, 1)

        PhaseInitStyle = lipgloss.NewStyle().
                Foreground(ColorWarning).
                Background(ColorSurface).
                Bold(true).
                Padding(0, 1)

        PhaseRunningStyle = lipgloss.NewStyle().
                Foreground(ColorPrimary).
                Background(ColorSurface).
                Bold(true).
                Padding(0, 1)

        PhaseResultsStyle = lipgloss.NewStyle().
                Foreground(ColorSuccess).
                Background(ColorSurface).
                Bold(true).
                Padding(0, 1)

        PhaseServersStyle = lipgloss.NewStyle().
                Foreground(ColorPrimary).
                Background(ColorSurface).
                Bold(true).
                Padding(0, 1)

        PhaseCheckStyle = lipgloss.NewStyle().
                Foreground(ColorPrimary).
                Background(ColorSurface).
                Bold(true).
                Padding(0, 1)

        PhaseTuneStyle = lipgloss.NewStyle().
                Foreground(ColorSecondary).
                Background(ColorSurface).
                Bold(true).
                Padding(0, 1)

        PhaseAboutStyle = lipgloss.NewStyle().
                Foreground(ColorSubtext).
                Background(ColorSurface).
                Bold(true).
                Padding(0, 1)

        PhaseErrorStyle = lipgloss.NewStyle().
                Foreground(ColorError).
                Background(ColorSurface).
                Bold(true).
                Padding(0, 1)

        HelpStyle = lipgloss.NewStyle().
                Foreground(ColorSubtext).
                Italic(true)

        ProtocolStyle = lipgloss.NewStyle().
                Foreground(ColorPrimary).
                Bold(true)

        RedAttentionStyle = lipgloss.NewStyle().
                Foreground(ColorError).
                Background(ColorSurface).
                Bold(true).
                Padding(1, 2).
                Border(lipgloss.RoundedBorder()).
                BorderForeground(ColorError)

        HeaderStyle = lipgloss.NewStyle().
                Foreground(ColorPrimary).
                Bold(true)
)
