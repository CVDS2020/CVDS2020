package lifecycle

import (
	"fmt"
	"github.com/CVDS2020/CVDS2020/common/id"
)

type StateError struct {
	TypeId uint64 `json:"code"`
	Runner string `json:"runner"`
}

var stateErrorTypeId uint64

var (
	StateClosedErrorTypeId     = id.Uint64Id(&stateErrorTypeId)
	StateRunningErrorTypeId    = id.Uint64Id(&stateErrorTypeId)
	StateClosingErrorTypeId    = id.Uint64Id(&stateErrorTypeId)
	StateRestartingErrorTypeId = id.Uint64Id(&stateErrorTypeId)
)

const (
	StateClosedErrorDescription     = "runner has been closed"
	StateRunningErrorDescription    = "runner is running"
	StateClosingErrorDescription    = "runner is closing"
	StateRestartingErrorDescription = "runner is restarting"
)

func NewStateError(e StateError, runner string) StateError {
	e.Runner = runner
	return e
}

func NewStateClosedError(runner string) StateError {
	return StateError{TypeId: StateClosingErrorTypeId, Runner: runner}
}

func NewStateRunningError(runner string) StateError {
	return StateError{TypeId: StateRunningErrorTypeId, Runner: runner}
}

func NewStateClosingError(runner string) StateError {
	return StateError{TypeId: StateClosingErrorTypeId, Runner: runner}
}

func NewStateRestartingError(runner string) StateError {
	return StateError{TypeId: StateRestartingErrorTypeId, Runner: runner}
}

func IsStateError(err error) bool {
	_, is := err.(*StateError)
	return is
}

func (e StateError) Description() string {
	switch e.TypeId {
	case StateClosedErrorTypeId:
		return StateClosedErrorDescription
	case StateRunningErrorTypeId:
		return StateRunningErrorDescription
	case StateClosingErrorTypeId:
		return StateClosingErrorDescription
	case StateRestartingErrorTypeId:
		return StateRestartingErrorDescription
	default:
		panic("invalid state error type id")
	}
}

func (e StateError) Error() string {
	switch e.TypeId {
	case StateClosedErrorTypeId:
		return fmt.Sprintf("runner %s has been closed", e.Runner)
	case StateRunningErrorTypeId:
		return fmt.Sprintf("runner %s is running", e.Runner)
	case StateClosingErrorTypeId:
		return fmt.Sprintf("runner %s is closing", e.Runner)
	case StateRestartingErrorTypeId:
		return fmt.Sprintf("runner %s is restarting", e.Runner)
	default:
		panic("invalid state error type id")
	}
}
