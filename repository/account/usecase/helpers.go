package usecase

import (
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

type accountIDSetter interface {
	SetID(uint64)
}

func setAccountID[T account.Model](obj T, id uint64) {
	if v, ok := any(obj).(accountIDSetter); ok {
		v.SetID(id)
	}
}

type userIDSetter interface {
	SetID(uint64)
}

func setUserID[T user.Model](obj T, id uint64) {
	if v, ok := any(obj).(userIDSetter); ok {
		v.SetID(id)
	}
}
