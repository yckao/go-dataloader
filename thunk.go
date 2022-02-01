package dataloader

import ("context")

type Thunk[V interface{}] struct {
    pending   chan bool
    value     chan V
    err       chan error
}

func NewThunk[V interface{}]() *Thunk[V] {
    thunk := &Thunk[V]{
        pending:  make(chan bool, 1),
        value:    make(chan V, 1),
        err:      make(chan error, 1),
    }

    thunk.pending <- true

    return thunk
}

func (a *Thunk[V]) Get(ctx context.Context) (V, error) {
    select {
    case <- ctx.Done():
        return *new(V), ctx.Err()
    case v := <- a.value:
        a.value <- v
        return v, nil
    case err := <- a.err:
        a.err <- err
        return *new(V), err
    }
}

func (a *Thunk[V]) set(ctx context.Context, value V) (V, error) {
    select {
    case <- a.err:
    case <- a.value:
    case <- a.pending:
    }

    a.value <- value

    return a.Get(ctx)
}

func (a *Thunk[V]) error(ctx context.Context, err error) (V, error) {
    select {
    case <- a.err:
    case <- a.value:
    case <- a.pending:
    }

    a.err <- err

    return a.Get(ctx)
}
