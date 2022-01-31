package dataloader

import (
  . "github.com/onsi/ginkgo/v2"
  . "github.com/onsi/gomega"

  "context"
  "errors"
  "time"
  "sync"
)

var _ = Describe("Thunk", func() {
	It("can wait value before it's set", func() {
		thunk := NewThunk[string]()
		expected := "foo"
		
		ctx, cancel := context.WithTimeout(context.TODO(), 1 * time.Second)

		go func() {
			defer GinkgoRecover()

			val, err := thunk.Get(ctx)
			Expect(err).To(BeNil())
			Expect(val).To(Equal(expected))
			cancel()
		}()

		thunk.set(ctx, expected)

		<- ctx.Done()
		Expect(ctx.Err()).To(Equal(context.Canceled))
	})

	It("can get value after it's set", func() {
		thunk := NewThunk[string]()
		expected := "foo"
		
		ctx, cancel := context.WithTimeout(context.TODO(), 1 * time.Second)

		thunk.set(ctx, expected)

		go func() {
			defer GinkgoRecover()

			val, err := thunk.Get(ctx)
			Expect(err).To(BeNil())
			Expect(val).To(Equal(expected))
			cancel()
		}()


		<- ctx.Done()
		Expect(ctx.Err()).To(Equal(context.Canceled))
	})

	It("can get value multiple times", func() {
		thunk := NewThunk[string]()
		expected := "foo"
		
		ctx, cancel := context.WithTimeout(context.TODO(), 1 * time.Second)
		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer GinkgoRecover()

			val, err := thunk.Get(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(expected))
			wg.Done()
		}()

		thunk.set(ctx, expected)

		go func() {
			defer GinkgoRecover()

			val, err := thunk.Get(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(expected))
			wg.Done()
		}()

		go func() {
			wg.Wait()
			cancel()
		}()

		select {
		case <- ctx.Done():
		}
		Expect(ctx.Err()).To(Equal(context.Canceled))
	})

	It("can get error multiple times", func() {
		thunk := NewThunk[string]()
		expected := errors.New("foo")
		
		ctx, cancel := context.WithTimeout(context.TODO(), 1 * time.Second)
		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer GinkgoRecover()

			val, err := thunk.Get(ctx)
			Expect(err).To(Equal(expected))
			Expect(val).To(Equal(""))
			wg.Done()
		}()

		thunk.error(ctx, expected)

		go func() {
			defer GinkgoRecover()

			val, err := thunk.Get(ctx)
			Expect(err).To(Equal(expected))
			Expect(val).To(Equal(""))
			wg.Done()
		}()

		go func() {
			wg.Wait()
			cancel()
		}()

		select {
		case <- ctx.Done():
		}
		Expect(ctx.Err()).To(Equal(context.Canceled))
	})

	It("can wait error before it's set", func() {
		thunk := NewThunk[string]()
		expected := errors.New("bar")
		
		ctx, cancel := context.WithTimeout(context.TODO(), 1 * time.Second)

		go func() {
			defer GinkgoRecover()

			val, err := thunk.Get(ctx)
			Expect(err).To(Equal(expected))
			Expect(val).To(Equal(""))
			cancel()
		}()

		thunk.error(ctx, expected)

		<- ctx.Done()
		Expect(ctx.Err()).To(Equal(context.Canceled))
	})

	It("can get error after it's set", func() {
		thunk := NewThunk[string]()
		expected := errors.New("bar")
		
		ctx, cancel := context.WithTimeout(context.TODO(), 1 * time.Second)

		thunk.error(ctx, expected)

		go func() {
			defer GinkgoRecover()

			val, err := thunk.Get(ctx)
			Expect(err).To(Equal(expected))
			Expect(val).To(Equal(""))
			cancel()
		}()


		<- ctx.Done()
		Expect(ctx.Err()).To(Equal(context.Canceled))
	})

	It("can override value with value", func() {
		thunk := NewThunk[string]()
		expected := "foo"
		
		ctx, cancel := context.WithTimeout(context.TODO(), 1 * time.Second)

		thunk.set(ctx, "bar")
		thunk.set(ctx, expected)

		go func() {
			defer GinkgoRecover()

			val, err := thunk.Get(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(expected))
			cancel()
		}()


		<- ctx.Done()
		Expect(ctx.Err()).To(Equal(context.Canceled))
	})

	It("can override value with error", func() {
		thunk := NewThunk[string]()
		expected := errors.New("foo")
		
		ctx, cancel := context.WithTimeout(context.TODO(), 1 * time.Second)

		thunk.set(ctx, "bar")
		thunk.error(ctx, expected)

		go func() {
			defer GinkgoRecover()

			val, err := thunk.Get(ctx)
			Expect(err).To(Equal(expected))
			Expect(val).To(Equal(""))
			cancel()
		}()


		<- ctx.Done()
		Expect(ctx.Err()).To(Equal(context.Canceled))
	})

	It("can override error with value", func() {
		thunk := NewThunk[string]()
		expected := "foo"
		
		ctx, cancel := context.WithTimeout(context.TODO(), 1 * time.Second)

		thunk.error(ctx, errors.New("bar"))
		thunk.set(ctx, expected)

		go func() {
			defer GinkgoRecover()

			val, err := thunk.Get(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(expected))
			cancel()
		}()


		<- ctx.Done()
		Expect(ctx.Err()).To(Equal(context.Canceled))
	})

	It("can override error with error", func() {
		thunk := NewThunk[string]()
		expected := errors.New("foo")
		
		ctx, cancel := context.WithTimeout(context.TODO(), 1 * time.Second)

		thunk.error(ctx, errors.New("bar"))
		thunk.error(ctx, expected)

		go func() {
			defer GinkgoRecover()

			val, err := thunk.Get(ctx)
			Expect(err).To(Equal(expected))
			Expect(val).To(Equal(""))
			cancel()
		}()


		<- ctx.Done()
		Expect(ctx.Err()).To(Equal(context.Canceled))
	})

	It("can cancel context", func() {
		thunk := NewThunk[string]()
		
		ctx, cancel := context.WithTimeout(context.TODO(), 1 * time.Second)

		done := make(chan struct{})

		go func() {
			defer GinkgoRecover()

			val, err := thunk.Get(ctx)
			Expect(val).To(Equal(""))
			Expect(err).To(Equal(context.Canceled))
			done <- struct{}{}
		}()

		cancel()

		<- done
	})
})
