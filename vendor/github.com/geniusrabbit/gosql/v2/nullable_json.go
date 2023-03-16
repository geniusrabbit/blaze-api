//
// @project GeniusRabbit
// @author Dmitry Ponomarev <demdxx@gmail.com> 2016 – 2018, 2022
//

package gosql

import (
	"database/sql/driver"
	"encoding/json"
)

// NullableJSON field
type NullableJSON[T any] struct {
	Data *T
}

// NewNullableJSON creates new JSON object
func NewNullableJSON[T any](val any) (*NullableJSON[T], error) {
	var obj NullableJSON[T]
	if err := obj.SetValue(val); err != nil {
		return nil, err
	}
	return &obj, nil
}

// String value
func (f *NullableJSON[T]) String() string {
	if f == nil || f.Data == nil {
		return "null"
	}
	data, _ := f.MarshalJSON()
	return string(data)
}

// SetValue of json
func (f *NullableJSON[T]) SetValue(value any) error {
	switch vl := value.(type) {
	case T:
		if f.Data == nil {
			f.Data = new(T)
		}
		*f.Data = vl
	case *T:
		if f.Data == nil {
			f.Data = new(T)
		}
		*f.Data = *vl
	case nil:
		f.Data = nil
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
func (f NullableJSON[T]) Value() (_ driver.Value, err error) {
	if v, err := f.MarshalJSON(); err == nil && v != nil {
		return string(v), nil
	}
	return nil, err
}

// Scan implements the driver.Valuer interface, json field interface
func (f *NullableJSON[T]) Scan(value any) error {
	var data []byte
	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	case nil:
		f.Data = nil
		return nil
	default:
		return ErrInvalidScan
	}
	if f.Data == nil {
		f.Data = new(T)
	}
	return json.Unmarshal(data, f.Data)
}

// MarshalJSON implements the json.Marshaler
func (f NullableJSON[T]) MarshalJSON() ([]byte, error) {
	if f.Data == nil {
		return []byte("null"), nil
	}
	return json.Marshal(f.Data)
}

// UnmarshalJSON implements the json.Unmarshaller
func (f *NullableJSON[T]) UnmarshalJSON(data []byte) error {
	target := new(T)
	err := json.Unmarshal(data, target)
	if err == nil {
		f.Data = target
	}
	return err
}

// DecodeValue implements the gocast.Decoder
func (f *NullableJSON[T]) DecodeValue(v any) error {
	switch val := v.(type) {
	case []byte:
		return f.UnmarshalJSON(val)
	case string:
		return f.UnmarshalJSON([]byte(val))
	default:
		return f.SetValue(v)
	}
}
