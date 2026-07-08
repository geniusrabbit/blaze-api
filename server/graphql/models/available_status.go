package models

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
)

// The list of statuses that shows is particular object is available
type AvailableStatus string

const (
	// All object by default have to be undefined
	AvailableStatusUndefined AvailableStatus = "UNDEFINED"
	// Status of the available object
	AvailableStatusAvailable AvailableStatus = "AVAILABLE"
	// Status of the unavailable object
	AvailableStatusUnavailable AvailableStatus = "UNAVAILABLE"
)

var AllAvailableStatus = []AvailableStatus{
	AvailableStatusUndefined,
	AvailableStatusAvailable,
	AvailableStatusUnavailable,
}

// AvailableStatusFrom model value
func AvailableStatusFrom(status pkgModels.AvailableStatus) AvailableStatus {
	switch status {
	case pkgModels.AvailableAvailableStatus:
		return AvailableStatusAvailable
	case pkgModels.UnavailableAvailableStatus:
		return AvailableStatusUnavailable
	}
	return AvailableStatusUndefined
}

func (e AvailableStatus) IsValid() bool {
	switch e {
	case AvailableStatusUndefined, AvailableStatusAvailable, AvailableStatusUnavailable:
		return true
	}
	return false
}

func (e AvailableStatus) String() string {
	return string(e)
}

func (e *AvailableStatus) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = AvailableStatus(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid AvailableStatus", str)
	}
	return nil
}

func (e AvailableStatus) MarshalGQL(w io.Writer) {
	_, _ = fmt.Fprint(w, strconv.Quote(e.String()))
}

func (e *AvailableStatus) UnmarshalJSON(b []byte) error {
	s, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	return e.UnmarshalGQL(s)
}

func (e AvailableStatus) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	e.MarshalGQL(&buf)
	return buf.Bytes(), nil
}

// ModelStatus returns status type from models
func (status *AvailableStatus) ModelStatus() pkgModels.AvailableStatus {
	if status == nil {
		return pkgModels.UndefinedAvailableStatus
	}
	switch *status {
	case AvailableStatusAvailable:
		return pkgModels.AvailableAvailableStatus
	case AvailableStatusUnavailable:
		return pkgModels.UnavailableAvailableStatus
	}
	return pkgModels.UndefinedAvailableStatus
}
