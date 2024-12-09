package acl

// RBACType represents RBAC type abstraction which can replace some non-standard ACL types & checks
// How to use:
// 1. Create a new type abstraction with the `RBACType` as an separate type
// 2. Define the `ResourceName` field with the name of the resource
// 3. While checking the permissions use WithAccountID, WithUserID, WithUserAccountID methods to set the account ID and user ID for the RBAC check
//
// Example:
// var MyRBACType = RBACType{
// 	ResourceName: `my_resource`,
// }
//
// if !acl.HaveAccessView(ctx, &MyRBACType) {
//   if acl.HaveAccessView(ctx, MyRBACType.WithAccountID(1)) {
//     filter.AccountID = 1
//   } else {
//     return ErrNoPermissions.WithMessage("no permissions to view the resource")
//   }
// }

type RBACType struct {
	AccountID    uint64
	UserID       uint64
	ResourceName string
}

// RBACResourceName returns the name of the resource for the RBAC
func (tp *RBACType) RBACResourceName() string {
	return tp.ResourceName
}

// RBACAccountID returns the account ID for the RBAC
func (tp *RBACType) WithAccountID(accountID uint64) *RBACType {
	nType := *tp
	nType.AccountID = accountID
	return &nType
}

// RBACUserID returns the user ID for the RBAC
func (tp *RBACType) WithUserID(userID uint64) *RBACType {
	nType := *tp
	nType.UserID = userID
	return &nType
}

// RBACWithUserAccountID returns the user ID and account ID for the RBAC
func (tp *RBACType) WithUserAccountID(userID, accountID uint64) *RBACType {
	nType := *tp
	nType.UserID = userID
	nType.AccountID = accountID
	return &nType
}
