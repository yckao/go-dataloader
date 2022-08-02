package main

import (
	"context"
	"fmt"

	"github.com/yckao/go-dataloader"
)

type ExampleData struct {
	Message string
}

func batchLoadFn(ctx context.Context, keys []string) []dataloader.Result[*ExampleData] {
	result := make([]dataloader.Result[*ExampleData], len(keys))
	for index, key := range keys {
		result[index] = dataloader.Result[*ExampleData]{
			Value: &ExampleData{
				Message: fmt.Sprintf("Hello %s", key),
			},
		}
	}
	return result
}

func NewCustomizeScheduler() (dataloader.BatchScheduleFn, chan struct{}) {
	// we can use custom trigger. In this example we use another to trigger dispatch
	channel := make(chan struct{})
	return func(ctx context.Context, batch dataloader.Batch, callback func()) {
		select {
		case <-ctx.Done():
			return
		case <-channel:
			callback()
		}
	}, channel
}

func main() {
	ctx := context.Background()
	scheduleFn, dispatch := NewCustomizeScheduler()
	loader := dataloader.New[string, *ExampleData, string](
		ctx, batchLoadFn,
		dataloader.WithMaxBatchSize[string, *ExampleData, string](0),
		dataloader.WithBatchScheduleFn[string, *ExampleData, string](scheduleFn),
	)

	thunk := loader.Load(ctx, "World")
	dispatch <- struct{}{}
	val, err := thunk.Get(ctx)
	fmt.Printf("value: %v, err: %v\n", val, err)
}
