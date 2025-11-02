//go:build postgres || alldb
// +build postgres alldb

package database

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func init() {
	registerDialector(&postgresDialector{}, "postgres", "postgresql")
}

type postgresDialector struct{ defaultDialector }

func (d *postgresDialector) Dialector(ctx context.Context, dsn string, config *gorm.Config) (gorm.Dialector, error) {
	parsedDSN, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	sslmode := "disable"
	if sslmodeVar := parsedDSN.Query().Get("sslmode"); sslmodeVar != "" {
		sslmode = sslmodeVar
	}
	password, _ := parsedDSN.User.Password()
	newDSN := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=%s",
		parsedDSN.Hostname(), parsedDSN.Port(),
		parsedDSN.User.Username(), password,
		strings.TrimLeft(parsedDSN.Path, "/"),
		sslmode,
	)
	return postgres.Open(newDSN), nil
}
