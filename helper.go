package dataloader

import (
	"context"
	"errors"
	"reflect"
	"time"
	"unsafe"
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
	kt := reflect.TypeOf(*new(K))
	ct := reflect.TypeOf(*new(C))
	emptyC := *new(C)

	if !kt.Comparable() {
		return func(ctx context.Context, key K) (C, error) {
			return emptyC, ErrUncomparableKey
		}
	}

	if !kt.ConvertibleTo(ct) {
		return func(ctx context.Context, key K) (C, error) {
			return emptyC, ErrUnconvertibleKey
		}
	}

	if kt == ct {
		return func(ctx context.Context, key K) (C, error) {
			return *(*C)(unsafe.Pointer(&key)), nil
		}
	}

	return func(ctx context.Context, key K) (C, error) {
		kv := reflect.ValueOf(key)
		return kv.Convert(ct).Interface().(C), nil
	}

}
