package generated

import (
	"time"

	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
)

// Model is the compile-time constraint for types used with Repository[T, TID] and Usecase[T, TID].
// T must expose its primary key via a value receiver so that T itself (not *T) satisfies the constraint.
type Model[TID comparable] interface {
	GetID() TID
}

// getModelID extracts the ID from obj by asserting it to Model[TID].
// Returns zero value of TID if the assertion fails.
func getModelID[TID comparable](obj any) TID {
	if v, ok := obj.(Model[TID]); ok {
		return v.GetID()
	}
	return *new(TID)
}

// ModelIDFieldGetter defines an interface for models that can override the primary key column name.
type ModelIDFieldGetter interface {
	GetIDField() string
}

// getModelIDField returns the primary key column name for obj.
// Defaults to "id" if the model does not implement ModelIDFieldGetter.
func getModelIDField(obj any) string {
	if v, ok := obj.(ModelIDFieldGetter); ok {
		return v.GetIDField()
	}
	return "id"
}

// ModelIDSetter defines an interface for models that can set their primary key.
type ModelIDSetter[TID any] interface {
	SetID(id TID)
}

// setModelID sets the ID on obj by asserting it to ModelIDSetter[TID].
// No-op if obj does not implement ModelIDSetter[TID].
func setModelID[TID any](obj any, id TID) {
	if v, ok := obj.(ModelIDSetter[TID]); ok {
		v.SetID(id)
	}
}

// ModelCreateTimeSetter defines an interface for models that can set their creation time.
type ModelCreateTimeSetter interface {
	SetCreatedAt(time.Time)
}

// setModelCreatedAt sets the creation time on obj.
// No-op if obj does not implement ModelCreateTimeSetter.
func setModelCreatedAt(obj any, t time.Time) {
	if v, ok := obj.(ModelCreateTimeSetter); ok {
		v.SetCreatedAt(t)
	}
}

// ModelUpdateTimeSetter defines an interface for models that can set their update time.
type ModelUpdateTimeSetter interface {
	SetUpdatedAt(time.Time)
}

// setModelUpdatedAt sets the update time on obj.
// No-op if obj does not implement ModelUpdateTimeSetter.
func setModelUpdatedAt(obj any, t time.Time) {
	if v, ok := obj.(ModelUpdateTimeSetter); ok {
		v.SetUpdatedAt(t)
	}
}

// ModelApproveStatusSetter defines an interface for models that support an approval workflow.
type ModelApproveStatusSetter interface {
	SetApproveStatus(status pkgModels.ApproveStatus)
}

// setModelApproveStatus sets the approval status on obj.
// No-op if obj does not implement ModelApproveStatusSetter.
func setModelApproveStatus(obj any, status pkgModels.ApproveStatus) {
	if v, ok := obj.(ModelApproveStatusSetter); ok {
		v.SetApproveStatus(status)
	}
}

// BaseModel is a convenience embed for domain models used with Repository[T, TID] and Usecase[T, TID].
// It provides GetID (value receiver, satisfies Model[TID] constraint) and SetID implementations.
//
// Usage:
//
//	type MyModel struct {
//	    generated.BaseModel[uint64]
//	    Name string
//	}
type BaseModel[TID comparable] struct {
	ID TID `gorm:"primaryKey" db:"id"`
}

// GetID returns the model's primary key.
// Value receiver ensures MyModel (not *MyModel) satisfies the Model[TID] constraint.
func (m BaseModel[TID]) GetID() TID { return m.ID }

// SetID sets the model's primary key.
func (m *BaseModel[TID]) SetID(id TID) { m.ID = id }

// BaseTimestamps is a convenience embed providing SetCreatedAt and SetUpdatedAt.
//
// Usage:
//
//	type MyModel struct {
//	    generated.BaseModel[uint64]
//	    generated.BaseTimestamps
//	    gorm.DeletedAt
//	}
type BaseTimestamps struct {
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// SetCreatedAt sets the creation timestamp.
func (m *BaseTimestamps) SetCreatedAt(t time.Time) { m.CreatedAt = t }

// SetUpdatedAt sets the update timestamp.
func (m *BaseTimestamps) SetUpdatedAt(t time.Time) { m.UpdatedAt = t }
