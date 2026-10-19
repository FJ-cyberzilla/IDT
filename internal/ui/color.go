package ui

import "fmt"

const (
	ColorReset  = "\033[0m"
	ColorBold   = "\033[1m"
	ColorDim    = "\033[2m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorRed    = "\033[31m"
	ColorCyan   = "\033[36m"
)

// Colorize wraps text with the given ANSI color code and resets.
func Colorize(colorCode, text string) string {
	return colorCode + text + ColorReset
}

// Green returns bold green text.
func Green(text string) string {
	return Colorize(ColorGreen, text)
}

// Yellow returns bold yellow text.
func Yellow(text string) string {
	return Colorize(ColorYellow, text)
}

// Red returns bold red text.
func Red(text string) string {
	return Colorize(ColorRed, text)
}

// Cyan returns cyan text.
func Cyan(text string) string {
	return Colorize(ColorCyan, text)
}

// Bold returns bold text.
func Bold(text string) string {
	return Colorize(ColorBold, text)
}

// Dim returns dim text.
func Dim(text string) string {
	return Colorize(ColorDim, text)
}

// FormatLatency returns color-coded latency status string based on duration.
func FormatLatency(ms int64) string {
	if ms < 100 {
		return Green(fmt.Sprintf("✅ optimal (%d ms)", ms))
	} else if ms < 300 {
		return Yellow(fmt.Sprintf("⚠ acceptable (%d ms)", ms))
	}
	return Yellow(fmt.Sprintf("⚠ slow (%d ms)", ms))
}
