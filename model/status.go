package model

// AvailableStatus type
type AvailableStatus int

// AvailableStatus option constants...
const (
	UndefinedAvailableStatus   = 0
	AvailableAvailableStatus   = 1
	UnavailableAvailableStatus = 2
)

// ApproveStatus of the model
type ApproveStatus int

// ApproveStatus option constants...
const (
	UndefinedApproveStatus   = 0
	ApprovedApproveStatus    = 1
	DisapprovedApproveStatus = 2
	BannedApproveStatus      = 3
)

func (s ApproveStatus) IsApproved() bool {
	return s == ApprovedApproveStatus
}
