package gosql

import (
	"bytes"
	"strconv"
	"strings"

	"golang.org/x/exp/constraints"
)

// Number general type
type Number interface {
	constraints.Integer | constraints.Float
}

// ArrayNumberDecode decodes array of type int
func ArrayNumberDecode[T Number](data any, begin, end byte) (result []T, err error) {
	var arr string
	switch vdata := data.(type) {
	case []byte:
		arr = string(vdata)
	case string:
		arr = vdata
	case []T:
		return vdata, nil
	case nil:
		return nil, nil
	default:
		return nil, ErrInvalidScan
	}
	if strings.EqualFold(arr, "null") {
		return nil, nil
	}
	if arr == "{}" || arr == "[]" || len(arr) == 0 {
		return []T{}, err
	}
	if vals := strings.Split(strings.Trim(arr, "{}[]"), ","); vals != nil {
		result = make([]T, 0, len(vals))
		for _, cid := range vals {
			switch any(T(0)).(type) {
			case int, int8, int16, int32, int64:
				if v, err := strconv.ParseInt(strings.Trim(cid, "'\""), 10, 64); err == nil {
					result = append(result, T(v))
				} else {
					return nil, err
				}
			case uint, uint8, uint16, uint32, uint64:
				if v, err := strconv.ParseUint(strings.Trim(cid, "'\""), 10, 64); err == nil {
					result = append(result, T(v))
				} else {
					return nil, err
				}
			case float32, float64:
				if v, err := strconv.ParseFloat(strings.Trim(cid, "'\""), 64); err == nil {
					switch sv := any(v).(type) {
					case T: // NOTE: Temporary hack of type casting
						result = append(result, sv)
					}
				} else {
					return nil, err
				}
			}
		} // end for
	}
	return result, err
}

// ArrayEncode encodes array of type int
func ArrayNumberEncode[T Number](begin, end byte, arr []T) *bytes.Buffer {
	var buff bytes.Buffer
	buff.WriteByte(begin)
	for i, v := range arr {
		if i > 0 {
			buff.WriteByte(',')
		}
		switch any(v).(type) {
		case int, int8, int16, int32, int64:
			buff.WriteString(strconv.FormatInt(int64(v), 10))
		case uint, uint8, uint16, uint32, uint64:
			buff.WriteString(strconv.FormatUint(uint64(v), 10))
		case float32, float64:
			buff.WriteString(strconv.FormatFloat(float64(v), 'G', -1, 64))
		}
	}
	buff.WriteByte(end)
	return &buff
}
