package models

import pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"

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
