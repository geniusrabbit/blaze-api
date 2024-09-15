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

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/context/ctxlogger"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/pkg/requestid"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
)

// Register gorm callbacks for history log
func Register(db *gorm.DB) (err error) {
	if cb := db.Callback(); cb != nil {
		err = multierr.Append(err, cb.Create().After("gorm:create").Register("historylog:create", Log(db, "create")))
		err = multierr.Append(err, cb.Update().Before("gorm:update").Register("historylog:update", Log(db, "update")))
		err = multierr.Append(err, cb.Delete().Before("gorm:delete").Register("historylog:delete", Log(db, "delete")))
	}
	return err
}

// Log action to history log
func Log(db *gorm.DB, name string) func(*gorm.DB) {
	return func(cdb *gorm.DB) {
		if false ||
			cdb.Statement == nil ||
			cdb.Statement.Schema == nil ||
			cdb.Statement.Schema.Name == "HistoryAction" ||
			!cdb.Statement.ReflectValue.IsValid() {
			return
		}

		switch cdb.Statement.ReflectValue.Kind() {
		case reflect.Slice, reflect.Array, reflect.Invalid:
			return
		}
		field := cdb.Statement.Schema.PrioritizedPrimaryField
		if field == nil {
			return
		}
		ctx := cdb.Statement.Context
		pkVal, _ := field.ValueOf(ctx, cdb.Statement.ReflectValue)
		user, acc := session.UserAccount(ctx)

		data := make(map[string]any, len(cdb.Statement.Schema.Fields))
		for _, field := range cdb.Statement.Schema.Fields {
			fLowName := strings.ToLower(field.DBName)
			// NOTE: Skip password and secret fields from history log as security reason
			if !strings.Contains(fLowName, "password") && !strings.Contains(fLowName, "secret") {
				data[field.DBName], _ = field.ValueOf(ctx, cdb.Statement.ReflectValue)
			}
		}

		jdata, _ := gosql.NewNullableJSON[map[string]any](data)
		if jdata == nil {
			jdata = &gosql.NullableJSON[map[string]any]{}
		}
		err := db.Create(&model.HistoryAction{
			ID:         uuid.New(),
			RequestID:  requestid.Get(ctx),
			Name:       gocast.Or(historylog.ActionFromContext(ctx), name),
			Message:    historylog.MessageFromContext(ctx),
			UserID:     user.ID,
			AccountID:  acc.ID,
			ObjectType: cdb.Statement.Schema.Name,
			ObjectID:   gocast.Number[uint64](pkVal),
			ObjectIDs:  gocast.Str(pkVal),
			Data:       *jdata,
			ActionAt:   time.Now(),
		}).Error
		if err != nil {
			ctxlogger.Get(ctx).
				Error("history log", zap.String("name", name), zap.Error(err))
		}
	}
}
