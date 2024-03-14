package models

import "github.com/geniusrabbit/blaze-api/model"

// // StatusFrom model value
// func StatusFrom(status model.ActiveStatus) ActiveStatus {
// 	switch status {
// 	case model.ActiveStatus:
// 		return ActiveStatusActive
// 	case model.PausedStatus:
// 		return ActiveStatusPaused
// 	}
// 	return ActiveStatusPaused
// }

// // Value model status
// func (status ActiveStatus) Value() model.Status {
// 	switch status {
// 	case ActiveStatusActive:
// 		return model.ActiveStatus
// 	case ActiveStatusPaused:
// 		return model.PausedStatus
// 	}
// 	return model.PausedStatus
// }

// ModelStatus returns status type from models
func (status *ApproveStatus) ModelStatus() model.ApproveStatus {
	if status == nil {
		return model.UndefinedApproveStatus
	}
	switch *status {
	case ApproveStatusApproved:
		return model.ApprovedApproveStatus
	case ApproveStatusRejected:
		return model.DisapprovedApproveStatus
	}
	return model.UndefinedApproveStatus
}

// ModelStatus returns status type from models
func (status *AvailableStatus) ModelStatus() model.AvailableStatus {
	if status == nil {
		return model.UndefinedAvailableStatus
	}
	switch *status {
	case AvailableStatusAvailable:
		return model.AvailableAvailableStatus
	case AvailableStatusUnavailable:
		return model.UnavailableAvailableStatus
	}
	return model.UndefinedAvailableStatus
}

// AvailableStatusFrom model value
func AvailableStatusFrom(status model.AvailableStatus) AvailableStatus {
	switch status {
	case model.AvailableAvailableStatus:
		return AvailableStatusAvailable
	case model.UnavailableAvailableStatus:
		return AvailableStatusUnavailable
	}
	return AvailableStatusUndefined
}

// ApproveStatusFrom model value
func ApproveStatusFrom(status model.ApproveStatus) ApproveStatus {
	switch status {
	case model.ApprovedApproveStatus:
		return ApproveStatusApproved
	case model.DisapprovedApproveStatus:
		return ApproveStatusRejected
	}
	return ApproveStatusPending
}
