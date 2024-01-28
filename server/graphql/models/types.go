package models

import (
	"time"

	"gorm.io/gorm"
)

func deletedAt(t gorm.DeletedAt) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}

func s4ptr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
