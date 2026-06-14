package types

import (
	"encoding/json"
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

// UnmarshalGQL implements method of interface graphql.Unmarshaler.
// gqlgen passes already-parsed values (map[string]any, []any, primitives)
// when variables are used, and strings/bytes when inlined as literals.
func (j *NullableJSON) UnmarshalGQL(v any) error {
	switch v := v.(type) {
	case []byte:
		return j.goJSON().UnmarshalJSON(v)
	case string:
		return j.goJSON().UnmarshalJSON([]byte(v))
	case nil:
		j.goJSON().Data = nil
		return nil
	default:
		// Already-parsed value from variables (map[string]any, []any, bool, float64, …)
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		return j.goJSON().UnmarshalJSON(data)
	}
}
