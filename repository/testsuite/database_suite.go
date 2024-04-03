package testsuite

import (
	"context"
	"database/sql"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/pkg/database"
	"github.com/geniusrabbit/blaze-api/pkg/permissions"
)

// DatabaseSuite implementation with database
type DatabaseSuite struct {
	suite.Suite

	Ctx    context.Context
	Mock   sqlmock.Sqlmock
	BaseDB *sql.DB
	DB     *gorm.DB
	Logger *zap.Logger
}

// SetupSuite inits context
func (s *DatabaseSuite) SetupSuite() {
	db, mock, err := sqlmock.New()
	if err != nil {
		s.T().Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	s.BaseDB = db
	s.Logger, err = zap.NewDevelopment()
	if err != nil {
		s.T().Fatalf("an error '%s' was not expected when create new logger object", err)
	}
	s.DB, err = gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		s.T().Fatalf("an error '%s' was not expected when create new gorm.database object", err)
	}
	s.Mock = mock
	s.Ctx = context.TODO()
	s.Ctx = permissions.WithManager(s.Ctx, permissions.NewTestManager(s.Ctx))
	s.Ctx = ctxlogger.WithLogger(s.Ctx, s.Logger)
	s.Ctx = database.WithDatabase(s.Ctx, s.DB, s.DB)
}

// TearDownSuite after finish
func (s *DatabaseSuite) TearDownSuite() {
	// s.Mock.ExpectClose()
	// assert.NoError(s.T(), s.BaseDB.Close())
}
