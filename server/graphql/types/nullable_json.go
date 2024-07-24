package types

import (
	"fmt"
	"io"

	"github.com/geniusrabbit/gosql/v2"
)

// NullableJSON implements IO custom type of JSON
type NullableJSON gosql.NullableJSON[any]

func NullableJSONFrom(v any) (*NullableJSON, error) {
	switch v := v.(type) {
	case nil:
	case *NullableJSON:
		return v, nil
	case *gosql.NullableJSON[any]:
		return (*NullableJSON)(v), nil
	}
	jobj, err := gosql.NewNullableJSON[any](v)
	return (*NullableJSON)(jobj), err
}

func MustNullableJSONFrom(v any) *NullableJSON {
	jobj, err := NullableJSONFrom(v)
	if err != nil {
		panic(err)
	}
	return jobj
}

func (j *NullableJSON) goJSON() *gosql.NullableJSON[any] {
	return (*gosql.NullableJSON[any])(j)
}

// Value object
func (j NullableJSON) Value() any {
	return j.goJSON().Data
}

// SetValue from any object
func (j *NullableJSON) SetValue(v any) error {
	return j.goJSON().SetValue(v)
}

// DataOr returns data or default value
func (j *NullableJSON) DataOr(def any) any {
	return j.goJSON().DataOr(def)
}

// MarshalGQL implements method of interface graphql.Marshaler
func (j NullableJSON) MarshalGQL(w io.Writer) {
	data, _ := j.goJSON().MarshalJSON()
	_, _ = w.Write(data)
}

// UnmarshalGQL implements method of interface graphql.Unmarshaler
func (j *NullableJSON) UnmarshalGQL(v any) error {
	switch v := v.(type) {
	case []byte:
		return j.goJSON().UnmarshalJSON(v)
	case string:
		return j.goJSON().UnmarshalJSON([]byte(v))
	default:
		return fmt.Errorf("%T is not a supported as a duration", v)
	}
}
