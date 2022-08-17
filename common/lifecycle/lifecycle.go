package lifecycle

type Lifecycle interface {
	OnceLifecycle

	AddRunningFuture(future chan error) chan error

	AddClosedFutureIfRunning(future chan error) chan error

	Run() error

	Start() error

	Restart() error
}

type Runner interface {
	OnceRunner

	DoStart() error
}

type defaultRunner struct {
	defaultOnceRunner
	doStart func() error
}

func (r *defaultRunner) DoStart() error {
	return r.doStart()
}

type DefaultRunner struct {
	abstractLifecycle
	defaultRunner
}

func (r *DefaultRunner) Lifecycle() Lifecycle {
	return r
}

type abstractLifecycle struct {
	abstractOnceLifecycle
	runningFutures []chan error
	startChecker   func() error
}

func (l *abstractLifecycle) SetStartChecker(checker func() error) {
	l.Lock()
	l.startChecker = checker
	l.Unlock()
}

func (l *abstractLifecycle) broadcastRunning(err error) {
	for _, c := range l.runningFutures {
		c <- err
	}
	l.runningFutures = l.runningFutures[:0]
}

func (l *abstractLifecycle) addRunningFuture(future chan error, makeIfNil bool) chan error {
	if future == nil && makeIfNil {
		future = make(chan error, 1)
	}
	if future != nil {
		l.runningFutures = append(l.runningFutures, future)
		return future
	}
	return nil
}

func (l *abstractLifecycle) AddRunningFuture(future chan error) chan error {
	l.Lock()
	defer l.Unlock()
	if l.Running() {
		return nil
	}
	return l.addRunningFuture(future, true)
}

func (l *abstractLifecycle) AddClosedFuture(future chan error) chan error {
	l.Lock()
	defer l.Unlock()
	l.addClosedFuture(future, true)
	return future
}

func (l *abstractLifecycle) AddClosedFutureIfRunning(future chan error) chan error {
	return l.abstractOnceLifecycle.AddClosedFuture(future)
}

func (l *abstractLifecycle) startCheck() error {
	if l.Restarting() {
		return NewStateRestartingError(l.name)
	} else if l.Running() {
		return NewStateRunningError(l.name)
	}
	if l.startChecker != nil {
		return l.startChecker()
	}
	return nil
}

func (l *abstractLifecycle) doStart() error {
	if err := l.runner.(Runner).DoStart(); err != nil {
		l.broadcastRunning(err)
		return err
	}
	l.ToRunning()
	l.broadcastRunning(nil)
	return nil
}

func (l *abstractLifecycle) start() error {
	l.Lock()
	defer l.Unlock()
	if err := l.startCheck(); err != nil {
		return err
	}
	return l.doStart()
}

func (l *abstractLifecycle) Run() error {
	if err := l.start(); err != nil {
		return err
	}
	return l.run()
}

func (l *abstractLifecycle) Start() error {
	if err := l.start(); err != nil {
		return err
	}
	go l.run()
	return nil
}

func (l *abstractLifecycle) Restart() error {
	l.Lock()
	if l.Restarting() {
		l.Unlock()
		return NewStateRestartingError(l.name)
	}

	l.ToRestarting()
	if l.Running() {
		future := make(chan error, 1)
		if err := l.doClose(future); err != nil {
			l.ToRestarted()
			l.Unlock()
			return err
		}
		l.Unlock()
		if err := <-future; err != nil {
			l.Lock()
			l.ToRestarted()
			l.Unlock()
			return err
		}
		l.Lock()
	}

	err := l.doStart()
	l.ToRestarted()
	l.Unlock()
	return err
}

func New(name string, doStart, doRun, doClose func() error, options ...Option) (*DefaultRunner, Lifecycle) {
	r := &DefaultRunner{
		abstractLifecycle: abstractLifecycle{
			abstractOnceLifecycle: abstractOnceLifecycle{
				name:  name,
				State: StateClosed,
			},
		},
		defaultRunner: defaultRunner{
			defaultOnceRunner: defaultOnceRunner{
				doRun:   doRun,
				doClose: doClose,
			},
			doStart: doStart,
		},
	}
	r.runner = r
	for _, option := range options {
		option.apply(r)
	}
	return r, &r.abstractLifecycle
}
