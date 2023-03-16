package repository

import (
	"strings"

	"github.com/google/uuid"
)

func newID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}
