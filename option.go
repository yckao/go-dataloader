package dataloader

type option[K any, V any, C comparable] func(*loader[K, V, C])

func WithBatch[K any, V any, C comparable](useBatch bool) option[K, V, C] {
	return func(l *loader[K, V, C]) {
		if !useBatch {
			l.maxBatchSize = 1
		}
	}
}

func WithMaxBatchSize[K any, V any, C comparable](maxBatchSize int) option[K, V, C] {
	return func(l *loader[K, V, C]) {
		l.maxBatchSize = maxBatchSize
	}
}

func WithBatchScheduleFn[K any, V any, C comparable](batchScheduleFn BatchScheduleFn) option[K, V, C] {
	return func(l *loader[K, V, C]) {
		l.batchScheduleFn = batchScheduleFn
	}
}

func WithCacheKeyFn[K any, V any, C comparable](cacheKeyFn CacheKeyFn[K, C]) option[K, V, C] {
	return func(l *loader[K, V, C]) {
		l.cacheKeyFn = cacheKeyFn
	}
}

func WithCacheMap[K any, V any, C comparable](cacheMap CacheMap[C, *Thunk[V]]) option[K, V, C] {
	return func(l *loader[K, V, C]) {
		<-l.cacheMap
		l.cacheMap <- cacheMap
	}
}

func WithHook[K any, V any, C comparable](hook Hook[K, V]) option[K, V, C] {
	return func(l *loader[K, V, C]) {
		l.hook = hook
	}
}
