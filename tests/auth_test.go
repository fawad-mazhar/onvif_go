package tests

import (
	"testing"

	"github.com/fawad-mazhar/onvif_go/auth"
)

func TestUsernameTokenValidation(t *testing.T) {
	// Create an authentication context
	authContext := &auth.ServiceContext{
		Username: "admin",
		Password: "admin123",
	}
	
	// Test valid username token with plain password
	validToken := auth.UsernameToken{
		Username: "admin",
		Password: "admin123",
		Nonce: "",
		Created: "",
	}
	
	if !authContext.ValidateUsernameToken(validToken) {
		t.Error("Expected valid token validation to pass, but it failed")
	}
	
	// Test invalid username
	invalidUsernameToken := auth.UsernameToken{
		Username: "invalid",
		Password: "admin123",
		Nonce: "",
		Created: "",
	}
	
	if authContext.ValidateUsernameToken(invalidUsernameToken) {
		t.Error("Expected invalid username token validation to fail, but it passed")
	}
	
	// Test invalid password
	invalidPasswordToken := auth.UsernameToken{
		Username: "admin",
		Password: "invalid",
		Nonce: "",
		Created: "",
	}
	
	if authContext.ValidateUsernameToken(invalidPasswordToken) {
		t.Error("Expected invalid password token validation to fail, but it passed")
	}
}

func TestNonceTimestampValidation(t *testing.T) {
	// Test valid timestamp (current time)
	validTimestamp := auth.GenerateCreatedTimestamp()
	if !auth.ValidateNonceTimestamp(validTimestamp, 300) {
		t.Error("Expected current timestamp validation to pass, but it failed")
	}
	
	// Test invalid timestamp (too old)
	invalidTimestamp := "2020-01-01T00:00:00Z"
	if auth.ValidateNonceTimestamp(invalidTimestamp, 300) {
		t.Error("Expected old timestamp validation to fail, but it passed")
	}
}
