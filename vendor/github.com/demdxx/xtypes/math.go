package xtypes

import "golang.org/x/exp/constraints"

// Min value
// Deprecated: use `min` in go1.21+ instead.
func Min[T constraints.Ordered](x, y T) T {
	if x < y {
		return x
	}
	return y
}

// Max value
// Deprecated: use `max` in go1.21+ instead.
func Max[T constraints.Ordered](x, y T) T {
	if x > y {
		return x
	}
	return y
}
