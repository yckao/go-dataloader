package dataloader

type option[K interface{}, V interface{}, C comparable] func(*loader[K, V, C])

func WithBatch[K interface{}, V interface{}, C comparable](useBatch bool) option[K, V, C] {
	return func(l *loader[K, V, C]) {
		l.maxBatchSize = 1
	}
}

func WithMaxBatchSize[K interface{}, V interface{}, C comparable](maxBatchSize int) option[K, V, C] {
	return func(l *loader[K, V, C]) {
		l.maxBatchSize = maxBatchSize
	}
}

func WithBatchScheduleFn[K interface{}, V interface{}, C comparable](batchScheduleFn BatchScheduleFn) option[K, V, C] {
	return func(l *loader[K, V, C]) {
		l.batchScheduleFn = batchScheduleFn
	}
}

func WithCacheKeyFn[K interface{}, V interface{}, C comparable](cacheKeyFn CacheKeyFn[K, C]) option[K, V, C] {
	return func(l *loader[K, V, C]) {
		l.cacheKeyFn = cacheKeyFn
	}
}

func WithCacheMap[K interface{}, V interface{}, C comparable](cacheMap CacheMap[C, *Thunk[V]]) option[K, V, C] {
	return func(l *loader[K, V, C]) {
		<-l.cacheMap
		l.cacheMap <- cacheMap
	}
}
