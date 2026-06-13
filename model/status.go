package model

import "github.com/geniusrabbit/blaze-api/pkg/models"

// AvailableStatus type
type AvailableStatus = models.AvailableStatus

// AvailableStatus option constants...
const (
	UndefinedAvailableStatus   = models.UndefinedAvailableStatus
	AvailableAvailableStatus   = models.AvailableAvailableStatus
	UnavailableAvailableStatus = models.UnavailableAvailableStatus
)

// ApproveStatus of the model
type ApproveStatus = models.ApproveStatus

// ApproveStatus option constants...
const (
	UndefinedApproveStatus   = models.UndefinedApproveStatus
	PendingApproveStatus     = models.PendingApproveStatus
	ApprovedApproveStatus    = models.ApprovedApproveStatus
	DisapprovedApproveStatus = models.DisapprovedApproveStatus
	BannedApproveStatus      = models.BannedApproveStatus
)
