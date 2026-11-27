package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	// StateNonceKey is the state parameter holding the random per-login nonce.
	StateNonceKey = "_n"
	// StateExpiresKey is the state parameter holding the expiration Unix time.
	StateExpiresKey = "_e"

	minStateSecretLen = 16
)

var (
	// ErrStateSecretTooShort is returned when the signing secret is too short.
	ErrStateSecretTooShort = errors.New("state secret must be at least 16 bytes")
	// ErrStateInvalid is returned for malformed, forged or tampered states.
	ErrStateInvalid = errors.New("invalid state")
	// ErrStateExpired is returned when the state lifetime is over.
	ErrStateExpired = errors.New("state is expired")
)

// StateCodec signs and verifies OAuth2 state values.
//
// The encoded value is `base64url(json) "." base64url(hmac-sha256)`. Every
// encoded state carries a random nonce (StateNonceKey) and an expiration time
// (StateExpiresKey). The payload is signed, not encrypted: never put secrets in it.
type StateCodec struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

// NewStateCodec creates a state codec with the given secret and state lifetime.
func NewStateCodec(secret []byte, ttl time.Duration) (*StateCodec, error) {
	if len(secret) < minStateSecretLen {
		return nil, ErrStateSecretTooShort
	}
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	return &StateCodec{secret: append([]byte(nil), secret...), ttl: ttl, now: time.Now}, nil
}

// TTL returns the state lifetime.
func (c *StateCodec) TTL() time.Duration { return c.ttl }

// Encode signs the state. It adds a fresh nonce and the expiration time,
// replacing any values of the reserved keys. The nonce is returned so the
// caller can bind it to the browser (e.g. a cookie).
func (c *StateCodec) Encode(state State) (encoded, nonce string, err error) {
	var nonceBytes [16]byte
	if _, err = rand.Read(nonceBytes[:]); err != nil {
		return "", "", err
	}
	nonce = hex.EncodeToString(nonceBytes[:])

	params := make(State, 0, len(state)+2)
	for _, p := range state {
		if p.Key != StateNonceKey && p.Key != StateExpiresKey {
			params = append(params, p)
		}
	}
	params = append(params,
		Param{Key: StateNonceKey, Value: nonce},
		Param{Key: StateExpiresKey, Value: strconv.FormatInt(c.now().Add(c.ttl).Unix(), 10)},
	)
	sort.Sort(params)

	payload, err := json.Marshal(stateToMap(params))
	if err != nil {
		return "", "", err
	}
	enc := base64.RawURLEncoding
	return enc.EncodeToString(payload) + "." + enc.EncodeToString(c.sign(payload)), nonce, nil
}

// Decode verifies the signature and expiration of the state and returns it.
func (c *StateCodec) Decode(encoded string) (State, error) {
	payloadPart, sigPart, ok := strings.Cut(encoded, ".")
	if !ok || payloadPart == "" || sigPart == "" {
		return nil, ErrStateInvalid
	}
	enc := base64.RawURLEncoding
	payload, err := enc.DecodeString(payloadPart)
	if err != nil {
		return nil, ErrStateInvalid
	}
	sig, err := enc.DecodeString(sigPart)
	if err != nil || !hmac.Equal(sig, c.sign(payload)) {
		return nil, ErrStateInvalid
	}

	var data map[string]string
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, ErrStateInvalid
	}
	state := make(State, 0, len(data))
	for k, v := range data {
		state = append(state, Param{Key: k, Value: v})
	}
	sort.Sort(state)

	if state.Get(StateNonceKey) == "" {
		return nil, ErrStateInvalid
	}
	expiresAt, err := strconv.ParseInt(state.Get(StateExpiresKey), 10, 64)
	if err != nil {
		return nil, ErrStateInvalid
	}
	if c.now().Unix() > expiresAt {
		return nil, ErrStateExpired
	}
	return state, nil
}

func (c *StateCodec) sign(payload []byte) []byte {
	mac := hmac.New(sha256.New, c.secret)
	_, _ = mac.Write(payload)
	return mac.Sum(nil)
}

func stateToMap(state State) map[string]string {
	data := make(map[string]string, len(state))
	for _, p := range state {
		data[p.Key] = p.Value
	}
	return data
}
