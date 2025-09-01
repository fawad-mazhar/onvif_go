package auth

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
	
	"github.com/fawad-mazhar/onvif_go/logger"
	"github.com/fawad-mazhar/onvif_go/utils"
)

// UsernameToken represents the WS-Security username token
type UsernameToken struct {
	Username  string
	Password  string
	Nonce     string
	Created   string
}

// ServiceContext holds the authentication context
type ServiceContext struct {
	Username  string
	Password  string
}

// ValidateUsernameToken validates the WS-Security username token
func (s *ServiceContext) ValidateUsernameToken(token UsernameToken) bool {
	// If no username/password is set in the configuration, skip validation
	if s.Username == "" || s.Password == "" {
		return true
	}
	
	// Check if the username matches
	if token.Username != s.Username {
		logger.Warn("Username mismatch: expected %s, got %s", s.Username, token.Username)
		return false
	}
	
	// If the password is provided in plain text, validate directly
	if token.Password != "" && token.Nonce == "" && token.Created == "" {
		if token.Password == s.Password {
			return true
		}
		logger.Warn("Password mismatch")
		return false
	}
	
	// Validate password digest
	if token.Nonce != "" && token.Created != "" {
		// Decode the nonce from base64
		nonce, err := base64.StdEncoding.DecodeString(token.Nonce)
		if err != nil {
			logger.Warn("Failed to decode nonce: %v", err)
			return false
		}
		
		// Create the digest according to ONVIF specification
		// Digest = Base64(SHA1(Nonce + Created + Password))
		h := sha1.New()
		h.Write(nonce)
		h.Write([]byte(token.Created))
		h.Write([]byte(s.Password))
		digest := base64.StdEncoding.EncodeToString(h.Sum(nil))
		
		// Compare with the provided password
		if token.Password == digest {
			return true
		}
		logger.Warn("Password digest mismatch")
		return false
	}
	
	return false
}

// ParseSOAPHeader parses the SOAP header to extract username token information
func ParseSOAPHeader(soapHeader string) (UsernameToken, error) {
	var token UsernameToken
	
	// Extract username
	usernameStart := strings.Index(soapHeader, "<Username>")
	if usernameStart == -1 {
		return token, fmt.Errorf("username not found in SOAP header")
	}
	usernameEnd := strings.Index(soapHeader, "</Username>")
	if usernameEnd == -1 || usernameEnd <= usernameStart {
		return token, fmt.Errorf("invalid username in SOAP header")
	}
	token.Username = soapHeader[usernameStart+10:usernameEnd]
	
	// Extract password
	passwordStart := strings.Index(soapHeader, "<Password>")
	if passwordStart == -1 {
		return token, fmt.Errorf("password not found in SOAP header")
	}
	passwordEnd := strings.Index(soapHeader, "</Password>")
	if passwordEnd == -1 || passwordEnd <= passwordStart {
		return token, fmt.Errorf("invalid password in SOAP header")
	}
	token.Password = soapHeader[passwordStart+10:passwordEnd]
	
	// Extract nonce (if present)
	nonceStart := strings.Index(soapHeader, "<Nonce>")
	if nonceStart != -1 {
		nonceEnd := strings.Index(soapHeader, "</Nonce>")
		if nonceEnd != -1 && nonceEnd > nonceStart {
			token.Nonce = soapHeader[nonceStart+7:nonceEnd]
		}
	}
	
	// Extract created timestamp (if present)
	createdStart := strings.Index(soapHeader, "<Created>")
	if createdStart != -1 {
		createdEnd := strings.Index(soapHeader, "</Created>")
		if createdEnd != -1 && createdEnd > createdStart {
			token.Created = soapHeader[createdStart+9:createdEnd]
		}
	}
	
	return token, nil
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
		logger.Warn("Failed to parse created timestamp: %v", err)
		return false
	}
	
	// Check if the timestamp is within the allowed age
	now := time.Now()
	diff := now.Sub(createdTime)
	if diff.Seconds() > float64(maxAgeSeconds) {
		logger.Warn("Nonce timestamp too old: %v", diff)
		return false
	}
	
	return true
}

// GenerateNonce creates a new nonce for authentication
func GenerateNonce() string {
	// Generate a random nonce
	nonce := utils.GenerateRandomString(16)
	return base64.StdEncoding.EncodeToString([]byte(nonce))
}

// GenerateCreatedTimestamp creates a timestamp for authentication
func GenerateCreatedTimestamp() string {
	// Generate current timestamp in RFC3339 format
	return time.Now().Format(time.RFC3339)
}
