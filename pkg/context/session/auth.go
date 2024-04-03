package session

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/demdxx/gocast/v2"
)

// CrossAuthHeader value as AccountID[:UserID]
const CrossAuthHeader = "auth.cross.account"

// NewCrossAuthHeader from Account & User ID
func NewCrossAuthHeader(accid, uid uint64) string {
	if uid <= 0 {
		return strconv.FormatUint(accid, 10)
	}
	return fmt.Sprintf("%d:%d", accid, uid)
}

// ParseCrossAuthHeader from value
func ParseCrossAuthHeader(val string) (accid, uid uint64) {
	connectIDs := strings.Split(val, ":")
	accid = gocast.Uint64(connectIDs[0])
	if len(connectIDs) == 2 {
		uid = gocast.Uint64(connectIDs[1])
	}
	return accid, uid
}
