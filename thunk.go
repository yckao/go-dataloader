package dataloader

import (
	"context"
)

type Thunk[V interface{}] struct {
	pending chan bool
	data    chan *thunkData[V]
}

type thunkData[V interface{}] struct {
	value V
	err   error
}

func NewThunk[V interface{}]() *Thunk[V] {
	Thunk := &Thunk[V]{
		pending: make(chan bool, 1),
		data:    make(chan *thunkData[V], 1),
	}

	Thunk.pending <- true

	return Thunk
}

func (a *Thunk[V]) Get(ctx context.Context) (V, error) {
	select {
	case <-ctx.Done():
		return *new(V), ctx.Err()
	case v := <-a.data:
		a.data <- v
		return v.value, v.err
	}
}

func (a *Thunk[V]) set(ctx context.Context, value V) (V, error) {
	select {
	case <-a.data:
	case <-a.pending:
	}

	a.data <- &thunkData[V]{value: value}

	return a.Get(ctx)
}

func (a *Thunk[V]) error(ctx context.Context, err error) (V, error) {
	select {
	case <-a.data:
	case <-a.pending:
	}

	a.data <- &thunkData[V]{err: err}

	return a.Get(ctx)
}
