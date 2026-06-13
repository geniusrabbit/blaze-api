package model

import accountModels "github.com/geniusrabbit/blaze-api/repository/account/models"

// AccountMember contains reference from user to account as member
type AccountMember = accountModels.AccountMember

// M2MAccountMemberRole m2m link between members and roles|permissions
type M2MAccountMemberRole = accountModels.M2MAccountMemberRole
