package dataloader

import (
	"context"
	"time"
	"reflect"
	"errors"
)

func NewTimeWindowScheduler(t time.Duration) BatchScheduleFn {
	return func(ctx context.Context, callback func()) {
		select {
		case <- ctx.Done():
			return
		case <- time.After(t):
			callback()
		}
	}
}

var (
	ErrUncomparableKey = errors.New("cannot use uncomparable key in mirror cache key function.")
	ErrUnconvertibleKey = errors.New("cannot use convert key to cache key function.")
)

func NewMirrorCacheKey[K interface{}, C comparable]() CacheKeyFn[K, C] {
	return func(ctx context.Context, key K) (C, error) {
		kt := reflect.TypeOf(key)
		ct := reflect.TypeOf(*new(C))
		kv := reflect.ValueOf(key)
		if kt.Comparable() == false {
			return *new(C), ErrUncomparableKey
		}

		if kt.ConvertibleTo(ct) == false {
			return *new(C), ErrUnconvertibleKey
		}

		return kv.Convert(ct).Interface().(C), nil
	}
}