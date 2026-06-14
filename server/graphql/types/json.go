package types

import (
	"encoding/json"
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

// UnmarshalGQL implements method of interface graphql.Unmarshaler.
// gqlgen passes already-parsed values (map[string]any, []any, primitives)
// when variables are used, and strings/bytes when inlined as literals.
func (j *JSON) UnmarshalGQL(v any) error {
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
