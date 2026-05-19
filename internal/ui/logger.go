package ui

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// ── Lapis Lazuli Color Palette (from DESIGN.md) ───────────────────────────────

var (
	// Primary (Lapis Blue) — step headers, box borders
	colorPrimary = color.New(color.FgBlue, color.Bold)

	// Active (Cyan Flash) — spinner, cursor, active step indicator
	colorActive = color.New(color.FgCyan, color.Bold)

	// Info (Wing Aqua) — file names, package names, neutral text
	colorInfo = color.New(color.FgHiCyan)

	// Success (Water Light) — passed tests, selected files
	colorSuccess = color.New(color.FgHiGreen, color.Bold)

	// Error (Cracked Gem) — failed tests, errors
	colorError = color.New(color.FgHiRed, color.Bold)

	// Warning (Storm Yellow) — heal attempts, warnings
	colorWarn = color.New(color.FgYellow)

	// Muted (Deep Current) — borders, dividers, hints
	colorMuted = color.New(color.FgHiBlack)

	// Emphasis (Gem Flash) — values, filenames in summary
	colorEmphasis = color.New(color.FgHiWhite, color.Bold)

	// Step — [N/4] headers (alias for Primary for backwards compat)
	colorStep = colorPrimary

	// Dim — faint secondary text
	colorDim = color.New(color.Faint)
)

// ── Logger Functions ──────────────────────────────────────────────────────────

// LogInfo prints a cyan informational message.
func LogInfo(format string, args ...any) {
	colorInfo.Printf("  "+format+"\n", args...)
}

// LogSuccess prints a green bold success message with ✓ prefix.
func LogSuccess(format string, args ...any) {
	colorSuccess.Printf("  ✓ "+format+"\n", args...)
}

// LogError prints a red bold error message with ✗ prefix.
func LogError(format string, args ...any) {
	colorError.Printf("  ✗ "+format+"\n", args...)
}

// LogWarn prints a yellow warning message with ⚠ prefix.
func LogWarn(format string, args ...any) {
	colorWarn.Printf("  ⚠ "+format+"\n", args...)
}

// LogStep prints a Lapis Blue step header: "[2/4] label".
func LogStep(step, total int, label string) {
	colorPrimary.Printf("\n[%d/%d] ", step, total)
	colorEmphasis.Printf("%s\n", label)
}

// LogDivider prints a muted horizontal rule.
func LogDivider() {
	colorMuted.Println("  " + strings.Repeat("─", 50))
}

// LogCodeBlock prints a titled, bordered code preview (no line limit - shows everything).
func LogCodeBlock(title, content string) {
	colorPrimary.Printf("\n  ┌─ %s %s\n", title, strings.Repeat("─", max(0, 44-len(title))))
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		colorMuted.Printf("  │ ")
		colorInfo.Printf("%s\n", line)
	}
	colorPrimary.Println("  └" + strings.Repeat("─", 51))
	fmt.Println()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
