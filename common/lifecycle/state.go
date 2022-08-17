package lifecycle

const (
	// StateClosed indicates instance has been closed, conflicts with
	// StateRunning and StateClosing, cannot change to StateClosing
	StateClosed = 1 << iota

	// StateRunning indicates instance is running, conflicts with
	// StateClosed
	StateRunning

	// StateClosing indicates instance is running, conflicts with
	// StateClosed, must coexist with StateRunning
	StateClosing

	// StateRestarting indicates instance is restart, can coexist with
	// any other state
	StateRestarting
)

type State int

func (s State) Is(state State) bool {
	return s&state != 0
}

func (s State) Closed() bool {
	return s.Is(StateClosed)
}

func (s State) Running() bool {
	return s.Is(StateRunning)
}

func (s State) Closing() bool {
	return s.Is(StateClosing)
}

func (s State) Restarting() bool {
	return s.Is(StateRestarting)
}

func (s *State) ToClosed() {
	*s &= ^(StateRunning | StateClosing)
	*s = StateClosed
}

func (s *State) ToRunning() {
	*s &= ^StateClosed
	*s |= StateRunning
}

func (s *State) ToClosing() {
	if s.Closed() {
		panic("internal error: lifecycle cannot be changed from Closed to Closing")
	}
	*s |= StateClosing
}

func (s *State) ToRestarting() {
	*s |= StateRestarting
}

func (s *State) ToRestarted() {
	*s &= ^StateRestarting
}
