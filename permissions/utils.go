package permissions

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"

	"github.com/demdxx/rbac"
	"github.com/demdxx/xtypes"
)

func AllRolesAndPermissions(roles []rbac.Role, perms []rbac.Permission) ([]rbac.Role, []rbac.Permission) {
	if len(roles) == 0 && len(perms) == 0 {
		return nil, nil
	}
	var (
		mapRoles = make(map[string]rbac.Role, len(roles))
		mapPerms = make(map[string]rbac.Permission, len(perms))
	)
	for _, role := range roles {
		mapRoles[kName(role.Name(), role.Ext())] = role
		nextMpRoles, nextMpPerms := AllRolesAndPermissions(role.ChildRoles(), role.ChildPermissions())
		for _, nextRole := range nextMpRoles {
			mapRoles[kName(nextRole.Name(), nextRole.Ext())] = nextRole
		}
		for _, nextPerm := range nextMpPerms {
			mapPerms[kName(nextPerm.Name(), nextPerm.Ext())] = nextPerm
		}
	}
	for _, perm := range perms {
		mapPerms[kName(perm.Name(), perm.Ext())] = perm
		_, nextMpPerms := AllRolesAndPermissions(nil, perm.ChildPermissions())
		for _, nextPerm := range nextMpPerms {
			mapPerms[kName(nextPerm.Name(), nextPerm.Ext())] = nextPerm
		}
	}
	return xtypes.Map[string, rbac.Role](mapRoles).Values(), xtypes.Map[string, rbac.Permission](mapPerms).Values()
}

func AllRolesAndPermissionsIDs(roles []rbac.Role, perms []rbac.Permission) []uint64 {
	roles, perms = AllRolesAndPermissions(roles, perms)
	ids := make([]uint64, 0, len(roles)+len(perms))
	for _, role := range roles {
		switch ext := role.Ext().(type) {
		case *ExtData:
			ids = append(ids, ext.ID)
		case nil:
		}
	}
	for _, perm := range perms {
		switch ext := perm.Ext().(type) {
		case *ExtData:
			ids = append(ids, ext.ID)
		case nil:
		}
	}
	return ids
}

func kName(name string, ext any) string {
	if ext != nil {
		return name + ":" + md5str(marshalJSON(ext))
	}
	return name
}

func marshalJSON(v any) string {
	if v == nil {
		return ""
	}
	if b, err := json.Marshal(v); err == nil {
		return string(b)
	}
	return ""
}

func md5str(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}
