package token

import (
	"os"
	"strings"
	"testing"

	"github.com/DoTuanAnh2k1/serverGoChi/models/config_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/config"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"

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
	tok, err := CreateToken("alice")
	if err != nil {
		t.Fatalf("CreateToken error: %v", err)
	}
	if !strings.HasPrefix(tok, "Basic ") {
		t.Errorf("token should start with 'Basic ', got %q", tok)
	}
}

func TestParseToken_RoundTrip(t *testing.T) {
	username := "alice"
	tok, err := CreateToken(username)
	if err != nil {
		t.Fatalf("CreateToken error: %v", err)
	}

	gotUser, err := ParseToken(tok)
	if err != nil {
		t.Fatalf("ParseToken error: %v", err)
	}
	if gotUser != username {
		t.Errorf("username: got %q, want %q", gotUser, username)
	}
}

func TestParseToken_StripsBasicPrefix(t *testing.T) {
	tok, _ := CreateToken("bob")
	tokWithoutPrefix := strings.TrimPrefix(tok, "Basic ")
	gotUser, err := ParseToken(tokWithoutPrefix)
	if err != nil {
		t.Fatalf("ParseToken without prefix error: %v", err)
	}
	if gotUser != "bob" {
		t.Errorf("username: got %q, want %q", gotUser, "bob")
	}
}

func TestParseToken_InvalidToken(t *testing.T) {
	if _, err := ParseToken("Basic not.a.valid.token"); err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

func TestParseToken_RandomString(t *testing.T) {
	if _, err := ParseToken("random-garbage"); err == nil {
		t.Fatal("expected error for random string, got nil")
	}
}

func TestParseToken_EmptyString(t *testing.T) {
	if _, err := ParseToken(""); err == nil {
		t.Fatal("expected error for empty token, got nil")
	}
}

func TestParseToken_WrongSigningKey(t *testing.T) {
	tok, _ := CreateToken("eve")
	fakeToken := "Basic eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJldmUiLCJleHAiOjk5OTk5OTk5OTl9.invalid_signature"
	if _, err := ParseToken(fakeToken); err == nil {
		t.Error("expected error for tampered token")
	}
	if _, err := ParseToken(tok); err != nil {
		t.Errorf("valid token should parse without error: %v", err)
	}
}
