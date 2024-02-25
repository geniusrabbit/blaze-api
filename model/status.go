package model

// AvailableStatus type
type AvailableStatus int

// AvailableStatus option constants...
const (
	UndefinedAvailableStatus   AvailableStatus = 0
	AvailableAvailableStatus   AvailableStatus = 1
	UnavailableAvailableStatus AvailableStatus = 2
)

// ApproveStatus of the model
type ApproveStatus int

// ApproveStatus option constants...
const (
	UndefinedApproveStatus   ApproveStatus = 0
	PendingApproveStatus     ApproveStatus = 0
	ApprovedApproveStatus    ApproveStatus = 1
	DisapprovedApproveStatus ApproveStatus = 2
	BannedApproveStatus      ApproveStatus = 3
)

func (s ApproveStatus) String() string {
	switch s {
	case ApprovedApproveStatus:
		return "Approved"
	case DisapprovedApproveStatus:
		return "Disapproved"
	case BannedApproveStatus:
		return "Banned"
	default:
		return "Undefined"
	}
}

func (s ApproveStatus) IsApproved() bool {
	return s == ApprovedApproveStatus
}

func (s ApproveStatus) IsRejected() bool {
	return s == DisapprovedApproveStatus
}

func (s ApproveStatus) IsUndefined() bool {
	return s == UndefinedApproveStatus
}
