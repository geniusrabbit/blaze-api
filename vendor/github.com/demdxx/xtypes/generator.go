package xtypes

import "context"

// GeneratorSimple returns generator which is not controlled from external
func GeneratorSimple[T any](size int, genFn func(prev T) (T, bool)) <-chan T {
	ch := make(chan T, size)
	go func() {
		defer close(ch)
		var (
			val T
			ok  bool
		)
		for {
			if val, ok = genFn(val); !ok {
				break
			}
			ch <- val
		}
	}()
	return (<-chan T)(ch)
}

// Generator returns generator with context Done support
func Generator[T any](ctx context.Context, size int, genFn func(ctx context.Context, prev T) (T, bool)) <-chan T {
	ch := make(chan T, size)
	go func() {
		defer close(ch)
		var (
			val T
			ok  bool
		)
		for {
			if val, ok = genFn(ctx, val); !ok {
				break
			}
			select {
			case <-ctx.Done():
				return
			case ch <- val:
			}
		}
	}()
	return (<-chan T)(ch)
}
