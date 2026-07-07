package appinit

import (
	"context"

	"go.uber.org/zap"

	"github.com/geniusrabbit/blaze-api/example/api/internal/domain"
	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	accountModels "github.com/geniusrabbit/blaze-api/repository/account/models"
	userModels "github.com/geniusrabbit/blaze-api/repository/user/models"
)

// EnsureSuperuser creates the superuser and system account if they do not already exist.
// Idempotent: returns nil immediately when a user with the given email is found.
func EnsureSuperuser(ctx context.Context, email, password string, deps *Deps) error {
	if email == "" || password == "" {
		return nil
	}

	// Bypass ACL checks — this runs at init time with no authenticated session.
	ctx = acl.WithNoPermCheck(ctx)

	lg := ctxlogger.Get(ctx)
	lg.Info("Ensuring superuser exists:", zap.String("email", email))

	// Check whether the user already exists.
	existing, _ := deps.UserModule.Repo.GetByEmail(ctx, email)
	if existing != nil && existing.GetID() != 0 {
		lg.Info("Superuser already exists, skipping creation", zap.String("email", email))
		return nil
	}

	// ===========================================================================
	// Create superuser and system account.
	//
	u := &domain.User{
		UserBase:  userModels.UserBase{Approve: pkgModels.ApprovedApproveStatus},
		UserEmail: userModels.UserEmail{Email: email},
	}
	userID, err := deps.UserModule.Repo.CreateWithPassword(ctx, u, password)
	if err != nil {
		lg.Error("Failed to create superuser", zap.String("email", email), zap.Error(err))
		return err
	}
	u.UserBase.ID = userID

	// ===========================================================================
	// Create system account and link superuser as admin member.
	//
	acc := &domain.Account{
		AccountBase: accountModels.AccountBase{Approve: pkgModels.ApprovedApproveStatus},
	}
	acc.ApplyProfile("system", "System account", "", "", "", "", nil)
	accID, err := deps.AccountRepo.Create(ctx, acc)
	if err != nil {
		lg.Error("Failed to create system account", zap.String("email", email), zap.Error(err))
		return err
	}
	acc.AccountBase.ID = accID

	// ===========================================================================
	// Link user as admin member of the system account.
	//
	err = deps.MemberRepo.LinkMember(ctx, acc, true, u)
	if err != nil {
		lg.Error("Failed to link user as admin member", zap.String("email", email), zap.Error(err))
		return err
	}

	// ===========================================================================
	// Assign system:admin role to the member.
	//
	if err = deps.MemberRepo.SetMemberRoles(ctx, acc, u, "system:admin"); err != nil {
		lg.Error("Failed to assign system:admin role", zap.String("email", email), zap.Error(err))
		return err
	}

	lg.Info("Superuser and system account created successfully",
		zap.String("email", email), zap.Uint64("user_id", userID), zap.Uint64("account_id", accID))

	return nil
}
