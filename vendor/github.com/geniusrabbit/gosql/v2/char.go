//
// @project GeniusRabbit
// @author Dmitry Ponomarev <demdxx@gmail.com> 2016, 2020
//

package gosql

import "database/sql/driver"

// Char type of field
type Char rune

// Value implements the driver.Valuer interface, char field
func (f Char) Value() (driver.Value, error) {
	if f == 0 {
		return " ", nil
	}
	return string(f), nil
}

// Scan implements the sql.Scanner interface, char field
func (f *Char) Scan(value interface{}) (err error) {
	*f, err = decodeChar(value)
	return err
}

// MarshalJSON implements the json.Marshaler
func (f Char) MarshalJSON() ([]byte, error) {
	if f == 0 {
		return []byte("\" \""), nil
	}
	return []byte{'"', byte(f), '"'}, nil
}

// UnmarshalJSON implements the json.Unmarshaller
func (f *Char) UnmarshalJSON(b []byte) (err error) {
	*f, err = decodeChar(b)
	return err
}

///////////////////////////////////////////////////////////////////////////////
/// Helpers
///////////////////////////////////////////////////////////////////////////////

func decodeChar(value interface{}) (Char, error) {
	if value == nil {
		return Char(0), ErrNullValueNotAllowed
	}

	switch v := value.(type) {
	case []byte:
		if len(v) > 0 {
			return Char(v[0]), nil
		}
	case string:
		if len(v) > 0 {
			return Char(v[0]), nil
		}
	}
	return Char(0), ErrInvalidScan
}
