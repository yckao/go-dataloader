package dataloader

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"context"
)

var _ = Describe("NoCache", func() {
	It("can invoke get and get zero value", func() {
		c1 := NewNoCache[string, string]()
		v1, err := c1.Get(context.TODO(), "key")
		Expect(v1).To(Equal(""))
		Expect(err).To(BeNil())

		c2 := NewNoCache[string, *Thunk[string]]()
		v2, err := c2.Get(context.TODO(), "key")
		Expect(v2).To(BeNil())
		Expect(err).To(BeNil())
	})

	It("can invoke set", func() {
		cache := NewNoCache[string, string]()
		err := cache.Set(context.TODO(), "key", "")
		Expect(err).To(BeNil())
	})

	It("can invoke delete", func() {
		cache := NewNoCache[string, string]()
		err := cache.Delete(context.TODO(), "key")
		Expect(err).To(BeNil())
	})

	It("can invoke clear", func() {
		cache := NewNoCache[string, string]()
		err := cache.Clear(context.TODO())
		Expect(err).To(BeNil())
	})
})

var _ = Describe("InMemoryCache", func() {
	It("can set value", func() {
		ctx := context.TODO()
		cache := NewInMemoryCache[string, string]()
		err := cache.Set(ctx, "foo", "bar")
		Expect(err).To(BeNil())
	})

	It("get zero value if not set", func() {
		ctx := context.TODO()
		c1 := NewInMemoryCache[string, string]()
		val, err := c1.Get(ctx, "foo")
		Expect(val)
		Expect(err).To(BeNil())

		c2 := NewNoCache[string, *Thunk[string]]()
		v2, err := c2.Get(ctx, "key")
		Expect(v2).To(BeNil())
		Expect(err).To(BeNil())
	})

	It("get correct value if set", func() {
		ctx := context.TODO()
		c1 := NewInMemoryCache[string, string]()
		c1.Set(ctx, "foo", "bar")

		v1, err := c1.Get(ctx, "foo")
		Expect(v1).To(Equal("bar"))
		Expect(err).To(BeNil())

		c2 := NewInMemoryCache[string, *Thunk[string]]()
		a2 := NewThunk[string]()
		c2.Set(ctx, "foobar", a2)

		v2, err := c2.Get(ctx, "foobar")
		Expect(v2).To(Equal(a2))
		Expect(err).To(BeNil())
	})

	It("get zero value if delete", func() {
		ctx := context.TODO()
		c1 := NewInMemoryCache[string, string]()

		c1.Set(ctx, "foo", "bar")
		c1.Delete(ctx, "foo")

		v1, err := c1.Get(ctx, "foo")
		Expect(v1).To(Equal(""))
		Expect(err).To(BeNil())
	})

	It("get zero value after clear", func() {
		ctx := context.TODO()
		c1 := NewInMemoryCache[string, string]()
		c1.Set(ctx, "foo", "bar")
		c1.Set(ctx, "foobar", "baz")
		c1.Clear(ctx)

		v1, err := c1.Get(ctx, "foo")
		Expect(v1).To(Equal(""))
		Expect(err).To(BeNil())

		v2, err := c1.Get(ctx, "foobar")
		Expect(v2).To(Equal(""))
		Expect(err).To(BeNil())
	})
})
