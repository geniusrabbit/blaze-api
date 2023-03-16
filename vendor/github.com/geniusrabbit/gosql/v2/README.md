# gosql

[![Build Status](https://github.com/geniusrabbit/gosql/workflows/run%20tests/badge.svg)](https://github.com/geniusrabbit/gosql/actions?workflow=run%20tests)
[![Go Report Card](https://goreportcard.com/badge/github.com/geniusrabbit/gosql)](https://goreportcard.com/report/github.com/geniusrabbit/gosql)
[![GoDoc](https://godoc.org/github.com/geniusrabbit/gosql?status.svg)](https://godoc.org/github.com/geniusrabbit/gosql)
[![Coverage Status](https://coveralls.io/repos/github/geniusrabbit/gosql/badge.svg)](https://coveralls.io/github/geniusrabbit/gosql)

Library of standart sql collections an types like:

 * Char
 * NumberArrays generic (Ordered, Nullable) Numbers only (int, float, ...)
 * StringArray
 * JSON generic with any type of value `structs`, `scalars`, ...

New version supports generics from `go1.8` to make easear using of data processing.

```go
import (
  "github.com/geniusrabbit/gosql/v2"
)

type Model struct {
  ID uint64
  Title string

  Configuration gosql.JSON[Config]
  Tags gosql.StringArray

  DebugMetricQuartiles gosql.NullableOrderedNumberArray[float64]
}
```

# License MIT

[LICENSE](LICENSE)
