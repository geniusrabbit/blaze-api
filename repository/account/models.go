package account

import (
	"github.com/geniusrabbit/blaze-api/repository/account/models"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

// import (
// 	"time"

// 	"github.com/geniusrabbit/blaze-api/model"
// )

// type AccountBase interface {
// 	IsNil() bool
// 	GetID() uint64
// 	SetID(id uint64)

// 	ExtendAdminUsers(ids ...uint64)
// 	SetPermissions(perm model.PermissionChecker)

// 	GetApprove() model.ApproveStatus
// 	SetApprove(status model.ApproveStatus)

// 	SetCreatedAt(createdAt time.Time)
// }

// type Account[AccountT AccountBase] interface {
// 	AccountBase
// 	New() AccountT
// 	NewBasicAccount(
// 		id uint64,
// 		title string,
// 		approve model.ApproveStatus,
// 		perms model.PermissionChecker,
// 		admins []uint64,
// 	) AccountT
// }

type (
	Account              = models.Account
	AccountMember        = models.AccountMember
	M2MAccountMemberRole = models.M2MAccountMemberRole
	User                 = user.User
)
