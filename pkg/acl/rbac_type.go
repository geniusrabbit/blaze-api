package acl

// RBACType represents RBAC type abstraction which can replace some non-standard ACL types & checks
type RBACType struct {
	AccountID    uint64
	UserID       uint64
	ResourceName string
}

// RBACResourceName returns the name of the resource for the RBAC
func (tp *RBACType) RBACResourceName() string {
	return tp.ResourceName
}
