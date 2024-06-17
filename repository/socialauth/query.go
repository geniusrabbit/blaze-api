package socialauth

import "gorm.io/gorm"

type Filter struct {
	UserID          []uint64
	SocialID        []string
	Provider        []string
	Email           []string
	RetrieveDeleted bool
}

func (fl *Filter) PrepareQuery(query *gorm.DB) *gorm.DB {
	if fl == nil {
		return query
	}
	if len(fl.UserID) > 0 {
		query = query.Where(`user_id IN (?)`, fl.UserID)
	}
	if len(fl.SocialID) > 0 {
		query = query.Where(`social_id IN (?)`, fl.SocialID)
	}
	if len(fl.Provider) > 0 {
		query = query.Where(`provider IN (?)`, fl.Provider)
	}
	if len(fl.Email) > 0 {
		query = query.Where(`email IN (?)`, fl.Email)
	}
	if fl.RetrieveDeleted {
		query = query.Unscoped()
	}
	return query
}
