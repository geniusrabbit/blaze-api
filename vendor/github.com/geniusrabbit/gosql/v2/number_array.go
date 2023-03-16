package gosql

import (
	"database/sql/driver"
	"sort"
)

// NumberArray[T Number] type of field
type NumberArray[T Number] []T

// Value implements the driver.Valuer interface, []int field
func (f NumberArray[T]) Value() (driver.Value, error) {
	if f == nil {
		return "{}", nil
	}
	return NullableNumberArray[T](f).Value()
}

// Scan implements the driver.Valuer interface, []int field
func (f *NumberArray[T]) Scan(value interface{}) error {
	if value == nil {
		return ErrNullValueNotAllowed
	}
	return (*NullableNumberArray[T])(f).Scan(value)
}

// UnmarshalJSON implements the json.Unmarshaller
func (f *NumberArray[T]) UnmarshalJSON(b []byte) error {
	if b == nil {
		return ErrNullValueNotAllowed
	}
	return (*NullableNumberArray[T])(f).UnmarshalJSON(b)
}

// MarshalJSON implements the json.Marshaler
func (f NumberArray[T]) MarshalJSON() ([]byte, error) {
	if f == nil {
		return []byte("[]"), nil
	}
	return ArrayNumberEncode('[', ']', f).Bytes(), nil
}

// DecodeValue implements the gocast.Decoder
func (f *NumberArray[T]) DecodeValue(v interface{}) error {
	if v == nil {
		return ErrNullValueNotAllowed
	}
	return (*NullableNumberArray[T])(f).DecodeValue(v)
}

// Sort ints NumberArray
func (f NumberArray[T]) Sort() NumberArray[T] {
	sort.Sort(f)
	return f
}

// Len of NumberArray
func (s NumberArray[T]) Len() int           { return len(s) }
func (s NumberArray[T]) Less(i, j int) bool { return s[i] < s[j] }
func (s NumberArray[T]) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// IndexOf NumberArray value
func (f NumberArray[T]) IndexOf(v T) int {
	return NullableNumberArray[T](f).IndexOf(v)
}

// OneOf value in NumberArray
func (f NumberArray[T]) OneOf(vals []T) bool {
	return NullableNumberArray[T](f).OneOf(vals)
}

// Ordered object
func (f NumberArray[T]) Ordered() OrderedNumberArray[T] {
	f.Sort()
	return OrderedNumberArray[T](f)
}

// Filter current NumberArray and create filtered copy
func (f NumberArray[T]) Filter(fn func(v T) bool) NumberArray[T] {
	return NumberArray[T](NullableNumberArray[T](f).Filter(fn))
}

// Map transforms every value into the target
func (f NumberArray[T]) Map(fn func(v T) (T, bool)) NumberArray[T] {
	return NumberArray[T](NullableNumberArray[T](f).Map(fn))
}
