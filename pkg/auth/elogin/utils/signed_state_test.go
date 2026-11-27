package utils

import (
	"errors"
	"testing"
	"time"
)

func TestStateCodec(t *testing.T) {
	t.Parallel()

	codec, err := NewStateCodec([]byte("0123456789abcdef"), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	encoded, nonce, err := codec.Encode(NewState(
		Param{Key: "r", Value: "/home"},
		Param{Key: StateNonceKey, Value: "attacker-chosen"},
	))
	if err != nil {
		t.Fatal(err)
	}

	state, err := codec.Decode(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if state.Get("r") != "/home" || state.Get(StateNonceKey) != nonce || len(nonce) != 32 {
		t.Fatalf("state = %+v, nonce = %q", state, nonce)
	}

	// Every encode gets a fresh nonce
	if _, nonce2, _ := codec.Encode(nil); nonce2 == nonce {
		t.Fatal("nonce is reused")
	}

	other, _ := NewStateCodec([]byte("fedcba9876543210"), time.Minute)
	for name, value := range map[string]string{
		"empty":           "",
		"no signature":    encoded[:len(encoded)-44],
		"tampered":        tamper(encoded),
		"bad signature":   encoded[:len(encoded)-2] + "AA",
		"legacy unsigned": NewState(Param{Key: "access_token", Value: "x"}).Encode(),
		"other secret":    func() string { v, _, _ := other.Encode(nil); return v }(),
		"garbage base64":  "!!!.!!!",
	} {
		if _, err := codec.Decode(value); !errors.Is(err, ErrStateInvalid) {
			t.Errorf("%s: error = %v, want ErrStateInvalid", name, err)
		}
	}

	expired, _ := NewStateCodec([]byte("0123456789abcdef"), time.Minute)
	expired.now = func() time.Time { return time.Now().Add(2 * time.Minute) }
	if _, err := expired.Decode(encoded); !errors.Is(err, ErrStateExpired) {
		t.Fatalf("expired state error = %v, want ErrStateExpired", err)
	}

	if _, err := NewStateCodec([]byte("short"), time.Minute); !errors.Is(err, ErrStateSecretTooShort) {
		t.Fatalf("short secret error = %v", err)
	}
}

// tamper changes one character of the payload part.
func tamper(encoded string) string {
	b := []byte(encoded)
	if b[3] == 'A' {
		b[3] = 'B'
	} else {
		b[3] = 'A'
	}
	return string(b)
}
