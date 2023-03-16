package gosql

import (
	"database/sql/driver"
	"sort"
)

// OrderedNumberArray type of field
type OrderedNumberArray[T Number] []T

// Value implements the driver.Valuer interface, []int field
func (f OrderedNumberArray[T]) Value() (driver.Value, error) {
	if f == nil {
		return "{}", nil
	}
	return NullableNumberArray[T](f).Value()
}

// Scan implements the driver.Valuer interface, []int field
func (f *OrderedNumberArray[T]) Scan(value interface{}) error {
	if value == nil {
		return ErrNullValueNotAllowed
	}
	return (*NullableNumberArray[T])(f).Scan(value)
}

// UnmarshalJSON implements the json.Unmarshaller
func (f *OrderedNumberArray[T]) UnmarshalJSON(b []byte) error {
	if b == nil {
		return ErrNullValueNotAllowed
	}
	return (*NullableNumberArray[T])(f).UnmarshalJSON(b)
}

// MarshalJSON implements the json.Marshaler
func (f OrderedNumberArray[T]) MarshalJSON() ([]byte, error) {
	return NumberArray[T](f).MarshalJSON()
}

// DecodeValue implements the gocast.Decoder
func (f *OrderedNumberArray[T]) DecodeValue(v interface{}) error {
	if v == nil {
		return ErrNullValueNotAllowed
	}
	return (*NullableNumberArray[T])(f).DecodeValue(v)
}

// Sort ints array
func (f OrderedNumberArray[T]) Sort() OrderedNumberArray[T] {
	sort.Sort(f)
	return f
}

// Len of array
func (s OrderedNumberArray[T]) Len() int           { return len(s) }
func (s OrderedNumberArray[T]) Less(i, j int) bool { return s[i] < s[j] }
func (s OrderedNumberArray[T]) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// IndexOf array value
func (f OrderedNumberArray[T]) IndexOf(v T) int {
	i := sort.Search(f.Len(), func(i int) bool { return f[i] >= v })
	if i >= 0 && i < f.Len() && f[i] == v {
		return i
	}
	return -1
}

// OneOf value in array
func (f OrderedNumberArray[T]) OneOf(vals []T) bool {
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
func (f OrderedNumberArray[T]) Filter(fn func(v T) bool) OrderedNumberArray[T] {
	return OrderedNumberArray[T](NullableNumberArray[T](f).Filter(fn))
}

// Map transforms every value into the target
func (f OrderedNumberArray[T]) Map(fn func(v T) (T, bool)) OrderedNumberArray[T] {
	return OrderedNumberArray[T](NullableNumberArray[T](f).Map(fn))
}
