/*
Package ui manages the terminal user interfaces and logging for the application.
This file provides an animated terminal spinner using Bubble Tea to show progress
during long-running operations (like LLM API calls), degrading gracefully for non-TTY.

Functions:
- NewSpinner: Creates a SpinnerModel with the given status message.
- RunSpinner: Runs work in a goroutine while displaying an animated spinner.
- isTTY: Checks if standard output is connected to an interactive terminal.
*/
package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type SpinnerModel struct {
	sp      spinner.Model
	message string
	done    bool
	err     error
}

type doneMsg struct{ err error }

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

		fmt.Printf("  → %s\n", msg)
		return work()
	}

	sm, ok := finalModel.(SpinnerModel)
	if ok && sm.err != nil {
		return sm.err
	}

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

func isTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
