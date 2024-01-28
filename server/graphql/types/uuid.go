package types

import (
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
)

// MarshalUUID redecalre the marshalel of standart scalar type UUID
func MarshalUUID(uid uuid.UUID) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.Quote(uid.String()))
	})
}

// UnmarshalUUID redecalre the unmarshalel of standart scalar type UUID
func UnmarshalUUID(v any) (uuid.UUID, error) {
	switch v := v.(type) {
	case []byte:
		return uuid.ParseBytes(v)
	case string:
		return uuid.Parse(v)
	default:
		return uuid.Nil, fmt.Errorf("%T is not a supported as a UUID", v)
	}
}
