package dataloader

import (
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"

    "context"
    "reflect"
)

var _ = Describe("Option", func () {
    It("can disable batch", func() {
        dl := New[string, string, string](context.TODO(), func(ctx context.Context, keys []string) []Result[string] { return []Result[string]{} }, WithBatch[string, string, string](false))
        Expect(dl.(*loader[string, string, string]).maxBatchSize).To(Equal(1))
    })

    It("can set max batch size", func() {
        dl := New[string, string, string](context.TODO(), func(ctx context.Context, keys []string) []Result[string] { return []Result[string]{} }, WithMaxBatchSize[string, string, string](10))
        Expect(dl.(*loader[string, string, string]).maxBatchSize).To(Equal(10))
    })

    It("can set schedule function", func() {
        scheduleFn := (func(ctx context.Context, callback func()) { callback() })
        dl := New[string, string, string](context.TODO(), func(ctx context.Context, keys []string) []Result[string] { return []Result[string]{} }, WithBatchScheduleFn[string, string, string](scheduleFn))

        pointer1 := reflect.ValueOf(dl.(*loader[string, string, string]).batchScheduleFn).Pointer()
        pointer2 := reflect.ValueOf(scheduleFn).Pointer()

        Expect(pointer1).To(Equal(pointer2))
    })

    It("can set cache key function", func() {
        cacheKeyFn := func(ctx context.Context, key string) (string, error) { return "", nil }
        dl := New[string, string, string](context.TODO(), func(ctx context.Context, keys []string) []Result[string] { return []Result[string]{} }, WithCacheKeyFn[string, string, string](cacheKeyFn))

        pointer1 := reflect.ValueOf(dl.(*loader[string, string, string]).cacheKeyFn).Pointer()
        pointer2 := reflect.ValueOf(cacheKeyFn).Pointer()

        Expect(pointer1).To(Equal(pointer2))
    })

    It("can set cache map", func() {
        cacheMap := NewNoCache[string, *Thunk[string]]()
        dl := New[string, string, string](context.TODO(), func(ctx context.Context, keys []string) []Result[string] { return []Result[string]{} }, WithCacheMap[string, string, string](cacheMap))
        acutal := <- dl.(*loader[string, string, string]).cacheMap
        Expect(acutal).To(Equal(cacheMap))
    })
})
