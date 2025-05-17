# go-dataloader

A clean, safe, user-friendly implementation of [GraphQL's Dataloader](https://github.com/graphql/dataloader), written with generics in go.

## Features

- written in generics with strong type
- nearly same interface with original nodejs version dataloader
- promise like thunk design, simply call `val, err := loader.Load(ctx, id).Get(ctx)`
- customizable cache, easily wrap lru.
- customizable scheduler, can manual dispatch, or use time window (default)

## Requirement

- go >= 1.18

## Getting Started

```go
package main

import (
    "context"
    "fmt"

    "github.com/yckao/go-dataloader"
)

type ExampleData struct {
    Message string
}

// batch load function is provided by user
// incoming context is the context when create the loader
// return value can set either Result.Value or Result.Error
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

    // context is the dataloader belong to
    // first generics type is the type of key
    // second generics type is the type of value
    // third generics type is cache key
    // user can provide custom cache function to map orignal key to other
    loader := dataloader.New[string, *ExampleData, string](ctx, batchLoadFn)

    // load is a synchronize function return a promise like thunk
    thunk := loader.Load(ctx, "World")

    // get will wait until thunk value or error is set
    val, err := thunk.Get(ctx)

    fmt.Printf("value: %v, err: %v\n", val, err)
}
```

## Hooks

The loader can emit observability events around each batch. Implement the
`Hook` interface and pass it to `New` via `WithHook` to be notified when a batch
is executed.

```go
type loggingHook struct{}

func (loggingHook) BeforeBatch(ctx context.Context, keys []string) {
    log.Printf("loading %v", keys)
}

func (loggingHook) AfterBatch(ctx context.Context, keys []string, results []dataloader.Result[*ExampleData]) {
    log.Printf("loaded %v entries", len(results))
}

func main() {
    ctx := context.Background()
    loader := dataloader.New[string, *ExampleData, string](ctx, batchLoadFn,
        dataloader.WithHook[string, *ExampleData, string](loggingHook{}),
    )
    // ...
}
```

## TODO

- [ ] Examples
- [ ] Docs
- [ ] Rewrite tests with mock
