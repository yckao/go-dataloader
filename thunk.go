package dataloader

import (
	"context"
)

type Thunk[V any] struct {
	pending chan bool
	data    chan *thunkData[V]
}

type thunkData[V any] struct {
	value V
	err   error
}

func NewThunk[V any]() *Thunk[V] {
	thunk := &Thunk[V]{
		pending: make(chan bool, 1),
		data:    make(chan *thunkData[V], 1),
	}

	thunk.pending <- true

	return thunk
}

func (t *Thunk[V]) Get(ctx context.Context) (V, error) {
	select {
	case <-ctx.Done():
		return *new(V), ctx.Err()
	case v := <-t.data:
		t.data <- v
		return v.value, v.err
	}
}

func (t *Thunk[V]) set(ctx context.Context, value V) (V, error) {
	select {
	case <-t.data:
	case <-t.pending:
	}

	t.data <- &thunkData[V]{value: value}

	return t.Get(ctx)
}

func (t *Thunk[V]) error(ctx context.Context, err error) (V, error) {
	select {
	case <-t.data:
	case <-t.pending:
	}

	t.data <- &thunkData[V]{err: err}

	return t.Get(ctx)
}
