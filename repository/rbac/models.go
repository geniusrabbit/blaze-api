package rbac

import "github.com/geniusrabbit/blaze-api/repository/rbac/models"

type (
	Role    = models.Role
	M2MRole = models.M2MRole
)

const (
	AccessLevelBasic       = models.AccessLevelBasic
	AccessLevelNoAnonymous = models.AccessLevelNoAnonymous
	AccessLevelAccount     = models.AccessLevelAccount
	AccessLevelSystem      = models.AccessLevelSystem
)
