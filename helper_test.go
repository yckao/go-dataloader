package dataloader

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"context"
	"time"
)

var _ = Describe("NewTimeWindowScheduler", func() {
	It("should run after specified duration", func() {
		scheduler := NewTimeWindowScheduler(200 * time.Millisecond)
		runned := false

		go scheduler(context.TODO(), func() { runned = true })
		Expect(runned).To(BeFalse())

		<-time.After(200 * time.Millisecond)
		Expect(runned).To(BeTrue())
	})

	It("can cancel", func() {
		scheduler := NewTimeWindowScheduler(200 * time.Millisecond)
		runned := false

		ctx, cancel := context.WithCancel(context.TODO())
		go scheduler(ctx, func() { runned = true })
		cancel()

		<-time.After(200 * time.Millisecond)
		Expect(runned).To(BeFalse())
	})
})

var _ = Describe("NewMirrorCacheKey", func() {
	It("uncomparable value", func() {
		cacheKeyFn := NewMirrorCacheKey[[]string, string]()
		val, err := cacheKeyFn(context.TODO(), []string{})
		Expect(err).To(Equal(ErrUncomparableKey))
		Expect(val).To(Equal(""))
	})

	It("unconvertible value", func() {
		cacheKeyFn := NewMirrorCacheKey[struct{}, string]()
		val, err := cacheKeyFn(context.TODO(), struct{}{})
		Expect(err).To(Equal(ErrUnconvertibleKey))
		Expect(val).To(Equal(""))
	})

	It("string to string", func() {
		cacheKeyFn := NewMirrorCacheKey[string, string]()
		val, err := cacheKeyFn(context.TODO(), "foo")
		Expect(err).To(BeNil())
		Expect(val).To(Equal("foo"))
	})

	It("int32 to int64", func() {
		cacheKeyFn := NewMirrorCacheKey[int32, int64]()
		val, err := cacheKeyFn(context.TODO(), int32(123))
		Expect(err).To(BeNil())
		Expect(val).To(Equal(int64(123)))
	})
})
