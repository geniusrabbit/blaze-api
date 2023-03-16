package gosql

import (
	"database/sql/driver"
	"sort"
)

// NullableNumberArray type of field
type NullableNumberArray[T Number] []T

// Value implements the driver.Valuer interface, []int field
func (f NullableNumberArray[T]) Value() (driver.Value, error) {
	if f == nil {
		return nil, nil
	}
	return ArrayNumberEncode('{', '}', f).String(), nil
}

// Scan implements the driver.Valuer interface, []int field
func (f *NullableNumberArray[T]) Scan(value interface{}) error {
	if res, err := ArrayNumberDecode[T](value, '{', '}'); err == nil {
		*f = NullableNumberArray[T](res)
	} else {
		return err
	}
	return nil
}

// MarshalJSON implements the json.Marshaler
func (f NullableNumberArray[T]) MarshalJSON() ([]byte, error) {
	if f == nil {
		return []byte("null"), nil
	}
	return ArrayNumberEncode('[', ']', f).Bytes(), nil
}

// UnmarshalJSON implements the json.Unmarshaller
func (f *NullableNumberArray[T]) UnmarshalJSON(b []byte) error {
	res, err := ArrayNumberDecode[T](b, '[', ']')
	*f = NullableNumberArray[T](res)
	return err
}

// DecodeValue implements the gocast.Decoder
func (f *NullableNumberArray[T]) DecodeValue(v interface{}) error {
	switch val := v.(type) {
	case []T:
		*f = NullableNumberArray[T](val)
	case NullableNumberArray[T]:
		*f = val
	case NumberArray[T]:
		*f = NullableNumberArray[T](val)
	case []byte, string:
		list, err := ArrayNumberDecode[T](v, '[', ']')
		if err != nil {
			return err
		}
		*f = NullableNumberArray[T](list)
	default:
		return ErrInvalidDecodeValue
	}
	return nil
}

// Sort ints array
func (f NullableNumberArray[T]) Sort() NullableNumberArray[T] {
	sort.Sort(f)
	return f
}

// Len of array
func (s NullableNumberArray[T]) Len() int           { return len(s) }
func (s NullableNumberArray[T]) Less(i, j int) bool { return s[i] < s[j] }
func (s NullableNumberArray[T]) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// IndexOf array value
func (f NullableNumberArray[T]) IndexOf(v T) int {
	if f == nil {
		return -1
	}
	for i, vl := range f {
		if vl == v {
			return i
		}
	}
	return -1
}

// OneOf value in array
func (f NullableNumberArray[T]) OneOf(vals []T) bool {
	if len(f) < 1 || len(vals) < 1 {
		return false
	}
	for _, v := range vals {
		if f.IndexOf(v) != -1 {
			return true
		}
	}
	return false
}

// Ordered object
func (f NullableNumberArray[T]) Ordered() NullableOrderedNumberArray[T] {
	f.Sort()
	return NullableOrderedNumberArray[T](f)
}

// Filter current array and create filtered copy
func (f NullableNumberArray[T]) Filter(fn func(v T) bool) NullableNumberArray[T] {
	resp := make(NullableNumberArray[T], 0, len(f))
	for _, v := range f {
		if fn(v) {
			resp = append(resp, v)
		}
	}
	return resp
}

// Map transforms every value into the target
func (f NullableNumberArray[T]) Map(fn func(v T) (T, bool)) NullableNumberArray[T] {
	resp := make(NullableNumberArray[T], 0, len(f))
	for _, v := range f {
		if vl, ok := fn(v); ok {
			resp = append(resp, vl)
		}
	}
	return resp
}
