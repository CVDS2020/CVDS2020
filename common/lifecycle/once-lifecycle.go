package lifecycle

import (
	"sync"
)

type OnceLifecycle interface {
	AddClosedFuture(future chan error) chan error

	Close(future chan error) error

	CloseWait() (closeErr error, exitErr error)
}

type OnceRunner interface {
	DoRun() error

	DoClose() error
}

type defaultOnceRunner struct {
	doRun   func() error
	doClose func() error
}

func (r *defaultOnceRunner) DoRun() error {
	return r.doRun()
}

func (r *defaultOnceRunner) DoClose() error {
	return r.doClose()
}

type DefaultOnceRunner struct {
	abstractOnceLifecycle
	defaultOnceRunner
}

func (r *DefaultOnceRunner) OnceLifecycle() OnceLifecycle {
	return r
}

type abstractOnceLifecycle struct {
	runner OnceRunner
	name   string

	closedFutures []chan error
	closeChecker  func() error

	State
	sync.Mutex
}

func (l *abstractOnceLifecycle) Name() string {
	return l.name
}

func (l *abstractOnceLifecycle) SetCloseChecker(checker func() error) {
	l.Lock()
	l.closeChecker = checker
	l.Unlock()
}

func (l *abstractOnceLifecycle) broadcastClosed(err error) {
	for _, c := range l.closedFutures {
		c <- err
	}
	l.closedFutures = l.closedFutures[:0]
}

func (l *abstractOnceLifecycle) addClosedFuture(future chan error, makeIfNil bool) chan error {
	if future == nil && makeIfNil {
		future = make(chan error, 1)
	}
	if future != nil {
		l.closedFutures = append(l.closedFutures, future)
		return future
	}
	return nil
}

func (l *abstractOnceLifecycle) run() error {
	err := l.runner.DoRun()
	l.Lock()
	l.ToClosed()
	l.broadcastClosed(err)
	l.Unlock()
	return err
}

func (l *abstractOnceLifecycle) AddClosedFuture(future chan error) chan error {
	l.Lock()
	defer l.Unlock()
	if !l.Running() {
		return nil
	}
	return l.addClosedFuture(future, true)
}

func (l *abstractOnceLifecycle) closeCheck() error {
	if l.Restarting() {
		return NewStateRestartingError(l.name)
	} else if l.Closed() {
		return NewStateClosedError(l.name)
	}
	if l.closeChecker != nil {
		return l.closeChecker()
	}
	return nil
}

func (l *abstractOnceLifecycle) doClose(future chan error) error {
	if l.Closing() {
		l.addClosedFuture(future, false)
		return nil
	}
	if err := l.runner.DoClose(); err != nil {
		return err
	}
	l.ToClosing()
	l.addClosedFuture(future, false)
	return nil
}

func (l *abstractOnceLifecycle) Close(future chan error) error {
	l.Lock()
	defer l.Unlock()
	if err := l.closeCheck(); err != nil {
		return err
	}
	return l.doClose(future)
}

func (l *abstractOnceLifecycle) CloseWait() (closeErr error, exitErr error) {
	future := make(chan error, 1)
	if err := l.Close(future); err != nil {
		return err, nil
	}
	return nil, <-future
}

type AbstractOnceLifecycle struct {
	abstractOnceLifecycle
}

func NewOnce(name string, runFn func() error, closeFn func() error, options ...OnceOption) (*DefaultOnceRunner, OnceLifecycle) {
	r := &DefaultOnceRunner{
		abstractOnceLifecycle: abstractOnceLifecycle{
			name:  name,
			State: StateRunning,
		},
		defaultOnceRunner: defaultOnceRunner{
			doRun:   runFn,
			doClose: closeFn,
		},
	}
	r.runner = r
	for _, option := range options {
		option.apply(r)
	}
	go r.run()
	return r, &r.abstractOnceLifecycle
}
