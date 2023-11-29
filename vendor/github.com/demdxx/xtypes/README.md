# XTypes

![License](https://img.shields.io/github/license/demdxx/xtypes)
[![GoDoc](https://godoc.org/github.com/demdxx/xtypes?status.svg)](https://godoc.org/github.com/demdxx/xtypes)
[![Testing Status](https://github.com/demdxx/xtypes/workflows/Tests/badge.svg)](https://github.com/demdxx/xtypes/actions?workflow=Tests)
[![Go Report Card](https://goreportcard.com/badge/github.com/demdxx/xtypes)](https://goreportcard.com/report/github.com/demdxx/xtypes)
[![Coverage Status](https://coveralls.io/repos/github/demdxx/xtypes/badge.svg?branch=main)](https://coveralls.io/github/demdxx/xtypes?branch=main)

Package represents basic go types and collections with extended functionalty.

## Collections

### Slice

```go
Slice[int]([]int{1, 2, 3}).
  Filter(func(val int) bool { return val > 1 }).
  // [2, 3]
  Apply(func(val int) int { return val * val }).
  // [4, 9]
  Sort(func(a, b int) bool { return a > b }).
  // [9, 4]
  ReduceIntoOne(func(val int, ret *int) { *ret += val })
  // 13

SliceReduce([]int{1, 2, 3}, func(val int, ret *float64) { *ret += 1/float64(val) })
// 1.83333...
```

### LazySlice

```go
NewLazySlice([]int{1, 2, 3}).
  Filter(func(val int) bool { return val > 1 }).
  // [2, 3]
  Apply(func(val int) int { return val * val }).
  // [4, 9]
  Commit()
```
