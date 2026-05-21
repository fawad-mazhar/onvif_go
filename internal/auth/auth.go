// Package auth provides authentication utilities for ONVIF services.
package auth

import (
	"crypto/sha1"
	"encoding/base64"
	"time"

	"github.com/fawad-mazhar/onvif-go/internal/logger"
	xmlpkg "github.com/fawad-mazhar/onvif-go/internal/xml"
)

// UsernameToken represents the WS-Security username token
type UsernameToken struct {
	Username string
	Password string
	Nonce    string
	Created  string
}

// ServiceContext holds the authentication context
type ServiceContext struct {
	Username string
	Password string
}

// ValidateUsernameToken validates the WS-Security username token
func (s *ServiceContext) ValidateUsernameToken(token UsernameToken) bool {
	// If no username/password is set in the configuration, skip validation
	if s.Username == "" || s.Password == "" {
		return true
	}

	// Check if the username matches
	if token.Username != s.Username {
		logger.Warnf("Username mismatch: expected %s, got %s", s.Username, token.Username)
		return false
	}

	// Validate password digest
	if token.Nonce != "" && token.Created != "" {
		// Decode the nonce from base64
		nonce, err := base64.StdEncoding.DecodeString(token.Nonce)
		if err != nil {
			logger.Warnf("Failed to decode nonce: %v", err)
			return false
		}

		// Create the digest according to ONVIF specification
		// Digest = Base64( SHA1( Base64Decode(Nonce) + Created + Password ) )
		h := sha1.New()
		h.Write(nonce)
		h.Write([]byte(token.Created))
		h.Write([]byte(s.Password))
		digest := base64.StdEncoding.EncodeToString(h.Sum(nil))

		// Compare with the provided password
		if token.Password == digest {
			return true
		}
		logger.Warnf("Password digest mismatch")
		return false
	}

	return false
}

// ParseSOAPHeader extracts the WS-Security UsernameToken fields from a
// SOAP envelope using a namespace-aware XML parser. Tolerates any prefix
// (wsse:, u:, etc.) and matches local-names case-sensitively per the WSS
// schema.
func ParseSOAPHeader(soapHeader string) (UsernameToken, error) {
	t, err := xmlpkg.ExtractUsernameToken([]byte(soapHeader))
	if err != nil {
		return UsernameToken{}, err
	}
	return UsernameToken{
		Username: t.Username,
		Password: t.Password,
		Nonce:    t.Nonce,
		Created:  t.Created,
	}, nil
}

// ValidateNonceTimestamp validates the nonce timestamp to prevent replay attacks
func ValidateNonceTimestamp(created string, maxAgeSeconds int) bool {
	// If no timestamp is provided, skip validation
	if created == "" {
		return true
	}

	// Parse the timestamp
	createdTime, err := time.Parse(time.RFC3339, created)
	if err != nil {
		logger.Warnf("Failed to parse created timestamp: %v", err)
		return false
	}

	// Check if the timestamp is within the allowed age
	now := time.Now()
	diff := now.Sub(createdTime)
	if diff.Seconds() > float64(maxAgeSeconds) {
		logger.Warnf("Nonce timestamp too old: %v", diff)
		return false
	}

	return true
}
