package session

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// CrossAuthHeader value as AccountID[:UserID]
const CrossAuthHeader = "auth.cross.account"

// ErrInvalidCrossAuthHeader is returned when the cross auth header value is malformed.
var ErrInvalidCrossAuthHeader = errors.New("invalid cross auth header value")

// NewCrossAuthHeader from Account & User ID
func NewCrossAuthHeader(accid, uid uint64) string {
	if uid <= 0 {
		return strconv.FormatUint(accid, 10)
	}
	return fmt.Sprintf("%d:%d", accid, uid)
}

// ParseCrossAuthHeader from value in the format AccountID[:UserID].
// Malformed values yield zero IDs; use ParseCrossAuthHeaderStrict to get the error.
func ParseCrossAuthHeader(val string) (accid, uid uint64) {
	accid, uid, _ = ParseCrossAuthHeaderStrict(val)
	return accid, uid
}

// ParseCrossAuthHeaderStrict parses a value in the format AccountID[:UserID].
// The account ID is required and must be positive; the user ID is optional.
func ParseCrossAuthHeaderStrict(val string) (accid, uid uint64, err error) {
	accPart, userPart, hasUser := strings.Cut(strings.TrimSpace(val), ":")
	if accid, err = strconv.ParseUint(strings.TrimSpace(accPart), 10, 64); err != nil || accid == 0 {
		return 0, 0, ErrInvalidCrossAuthHeader
	}
	if hasUser {
		if uid, err = strconv.ParseUint(strings.TrimSpace(userPart), 10, 64); err != nil || uid == 0 {
			return 0, 0, ErrInvalidCrossAuthHeader
		}
	}
	return accid, uid, nil
}
