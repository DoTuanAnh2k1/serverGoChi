package token

import (
	"os"
	"strings"
	"testing"

	"github.com/DoTuanAnh2k1/serverGoChi/internal/config"
	"github.com/DoTuanAnh2k1/serverGoChi/internal/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/models/config_models"

	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	l := logrus.New()
	l.SetLevel(logrus.ErrorLevel)
	logger.Logger = l
	logger.DbLogger = l

	config.Init(&config_models.Config{
		Token: config_models.TokenConfig{
			SecretKey:   "test-secret-key-for-unit-tests",
			ExpiryHours: 1,
		},
	})
	os.Exit(m.Run())
}

func TestCreateToken_ReturnsBasicPrefixedToken(t *testing.T) {
	tok, err := CreateToken("alice", "admin")
	if err != nil {
		t.Fatalf("CreateToken error: %v", err)
	}
	if !strings.HasPrefix(tok, "Basic ") {
		t.Errorf("token should start with 'Basic ', got %q", tok)
	}
}

func TestParseToken_RoundTrip(t *testing.T) {
	username := "alice"
	roles := "admin operator"

	tok, err := CreateToken(username, roles)
	if err != nil {
		t.Fatalf("CreateToken error: %v", err)
	}

	gotUser, gotRoles, err := ParseToken(tok)
	if err != nil {
		t.Fatalf("ParseToken error: %v", err)
	}
	if gotUser != username {
		t.Errorf("username: got %q, want %q", gotUser, username)
	}
	if gotRoles != roles {
		t.Errorf("roles: got %q, want %q", gotRoles, roles)
	}
}

func TestParseToken_StripsBasicPrefix(t *testing.T) {
	tok, _ := CreateToken("bob", "viewer")
	// ParseToken should work with or without "Basic " prefix
	tokWithoutPrefix := strings.TrimPrefix(tok, "Basic ")
	gotUser, _, err := ParseToken(tokWithoutPrefix)
	if err != nil {
		t.Fatalf("ParseToken without prefix error: %v", err)
	}
	if gotUser != "bob" {
		t.Errorf("username: got %q, want %q", gotUser, "bob")
	}
}

func TestParseToken_InvalidToken(t *testing.T) {
	_, _, err := ParseToken("Basic not.a.valid.token")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

func TestParseToken_RandomString(t *testing.T) {
	_, _, err := ParseToken("random-garbage")
	if err == nil {
		t.Fatal("expected error for random string, got nil")
	}
}

func TestParseToken_EmptyString(t *testing.T) {
	_, _, err := ParseToken("")
	if err == nil {
		t.Fatal("expected error for empty token, got nil")
	}
}

func TestParseToken_WrongSigningKey(t *testing.T) {
	// Create token with current key, then change key and try to parse
	tok, _ := CreateToken("eve", "admin")

	// Temporarily use a different key by creating a fake token with another key
	// Test that a token signed with a different key is rejected
	fakeToken := "Basic eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJldmUiLCJhdWQiOiJhZG1pbiIsImV4cCI6OTk5OTk5OTk5OX0.invalid_signature"
	_, _, err := ParseToken(fakeToken)
	if err == nil {
		t.Error("expected error for tampered token")
	}
	// the valid one should still work
	_, _, err2 := ParseToken(tok)
	if err2 != nil {
		t.Errorf("valid token should parse without error: %v", err2)
	}
}

func TestCreateToken_EmptyRoles(t *testing.T) {
	tok, err := CreateToken("alice", "")
	if err != nil {
		t.Fatalf("CreateToken with empty roles error: %v", err)
	}
	_, roles, err := ParseToken(tok)
	if err != nil {
		t.Fatalf("ParseToken error: %v", err)
	}
	if roles != "" {
		t.Errorf("roles: got %q, want empty", roles)
	}
}
