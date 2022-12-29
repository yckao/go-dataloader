package dataloader

import (
	"context"
	"errors"
	"reflect"
	"time"
)

func NewTimeWindowScheduler(t time.Duration) BatchScheduleFn {
	return func(ctx context.Context, batch Batch, callback func()) {
		timer := time.NewTimer(t)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return
		case <-batch.Dispatch():
			callback()
		case <-batch.Full():
			callback()
		case <-timer.C:
			callback()
		}
	}
}

var (
	ErrUncomparableKey  = errors.New("cannot use uncomparable key in mirror cache key function")
	ErrUnconvertibleKey = errors.New("cannot use convert key to cache key function")
)

func NewMirrorCacheKey[K interface{}, C comparable]() CacheKeyFn[K, C] {
	return func(ctx context.Context, key K) (C, error) {
		kt := reflect.TypeOf(key)
		ct := reflect.TypeOf(*new(C))
		kv := reflect.ValueOf(key)
		if !kt.Comparable() {
			return *new(C), ErrUncomparableKey
		}

		if !kt.ConvertibleTo(ct) {
			return *new(C), ErrUnconvertibleKey
		}

		return kv.Convert(ct).Interface().(C), nil
	}
}
