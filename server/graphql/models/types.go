package models

import (
	"time"

	"gorm.io/gorm"
)

// DeletedAt returns a pointer to the time.Time value if the gorm.DeletedAt is valid, otherwise it returns nil.
func DeletedAt(t gorm.DeletedAt) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}
