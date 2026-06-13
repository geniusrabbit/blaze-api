package model

import rbacModels "github.com/geniusrabbit/blaze-api/repository/rbac/models"

// Role base model
type Role = rbacModels.Role

// M2MRole link parent and child role
type M2MRole = rbacModels.M2MRole

// Access level constants
const (
	AccessLevelBasic       = rbacModels.AccessLevelBasic
	AccessLevelNoAnonymous = rbacModels.AccessLevelNoAnonymous
	AccessLevelAccount     = rbacModels.AccessLevelAccount
	AccessLevelSystem      = rbacModels.AccessLevelSystem
)

