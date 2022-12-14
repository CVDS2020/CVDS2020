package pool

import (
	"sync"
)

type Pool[X any] struct {
	p     sync.Pool
	newFn func() X
}

func (p *Pool[X]) new() any {
	return p.newFn()
}

func NewPool[X any](new func() X) *Pool[X] {
	p := &Pool[X]{newFn: new}
	p.p.New = p.new
	return p
}

func (p *Pool[X]) Get() X {
	return p.p.Get().(X)
}

func (p *Pool[X]) Put(val X) {
	p.p.Put(val)
}
