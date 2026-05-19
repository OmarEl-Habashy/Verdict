package ui

import (
	"fmt"
)

// PostRunDiagnostic provides a brief explanation of why a test failed
func PostRunDiagnostic(errorType string, stderr string) string {
	// Show full stderr in diagnostic (no truncation for user visibility)

	var msg string
	switch errorType {
	case "missing_import":
		msg = "Missing import or undefined symbol in generated test.\n" +
			"The LLM forgot to import a required package.\n\n" +
			"Fix: This usually heals on the next attempt.\n"
	case "syntax_error":
		msg = "Go syntax error in generated test code.\n" +
			"Check brackets, parentheses, or formatting.\n"
	case "compilation":
		msg = "Compilation failed in generated test.\n" +
			"The LLM produced invalid Go code.\n"
	case "runtime_error":
		msg = "Test failed at runtime (panic or failure).\n" +
			"The generated test logic doesn't match the source.\n"
	case "low_coverage":
		msg = "Tests compile and run but achieve <80% coverage.\n" +
			"The LLM needs to add more branches and edge cases.\n"
	default:
		msg = "Unknown error. Check the error details below.\n"
	}

	// Return message with capped stderr
	return msg + "\n" + stderr
}

// PostRunMenuChoice represents the user's choice after a test run
type PostRunMenuChoice int

const (
	ChoiceNext PostRunMenuChoice = iota
	ChoiceExtendHeal
	ChoiceReturnMenu
)

// PostRunMenu displays options after a test completes
// Accepts: passed status, test file path, total files count, current file index
// Returns the user's choice; file cleanup should be handled by caller
func PostRunMenu(passed bool, testFilePath string, totalFiles int, currentFileIndex int) PostRunMenuChoice {
	fmt.Println()
	colorEmphasis.Println("  [Post-Run Menu]")
	fmt.Println()

	// Check if there are more files after this one
	hasMoreFiles := currentFileIndex < totalFiles-1

	if passed {
		colorSuccess.Println("  ✓ Test passed!")
		fmt.Println()
		fmt.Println("  Options:")

		if hasMoreFiles {
			fmt.Println(colorInfo.Sprint("    1. Continue to next file"))
			fmt.Println(colorInfo.Sprint("    2. Return to main menu"))
			fmt.Println()
			fmt.Print(colorMuted.Sprint("  Choose (1-2): "))

			var choice string
			fmt.Scanln(&choice)

			if choice == "2" {
				return ChoiceReturnMenu
			}
			return ChoiceNext
		}

		// Last file - no continue option
		fmt.Println(colorInfo.Sprint("    1. Return to main menu"))
		fmt.Println()
		fmt.Print(colorMuted.Sprint("  Choose (1): "))
		var choice string
		fmt.Scanln(&choice)
		return ChoiceReturnMenu
	}

	// Failed test
	colorError.Println("  ✗ Test failed")
	fmt.Println()
	fmt.Println("  Options:")

	if hasMoreFiles {
		fmt.Println(colorInfo.Sprint("    1. Continue to next file (skip this one)"))
		fmt.Println(colorWarn.Sprint("    2. Try 2 more healing attempts"))
		fmt.Println(colorInfo.Sprint("    3. Return to main menu"))
		fmt.Println()
		fmt.Print(colorMuted.Sprint("  Choose (1-3): "))

		var choice string
		fmt.Scanln(&choice)

		switch choice {
		case "2":
			return ChoiceExtendHeal
		case "3":
			return ChoiceReturnMenu
		default:
			return ChoiceNext
		}
	}

	// Last file - no continue option
	fmt.Println(colorWarn.Sprint("    1. Try 2 more healing attempts"))
	fmt.Println(colorInfo.Sprint("    2. Return to main menu"))
	fmt.Println()
	fmt.Print(colorMuted.Sprint("  Choose (1-2): "))

	var choice string
	fmt.Scanln(&choice)

	switch choice {
	case "1":
		return ChoiceExtendHeal
	default:
		return ChoiceReturnMenu
	}
}
