package query

import (
	"context"
	"strings"

	"github.com/demdxx/xtypes"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/account/models"
)

// UserExtFilter of the objects list with extended fields
type UserExtFilter struct {
	AccountID []uint64
	UserID    []uint64
	Emails    []string
	Roles     []uint64
}

// AdjustPermissions adjusts the filter based on the current user's permissions.
func (fl *UserExtFilter) AdjustPermissions(ctx context.Context) error {
	fl.AccountID = []uint64{session.Account(ctx).ID}
	return nil
}

// PrepareQuery returns the query with applied filters
func (fl *UserExtFilter) PrepareQuery(q *gorm.DB) *gorm.DB {
	if fl == nil {
		return q
	}
	if len(fl.Roles) > 0 {
		qstr := `SELECT member_id FROM ` +
			(*models.M2MAccountMemberRole)(nil).TableName() + ` WHERE role_id IN (?)`
		if len(fl.AccountID) > 0 {
			q = q.Where(`id IN (SELECT user_id FROM `+(*models.AccountMember)(nil).TableName()+
				` WHERE account_id IN (?) OR id IN (`+qstr+`))`, fl.AccountID, fl.Roles)
		} else {
			q = q.Where(`id IN (SELECT user_id FROM `+(*models.AccountMember)(nil).TableName()+
				` WHERE id IN (`+qstr+`))`, fl.Roles)
		}
	} else if len(fl.AccountID) > 0 {
		q = q.Where(`id IN (SELECT user_id FROM `+
			(*models.AccountMember)(nil).TableName()+` WHERE account_id IN (?))`, fl.AccountID)
	}
	if len(fl.UserID) > 0 {
		q = q.Where(`id IN (?)`, fl.UserID)
	}
	if len(fl.Emails) > 0 {
		q = q.Where(`lower(email) IN (?)`, xtypes.SliceApply(fl.Emails, func(v string) string {
			return strings.ToLower(v)
		}))
	}
	return q
}
