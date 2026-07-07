package repository

import (
	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
)

type approveModel interface {
	SetApprove(pkgModels.ApproveStatus)
}

func setApproveOnModel(obj any, status pkgModels.ApproveStatus) {
	if v, ok := obj.(approveModel); ok {
		v.SetApprove(status)
	}
}
