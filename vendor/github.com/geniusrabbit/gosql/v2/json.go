//
// @author Dmitry Ponomarev <demdxx@gmail.com> 2016 – 2018, 2022
//

package gosql

import (
	"database/sql/driver"
	"encoding/json"
)

// JSON field
type JSON[T any] struct {
	Data T
}

// NewJSON creates new JSON object
func NewJSON[T any](val any) (*JSON[T], error) {
	var obj JSON[T]
	if err := obj.SetValue(val); err != nil {
		return nil, err
	}
	return &obj, nil
}

// String value
func (f *JSON[T]) String() string {
	if f == nil {
		return "{}"
	}
	data, _ := f.MarshalJSON()
	return string(data)
}

// SetValue of json
func (f *JSON[T]) SetValue(value any) error {
	switch vl := value.(type) {
	case T:
		f.Data = vl
	case *T:
		f.Data = *vl
	case nil:
		return ErrNullValueNotAllowed
	case string:
		return f.UnmarshalJSON([]byte(vl))
	case []byte:
		return f.UnmarshalJSON(vl)
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		return f.UnmarshalJSON(data)
	}
	return nil
}

// Value implements the driver.Valuer interface, json field interface
func (f JSON[T]) Value() (_ driver.Value, err error) {
	if v, err := f.MarshalJSON(); err == nil && v != nil {
		return string(v), nil
	}
	return nil, err
}

// Scan implements the driver.Valuer interface, json field interface
func (f *JSON[T]) Scan(value any) error {
	var data []byte
	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	case nil:
		return ErrNullValueNotAllowed
	default:
		return ErrInvalidScan
	}
	return f.UnmarshalJSON(data)
}

// MarshalJSON implements the json.Marshaler
func (f JSON[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Data)
}

// UnmarshalJSON implements the json.Unmarshaller
func (f *JSON[T]) UnmarshalJSON(data []byte) error {
	f.Data = *new(T)
	return json.Unmarshal(data, &f.Data)
}

// DecodeValue implements the gocast.Decoder
func (f *JSON[T]) DecodeValue(v any) error {
	switch val := v.(type) {
	case []byte:
		return f.UnmarshalJSON(val)
	case string:
		return f.UnmarshalJSON([]byte(val))
	default:
		return f.SetValue(v)
	}
}
