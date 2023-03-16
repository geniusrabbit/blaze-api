package gosql

import (
	"database/sql/driver"
	"sort"
)

// NullableOrderedNumberArray[T Number] type of field
type NullableOrderedNumberArray[T Number] []T

// Value implements the driver.Valuer interface, []int field
func (f NullableOrderedNumberArray[T]) Value() (driver.Value, error) {
	if f == nil {
		return nil, nil
	}
	return NullableNumberArray[T](f).Value()
}

// Scan implements the driver.Valuer interface, []int field
func (f *NullableOrderedNumberArray[T]) Scan(value interface{}) error {
	if value == nil {
		return ErrNullValueNotAllowed
	}
	return (*NullableNumberArray[T])(f).Scan(value)
}

// UnmarshalJSON implements the json.Unmarshaller
func (f *NullableOrderedNumberArray[T]) UnmarshalJSON(b []byte) error {
	if b == nil {
		return ErrNullValueNotAllowed
	}
	return (*NullableNumberArray[T])(f).UnmarshalJSON(b)
}

// MarshalJSON implements the json.Marshaler
func (f NullableOrderedNumberArray[T]) MarshalJSON() ([]byte, error) {
	return (NullableNumberArray[T](f)).MarshalJSON()
}

// DecodeValue implements the gocast.Decoder
func (f *NullableOrderedNumberArray[T]) DecodeValue(v interface{}) error {
	if v == nil {
		return ErrNullValueNotAllowed
	}
	return (*NullableNumberArray[T])(f).DecodeValue(v)
}

// Sort ints array
func (f NullableOrderedNumberArray[T]) Sort() NullableOrderedNumberArray[T] {
	sort.Sort(f)
	return f
}

// Len of array
func (s NullableOrderedNumberArray[T]) Len() int           { return len(s) }
func (s NullableOrderedNumberArray[T]) Less(i, j int) bool { return s[i] < s[j] }
func (s NullableOrderedNumberArray[T]) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// IndexOf array value
func (f NullableOrderedNumberArray[T]) IndexOf(v T) int {
	i := sort.Search(f.Len(), func(i int) bool { return f[i] >= v })
	if i >= 0 && i < f.Len() && f[i] == v {
		return i
	}
	return -1
}

// OneOf value in array
func (f NullableOrderedNumberArray[T]) OneOf(vals []T) bool {
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

// Filter current array and create filtered copy
func (f NullableOrderedNumberArray[T]) Filter(fn func(v T) bool) NullableOrderedNumberArray[T] {
	return NullableOrderedNumberArray[T](NullableNumberArray[T](f).Filter(fn))
}

// Map transforms every value into the target
func (f NullableOrderedNumberArray[T]) Map(fn func(v T) (T, bool)) NullableOrderedNumberArray[T] {
	return NullableOrderedNumberArray[T](NullableNumberArray[T](f).Map(fn))
}
