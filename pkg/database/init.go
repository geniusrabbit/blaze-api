package database

import (
	"context"
	"fmt"
	"net/url"

	"github.com/demdxx/gocast/v2"
	"gorm.io/gorm"
)

type dialectorExt interface {
	IsDebug(ctx context.Context, dns string) bool
	PrepareDB(ctx context.Context, db *gorm.DB) (*gorm.DB, error)
	Dialector(ctx context.Context, dns string, config *gorm.Config) (gorm.Dialector, error)
}

var dialectors = map[string]dialectorExt{}

//lint:ignore U1000 ignore unused for build tags
func registerDialector(open dialectorExt, names ...string) {
	for _, name := range names {
		dialectors[name] = open
	}
}

//lint:ignore U1000 ignore unused for build tags
type defaultDialector struct{}

func (d *defaultDialector) IsDebug(ctx context.Context, dsn string) bool {
	connURL, err := url.Parse(dsn)
	if err != nil {
		return false
	}
	query := connURL.Query()
	debug := gocast.Bool(query.Get("debug"))
	return debug
}

func (d *defaultDialector) PrepareDB(ctx context.Context, db *gorm.DB) (*gorm.DB, error) {
	return db, nil
}

func (d *defaultDialector) Dialector(ctx context.Context, dsn string, config *gorm.Config) (gorm.Dialector, error) {
	return nil, fmt.Errorf("dialector not implemented")
}
