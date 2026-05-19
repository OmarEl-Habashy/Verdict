/*
Package ui manages the terminal user interfaces and logging for the application.
This file defines the application's color palette (Lapis Lazuli) and provides
styled logging functions for various levels of severity and structure.

Functions:
- LogInfo: Prints a cyan informational message.
- LogSuccess: Prints a green bold success message.
- LogError: Prints a red bold error message.
- LogWarn: Prints a yellow warning message.
- LogStep: Prints a styled step header for the CLI workflow.
- LogDivider: Prints a muted horizontal rule.
- LogCodeBlock: Prints a titled, bordered code preview.
- max: Simple helper function returning the maximum of two integers.
*/
package ui

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

var (
	colorPrimary = color.New(color.FgBlue, color.Bold)

	colorActive = color.New(color.FgCyan, color.Bold)

	colorInfo = color.New(color.FgHiCyan)

	colorSuccess = color.New(color.FgHiGreen, color.Bold)

	colorError = color.New(color.FgHiRed, color.Bold)

	colorWarn = color.New(color.FgYellow)

	colorMuted = color.New(color.FgHiBlack)

	colorEmphasis = color.New(color.FgHiWhite, color.Bold)

	colorStep = colorPrimary

	colorDim = color.New(color.Faint)
)

func LogInfo(format string, args ...any) {
	colorInfo.Printf("  "+format+"\n", args...)
}

func LogSuccess(format string, args ...any) {
	colorSuccess.Printf("  ✓ "+format+"\n", args...)
}

func LogError(format string, args ...any) {
	colorError.Printf("  ✗ "+format+"\n", args...)
}

func LogWarn(format string, args ...any) {
	colorWarn.Printf("  ⚠ "+format+"\n", args...)
}

func LogStep(step, total int, label string) {
	colorPrimary.Printf("\n[%d/%d] ", step, total)
	colorEmphasis.Printf("%s\n", label)
}

func LogDivider() {
	colorMuted.Println("  " + strings.Repeat("─", 50))
}

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
