package repository_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/geniusrabbit/blaze-api/pkg/acl"
	"github.com/geniusrabbit/blaze-api/pkg/context/database"
	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/account/models"
	accountrepo "github.com/geniusrabbit/blaze-api/repository/account/repository"
	rbacModels "github.com/geniusrabbit/blaze-api/repository/rbac/models"
	"github.com/geniusrabbit/blaze-api/repository/user/testutil"
)

type memberEnv struct {
	db   *gorm.DB
	ctx  context.Context
	repo account.MemberRepository[*testutil.User, *testAccount]
	acc  *testAccount
}

func newMemberEnv(t *testing.T) *memberEnv {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())),
		&gorm.Config{Logger: logger.Discard, SkipDefaultTransaction: true})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&models.MemberBase{}, &rbacModels.Role{}, &models.M2MAccountMemberRole{}); err != nil {
		t.Fatal(err)
	}
	// Same constraint as the SQL migration
	if err := db.Exec(`CREATE UNIQUE INDEX uq_account_member ON account_member (account_id, user_id)`).Error; err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"admin", "viewer"} {
		if err := db.Create(&rbacModels.Role{Name: name, Title: name}).Error; err != nil {
			t.Fatal(err)
		}
	}
	acc := &testAccount{}
	acc.ID = 10
	return &memberEnv{
		db:  db,
		ctx: acl.WithNoPermCheck(database.WithDatabase(context.Background(), db, nil)),
		repo: accountrepo.NewMemberRepositoryFor(func() *account.Member[*testutil.User, *testAccount] {
			return new(account.Member[*testutil.User, *testAccount])
		}),
		acc: acc,
	}
}

func (env *memberEnv) member(t *testing.T, userID uint64) models.MemberBase {
	t.Helper()
	var m models.MemberBase
	if err := env.db.Unscoped().Where("account_id=? AND user_id=?", env.acc.ID, userID).Take(&m).Error; err != nil {
		t.Fatalf("member %d: %v", userID, err)
	}
	return m
}

func TestLinkMemberKeepsActiveAdmin(t *testing.T) {
	t.Parallel()

	env := newMemberEnv(t)
	if err := env.repo.LinkMember(env.ctx, env.acc, true, testutil.Stub(1)); err != nil {
		t.Fatal(err)
	}
	// Before the fix re-linking (e.g. an invite) reset is_admin to false
	if err := env.repo.LinkMember(env.ctx, env.acc, false, testutil.Stub(1)); err != nil {
		t.Fatal(err)
	}
	if m := env.member(t, 1); !m.IsAdmin {
		t.Fatal("re-linking demoted an admin")
	}
}

func TestLinkMemberRestoresRemovedMember(t *testing.T) {
	t.Parallel()

	env := newMemberEnv(t)
	for _, uid := range []uint64{1, 2} {
		if err := env.repo.LinkMember(env.ctx, env.acc, true, testutil.Stub(uid)); err != nil {
			t.Fatal(err)
		}
	}
	if err := env.repo.UnlinkMember(env.ctx, env.acc, testutil.Stub(2)); err != nil {
		t.Fatal(err)
	}
	if err := env.repo.LinkMember(env.ctx, env.acc, false, testutil.Stub(2)); err != nil {
		t.Fatal(err)
	}
	m := env.member(t, 2)
	if m.DeletedAt.Valid || m.IsAdmin || m.Approve != pkgModels.ApprovedApproveStatus {
		t.Fatalf("restored member = %+v, want active non-admin", m)
	}
}

func TestUnlinkLastAdminIsRefused(t *testing.T) {
	t.Parallel()

	env := newMemberEnv(t)
	_ = env.repo.LinkMember(env.ctx, env.acc, true, testutil.Stub(1))
	_ = env.repo.LinkMember(env.ctx, env.acc, false, testutil.Stub(2))

	// Before the fix UnlinkMember had no admin guard
	if err := env.repo.UnlinkMember(env.ctx, env.acc, testutil.Stub(1)); !errors.Is(err, accountrepo.ErrAccountHaveToHaveAdmin) {
		t.Fatalf("unlink last admin: err = %v", err)
	}
	if err := env.repo.UnlinkMember(env.ctx, env.acc, testutil.Stub(2)); err != nil {
		t.Fatalf("unlink a regular member: %v", err)
	}
}

func TestConcurrentDemotionsKeepAnAdmin(t *testing.T) {
	t.Parallel()

	env := newMemberEnv(t)
	_ = env.repo.LinkMember(env.ctx, env.acc, true, testutil.Stub(1), testutil.Stub(2))

	// Each demotion alone is allowed; together they would leave no admin.
	// The check now runs inside the transaction with the admin rows locked.
	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i, uid := range []uint64{1, 2} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = env.repo.SetMemberRoles(env.ctx, env.acc, testutil.Stub(uid), "viewer")
		}()
	}
	wg.Wait()

	admins := 0
	for _, uid := range []uint64{1, 2} {
		if env.member(t, uid).IsAdmin {
			admins++
		}
	}
	if admins == 0 {
		t.Fatalf("both admins were demoted: errs = %v", errs)
	}
	if !errors.Is(errs[0], accountrepo.ErrAccountHaveToHaveAdmin) && !errors.Is(errs[1], accountrepo.ErrAccountHaveToHaveAdmin) {
		t.Fatalf("errs = %v, want one ErrAccountHaveToHaveAdmin", errs)
	}
}
