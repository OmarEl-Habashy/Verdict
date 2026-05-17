package statemachine

import (
	"errors"
	"fmt"
)

// State represents a state in the state machine
type State int

const (
	StateIdle State = iota
	StateRunning
	StatePaused
	StateStopped
	StateError
	StateTerminated
)

// StateMachine manages state transitions
type StateMachine struct {
	currentState State
	history      []State
}

// NewStateMachine creates a new state machine
func NewStateMachine() *StateMachine {
	return &StateMachine{
		currentState: StateIdle,
		history:      []State{StateIdle},
	}
}

// Start transitions from Idle to Running
func (sm *StateMachine) Start() error {
	if sm.currentState != StateIdle {
		return errors.New("cannot start: not in idle state")
	}

	sm.currentState = StateRunning
	sm.history = append(sm.history, StateRunning)
	return nil
}

// Pause transitions from Running to Paused
func (sm *StateMachine) Pause() error {
	if sm.currentState != StateRunning {
		return errors.New("cannot pause: not in running state")
	}

	sm.currentState = StatePaused
	sm.history = append(sm.history, StatePaused)
	return nil
}

// Resume transitions from Paused to Running
func (sm *StateMachine) Resume() error {
	if sm.currentState != StatePaused {
		return errors.New("cannot resume: not in paused state")
	}

	sm.currentState = StateRunning
	sm.history = append(sm.history, StateRunning)
	return nil
}

// Stop transitions to Stopped
func (sm *StateMachine) Stop() error {
	if sm.currentState == StateIdle || sm.currentState == StateStopped || sm.currentState == StateTerminated {
		return errors.New("cannot stop: already in terminal state")
	}

	sm.currentState = StateStopped
	sm.history = append(sm.history, StateStopped)
	return nil
}

// ErrorOccurred transitions to Error state
func (sm *StateMachine) ErrorOccurred() error {
	if sm.currentState == StateTerminated {
		return errors.New("cannot error: already terminated")
	}

	sm.currentState = StateError
	sm.history = append(sm.history, StateError)
	return nil
}

// Recover transitions from Error back to Running
func (sm *StateMachine) Recover() error {
	if sm.currentState != StateError {
		return errors.New("cannot recover: not in error state")
	}

	sm.currentState = StateRunning
	sm.history = append(sm.history, StateRunning)
	return nil
}

// Terminate transitions to Terminated
func (sm *StateMachine) Terminate() error {
	if sm.currentState == StateTerminated {
		return errors.New("already terminated")
	}

	sm.currentState = StateTerminated
	sm.history = append(sm.history, StateTerminated)
	return nil
}

// GetCurrentState returns the current state
func (sm *StateMachine) GetCurrentState() State {
	return sm.currentState
}

// IsRunning checks if machine is in Running state
func (sm *StateMachine) IsRunning() bool {
	return sm.currentState == StateRunning
}

// IsPaused checks if machine is in Paused state
func (sm *StateMachine) IsPaused() bool {
	return sm.currentState == StatePaused
}

// IsInError checks if machine is in Error state
func (sm *StateMachine) IsInError() bool {
	return sm.currentState == StateError
}

// GetStateString returns string representation of state
func GetStateString(s State) string {
	switch s {
	case StateIdle:
		return "Idle"
	case StateRunning:
		return "Running"
	case StatePaused:
		return "Paused"
	case StateStopped:
		return "Stopped"
	case StateError:
		return "Error"
	case StateTerminated:
		return "Terminated"
	default:
		return "Unknown"
	}
}

// PrintHistory prints the state transition history
func (sm *StateMachine) PrintHistory() string {
	var result string
	for i, state := range sm.history {
		if i > 0 {
			result += " -> "
		}
		result += GetStateString(state)
	}
	return result
}

// GetTransitionCount returns number of state transitions
func (sm *StateMachine) GetTransitionCount() int {
	return len(sm.history) - 1
}

// CanTransitionTo checks if transition to target state is valid
func (sm *StateMachine) CanTransitionTo(targetState State) bool {
	switch sm.currentState {
	case StateIdle:
		return targetState == StateRunning
	case StateRunning:
		return targetState == StatePaused || targetState == StateStopped || targetState == StateError
	case StatePaused:
		return targetState == StateRunning || targetState == StateStopped || targetState == StateError
	case StateError:
		return targetState == StateRunning || targetState == StateTerminated
	case StateStopped, StateTerminated:
		return false
	}
	return false
}

// Execute performs a sequence of operations
func (sm *StateMachine) Execute(operations []string) error {
	for _, op := range operations {
		switch op {
		case "start":
			if err := sm.Start(); err != nil {
				return fmt.Errorf("start failed: %w", err)
			}
		case "pause":
			if err := sm.Pause(); err != nil {
				return fmt.Errorf("pause failed: %w", err)
			}
		case "resume":
			if err := sm.Resume(); err != nil {
				return fmt.Errorf("resume failed: %w", err)
			}
		case "stop":
			if err := sm.Stop(); err != nil {
				return fmt.Errorf("stop failed: %w", err)
			}
		case "terminate":
			if err := sm.Terminate(); err != nil {
				return fmt.Errorf("terminate failed: %w", err)
			}
		default:
			return fmt.Errorf("unknown operation: %s", op)
		}
	}
	return nil
}
