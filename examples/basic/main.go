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

func main() {
	ctx := context.Background()
	loader := dataloader.New[string, *ExampleData, string](ctx, batchLoadFn)

	thunk := loader.Load(ctx, "World")
	val, err := thunk.Get(ctx)
	fmt.Printf("value: %v, err: %v\n", val, err)
}
