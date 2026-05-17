package ui

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

var (
	colorInfo    = color.New(color.FgCyan)
	colorSuccess = color.New(color.FgGreen, color.Bold)
	colorError   = color.New(color.FgRed, color.Bold)
	colorWarn    = color.New(color.FgYellow)
	colorStep    = color.New(color.FgWhite, color.Bold)
	colorDim     = color.New(color.Faint)
)

// LogInfo prints a cyan informational message.
func LogInfo(format string, args ...any) {
	colorInfo.Printf("  "+format+"\n", args...)
}

// LogSuccess prints a green bold success message.
func LogSuccess(format string, args ...any) {
	colorSuccess.Printf("  ✓ "+format+"\n", args...)
}

// LogError prints a red bold error message.
func LogError(format string, args ...any) {
	colorError.Printf("  ✗ "+format+"\n", args...)
}

// LogWarn prints a yellow warning message.
func LogWarn(format string, args ...any) {
	colorWarn.Printf("  ⚠ "+format+"\n", args...)
}

// LogStep prints a white bold step header: "[2/4] label".
func LogStep(step, total int, label string) {
	colorStep.Printf("\n[%d/%d] %s\n", step, total, label)
}

// LogDivider prints a dim horizontal rule.
func LogDivider() {
	colorDim.Println("  " + strings.Repeat("─", 50))
}

// LogCodeBlock prints a titled, bordered code preview (capped at 30 lines).
func LogCodeBlock(title, content string) {
	colorDim.Printf("\n  ┌─ %s ", title)
	colorDim.Println(strings.Repeat("─", max(0, 44-len(title))))

	lines := strings.Split(content, "\n")
	limit := 30
	if len(lines) < limit {
		limit = len(lines)
	}
	for _, line := range lines[:limit] {
		colorDim.Printf("  │ %s\n", line)
	}
	if len(lines) > 30 {
		colorDim.Printf("  │ ... (%d more lines)\n", len(lines)-30)
	}

	colorDim.Println("  └" + strings.Repeat("─", 51))
	fmt.Println()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
