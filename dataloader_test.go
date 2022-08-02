package dataloader

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

type ErrorCache[C comparable, V interface{}] struct {
	err   error
	errOn string
}

func NewErrorCache[C comparable, V interface{}](err error, errOn string) *ErrorCache[C, V] {
	return &ErrorCache[C, V]{
		err:   err,
		errOn: errOn,
	}
}
func (c *ErrorCache[C, V]) Get(ctx context.Context, key C) (V, error) {
	if c.errOn == "Get" {
		return *new(V), c.err
	} else {
		return *new(V), nil
	}
}
func (c *ErrorCache[C, V]) Set(ctx context.Context, key C, val V) error {
	if c.errOn == "Set" {
		return c.err
	} else {
		return nil
	}
}
func (c *ErrorCache[C, V]) Delete(ctx context.Context, key C) error {
	if c.errOn == "Delete" {
		return c.err
	} else {
		return nil
	}
}
func (c *ErrorCache[C, V]) Clear(ctx context.Context) error {
	if c.errOn == "Clear" {
		return c.err
	} else {
		return nil
	}
}

var _ = Describe("DataLoader", func() {
	It("can batch load request", func() {
		ctx := context.TODO()
		loadCount := 0
		batchLoadFn := func(ctx context.Context, keys []string) []Result[string] {
			defer GinkgoRecover()

			loadCount += 1
			result := make([]Result[string], len(keys))
			for index, key := range keys {
				result[index] = Result[string]{
					Value: "res:" + key,
				}
			}

			return result
		}

		loader := New[string, string, string](ctx, batchLoadFn)
		wg := &sync.WaitGroup{}
		wg.Add(4)

		for i := 0; i < 4; i++ {
			k := fmt.Sprintf("key%d", i)
			r := fmt.Sprintf("res:key%d", i)
			go func() {
				defer GinkgoRecover()
				defer wg.Done()

				val, err := loader.Load(ctx, k).Get(ctx)
				Expect(err).To(BeNil())
				Expect(val).To(Equal(r))
			}()
		}

		wg.Wait()
		Expect(loadCount).To(Equal(1))
	})

	It("split batch if over max batch size", func() {
		ctx := context.TODO()
		loadCount := 0
		batchLoadFn := func(ctx context.Context, keys []string) []Result[string] {
			loadCount += 1
			result := make([]Result[string], len(keys))
			for index, key := range keys {
				result[index] = Result[string]{
					Value: "res:" + key,
				}
			}
			return result
		}

		loader := New[string, string, string](
			ctx, batchLoadFn,
			WithMaxBatchSize[string, string, string](2),
			WithBatchScheduleFn[string, string, string](NewTimeWindowScheduler(1*time.Second)),
		)
		wg := &sync.WaitGroup{}
		wg.Add(5)

		for i := 0; i < 5; i++ {
			k := fmt.Sprintf("key%d", i)
			r := fmt.Sprintf("res:key%d", i)
			go func() {
				defer GinkgoRecover()
				defer wg.Done()

				val, err := loader.Load(ctx, k).Get(ctx)
				Expect(err).To(BeNil())
				Expect(val).To(Equal(r))
			}()
		}

		wg.Wait()
		Expect(loadCount).To(Equal(3))
	})

	It("use cached thunk before loaded", func() {
		ctx := context.TODO()
		loadCount := 0
		batchLoadFn := func(ctx context.Context, keys []string) []Result[string] {
			defer GinkgoRecover()
			Expect(len(keys)).To(Equal(1))

			loadCount += 1
			result := make([]Result[string], len(keys))
			for index, _ := range keys {
				result[index] = Result[string]{
					Value: "bar",
				}
			}

			return result
		}

		loader := New[string, string, string](ctx, batchLoadFn)
		wg := &sync.WaitGroup{}
		wg.Add(4)

		for i := 0; i < 4; i++ {
			k := "foo"
			r := "bar"
			go func() {
				defer GinkgoRecover()
				defer wg.Done()

				val, err := loader.Load(ctx, k).Get(ctx)
				Expect(err).To(BeNil())
				Expect(val).To(Equal(r))
			}()
		}

		wg.Wait()
		Expect(loadCount).To(Equal(1))
	})

	It("use cached thunk after loaded", func() {
		ctx := context.TODO()
		loadCount := 0
		batchLoadFn := func(ctx context.Context, keys []string) []Result[string] {
			defer GinkgoRecover()
			Expect(len(keys)).To(Equal(1))

			loadCount += 1
			result := make([]Result[string], len(keys))
			for index, _ := range keys {
				result[index] = Result[string]{
					Value: "bar",
				}
			}

			return result
		}

		loader := New[string, string, string](ctx, batchLoadFn)
		wg := &sync.WaitGroup{}
		wg.Add(4)

		k := "foo"
		r := "bar"

		v1, err := loader.Load(ctx, k).Get(ctx)
		Expect(err).To(BeNil())
		Expect(v1).To(Equal(r))

		v2, err := loader.Load(ctx, k).Get(ctx)
		Expect(err).To(BeNil())
		Expect(v2).To(Equal(r))

		Expect(loadCount).To(Equal(1))
	})

	It("when cache key function failed", func() {
		ctx := context.TODO()
		loadCount := 0
		batchLoadFn := func(ctx context.Context, keys []string) []Result[string] {
			defer GinkgoRecover()

			loadCount += 1
			result := make([]Result[string], len(keys))
			for index, key := range keys {
				result[index] = Result[string]{
					Value: "res:" + key,
				}
			}

			return result
		}

		expected := fmt.Errorf("expected error")
		cacheKeyFn := func(ctx context.Context, key string) (string, error) {
			return "", expected
		}

		loader := New[string, string, string](ctx, batchLoadFn, WithCacheKeyFn[string, string, string](cacheKeyFn))
		val, err := loader.Load(ctx, "foo").Get(ctx)
		Expect(err).To(Equal(expected))
		Expect(val).To(Equal(""))
		Expect(loadCount).To(Equal(0))
	})

	It("when cache get failed", func() {
		ctx := context.TODO()
		loadCount := 0
		batchLoadFn := func(ctx context.Context, keys []string) []Result[string] {
			defer GinkgoRecover()

			loadCount += 1
			result := make([]Result[string], len(keys))

			return result
		}
		expected := fmt.Errorf("expected error")
		cacheMap := NewErrorCache[string, *Thunk[string]](expected, "Get")

		loader := New[string, string, string](ctx, batchLoadFn, WithCacheMap[string, string, string](cacheMap))
		val, err := loader.Load(ctx, "foo").Get(ctx)
		Expect(err).To(Equal(expected))
		Expect(val).To(Equal(""))
		Expect(loadCount).To(Equal(0))
	})

	It("when cache set failed", func() {
		ctx := context.TODO()
		loadCount := 0
		batchLoadFn := func(ctx context.Context, keys []string) []Result[string] {
			defer GinkgoRecover()

			loadCount += 1
			result := make([]Result[string], len(keys))

			return result
		}
		expected := fmt.Errorf("expected error")
		cacheMap := NewErrorCache[string, *Thunk[string]](expected, "Set")

		loader := New[string, string, string](ctx, batchLoadFn, WithCacheMap[string, string, string](cacheMap))
		val, err := loader.Load(ctx, "foo").Get(ctx)
		Expect(err).To(Equal(expected))
		Expect(val).To(Equal(""))
		Expect(loadCount).To(Equal(0))
	})

	It("when batch load function failed", func() {
		ctx := context.TODO()
		expected := fmt.Errorf("expected error")

		loadCount := 0
		batchLoadFn := func(ctx context.Context, keys []string) []Result[string] {
			defer GinkgoRecover()

			loadCount += 1
			result := make([]Result[string], len(keys))
			for index, _ := range keys {
				result[index] = Result[string]{
					Error: expected,
				}
			}

			return result
		}

		loader := New[string, string, string](ctx, batchLoadFn)
		wg := &sync.WaitGroup{}
		wg.Add(4)

		for i := 0; i < 4; i++ {
			k := fmt.Sprintf("key%d", i)
			go func() {
				defer GinkgoRecover()
				defer wg.Done()

				val, err := loader.Load(ctx, k).Get(ctx)
				Expect(err).To(Equal(expected))
				Expect(val).To(Equal(""))
			}()
		}

		wg.Wait()
	})

	It("can load many", func() {
		ctx := context.TODO()
		loadCount := 0
		keys := []string{"foo", "bar", "baz"}
		values := map[string]string{
			"foo": "foobar",
			"bar": "barbaz",
			"baz": "bazfoo",
		}

		batchLoadFn := func(ctx context.Context, keys []string) []Result[string] {
			defer GinkgoRecover()

			loadCount += 1
			result := make([]Result[string], len(keys))
			for index, key := range keys {
				result[index] = Result[string]{
					Value: values[key],
				}
			}

			return result
		}

		loader := New[string, string, string](ctx, batchLoadFn)
		thunks := loader.LoadMany(ctx, keys)
		for index, thunk := range thunks {
			key := keys[index]
			val, err := thunk.Get(ctx)
			Expect(err).To(BeNil())
			Expect(val).To(Equal(values[key]))
		}
		Expect(loadCount).To(Equal(1))
	})

	It("reload after clear", func() {
		ctx := context.TODO()
		loadCount := 0
		values := map[string]string{
			"foo": "foobar",
		}

		batchLoadFn := func(ctx context.Context, keys []string) []Result[string] {
			defer GinkgoRecover()

			loadCount += 1
			result := make([]Result[string], len(keys))
			for index, key := range keys {
				result[index] = Result[string]{
					Value: values[key],
				}
			}

			return result
		}

		loader := New[string, string, string](ctx, batchLoadFn)
		v1, err := loader.Load(ctx, "foo").Get(ctx)
		Expect(v1).To(Equal(values["foo"]))
		Expect(err).To(BeNil())
		Expect(loadCount).To(Equal(1))

		loader.Clear(ctx, "foo")

		v2, err := loader.Load(ctx, "foo").Get(ctx)
		Expect(v2).To(Equal(values["foo"]))
		Expect(err).To(BeNil())
		Expect(loadCount).To(Equal(2))
	})

	It("reload after clear all", func() {
		ctx := context.TODO()
		loadCount := 0
		values := map[string]string{
			"foo": "foobar",
		}

		batchLoadFn := func(ctx context.Context, keys []string) []Result[string] {
			defer GinkgoRecover()

			loadCount += 1
			result := make([]Result[string], len(keys))
			for index, key := range keys {
				result[index] = Result[string]{
					Value: values[key],
				}
			}

			return result
		}

		loader := New[string, string, string](ctx, batchLoadFn)
		v1, err := loader.Load(ctx, "foo").Get(ctx)
		Expect(v1).To(Equal(values["foo"]))
		Expect(err).To(BeNil())
		Expect(loadCount).To(Equal(1))

		loader.ClearAll(ctx)

		v2, err := loader.Load(ctx, "foo").Get(ctx)
		Expect(v2).To(Equal(values["foo"]))
		Expect(err).To(BeNil())
		Expect(loadCount).To(Equal(2))
	})

	It("use cached after set prime", func() {
		ctx := context.TODO()
		loadCount := 0
		batchLoadFn := func(ctx context.Context, keys []string) []Result[string] {
			defer GinkgoRecover()

			result := make([]Result[string], len(keys))
			return result
		}

		loader := New[string, string, string](ctx, batchLoadFn)
		loader.Prime(ctx, "foo", "bar")
		val, err := loader.Load(ctx, "foo").Get(ctx)
		Expect(val).To(Equal("bar"))
		Expect(err).To(BeNil())
		Expect(loadCount).To(Equal(0))
	})

	It("can manually dispatch batch", func() {
		ctx := context.TODO()
		loadCount := 0
		values := map[string]string{
			"foo": "foobar",
		}

		batchLoadFn := func(ctx context.Context, keys []string) []Result[string] {
			defer GinkgoRecover()

			loadCount += 1
			result := make([]Result[string], len(keys))
			for index, key := range keys {
				result[index] = Result[string]{
					Value: values[key],
				}
			}

			return result
		}

		loader := New[string, string, string](ctx, batchLoadFn, WithMaxBatchSize[string, string, string](200), WithBatchScheduleFn[string, string, string](NewTimeWindowScheduler(1*time.Second)))
		start := time.Now()
		thunk := loader.Load(ctx, "foo")
		loader.Dispatch()
		val, err := thunk.Get(ctx)
		Expect(time.Now().Before(start.Add(500 * time.Millisecond))).To(BeTrue())
		Expect(val).To(Equal(values["foo"]))
		Expect(err).To(BeNil())
		Expect(loadCount).To(Equal(1))
	})

	It("can early dispatch if full", func() {
		ctx := context.TODO()
		loadCount := 0
		values := map[string]string{
			"foo": "foobar",
		}

		batchLoadFn := func(ctx context.Context, keys []string) []Result[string] {
			defer GinkgoRecover()

			loadCount += 1
			result := make([]Result[string], len(keys))
			for index, key := range keys {
				result[index] = Result[string]{
					Value: values[key],
				}
			}

			return result
		}

		loader := New[string, string, string](ctx, batchLoadFn, WithMaxBatchSize[string, string, string](0), WithBatchScheduleFn[string, string, string](NewTimeWindowScheduler(1*time.Second)))
		start := time.Now()
		thunk := loader.Load(ctx, "foo")
		val, err := thunk.Get(ctx)
		Expect(time.Now().Before(start.Add(500 * time.Millisecond))).To(BeTrue())
		Expect(val).To(Equal(values["foo"]))
		Expect(err).To(BeNil())
		Expect(loadCount).To(Equal(1))
	})

	It("dispatch multiple time should not crash dataloader", func() {
		ctx := context.TODO()
		batchLoadFn := func(ctx context.Context, keys []string) []Result[string] {
			defer GinkgoRecover()

			result := make([]Result[string], len(keys))
			return result
		}

		loader := New[string, string, string](ctx, batchLoadFn).(*loader[string, string, string])
		loader.dispatch()
		loader.dispatch()
	})
})

func TestDataloader(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dataloader Suite")
}
