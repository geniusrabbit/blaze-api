package appcontext

import (
	"strings"
	"testing"
)

func TestConfigStringMasksSecrets(t *testing.T) {
	t.Parallel()

	var cfg ConfigType
	cfg.System.Storage.MasterConnect = "postgres://user:db-pass-123@db/x"
	cfg.Superuser.Password = "super-pass-123"
	cfg.OAuth2.Secret = "oauth-secret-123"
	cfg.Session.DevToken = "dev-token-123"
	cfg.SocialAuth.Facebook.ClientSecret = "fb-secret-123"
	cfg.Superuser.Email = "admin@example.com"

	out := cfg.String()
	for _, secret := range []string{"db-pass-123", "super-pass-123", "oauth-secret-123", "dev-token-123", "fb-secret-123"} {
		if strings.Contains(out, secret) {
			t.Errorf("config dump contains %q", secret)
		}
	}
	if !strings.Contains(out, "admin@example.com") {
		t.Error("non-secret values are missing from the dump")
	}
}

func TestDebugIsExplicit(t *testing.T) {
	t.Parallel()

	// Before the fix debug followed LOG_LEVEL=debug, which was the default
	cfg := ConfigType{LogLevel: "debug"}
	if cfg.IsDebug() {
		t.Fatal("LOG_LEVEL=debug enabled debug mode")
	}
	cfg.Debug = true
	if !cfg.IsDebug() {
		t.Fatal("DEBUG=true didn't enable debug mode")
	}
}

func TestCORSOrigins(t *testing.T) {
	t.Parallel()

	var cfg ConfigType
	cfg.Server.HTTP.CORSOrigins = " https://a.example , ,https://b.example"
	if got := cfg.CORSOrigins(); len(got) != 2 || got[0] != "https://a.example" || got[1] != "https://b.example" {
		t.Fatalf("origins = %q", got)
	}
}
