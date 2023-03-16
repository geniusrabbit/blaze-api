//
// @project GeniusRabbit
// @author Dmitry Ponomarev <demdxx@gmail.com> 2016, 2020 – 2021
//

package gosql

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"strings"
)

// NullableStringArray implementation
type NullableStringArray []string

// Join array to string
func (f NullableStringArray) Join(sep string) string {
	if f == nil {
		return ""
	}
	return strings.Join(f, sep)
}

// SetArray value
func (f *NullableStringArray) SetArray(arr []string) *NullableStringArray {
	*f = arr
	return f
}

// Value implements the driver.Valuer interface, []string field
func (f NullableStringArray) Value() (driver.Value, error) {
	if f == nil {
		return nil, nil
	}
	return encodeNullableStringArray('{', '}', '"', `""`, f).String(), nil
}

// Scan implements the driver.Valuer interface, []string field
func (f *NullableStringArray) Scan(value interface{}) error {
	switch val := value.(type) {
	case []byte:
		*f = decodeNullableStringArray(string(val), '{', '}', '"', `""`)
	case string:
		*f = decodeNullableStringArray(val, '{', '}', '"', `""`)
	case nil:
		*f = nil
	}
	return nil
}

// MarshalJSON implements the json.Marshaler
func (f NullableStringArray) MarshalJSON() ([]byte, error) {
	if f == nil {
		return []byte("null"), nil
	}
	return json.Marshal([]string(f))
}

// UnmarshalJSON implements the json.Unmarshaller
func (f *NullableStringArray) UnmarshalJSON(b []byte) error {
	var list []string
	if err := json.Unmarshal(b, &list); err != nil {
		return err
	}
	*f = list
	return nil
}

// DecodeValue implements the gocast.Decoder
func (f *NullableStringArray) DecodeValue(v interface{}) error {
	switch val := v.(type) {
	case []string:
		*f = NullableStringArray(val)
	case NullableStringArray:
		*f = val
	case StringArray:
		*f = NullableStringArray(val)
	case []byte:
		*f = NullableStringArray(strings.Split(string(val), ","))
	case string:
		*f = NullableStringArray(strings.Split(val, ","))
	default:
		return ErrInvalidDecodeValue
	}
	return nil
}

// Len of array
func (f NullableStringArray) Len() int {
	return len(f)
}

// IndexOf array value
func (f NullableStringArray) IndexOf(v string) int {
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
func (f NullableStringArray) OneOf(vals []string) bool {
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

///////////////////////////////////////////////////////////////////////////////

// StringArray implementation
type StringArray NullableStringArray

// Join array to string
func (f StringArray) Join(sep string) string {
	return NullableStringArray(f).Join(sep)
}

// SetArray value
func (f *StringArray) SetArray(arr []string) *StringArray {
	if arr == nil {
		arr = []string{}
	}
	*f = arr
	return f
}

// Value implements the driver.Valuer interface, []string field
func (f StringArray) Value() (driver.Value, error) {
	if f == nil {
		return "{}", ErrNullValueNotAllowed
	}
	return NullableStringArray(f).Value()
}

// Scan implements the driver.Valuer interface, []string field
func (f *StringArray) Scan(value interface{}) error {
	if value == nil {
		return ErrNullValueNotAllowed
	}
	return (*NullableStringArray)(f).Scan(value)
}

// UnmarshalJSON implements the json.Unmarshaller
func (f *StringArray) UnmarshalJSON(b []byte) error {
	if b == nil {
		return ErrNullValueNotAllowed
	}
	return (*NullableStringArray)(f).UnmarshalJSON(b)
}

// DecodeValue implements the gocast.Decoder
func (f *StringArray) DecodeValue(v interface{}) error {
	if v == nil {
		return ErrNullValueNotAllowed
	}
	return (*NullableStringArray)(f).DecodeValue(v)
}

// Len of array
func (f StringArray) Len() int {
	return len(f)
}

// IndexOf array value
func (f StringArray) IndexOf(v string) int {
	return (NullableStringArray)(f).IndexOf(v)
}

// OneOf value in array
func (f StringArray) OneOf(vals []string) bool {
	return (NullableStringArray)(f).OneOf(vals)
}

///////////////////////////////////////////////////////////////////////////////
/// Helpers
///////////////////////////////////////////////////////////////////////////////

func decodeNullableStringArray(arrSrc string, begin, end, border byte, escape string) []string {
	if strings.EqualFold(arrSrc, "null") {
		return nil
	}
	if arrSrc == string([]byte{begin, end}) {
		return []string{}
	}
	arr := strings.Split(strings.Trim(arrSrc, "{}[]"), ",")
	for i, val := range arr {
		val = strings.TrimSpace(val)
		if val == string([]byte{border, border}) {
			arr[i] = ""
		} else {
			if strings.HasPrefix(val, `"`) && strings.HasSuffix(val, `"`) {
				val = val[1 : len(val)-1]
			}
			arr[i] = strings.ReplaceAll(val, escape, string(border))
		}
	}
	return arr
}

func encodeNullableStringArray(begin, end, border byte, escape string, arr []string) *bytes.Buffer {
	var buff bytes.Buffer
	buff.WriteByte(begin)

	for i, v := range arr {
		if i > 0 {
			buff.WriteByte(',')
		}
		if byte(0) != border {
			buff.WriteByte(border)
			if escape != "" {
				v = strings.ReplaceAll(v, string(border), escape)
			}
		}
		buff.WriteString(v)
		if byte(0) != border {
			buff.WriteByte(border)
		}
	}

	buff.WriteByte(end)
	return &buff
}
