package types

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
)

const dateTimeFormat = "2006-01-02T15:04:05.999999999Z"

// DateTime implements IO custom type of time with custom format
type DateTime time.Time

// DateTimeFromPtr returns DateTime object from pointer
func DateTimeFromPtr(tm *time.Time) DateTime {
	if tm == nil {
		return DateTime{}
	}
	return DateTime(*tm)
}

// GetTime object from DateTime
func (t DateTime) GetTime() time.Time {
	return (time.Time)(t)
}

// SetTime from time object
func (t *DateTime) SetTime(tm time.Time) {
	*t = (DateTime)(tm)
}

// MarshalGQL implements method of interface graphql.Marshaler
func (t DateTime) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(strconv.Quote(
		t.GetTime().Format(dateTimeFormat),
	)))
}

// UnmarshalGQL implements method of interface graphql.Unmarshaler
func (t *DateTime) UnmarshalGQL(v any) error {
	switch v := v.(type) {
	case string:
		tm, err := parseDate(v)
		if err != nil {
			return err
		}
		t.SetTime(tm)
		return nil
	default:
		return fmt.Errorf("%T is not a supported as a date", v)
	}
}

// MarshalTime redecalre the marshalel of standart scalar type Time
func MarshalTime(t time.Time) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.Quote(t.Format(dateTimeFormat)))
	})
}

// UnmarshalTime redecalre the unmarshalel of standart scalar type Time
func UnmarshalTime(v any) (time.Time, error) {
	switch v := v.(type) {
	case string:
		return parseDate(v)
	default:
		return time.Time{}, fmt.Errorf("%T is not a supported as a date", v)
	}
}

var timeFormats = []string{
	time.RFC1123Z,
	time.RFC3339Nano,
	time.UnixDate,
	time.RubyDate,
	time.RFC1123,
	time.RFC3339,
	time.RFC822,
	time.RFC850,
	time.RFC822Z,
	"2006-01-02",
	"2006-01-02 15:04:05",
	"2006/01/02",
	"2006/01/02 15:04:05",
}

func parseDate(tm string) (t time.Time, err error) {
	for _, f := range timeFormats {
		if t, err = time.Parse(f, tm); err == nil {
			break
		}
	}
	return t, err
}
