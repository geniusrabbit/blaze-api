package models

import (
	"time"

	"gorm.io/gorm"
)

func DeletedAt(t gorm.DeletedAt) *time.Time {
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

func s2ptr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
