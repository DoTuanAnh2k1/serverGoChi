package bcrypt

import (
	"os"
	"testing"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/testutil"
)

func TestMain(m *testing.M) {
	testutil.InitTestLogger()
	os.Exit(m.Run())
}

func TestEncode_ReturnsNonEmptyHash(t *testing.T) {
	hash := Encode("mypassword")
	if hash == "" {
		t.Fatal("Encode returned empty string")
	}
	if hash == "mypassword" {
		t.Fatal("Encode returned plaintext, expected hash")
	}
}

func TestEncode_DifferentInputsDifferentHashes(t *testing.T) {
	h1 := Encode("password1")
	h2 := Encode("password2")
	if h1 == h2 {
		t.Error("different passwords should produce different hashes")
	}
}

func TestEncode_SameInputDifferentHashes(t *testing.T) {
	// bcrypt includes a random salt, so same input → different hashes
	h1 := Encode("same")
	h2 := Encode("same")
	if h1 == h2 {
		t.Error("bcrypt should produce different hashes for the same input due to random salt")
	}
}

func TestMatches_CorrectPassword(t *testing.T) {
	password := "secret123"
	hash := Encode(password)
	if !Matches(password, hash) {
		t.Error("Matches returned false for correct password")
	}
}

func TestMatches_WrongPassword(t *testing.T) {
	hash := Encode("correctpassword")
	if Matches("wrongpassword", hash) {
		t.Error("Matches returned true for wrong password")
	}
}

func TestMatches_EmptyPasswordAgainstHash(t *testing.T) {
	hash := Encode("nonempty")
	if Matches("", hash) {
		t.Error("Matches returned true for empty password")
	}
}

func TestMatches_InvalidHash(t *testing.T) {
	if Matches("password", "not-a-valid-bcrypt-hash") {
		t.Error("Matches returned true for an invalid hash")
	}
}

func TestEncode_EmptyString(t *testing.T) {
	hash := Encode("")
	// empty string is still a valid bcrypt input
	if hash == "" {
		t.Fatal("Encode returned empty for empty input")
	}
	if !Matches("", hash) {
		t.Error("Matches failed for empty password round-trip")
	}
}
