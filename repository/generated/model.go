package generated

import (
	"time"

	"github.com/geniusrabbit/blaze-api/model"
)

// ModelIDGetter defines an interface for models that can return their ID
type ModelIDGetter[TID any] interface {
	GetID() TID
}

// getModelID extracts the ID from a model that implements ModelIDGetter
// Returns zero value if the model doesn't implement the interface
//
//go:inline
func getModelID[TID any](obj any) TID {
	switch v := obj.(type) {
	case ModelIDGetter[TID]:
		return v.GetID()
	}
	return *new(TID)
}

// ModelIDFieldGetter defines an interface for models that can return their ID field name
type ModelIDFieldGetter interface {
	GetIDField() string
}

// getModelIDField extracts the ID field name from a model that implements ModelIDFieldGetter
// Returns empty string if the model doesn't implement the interface
//
//go:inline
func getModelIDField(obj any) string {
	switch v := obj.(type) {
	case ModelIDFieldGetter:
		return v.GetIDField()
	}
	return "id"
}

// ModelIDSetter defines an interface for models that can set their ID
type ModelIDSetter[TID any] interface {
	SetID(id TID)
}

// setModelID sets the ID on a model that implements ModelIDSetter
//
//go:inline
func setModelID[TID any](obj any, id TID) {
	switch v := obj.(type) {
	case ModelIDSetter[TID]:
		v.SetID(id)
	}
}

// ModelCreateTimeSetter defines an interface for models that can set their creation time
type ModelCreateTimeSetter interface {
	SetCreatedAt(time.Time)
}

// setModelCreatedAt sets the creation time on a model that implements ModelCreateTimeSetter
//
//go:inline
func setModelCreatedAt(obj any, t time.Time) {
	switch v := obj.(type) {
	case ModelCreateTimeSetter:
		v.SetCreatedAt(t)
	}
}

// ModelUpdateTimeSetter defines an interface for models that can set their update time
type ModelUpdateTimeSetter interface {
	SetUpdatedAt(time.Time)
}

// setModelUpdatedAt sets the update time on a model that implements ModelUpdateTimeSetter
//
//go:inline
func setModelUpdatedAt(obj any, t time.Time) {
	switch v := obj.(type) {
	case ModelUpdateTimeSetter:
		v.SetUpdatedAt(t)
	}
}

// ModelApproveStatusSetter defines an interface for models that can set their approval status
type ModelApproveStatusSetter interface {
	SetApproveStatus(status model.ApproveStatus)
}

// setModelApproveStatus sets the approval status on a model that implements ModelApproveStatusSetter
//
//go:inline
func setModelApproveStatus(obj any, status model.ApproveStatus) {
	switch v := obj.(type) {
	case ModelApproveStatusSetter:
		v.SetApproveStatus(status)
	}
}
