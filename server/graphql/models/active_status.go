package models

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
)

// The list of statuses that shows is particular object active or paused
type ActiveStatus string

const (
	// All object by default have to be paused
	ActiveStatusPaused ActiveStatus = "PAUSED"
	// Status of the active object
	ActiveStatusActive ActiveStatus = "ACTIVE"
)

var AllActiveStatus = []ActiveStatus{
	ActiveStatusPaused,
	ActiveStatusActive,
}

func (e ActiveStatus) IsValid() bool {
	switch e {
	case ActiveStatusPaused, ActiveStatusActive:
		return true
	}
	return false
}

func (e ActiveStatus) String() string {
	return string(e)
}

func (e *ActiveStatus) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ActiveStatus(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ActiveStatus", str)
	}
	return nil
}

func (e ActiveStatus) MarshalGQL(w io.Writer) {
	_, _ = fmt.Fprint(w, strconv.Quote(e.String()))
}

func (e *ActiveStatus) UnmarshalJSON(b []byte) error {
	s, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	return e.UnmarshalGQL(s)
}

func (e ActiveStatus) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	e.MarshalGQL(&buf)
	return buf.Bytes(), nil
}
