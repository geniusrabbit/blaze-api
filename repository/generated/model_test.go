package generated

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
)

// ---------------------------------------------------------------------------
// Helpers — minimal in-test models
// ---------------------------------------------------------------------------

// idModel implements Model[uint64] (value receiver) + ModelIDSetter[uint64].
type idModel struct {
	ID uint64
}

func (m idModel) GetID() uint64   { return m.ID }
func (m *idModel) SetID(v uint64) { m.ID = v }

// stringIDModel uses string as TID.
type stringIDModel struct {
	ID string
}

func (m stringIDModel) GetID() string   { return m.ID }
func (m *stringIDModel) SetID(v string) { m.ID = v }

// customFieldModel overrides the primary key column name.
type customFieldModel struct{}

func (m customFieldModel) GetID() uint64      { return 0 }
func (m customFieldModel) GetIDField() string { return "custom_id" }

// timestampModel implements both time setters.
type timestampModel struct {
	Created time.Time
	Updated time.Time
}

func (m *timestampModel) SetCreatedAt(t time.Time) { m.Created = t }
func (m *timestampModel) SetUpdatedAt(t time.Time) { m.Updated = t }

// approveModel implements ModelApproveStatusSetter.
type approveModel struct {
	Status pkgModels.ApproveStatus
}

func (m *approveModel) SetApproveStatus(s pkgModels.ApproveStatus) { m.Status = s }

// plainModel implements none of the optional interfaces.
type plainModel struct{ Val int }

// ---------------------------------------------------------------------------
// BaseModel
// ---------------------------------------------------------------------------

func TestBaseModel_GetID(t *testing.T) {
	m := BaseModel[uint64]{ID: 42}
	assert.Equal(t, uint64(42), m.GetID())
}

func TestBaseModel_SetID(t *testing.T) {
	m := BaseModel[uint64]{}
	m.SetID(99)
	assert.Equal(t, uint64(99), m.ID)
}

func TestBaseModel_StringTID(t *testing.T) {
	m := BaseModel[string]{ID: "abc"}
	assert.Equal(t, "abc", m.GetID())
	m.SetID("xyz")
	assert.Equal(t, "xyz", m.ID)
}

// Value receiver satisfies Model[TID] — T (not *T) can be used as constraint.
func TestBaseModel_ValueReceiverSatisfiesConstraint(t *testing.T) {
	var _ Model[uint64] = BaseModel[uint64]{}
	var _ Model[string] = BaseModel[string]{}
}

// ---------------------------------------------------------------------------
// BaseTimestamps
// ---------------------------------------------------------------------------

func TestBaseTimestamps_SetCreatedAt(t *testing.T) {
	m := &BaseTimestamps{}
	ts := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	m.SetCreatedAt(ts)
	assert.Equal(t, ts, m.CreatedAt)
}

func TestBaseTimestamps_SetUpdatedAt(t *testing.T) {
	m := &BaseTimestamps{}
	ts := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	m.SetUpdatedAt(ts)
	assert.Equal(t, ts, m.UpdatedAt)
}

// ---------------------------------------------------------------------------
// getModelID
// ---------------------------------------------------------------------------

func TestGetModelID_WithInterface(t *testing.T) {
	m := &idModel{ID: 7}
	assert.Equal(t, uint64(7), getModelID[uint64](m))
}

func TestGetModelID_ValueReceiver(t *testing.T) {
	// value receiver: both T and *T satisfy Model[TID]
	m := idModel{ID: 3}
	assert.Equal(t, uint64(3), getModelID[uint64](m))
}

func TestGetModelID_StringID(t *testing.T) {
	m := &stringIDModel{ID: "hello"}
	assert.Equal(t, "hello", getModelID[string](m))
}

func TestGetModelID_NoInterface_ReturnsZero(t *testing.T) {
	m := &plainModel{Val: 5}
	assert.Equal(t, uint64(0), getModelID[uint64](m))
}

// ---------------------------------------------------------------------------
// setModelID
// ---------------------------------------------------------------------------

func TestSetModelID_WithInterface(t *testing.T) {
	m := &idModel{}
	setModelID(m, uint64(55))
	assert.Equal(t, uint64(55), m.ID)
}

func TestSetModelID_NoInterface_NoOp(t *testing.T) {
	m := &plainModel{Val: 1}
	// Must not panic; Val must remain unchanged.
	setModelID(m, uint64(99))
	assert.Equal(t, 1, m.Val)
}

// ---------------------------------------------------------------------------
// getModelIDField
// ---------------------------------------------------------------------------

func TestGetModelIDField_Default(t *testing.T) {
	assert.Equal(t, "id", getModelIDField(&plainModel{}))
}

func TestGetModelIDField_Custom(t *testing.T) {
	assert.Equal(t, "custom_id", getModelIDField(&customFieldModel{}))
}

// ---------------------------------------------------------------------------
// setModelCreatedAt
// ---------------------------------------------------------------------------

func TestSetModelCreatedAt_WithInterface(t *testing.T) {
	m := &timestampModel{}
	ts := time.Now()
	setModelCreatedAt(m, ts)
	assert.Equal(t, ts, m.Created)
}

func TestSetModelCreatedAt_NoInterface_NoOp(t *testing.T) {
	m := &plainModel{}
	assert.NotPanics(t, func() {
		setModelCreatedAt(m, time.Now())
	})
}

// ---------------------------------------------------------------------------
// setModelUpdatedAt
// ---------------------------------------------------------------------------

func TestSetModelUpdatedAt_WithInterface(t *testing.T) {
	m := &timestampModel{}
	ts := time.Now()
	setModelUpdatedAt(m, ts)
	assert.Equal(t, ts, m.Updated)
}

func TestSetModelUpdatedAt_NoInterface_NoOp(t *testing.T) {
	m := &plainModel{}
	assert.NotPanics(t, func() {
		setModelUpdatedAt(m, time.Now())
	})
}

// ---------------------------------------------------------------------------
// setModelApproveStatus
// ---------------------------------------------------------------------------

func TestSetModelApproveStatus_WithInterface(t *testing.T) {
	m := &approveModel{}
	setModelApproveStatus(m, pkgModels.ApprovedApproveStatus)
	assert.Equal(t, pkgModels.ApprovedApproveStatus, m.Status)
}

func TestSetModelApproveStatus_NoInterface_NoOp(t *testing.T) {
	m := &plainModel{Val: 2}
	assert.NotPanics(t, func() {
		setModelApproveStatus(m, pkgModels.ApprovedApproveStatus)
	})
	assert.Equal(t, 2, m.Val)
}
