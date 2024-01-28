package types

import (
	"fmt"
	"io"

	"github.com/geniusrabbit/gosql/v2"
)

// JSON implements IO custom type of JSON
type JSON gosql.JSON[any]

func JSONFrom(v any) (*JSON, error) {
	jobj, err := gosql.NewJSON[any](v)
	return (*JSON)(jobj), err
}

func MustJSONFrom(v any) *JSON {
	jobj, err := JSONFrom(v)
	if err != nil {
		panic(err)
	}
	return jobj
}

func (j *JSON) goJSON() *gosql.JSON[any] {
	return (*gosql.JSON[any])(j)
}

// Value object
func (j JSON) Value() any {
	return j.goJSON().Data
}

// SetValue from any object
func (j *JSON) SetValue(v any) error {
	return j.goJSON().SetValue(v)
}

// MarshalGQL implements method of interface graphql.Marshaler
func (j JSON) MarshalGQL(w io.Writer) {
	data, _ := j.goJSON().MarshalJSON()
	_, _ = w.Write(data)
}

// UnmarshalGQL implements method of interface graphql.Unmarshaler
func (j *JSON) UnmarshalGQL(v any) error {
	switch v := v.(type) {
	case []byte:
		return j.goJSON().UnmarshalJSON(v)
	case string:
		return j.goJSON().UnmarshalJSON([]byte(v))
	default:
		return fmt.Errorf("%T is not a supported as a duration", v)
	}
}
