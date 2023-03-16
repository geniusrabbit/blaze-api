package types

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
)

// TimeDuration implements IO custom type of time with custom format
type TimeDuration time.Duration

// Duration object from TimeDuration
func (t TimeDuration) Duration() time.Duration {
	return (time.Duration)(t)
}

// SetDuration from time object
func (t *TimeDuration) SetDuration(d time.Duration) {
	*t = (TimeDuration)(d)
}

// MarshalGQL implements method of interface graphql.Marshaler
func (t TimeDuration) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(
		strconv.Quote(t.Duration().String()),
	))
}

// UnmarshalGQL implements method of interface graphql.Unmarshaler
func (t *TimeDuration) UnmarshalGQL(v any) error {
	switch v := v.(type) {
	case string:
		d, err := time.ParseDuration(v)
		if err != nil {
			return err
		}
		t.SetDuration(d)
		return nil
	default:
		return fmt.Errorf("%T is not a supported as a duration", v)
	}
}

// MarshalDuration redecalre the marshalel of standart scalar type Duration
func MarshalDuration(d time.Duration) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.Quote(d.String()))
	})
}

// UnmarshalDuration redecalre the unmarshalel of standart scalar type Duration
func UnmarshalDuration(v any) (time.Duration, error) {
	switch v := v.(type) {
	case string:
		return time.ParseDuration(v)
	default:
		return time.Duration(0), fmt.Errorf("%T is not a supported as a duration", v)
	}
}
