package appinit

import (
	"time"

	"github.com/alexedwards/scs/v2"
)

// SessionManager returns new session manager
func SessionManager(cookieName string, lifetime time.Duration) *scs.SessionManager {
	manager := scs.New()
	manager.Lifetime = lifetime
	manager.Cookie.Name = cookieName
	return manager
}
