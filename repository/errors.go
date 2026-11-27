package repository

import "gorm.io/gorm"

// ErrNotFound is returned by repository Get methods when the object doesn't
// exist (or is soft-deleted). It is gorm.ErrRecordNotFound, so checks with
// either value work.
var ErrNotFound = gorm.ErrRecordNotFound
