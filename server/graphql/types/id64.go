package types

import (
	"encoding/json"
	"io"

	"github.com/99designs/gqlgen/graphql"
	"github.com/demdxx/gocast/v2"
)

// MarshalID64 redecalre the marshalel of standart scalar type uint64 as ID64
func MarshalID64(id uint64) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, gocast.Str(id))
	})
}

// UnmarshalID64 redecalre the unmarshalel of standart scalar type uint64 as ID64
func UnmarshalID64(v any) (uint64, error) {
	switch vt := v.(type) {
	case json.Number:
		nv, err := vt.Int64()
		return uint64(nv), err
	}
	return gocast.TryNumber[uint64](v)
}
