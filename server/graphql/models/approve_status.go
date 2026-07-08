package models

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
)

// The list of statuses that shows is object approved or not
type ApproveStatus string

const (
	// Pending status of the just inited objects
	ApproveStatusPending ApproveStatus = "PENDING"
	// Approved status of object could be obtained from the some authorized user who have permissions
	ApproveStatusApproved ApproveStatus = "APPROVED"
	// Rejected status of object could be obtained from the some authorized user who have permissions
	ApproveStatusRejected ApproveStatus = "REJECTED"
)

var AllApproveStatus = []ApproveStatus{
	ApproveStatusPending,
	ApproveStatusApproved,
	ApproveStatusRejected,
}

// ApproveStatusFrom model value
func ApproveStatusFrom(status pkgModels.ApproveStatus) ApproveStatus {
	switch status {
	case pkgModels.ApprovedApproveStatus:
		return ApproveStatusApproved
	case pkgModels.DisapprovedApproveStatus:
		return ApproveStatusRejected
	}
	return ApproveStatusPending
}

func (e ApproveStatus) IsValid() bool {
	switch e {
	case ApproveStatusPending, ApproveStatusApproved, ApproveStatusRejected:
		return true
	}
	return false
}

func (e ApproveStatus) String() string {
	return string(e)
}

func (e *ApproveStatus) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ApproveStatus(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ApproveStatus", str)
	}
	return nil
}

func (e ApproveStatus) MarshalGQL(w io.Writer) {
	_, _ = fmt.Fprint(w, strconv.Quote(e.String()))
}

func (e *ApproveStatus) UnmarshalJSON(b []byte) error {
	s, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	return e.UnmarshalGQL(s)
}

func (e ApproveStatus) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	e.MarshalGQL(&buf)
	return buf.Bytes(), nil
}

// ModelStatus returns status type from models
func (status *ApproveStatus) ModelStatus() pkgModels.ApproveStatus {
	if status == nil {
		return pkgModels.UndefinedApproveStatus
	}
	switch *status {
	case ApproveStatusApproved:
		return pkgModels.ApprovedApproveStatus
	case ApproveStatusRejected:
		return pkgModels.DisapprovedApproveStatus
	}
	return pkgModels.UndefinedApproveStatus
}
