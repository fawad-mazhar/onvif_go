package tests

import (
	"crypto/sha1"
	"encoding/base64"
	"testing"
	"time"

	"github.com/fawad-mazhar/onvif-go/internal/auth"
)

// makeDigestToken computes a WS-Security UsernameToken digest:
// Base64(SHA1(nonce + created + password)) per ONVIF spec.
func makeDigestToken(username, password string, nonce []byte, created string) auth.UsernameToken {
	h := sha1.New()
	h.Write(nonce)
	h.Write([]byte(created))
	h.Write([]byte(password))
	return auth.UsernameToken{
		Username: username,
		Password: base64.StdEncoding.EncodeToString(h.Sum(nil)),
		Nonce:    base64.StdEncoding.EncodeToString(nonce),
		Created:  created,
	}
}

func TestUsernameTokenValidation(t *testing.T) {
	authContext := &auth.ServiceContext{
		Username: "admin",
		Password: "admin123",
	}

	nonce := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	created := time.Now().UTC().Format(time.RFC3339)

	// Valid digest token
	validToken := makeDigestToken("admin", "admin123", nonce, created)
	if !authContext.ValidateUsernameToken(validToken) {
		t.Error("Expected valid digest token validation to pass, but it failed")
	}

	// Invalid username — digest is correct for the right password but username is wrong
	invalidUsernameToken := makeDigestToken("invalid", "admin123", nonce, created)
	if authContext.ValidateUsernameToken(invalidUsernameToken) {
		t.Error("Expected invalid username token validation to fail, but it passed")
	}

	// Wrong password — valid nonce/created, digest computed against wrong password
	invalidPasswordToken := makeDigestToken("admin", "wrongpassword", nonce, created)
	if authContext.ValidateUsernameToken(invalidPasswordToken) {
		t.Error("Expected wrong-password digest token validation to fail, but it passed")
	}

	// Plain-text password (no Nonce/Created) must be rejected — G-002
	plainToken := auth.UsernameToken{Username: "admin", Password: "admin123"}
	if authContext.ValidateUsernameToken(plainToken) {
		t.Error("Expected plain-text password to be rejected (G-002), but it passed")
	}
}

func TestNonceTimestampValidation(t *testing.T) {
	// Test valid timestamp (current time)
	validTimestamp := time.Now().Format(time.RFC3339)
	if !auth.ValidateNonceTimestamp(validTimestamp, 300) {
		t.Error("Expected current timestamp validation to pass, but it failed")
	}

	// Test invalid timestamp (too old)
	invalidTimestamp := "2020-01-01T00:00:00Z"
	if auth.ValidateNonceTimestamp(invalidTimestamp, 300) {
		t.Error("Expected old timestamp validation to fail, but it passed")
	}
}
