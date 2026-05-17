package ui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
)

// SpinnerModel implements tea.Model for the bubbletea spinner.
// The bubbletea interface is the ONLY permitted interface in this codebase (Rule 1 exception).
type SpinnerModel struct {
	sp      spinner.Model
	message string
	done    bool
	err     error
}

type doneMsg struct{ err error }

// NewSpinner creates a SpinnerModel with the given status message.
func NewSpinner(msg string) SpinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return SpinnerModel{sp: s, message: msg}
}

func (m SpinnerModel) Init() tea.Cmd {
	return m.sp.Tick
}

func (m SpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case doneMsg:
		m.done = true
		m.err = msg.err
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.sp, cmd = m.sp.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m SpinnerModel) View() string {
	if m.done {
		return ""
	}
	return fmt.Sprintf("  %s %s\n", m.sp.View(), m.message)
}

// RunSpinner runs work() in a goroutine while displaying an animated spinner.
// Degrades gracefully to a plain fmt.Println when stdout is not a TTY.
func RunSpinner(msg string, work func() error) error {
	if !isTTY() {
		fmt.Printf("  → %s\n", msg)
		return work()
	}

	m := NewSpinner(msg)
	p := tea.NewProgram(m)

	errCh := make(chan error, 1)
	go func() {
		err := work()
		p.Send(doneMsg{err: err})
		errCh <- err
	}()

	finalModel, runErr := p.Run()
	if runErr != nil {
		// Bubbletea failed (e.g., terminal issue) — fall back to plain output.
		fmt.Printf("  → %s\n", msg)
		return work()
	}

	sm, ok := finalModel.(SpinnerModel)
	if ok && sm.err != nil {
		return sm.err
	}

	// Drain the channel (work may have already sent).
	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

// isTTY returns true if stdout is an interactive terminal.
func isTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
