/*
Package ui manages the terminal user interfaces and logging for the application.
This file formats and displays the final test results, either as a stylish ASCII
box summary or as a machine-readable JSON object depending on user settings.

Functions:
- PrintSummary: Renders an ASCII summary box of the test run to standard output.
- PrintJSON: Emits a single JSON line to standard output for machine consumption.
- repeat: Helper function to repeat a character multiple times.
- center: Helper function to center a string within a given width.
- LogBatchSummary: Prints a summary box for batch mode runs.
*/
package ui

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/fatih/color"
)

type JSONOutput struct {
	Passed    bool    `json:"passed"`
	Attempts  int     `json:"attempts"`
	TestFile  string  `json:"test_file"`
	ElapsedMs int64   `json:"elapsed_ms"`
	Coverage  float64 `json:"coverage,omitempty"`
	Error     string  `json:"error,omitempty"`
}

type SummaryResult struct {
	TestFile   string
	Passed     bool
	Attempts   int
	FinalError string
	ElapsedMs  int64
	Coverage   float64
}

func PrintSummary(r SummaryResult, elapsed time.Duration) {
	const width = 42

	statusStr := color.GreenString("✓ PASSED")
	if !r.Passed {
		statusStr = color.RedString("✗ FAILED")
	}

	testFile := r.TestFile
	if testFile == "" {
		testFile = "—"
	} else {
		testFile = filepath.Base(testFile)
	}

	elapsedStr := fmt.Sprintf("%.1fs", elapsed.Seconds())

	fmt.Println()
	fmt.Println("  ╔" + repeat("═", width) + "╗")
	fmt.Println("  ║" + center("Q A G E N T  R E S U L T", width) + "║")
	fmt.Println("  ╠" + repeat("═", width) + "╣")
	fmt.Printf("  ║  %-10s: %-27s║\n", "Status", statusStr)
	fmt.Printf("  ║  %-10s: %-27s║\n", "File", testFile)
	fmt.Printf("  ║  %-10s: %-27s║\n", "Attempts", fmt.Sprintf("%d / %d max", r.Attempts, r.Attempts))
	if r.Coverage > 0 {
		coverageStr := fmt.Sprintf("%.1f%%", r.Coverage)
		fmt.Printf("  ║  %-10s: %-27s║\n", "Coverage", coverageStr)
	}
	fmt.Printf("  ║  %-10s: %-27s║\n", "Time", elapsedStr)
	fmt.Println("  ╚" + repeat("═", width) + "╝")
	fmt.Println()

	if !r.Passed && r.FinalError != "" {
		colorError.Println("  Final error:")
		colorDim.Printf("  %s\n\n", r.FinalError)
	}
}

func PrintJSON(r SummaryResult, elapsed time.Duration) {
	out := JSONOutput{
		Passed:    r.Passed,
		Attempts:  r.Attempts,
		TestFile:  r.TestFile,
		ElapsedMs: elapsed.Milliseconds(),
		Coverage:  r.Coverage,
	}
	if r.FinalError != "" {
		out.Error = r.FinalError
	}
	data, _ := json.Marshal(out)
	fmt.Println(string(data))
}

func repeat(ch string, n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += ch
	}
	return s
}

func center(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	pad := width - len(s)
	left := pad / 2
	right := pad - left
	return repeat(" ", left) + s + repeat(" ", right)
}

func LogBatchSummary(passed, failed int) {
	fmt.Println("  ╔════════════════════════════════════════╗")
	fmt.Printf("  ║  Results: %d passed, %d failed        ║\n", passed, failed)
	fmt.Println("  ╚════════════════════════════════════════╝")
}
