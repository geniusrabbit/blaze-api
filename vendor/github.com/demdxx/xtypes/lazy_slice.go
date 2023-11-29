package xtypes

type itemTransformator[T any] struct {
	f func(T) bool
	a func(T) T
}

// LazySlice type extended with banch of processing methods
type LazySlice[T any] struct {
	slice          []T
	transformators []itemTransformator[T]
}

// NewLazySlice creates new lazy slice
func NewLazySlice[T any](sl []T) *LazySlice[T] {
	return &LazySlice[T]{slice: sl}
}

// Filter slice values and return new slice without excluded values
func (sl *LazySlice[T]) Filter(filter func(val T) bool) *LazySlice[T] {
	return &LazySlice[T]{
		slice:          sl.slice,
		transformators: append(sl.transformators, itemTransformator[T]{f: filter}),
	}
}

// Apply the function to each element of the slice
func (sl *LazySlice[T]) Apply(apply func(val T) T) *LazySlice[T] {
	return &LazySlice[T]{
		slice:          sl.slice,
		transformators: append(sl.transformators, itemTransformator[T]{a: apply}),
	}
}

// Each iterates every element in the list
func (sl *LazySlice[T]) Each(iter func(val T)) *LazySlice[T] {
main:
	for _, val := range sl.slice {
		for _, t := range sl.transformators {
			if t.f != nil && !t.f(val) {
				continue main
			}
			if t.a != nil {
				val = t.a(val)
			}
		}
		iter(val)
	}
	return sl
}

// Commit all changes and return new slice
func (sl *LazySlice[T]) Commit() Slice[T] {
	nSlice := make([]T, 0, Min(len(sl.slice)/2, 10))
	sl.Each(func(val T) { nSlice = append(nSlice, val) })
	return nSlice
}
