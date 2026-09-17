package serverprovider

import (
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
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

func TestStorageSuite(t *testing.T) {
	suite.Run(t, &storageSuite{})
}
