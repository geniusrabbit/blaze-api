package repository

import "testing"

func TestAllowedForRejectedSession(t *testing.T) {
	t.Parallel()

	// Before the fix the filter used `||` and kept every role
	for name, want := range map[string]bool{
		"system:admin":   false,
		"account:admin":  false,
		"account:editor": false,
		"viewer":         true,
		"user:basic":     true,
	} {
		if got := allowedForRejectedSession(name); got != want {
			t.Errorf("allowedForRejectedSession(%q) = %v, want %v", name, got, want)
		}
	}
}
