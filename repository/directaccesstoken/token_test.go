package directaccesstoken

import (
	"regexp"
	"testing"
)

var urlSafeToken = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func TestGenerateTokenIsURLSafe(t *testing.T) {
	t.Parallel()

	seen := make(map[string]struct{})
	for range 200 {
		token, err := GenerateToken(32)
		if err != nil {
			t.Fatalf("generate: %v", err)
		}
		if len(token) < 43 {
			t.Fatalf("token %q has %d characters, want at least 43", token, len(token))
		}
		if !urlSafeToken.MatchString(token) {
			t.Fatalf("token %q has characters outside the URL-safe alphabet", token)
		}
		if _, duplicate := seen[token]; duplicate {
			t.Fatalf("token %q generated twice", token)
		}
		seen[token] = struct{}{}
	}
}
