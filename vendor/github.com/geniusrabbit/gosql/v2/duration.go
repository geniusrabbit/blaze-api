package gosql

import (
	"database/sql/driver"
	"strconv"
	"strings"
	"time"
)

// Duration constants
const (
	Nanosecond  Duration = Duration(time.Nanosecond)
	Microsecond          = Duration(time.Microsecond)
	Millisecond          = Duration(time.Millisecond)
	Second               = Duration(time.Second)
	Minute               = Duration(time.Minute)
	Hour                 = Duration(time.Hour)
)

// ParseDuration parses a duration string.
func ParseDuration(s string) (Duration, error) {
	var (
		additional time.Duration
		dur        time.Duration
		err        error
	)
	if strings.Contains(s, "w") {
		v := strings.Split(s, "w")
		i, err := strconv.Atoi(v[0])
		if err != nil {
			return 0, err
		}
		additional = time.Duration(i) * 7 * 24 * time.Hour
		s = v[1]
	}
	if strings.Contains(s, "d") {
		v := strings.Split(s, "d")
		i, err := strconv.Atoi(v[0])
		if err != nil {
			return 0, err
		}
		additional = time.Duration(i) * 24 * time.Hour
		s = v[1]
	}
	if s != "" {
		dur, err = time.ParseDuration(s)
	}
	return Duration(dur + additional), err
}

// Duration is a wrapper around time.Duration that allows us to
type Duration time.Duration

// String implements the Stringer interface.
func (d Duration) String() string {
	return time.Duration(d).String()
}

// Duration returns the time.Duration value.
func (d Duration) Duration() time.Duration { return time.Duration(d) }

// Scan implements the Scanner interface.
func (d *Duration) Scan(value any) error {
	switch v := value.(type) {
	case int64:
		*d = Duration(time.Duration(v))
	case []byte:
		dur, err := ParseDuration(string(v))
		if err != nil {
			return err
		}
		*d = dur
	case string:
		dur, err := ParseDuration(v)
		if err != nil {
			return err
		}
		*d = dur
	default:
		return ErrInvalidScanValue
	}
	return nil
}

// Value implements the driver Valuer interface.
func (d Duration) Value() (driver.Value, error) {
	return time.Duration(d).String(), nil
}

// Nanoseconds returns the duration as an integer nanosecond count.
func (d Duration) Nanoseconds() int64 { return int64(d) }

// Microseconds returns the duration as an integer microsecond count.
func (d Duration) Microseconds() int64 { return int64(d) / 1e3 }

// Milliseconds returns the duration as an integer millisecond count.
func (d Duration) Milliseconds() int64 { return int64(d) / 1e6 }

// Seconds returns the duration as a floating point number of seconds.
func (d Duration) Seconds() float64 { return time.Duration(d).Seconds() }

// Minutes returns the duration as a floating point number of minutes.
func (d Duration) Minutes() float64 { return time.Duration(d).Minutes() }

// Hours returns the duration as a floating point number of hours.
func (d Duration) Hours() float64 { return time.Duration(d).Hours() }

// Truncate returns the result of rounding d toward zero to a multiple of m.
// If m <= 0, Truncate returns d unchanged.
func (d Duration) Truncate(m Duration) Duration {
	return Duration(time.Duration(d).Truncate(m.Duration()))
}

// Round returns the result of rounding d to the nearest multiple of m.
// The rounding behavior for halfway values is to round away from zero.
// If the result exceeds the maximum (or minimum)
// value that can be stored in a Duration,
// Round returns the maximum (or minimum) duration.
// If m <= 0, Round returns d unchanged.
func (d Duration) Round(m Duration) Duration {
	return Duration(time.Duration(d).Round(m.Duration()))
}

// Abs returns the absolute value of d.
// As a special case, math.MinInt64 is converted to math.MaxInt64.
func (d Duration) Abs() Duration { return Duration(time.Duration(d).Abs()) }

// MarshalJSON implements the json.Marshaler interface.
func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *Duration) UnmarshalJSON(data []byte) error {
	if len(data) < 2 {
		return ErrInvalidDecodeValue
	}
	if data[0] != '"' || data[len(data)-1] != '"' {
		return ErrInvalidDecodeValue
	}
	return d.Scan(string(data[1 : len(data)-1]))
}
