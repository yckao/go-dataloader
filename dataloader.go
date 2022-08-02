package dataloader

import (
	"context"
	"time"
)

type DataLoader[K interface{}, V interface{}, C comparable] interface {
	Load(context.Context, K) *Thunk[V]
	LoadMany(context.Context, []K) []*Thunk[V]
	Clear(context.Context, K) DataLoader[K, V, C]
	ClearAll(ctx context.Context) DataLoader[K, V, C]
	Prime(context.Context, K, V) DataLoader[K, V, C]
	Dispatch()
}

type Result[V interface{}] struct {
	Value V
	Error error
}

type BatchLoadFn[K interface{}, V interface{}] func(context.Context, []K) []Result[V]
type BatchScheduleFn func(ctx context.Context, batch Batch, callback func())
type CacheKeyFn[K interface{}, C comparable] func(ctx context.Context, key K) (C, error)

func New[K interface{}, V interface{}, C comparable](ctx context.Context, batchLoadFn BatchLoadFn[K, V], options ...option[K, V, C]) DataLoader[K, V, C] {
	l := &loader[K, V, C]{
		ctx:             ctx,
		batches:         make(chan []*batch[K, V], 1),
		batchLoadFn:     batchLoadFn,
		batchScheduleFn: NewTimeWindowScheduler(16 * time.Millisecond),
		cacheKeyFn:      NewMirrorCacheKey[K, C](),
		cacheMap:        make(chan CacheMap[C, *Thunk[V]], 1),
		maxBatchSize:    100,
	}

	l.cacheMap <- NewInMemoryCache[C, *Thunk[V]]()
	l.batches <- []*batch[K, V]{}

	for _, option := range options {
		option(l)
	}

	return l
}

type loader[K interface{}, V interface{}, C comparable] struct {
	ctx             context.Context
	batches         chan []*batch[K, V]
	batchLoadFn     BatchLoadFn[K, V]
	batchScheduleFn BatchScheduleFn
	cacheKeyFn      CacheKeyFn[K, C]
	cacheMap        chan CacheMap[C, *Thunk[V]]
	maxBatchSize    int
}

type batch[K interface{}, V interface{}] struct {
	full     chan struct{}
	dispatch chan struct{}
	keys     []K
	thunks   []*Thunk[V]
}

func (b *batch[K, V]) Full() <-chan struct{} {
	return b.full
}

func (b *batch[K, V]) Dispatch() <-chan struct{} {
	return b.dispatch
}

type Batch interface {
	Full() <-chan struct{}
	Dispatch() <-chan struct{}
}

func (l *loader[K, V, C]) Load(ctx context.Context, key K) *Thunk[V] {
	cacheKey, err := l.cacheKeyFn(ctx, key)
	if err != nil {
		thunk := NewThunk[V]()
		thunk.error(ctx, err)
		return thunk
	}

	batches := <-l.batches

	cacheMap := <-l.cacheMap
	cached, err := cacheMap.Get(ctx, cacheKey)

	if err != nil {
		l.cacheMap <- cacheMap
		l.batches <- batches
		thunk := NewThunk[V]()
		thunk.error(ctx, err)
		return thunk
	}

	if cached != nil {
		l.cacheMap <- cacheMap
		l.batches <- batches
		return cached
	}

	thunk := NewThunk[V]()
	err = cacheMap.Set(ctx, cacheKey, thunk)
	if err != nil {
		l.cacheMap <- cacheMap
		l.batches <- batches
		thunk.error(ctx, err)
		return thunk
	}

	l.cacheMap <- cacheMap

	if len(batches) == 0 || len(batches[len(batches)-1].keys) >= l.maxBatchSize {
		b := &batch[K, V]{
			full:     make(chan struct{}),
			dispatch: make(chan struct{}),
			keys:     []K{},
			thunks:   []*Thunk[V]{},
		}

		batches = append(batches, b)

		go l.batchScheduleFn(l.ctx, b, l.dispatch)
	}

	bat := batches[len(batches)-1]
	bat.keys = append(bat.keys, key)
	bat.thunks = append(bat.thunks, thunk)

	if len(batches) != 0 && len(batches[len(batches)-1].keys) >= l.maxBatchSize {
		close(batches[len(batches)-1].full)
	}

	l.batches <- batches

	return thunk
}

func (l *loader[K, V, C]) LoadMany(ctx context.Context, keys []K) []*Thunk[V] {
	thunks := make([]*Thunk[V], len(keys))
	for index, key := range keys {
		thunks[index] = l.Load(ctx, key)
	}

	return thunks
}

func (l *loader[K, V, C]) Dispatch() {
	batches := <-l.batches
	for _, batch := range batches {
		select {
		case <-batch.full:
		case <-batch.dispatch:
		default:
			close(batch.dispatch)
		}
	}
	l.batches <- batches
}

func (l *loader[K, V, C]) dispatch() {
	ctx := l.ctx
	batches := <-l.batches

	if len(batches) == 0 {
		l.batches <- batches
		return
	}
	batch := batches[0]

	l.batches <- batches[1:]

	results := l.batchLoadFn(ctx, batch.keys)

	for index, res := range results {
		if res.Error != nil {
			batch.thunks[index].error(ctx, res.Error)
		} else {
			batch.thunks[index].set(ctx, res.Value)
		}
	}
}

func (l *loader[K, V, C]) Clear(ctx context.Context, key K) DataLoader[K, V, C] {
	cacheKey, _ := l.cacheKeyFn(ctx, key)
	cacheMap := <-l.cacheMap
	cacheMap.Delete(ctx, cacheKey)
	l.cacheMap <- cacheMap
	return l
}

func (l *loader[K, V, C]) ClearAll(ctx context.Context) DataLoader[K, V, C] {
	cacheMap := <-l.cacheMap
	cacheMap.Clear(ctx)
	l.cacheMap <- cacheMap
	return l
}

func (l *loader[K, V, C]) Prime(ctx context.Context, key K, value V) DataLoader[K, V, C] {
	cacheKey, _ := l.cacheKeyFn(ctx, key)
	thunk := NewThunk[V]()
	thunk.set(ctx, value)

	cacheMap := <-l.cacheMap
	cacheMap.Set(ctx, cacheKey, thunk)
	l.cacheMap <- cacheMap

	return l
}
