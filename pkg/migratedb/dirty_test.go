package migratedb

import (
	"errors"
	"testing"

	mdatabase "github.com/golang-migrate/migrate/v4/database"
)

type fakeDriver struct {
	mdatabase.Driver
	version int
	dirty   bool
	setTo   *int
}

func (d *fakeDriver) Version() (int, bool, error) { return d.version, d.dirty, nil }
func (d *fakeDriver) SetVersion(v int, dirty bool) error {
	d.setTo, d.dirty = &v, dirty
	return nil
}

func TestResolveDirty(t *testing.T) {
	t.Parallel()

	clean := &fakeDriver{version: 5}
	if err := resolveDirty(clean, MigrateSource{}, "schema_migrations"); err != nil || clean.setTo != nil {
		t.Fatalf("clean: err = %v, setTo = %v", err, clean.setTo)
	}

	// Before the fix a dirty version was silently reset and re-run
	dirty := &fakeDriver{version: 5, dirty: true}
	if err := resolveDirty(dirty, MigrateSource{}, "schema_migrations"); !errors.Is(err, ErrDirtyDatabase) || dirty.setTo != nil {
		t.Fatalf("dirty: err = %v, setTo = %v", err, dirty.setTo)
	}

	allowed := &fakeDriver{version: 5, dirty: true}
	if err := resolveDirty(allowed, MigrateSource{ResetDirty: true}, "schema_migrations"); err != nil || allowed.setTo == nil || *allowed.setTo != 4 {
		t.Fatalf("reset allowed: err = %v, setTo = %v", err, allowed.setTo)
	}
}
