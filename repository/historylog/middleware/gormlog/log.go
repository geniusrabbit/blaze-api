package gormlog

import (
	"reflect"
	"strings"
	"time"

	"github.com/demdxx/gocast/v2"
	"github.com/geniusrabbit/gosql/v2"
	"github.com/google/uuid"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/pkg/requestid"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
	historylogModels "github.com/geniusrabbit/blaze-api/repository/historylog/models"
)

// Register gorm callbacks for history log.
//
// History rows are written after the statement succeeded, through the same
// connection or transaction: a rolled-back change leaves no history, and a
// failed or no-op statement is not logged.
func Register(db *gorm.DB) (err error) {
	if cb := db.Callback(); cb != nil {
		err = multierr.Append(err, cb.Create().After("gorm:create").Register("historylog:create", Log(db, "create")))
		err = multierr.Append(err, cb.Update().After("gorm:update").Register("historylog:update", Log(db, "update")))
		err = multierr.Append(err, cb.Delete().After("gorm:delete").Register("historylog:delete", Log(db, "delete")))
	}
	return err
}

// sensitiveFieldParts mark columns that are never stored in the history.
// A field can also be excluded with the struct tag `historylog:"-"`.
var sensitiveFieldParts = []string{"password", "secret", "token", "salt", "hash", "api_key", "apikey", "private_key"}

// IsSensitiveField reports whether the column must not be stored in the history.
func IsSensitiveField(field *schema.Field) bool {
	if field.Tag.Get("historylog") == "-" {
		return true
	}
	name := strings.ToLower(field.DBName)
	for _, part := range sensitiveFieldParts {
		if strings.Contains(name, part) {
			return true
		}
	}
	return false
}

// Log action to history log.
// The db argument is kept for compatibility; the history row is written with
// the statement's own connection (and transaction, if any).
func Log(_ *gorm.DB, name string) func(*gorm.DB) {
	return func(cdb *gorm.DB) {
		if cdb.Error != nil || cdb.RowsAffected == 0 ||
			cdb.Statement == nil ||
			cdb.Statement.Schema == nil ||
			cdb.Statement.Schema.Name == "HistoryAction" ||
			cdb.Statement.Schema.PrioritizedPrimaryField == nil {
			return
		}

		var (
			pkVal any
			ctx   = cdb.Statement.Context
			field = cdb.Statement.Schema.PrioritizedPrimaryField
			rv    = cdb.Statement.ReflectValue
			data  = make(map[string]any, len(cdb.Statement.Schema.Fields))
		)
		if rv.Kind() == reflect.Pointer {
			rv = reflect.Indirect(rv)
		}
		if rv.Kind() == reflect.Struct {
			if v, isZero := field.ValueOf(ctx, rv); !isZero {
				pkVal = v
			}
			for _, field := range cdb.Statement.Schema.Fields {
				if field.DBName == "" || IsSensitiveField(field) {
					continue
				}
				data[field.DBName], _ = field.ValueOf(ctx, rv)
			}
		}
		// Statements by condition (e.g. Delete(&Model{}, id)) carry no PK in the value
		if pkVal == nil {
			pkVal = historylog.PKFromContext(ctx)
		}

		// Skip if primary key not found
		if pkVal == nil {
			ctxlogger.Get(ctx).Warn("history log: primary key not found",
				zap.String("action", name),
				zap.String("obj_type", cdb.Statement.Schema.Name),
				zap.String("message", historylog.MessageFromContext(ctx)),
			)
			return
		}

		user, acc := session.UserAccount(ctx)

		jdata, _ := gosql.NewNullableJSON[map[string]any](data)
		if jdata == nil {
			jdata = &gosql.NullableJSON[map[string]any]{}
		}

		// Same connection pool / transaction and context as the logged statement
		tx := cdb.Session(&gorm.Session{NewDB: true})

		// Inside a transaction a failed INSERT would abort the whole transaction
		// (Postgres), so the history write is isolated in a savepoint.
		_, inTx := cdb.Statement.ConnPool.(gorm.TxCommitter)
		if inTx {
			if err := tx.SavePoint(savepointName).Error; err != nil {
				ctxlogger.Get(ctx).Error("history log savepoint", zap.String("name", name), zap.Error(err))
				return
			}
		}
		err := tx.Create(&historylogModels.HistoryAction{
			ID:         uuid.New(),
			RequestID:  requestid.Get(ctx),
			Name:       gocast.Or(historylog.ActionFromContext(ctx), name),
			Message:    historylog.MessageFromContext(ctx),
			UserID:     idOf(user),
			AccountID:  idOf(acc),
			ObjectType: cdb.Statement.Schema.Name,
			ObjectID:   gocast.Uint64(pkVal),
			ObjectIDs:  gocast.Str(pkVal),
			Data:       *jdata,
			ActionAt:   time.Now(),
		}).Error

		if err != nil {
			ctxlogger.Get(ctx).
				Error("history log", zap.String("name", name), zap.Error(err))
			if inTx {
				_ = tx.RollbackTo(savepointName).Error
			}
		}
	}
}

const savepointName = "historylog"

func idOf(obj interface{ GetID() uint64 }) uint64 {
	if gocast.IsNil(obj) {
		return 0
	}
	return obj.GetID()
}
