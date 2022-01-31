package dataloader

import (
    "context"
)

type CacheKeyFunc[K interface{}, C comparable] func(ctx context.Context, key K) (C, error)

type CacheMap[C comparable, V interface{}] interface {
    Get(ctx context.Context, key C) (V, error)
    Set(ctx context.Context, key C, val V) error
    Delete(ctx context.Context, key C) error
    Clear(ctx context.Context) error
}

type NoCache[C comparable, V interface{}] struct {}
func NewNoCache[C comparable, V interface{}]() *NoCache[C, V] { return &NoCache[C,V]{} }
func (c *NoCache[C, V]) Get(ctx context.Context, key C) (V, error) { return *new(V), nil }
func (c *NoCache[C, V]) Set(ctx context.Context, key C, val V) error { return nil }
func (c *NoCache[C, V]) Delete(ctx context.Context, key C) error { return nil }
func (c *NoCache[C, V]) Clear(ctx context.Context) error { return nil }


type InMemoryCache[C comparable, V interface{}] struct {
    items map[C]V
}

func NewInMemoryCache[C comparable, V interface{}]() *InMemoryCache[C, V] {
    return &InMemoryCache[C, V]{
        items: make(map[C]V),
    }
}

func (c *InMemoryCache[C, V]) Get(ctx context.Context, key C) (V, error) {
    return c.items[key], nil
}

func (c *InMemoryCache[C, V]) Set(ctx context.Context, key C, val V) error {
    c.items[key] = val
    return nil
}

func (c *InMemoryCache[C, V]) Delete(ctx context.Context, key C) error {
    delete(c.items, key)
    return nil
}

func (c *InMemoryCache[C, V]) Clear(ctx context.Context) error {
    c.items = make(map[C]V)
    return nil
}