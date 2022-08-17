package lifecycle

type OnceOption interface {
	apply(runner *DefaultOnceRunner)
}

type onceOptionFunc func(runner *DefaultOnceRunner)

func (f onceOptionFunc) apply(runner *DefaultOnceRunner) {
	f(runner)
}

func OnceCloseChecker(checker func() error) OnceOption {
	return onceOptionFunc(func(runner *DefaultOnceRunner) {
		runner.SetCloseChecker(checker)
	})
}

// An Option configures a Logger.
type Option interface {
	apply(runner *DefaultRunner)
}

// optionFunc wraps a func so it satisfies the Option interface.
type optionFunc func(runner *DefaultRunner)

func (f optionFunc) apply(runner *DefaultRunner) {
	f(runner)
}

func StartChecker(checker func() error) Option {
	return optionFunc(func(runner *DefaultRunner) {
		runner.SetStartChecker(checker)
	})
}

func CloseChecker(checker func() error) Option {
	return optionFunc(func(runner *DefaultRunner) {
		runner.SetCloseChecker(checker)
	})
}
