package serverprovider

import (
	"database/sql/driver"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/ory/fosite"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/geniusrabbit/blaze-api/pkg/cache/dummy"
	"github.com/geniusrabbit/blaze-api/repository/testsuite"
)

type storageSuite struct {
	testsuite.DatabaseSuite

	storage *DatabaseStorage
}

func (s *storageSuite) SetupSuite() {
	s.DatabaseSuite.SetupSuite()
	s.storage = NewDatabaseStorage(s.DB, nil, dummy.New(), time.Hour)
}

func (s *storageSuite) TestGetClientHyphenatedID() {
	expiresAt := time.Now().Add(time.Hour)
	s.Mock.ExpectQuery(`SELECT .* FROM "auth_client" WHERE id = \$1`).
		WithArgs("mcp-client").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "account_id", "user_id", "title", "secret", "expires_at"}).
				AddRow("mcp-client", 1, 1, "mcp", "secret", expiresAt),
		)

	client, err := s.storage.GetClient(NewContext(s.Ctx), "mcp-client")
	s.NoError(err)
	s.Equal("mcp-client", client.GetID())
}

func (s *storageSuite) TestCreatePKCERequestSessionUpdatesExistingSession() {
	const (
		requestID = "req-1"
		codeSig   = "code-sig"
		rowID     = uint64(1)
	)
	req := testOAuthRequest(requestID)
	req.Form = map[string][]string{
		"code_challenge":        {"challenge123"},
		"code_challenge_method": {"S256"},
	}
	ctx := NewContext(s.Ctx)

	// First, authorize code session creates the row
	s.expectAuthSessionInsert(codeSig, requestID)
	s.NoError(s.storage.CreateAuthorizeCodeSession(ctx, codeSig, req))

	// PKCE session updates the same row (same code signature)
	s.Mock.ExpectQuery(`SELECT \* FROM "auth_session"`).
		WithArgs(codeSig).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "request_id", "access_token", "form", "client_id", "username", "subject", "active", "access_token_expires_at", "refresh_token_expires_at", "created_at"}).
				AddRow(rowID, requestID, codeSig, "existing=value", "mcp-client", "user", "subject", true, time.Now(), time.Now(), time.Now()),
		)
	s.Mock.ExpectExec(`UPDATE "auth_session" SET`).
		WithArgs(sqlmock.AnyArg(), rowID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	s.NoError(s.storage.CreatePKCERequestSession(ctx, codeSig, req))
}

func (s *storageSuite) TestCreateRefreshTokenSessionUpdatesByPrimaryKey() {
	const (
		requestID = "req-1"
		rowID     = uint64(2)
	)
	req := testOAuthRequest(requestID)
	req.GetSession().SetExpiresAt(fosite.RefreshToken, time.Now().Add(time.Hour))

	s.Mock.ExpectQuery(`SELECT \* FROM "auth_session" WHERE request_id=\$1`).
		WithArgs(requestID, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "request_id", "access_token"}).
				AddRow(rowID, requestID, "access-sig"),
		)
	s.Mock.ExpectExec(`UPDATE "auth_session" SET .*WHERE id=\$`).
		WithArgs(requestID, "access-sig", "refresh-sig", sqlmock.AnyArg(), rowID, rowID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := s.storage.CreateRefreshTokenSession(NewContext(s.Ctx), "refresh-sig", "access-sig", req)
	s.NoError(err)
}

func (s *storageSuite) expectAuthSessionInsert(accessToken, requestID string) {
	var got []driver.Value
	cap := captureArg{dst: &got}
	args := make([]driver.Value, 16)
	for i := range args {
		args[i] = cap
	}
	s.Mock.ExpectQuery(`INSERT INTO "auth_session"`).
		WithArgs(args...).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1)).
		WillReturnError(nil)
	s.T().Cleanup(func() {
		require.Contains(s.T(), stringsOf(got), accessToken)
		require.Contains(s.T(), stringsOf(got), requestID)
	})
}

func testOAuthRequest(id string) *fosite.Request {
	return &fosite.Request{
		ID:     id,
		Client: &fosite.DefaultClient{ID: "mcp-client"},
		Session: &fosite.DefaultSession{
			Username: "user",
			Subject:  "subject",
		},
	}
}

type captureArg struct {
	dst *[]driver.Value
}

func (c captureArg) Match(v driver.Value) bool {
	*c.dst = append(*c.dst, v)
	return true
}

func stringsOf(vals []driver.Value) []string {
	out := make([]string, 0, len(vals))
	for _, v := range vals {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func TestStorageSuite(t *testing.T) {
	suite.Run(t, &storageSuite{})
}
