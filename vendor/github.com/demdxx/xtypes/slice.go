package xtypes

import "sort"

// Slice type extended with banch of processing methods
type Slice[T any] []T

// SliceApply the function to each element of the slice
func SliceApply[T any, N any](sl []T, apply func(val T) N) Slice[N] {
	nSlice := make(Slice[N], 0, len(sl))
	for _, val := range sl {
		nSlice = append(nSlice, apply(val))
	}
	return nSlice
}

// SliceReduce slice and return new value
func SliceReduce[T any, R any](sl []T, reduce func(val T, ret *R)) R {
	ret := new(R)
	for _, val := range sl {
		reduce(val, ret)
	}
	return *ret
}

// SliceUnique return new slice without duplicated values
func SliceUnique[T comparable](sl []T) []T {
	m := make(map[T]bool)
	for _, val := range sl {
		m[val] = true
	}
	ret := make([]T, 0, len(m))
	for val := range m {
		ret = append(ret, val)
	}
	return ret
}

// Len of slice
func (sl Slice[T]) Len() int {
	return len(sl)
}

// First value from slice
func (sl Slice[T]) First() *T {
	if len(sl) == 0 {
		return nil
	}
	return &sl[0]
}

// FirstOr value from slice or default value
func (sl Slice[T]) FirstOr(def T) T {
	if v := sl.First(); v != nil {
		return *v
	}
	return def
}

// Last value from slice
func (sl Slice[T]) Last() *T {
	if len(sl) == 0 {
		return nil
	}
	return &sl[len(sl)-1]
}

// LastOr value from slice or default value
func (sl Slice[T]) LastOr(def T) T {
	if v := sl.Last(); v != nil {
		return *v
	}
	return def
}

// Append value to slice
func (sl Slice[T]) Append(val T) Slice[T] {
	return append(sl, val)
}

// Prepend value to slice
func (sl Slice[T]) Prepend(val T) Slice[T] {
	return append([]T{val}, sl...)
}

// ValueOr return value from slice or default value
func (sl Slice[T]) ValueOr(i int, def T) T {
	if i < 0 || i >= len(sl) {
		return def
	}
	return sl[i]
}

// RemoveAt value at index from slice
func (sl Slice[T]) RemoveAt(i int) Slice[T] {
	if i < 0 || i >= len(sl) {
		return sl
	}
	return append(sl[:i], sl[i+1:]...)
}

// RemoveRange values from slice
func (sl Slice[T]) RemoveRange(i, j int) Slice[T] {
	if i > j {
		i, j = j, i
	}
	if i < 0 || i >= len(sl) {
		return sl
	}
	if j >= len(sl) {
		return sl[:i]
	}
	return append(sl[:i], sl[j:]...)
}

// Copy slice
func (sl Slice[T]) Copy() Slice[T] {
	return append(Slice[T]{}, sl...)
}

// Filter slice values and return new slice without excluded values
func (sl Slice[T]) Filter(filter func(val T) bool) Slice[T] {
	nSlice := make([]T, 0, len(sl))
	for _, val := range sl {
		if filter(val) {
			nSlice = append(nSlice, val)
		}
	}
	return nSlice
}

// Apply the function to each element of the slice
func (sl Slice[T]) Apply(apply func(val T) T) Slice[T] {
	return SliceApply(sl, apply)
}

// ReduceIntoOne slice and return new single value
func (sl Slice[T]) ReduceIntoOne(apply func(val T, ret *T)) T {
	return SliceReduce(sl, apply)
}

// Sort slice values
func (sl Slice[T]) Sort(cmp func(a, b T) bool) Slice[T] {
	sort.Slice(sl, func(i, j int) bool { return cmp(sl[i], sl[j]) })
	return sl
}

// Each iterates every element in the list
func (sl Slice[T]) Each(iter func(val T)) Slice[T] {
	for _, val := range sl {
		iter(val)
	}
	return sl
}

// IndexOf slice values
func (sl Slice[T]) IndexOf(fn func(val T) bool) int {
	for i, val := range sl {
		if fn(val) {
			return i
		}
	}
	return -1
}

// Has slice values
func (sl Slice[T]) Has(fn func(val T) bool) bool {
	return sl.IndexOf(fn) > -1
}

// BinarySearch slice values
func (sl Slice[T]) BinarySearch(fn func(val T) bool) int {
	ret := sort.Search(len(sl), func(i int) bool { return fn(sl[i]) })
	if ret < len(sl) && fn(sl[ret]) {
		return ret
	}
	return -1
}
