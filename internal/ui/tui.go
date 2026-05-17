package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// TUIModel represents the main TUI state machine
type TUIModel struct {
	mode     string // "input", "browser", "files", "results"
	input    string // current directory input
	rootDir  string // selected root directory
	allFiles []string
	selected []string
	err      error
}

// item implements list.Item for file/folder display
type item struct {
	title string
	desc  string
}

func (i item) FilterValue() string { return i.title }
func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }

// InitialModel creates the initial TUI state
func InitialModel() TUIModel {
	return TUIModel{
		mode:  "input",
		input: "",
	}
}

// Init implements tea.Model (required)
func (m TUIModel) Init() tea.Cmd {
	return nil
}

// Update handles user input
func (m TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			if m.mode == "input" {
				// Validate and set root directory
				if m.input == "" {
					m.input = "."
				}
				info, err := os.Stat(m.input)
				if err != nil || !info.IsDir() {
					m.err = fmt.Errorf("invalid directory: %s", m.input)
					return m, nil
				}
				m.rootDir = m.input
				m.mode = "browser"
				m.scanGoFiles()
				return m, nil
			}
			if m.mode == "results" {
				// Ready to exit with selections
				return m, tea.Quit
			}

		case "backspace":
			if m.mode == "input" && len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}

		case "tab":
			if m.mode == "browser" {
				m.mode = "results"
				return m, nil
			}
			if m.mode == "results" {
				m.mode = "input"
				return m, nil
			}

		case "b":
			if m.mode == "results" {
				m.mode = "browser"
				return m, nil
			}

		default:
			if m.mode == "input" {
				// Add character to input
				if len(msg.Runes) > 0 {
					m.input += string(msg.Runes[0])
				}
			}
		}
	}
	return m, nil
}

// View renders the TUI
func (m TUIModel) View() string {
	var output strings.Builder

	output.WriteString("\n")

	switch m.mode {
	case "input":
		output.WriteString(m.renderInputScreen())
	case "browser":
		output.WriteString(m.renderBrowserScreen())
	case "results":
		output.WriteString(m.renderResultsScreen())
	}

	if m.err != nil {
		output.WriteString("\n")
		colorError.Fprintf(&output, "  Error: %v\n", m.err)
	}

	output.WriteString("\n")
	return output.String()
}

// renderInputScreen shows the directory input
func (m TUIModel) renderInputScreen() string {
	var output strings.Builder

	output.WriteString("  " + colorStep.Sprint("QAgent — Test Generation TUI") + "\n\n")

	output.WriteString("  Enter directory to scan for Go files:\n")
	output.WriteString("  " + colorInfo.Sprint("> "+m.input) + colorDim.Sprint("█") + "\n\n")

	output.WriteString("  " + colorDim.Sprint("Hints:") + "\n")
	output.WriteString("  " + colorDim.Sprint("  • Use . for current directory") + "\n")
	output.WriteString("  " + colorDim.Sprint("  • Press Enter to continue") + "\n")
	output.WriteString("  " + colorDim.Sprint("  • Press Ctrl+C to exit") + "\n")

	return output.String()
}

// renderBrowserScreen shows the file browser
func (m TUIModel) renderBrowserScreen() string {
	var output strings.Builder

	output.WriteString("  " + colorStep.Sprint("QAgent — File Browser") + "\n\n")
	output.WriteString("  " + colorDim.Sprint("Directory: ") + colorInfo.Sprint(m.rootDir) + "\n")
	output.WriteString("  " + colorDim.Sprint(fmt.Sprintf("Found %d Go files", len(m.allFiles))) + "\n\n")

	// Show files list (simplified - in real app would be interactive)
	for i, file := range m.allFiles {
		if i >= 10 { // Show first 10
			output.WriteString("  " + colorDim.Sprintf("... and %d more\n", len(m.allFiles)-10))
			break
		}
		output.WriteString("  " + colorInfo.Sprint("▪ ") + filepath.Base(file) + "\n")
	}

	output.WriteString("\n  " + colorDim.Sprint("Press Tab to review files, Ctrl+C to exit") + "\n")

	return output.String()
}

// renderResultsScreen shows test results
func (m TUIModel) renderResultsScreen() string {
	var output strings.Builder

	output.WriteString("  " + colorStep.Sprint("QAgent — Ready to Test") + "\n\n")
	output.WriteString("  " + colorDim.Sprint(fmt.Sprintf("Directory: %s", m.rootDir)) + "\n")
	output.WriteString("  " + colorDim.Sprint(fmt.Sprintf("Files to test: %d", len(m.allFiles))) + "\n\n")

	output.WriteString("  " + colorSuccess.Sprint("✓") + " Ready to generate tests\n\n")

	output.WriteString("  " + colorDim.Sprint("Next steps:") + "\n")
	output.WriteString("  " + colorDim.Sprint("  • Press Enter to start testing") + "\n")
	output.WriteString("  " + colorDim.Sprint("  • Press B to go back") + "\n")

	return output.String()
}

// scanGoFiles finds all .go files (excluding _test.go) in the root directory
func (m *TUIModel) scanGoFiles() {
	m.allFiles = []string{}

	filepath.Walk(m.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			m.allFiles = append(m.allFiles, path)
		}
		return nil
	})

	sort.Strings(m.allFiles)
}

// RunTUI launches the interactive TUI
func RunTUI() (string, []string, error) {
	model := InitialModel()

	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return "", nil, err
	}

	m := finalModel.(TUIModel)
	return m.rootDir, m.allFiles, nil
}
