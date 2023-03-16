package models

import "github.com/geniusrabbit/api-template-base/model"

// FromUserModel to local graphql model
func FromUserModel(u *model.User) *User {
	return &User{
		ID:        int(u.ID),
		Username:  u.Email,
		Status:    ApproveStatusFrom(u.Approve),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// FromUserModelList converts model list to local model list
func FromUserModelList(list []*model.User) []*User {
	users := make([]*User, 0, len(list))
	for _, u := range list {
		users = append(users, FromUserModel(u))
	}
	return users
}
