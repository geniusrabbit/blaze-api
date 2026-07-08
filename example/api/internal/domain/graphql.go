package domain

import (
	"github.com/demdxx/gocast/v2"

	exmodels "github.com/geniusrabbit/blaze-api/example/api/internal/server/graphql/models"
	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/repository/user"
	basemodels "github.com/geniusrabbit/blaze-api/server/graphql/models"
)

// AccountToGraphQL maps example domain Account to extended GraphQL Account.
func AccountToGraphQL(acc *Account) *exmodels.Account {
	if acc == nil {
		return nil
	}
	return &exmodels.Account{
		ID:                acc.GetID(),
		Status:            basemodels.ApproveStatusFrom(acc.GetApprove()),
		Title:             acc.GetTitle(),
		Description:       acc.GetDescription(),
		LogoURI:           acc.GetLogoURI(),
		PolicyURI:         acc.GetPolicyURI(),
		TermsOfServiceURI: acc.GetTermsOfServiceURI(),
		ClientURI:         acc.GetClientURI(),
		Contacts:          acc.GetContacts(),
		CreatedAt:         acc.GetCreatedAt(),
		UpdatedAt:         acc.GetUpdatedAt(),
	}
}

func applyAccountProfile(
	dest *Account,
	title, description, logoURI, policyURI, termsOfServiceURI, clientURI *string,
	contacts []string,
) {
	if dest == nil {
		return
	}
	dest.ApplyProfile(
		gocast.PtrAsValue(title, ""),
		gocast.PtrAsValue(description, ""),
		gocast.PtrAsValue(logoURI, ""),
		gocast.PtrAsValue(policyURI, ""),
		gocast.PtrAsValue(termsOfServiceURI, ""),
		gocast.PtrAsValue(clientURI, ""),
		append([]string{}, contacts...),
	)
}

// FillAccountFromCreateInput copies account create input into domain Account.
func FillAccountFromCreateInput(dest *Account, input *exmodels.AccountCreateInput, appStatus ...pkgModels.ApproveStatus) *Account {
	if dest == nil || input == nil {
		return dest
	}

	status := pkgModels.UndefinedApproveStatus
	if len(appStatus) > 0 {
		status = appStatus[0]
	} else if input.Status != nil {
		status = input.Status.ModelStatus()
	}
	dest.SetApprove(status)
	applyAccountProfile(dest,
		input.Title,
		input.Description,
		input.LogoURI,
		input.PolicyURI,
		input.TermsOfServiceURI,
		input.ClientURI,
		input.Contacts,
	)
	return dest
}

// FillAccountFromUpdateInput applies account update input to a domain Account.
func FillAccountFromUpdateInput(dest *Account, input *exmodels.AccountUpdateInput, appStatus ...pkgModels.ApproveStatus) *Account {
	if dest == nil {
		return dest
	}
	if len(appStatus) > 0 {
		dest.SetApprove(appStatus[0])
	} else if input != nil && input.Status != nil {
		dest.SetApprove(input.Status.ModelStatus())
	}
	if input != nil {
		applyAccountProfile(dest,
			input.Title,
			input.Description,
			input.LogoURI,
			input.PolicyURI,
			input.TermsOfServiceURI,
			input.ClientURI,
			input.Contacts,
		)
	}
	return dest
}

// UserToGraphQL maps example domain User to base GraphQL User.
func UserToGraphQL(u *User) *exmodels.User {
	if u == nil {
		return nil
	}
	return &exmodels.User{
		ID:        u.GetID(),
		Email:     u.GetEmail(),
		Status:    basemodels.ApproveStatusFrom(u.GetApprove()),
		CreatedAt: u.GetCreatedAt(),
		UpdatedAt: u.GetUpdatedAt(),
	}
}

// UserToGraphQLPtr maps example domain User to base GraphQL User pointer (account/member payloads).
func UserToGraphQLPtr(u *User) *exmodels.User {
	if u == nil {
		return nil
	}
	return UserToGraphQL(u)
}

// UserFromCreateInput builds a new domain User from GraphQL create input.
func UserFromCreateInput(input *exmodels.UserCreateInput, appStatus ...pkgModels.ApproveStatus) *User {
	if input == nil {
		return nil
	}
	u := new(User)
	var status pkgModels.ApproveStatus
	if len(appStatus) > 0 {
		status = appStatus[0]
	}
	u.SetApprove(status)
	u.SetEmail(input.Email)
	return u
}

// FillUserFromInput merges GraphQL update input into an existing domain User.
func UserFromUpdateInput(input *exmodels.UserUpdateInput, dest *User, appStatus ...pkgModels.ApproveStatus) *User {
	if dest == nil {
		return nil
	}
	if len(appStatus) > 0 {
		dest.SetApprove(appStatus[0])
	} else if input != nil && input.Status != nil {
		dest.SetApprove(input.Status.ModelStatus())
	}
	if input != nil && input.Email != nil {
		dest.SetEmail(*input.Email)
	}
	return dest
}

// UserListFilterMapper converts extended user list filter to domain query option.
func UserListFilterMapper(fl *exmodels.UserListFilter) user.QOption {
	if fl == nil {
		return nil
	}
	return fl.Filter()
}

// UserListOrderMapper converts extended user list order to domain query option.
func UserListOrderMapper(ord *exmodels.UserListOrder) user.QOption {
	if ord == nil {
		return nil
	}
	return ord.Order()
}
