package account

// AdminRoleNames are the role names that make a member an account admin.
var AdminRoleNames = []string{"admin", "account:admin", "system:admin"}

// IsAdminRole reports whether the role name grants account admin rights.
func IsAdminRole(name string) bool {
	for _, n := range AdminRoleNames {
		if n == name {
			return true
		}
	}
	return false
}

// HasAdminRole reports whether any of the role names is an admin role.
func HasAdminRole(names ...string) bool {
	for _, name := range names {
		if IsAdminRole(name) {
			return true
		}
	}
	return false
}
