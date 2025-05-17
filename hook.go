package dataloader

import "context"

// Hook is used to observe batch execution.
type Hook[K any, V any] interface {
	// BeforeBatch will be called before executing a batch with the keys to load.
	BeforeBatch(ctx context.Context, keys []K)
	// AfterBatch will be called after executing a batch. It receives the same
	// keys and the results produced by the batch load function.
	AfterBatch(ctx context.Context, keys []K, results []Result[V])
}
