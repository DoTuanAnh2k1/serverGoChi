package service_test

import (
	"os"
	"testing"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/config_models"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/config"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/testutil"
)

// Ensure each test starts from a clean mock store + a token config good
// enough for JWT signing.
func installFreshMock(t *testing.T) *testutil.MockStore {
	t.Helper()
	config.Init(&config_models.Config{
		Token: config_models.TokenConfig{SecretKey: "test-secret", ExpiryHours: 1},
	})
	m := testutil.NewMockStore()
	store.SetSingleton(m)
	return m
}

func TestMainSuite(t *testing.T) {
	// Ensure logger global is initialized so service calls don't panic.
	testutil.InitTestLogger()
	_ = os.Stdout
}

func TestAuthenticate_HappyPath(t *testing.T) {
	testutil.InitTestLogger()
	m := installFreshMock(t)

	u := &db_models.User{Username: "alice", IsEnabled: true}
	if err := service.CreateUser(u, "Str0ng-pass!"); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	tok, err := service.Authenticate("alice", "Str0ng-pass!", "127.0.0.1")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if tok == "" {
		t.Fatal("expected a non-empty token")
	}
	// A successful auth should also stamp LastLoginAt.
	got, _ := m.GetUserByUsername("alice")
	if got == nil || got.LastLoginAt == nil {
		t.Fatalf("expected LastLoginAt to be set")
	}
}

func TestAuthenticate_WrongPasswordIncrementsFailureCount(t *testing.T) {
	testutil.InitTestLogger()
	m := installFreshMock(t)

	u := &db_models.User{Username: "bob", IsEnabled: true}
	if err := service.CreateUser(u, "Str0ng-pass!"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Authenticate("bob", "wrong", "127.0.0.1"); err == nil {
		t.Fatal("expected error on wrong password")
	}
	got, _ := m.GetUserByUsername("bob")
	if got.LoginFailureCount != 1 {
		t.Errorf("LoginFailureCount: got %d, want 1", got.LoginFailureCount)
	}
}

func TestAuthenticate_LocksAfterMaxFailures(t *testing.T) {
	testutil.InitTestLogger()
	m := installFreshMock(t)

	if err := service.UpsertPasswordPolicy(&db_models.PasswordPolicy{
		MinLength: 1, MaxLoginFailure: 3, LockoutMinutes: 10,
	}); err != nil {
		t.Fatal(err)
	}

	u := &db_models.User{Username: "charlie", IsEnabled: true}
	if err := service.CreateUser(u, "pw"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		_, _ = service.Authenticate("charlie", "bad", "127.0.0.1")
	}
	got, _ := m.GetUserByUsername("charlie")
	if got.LockedAt == nil {
		t.Fatal("expected account to be locked after threshold")
	}
	// A subsequent correct password should still be denied while locked.
	if _, err := service.Authenticate("charlie", "pw", "127.0.0.1"); err == nil {
		t.Fatal("expected lockout to block even a correct password")
	}
}

func TestAuthenticate_DisabledUser(t *testing.T) {
	testutil.InitTestLogger()
	m := installFreshMock(t)

	u := &db_models.User{Username: "diana", IsEnabled: false}
	_ = m.CreateUser(u)
	// Set a hash so the bcrypt check isn't what fails.
	hash, _ := service.HashPassword("pw")
	u.PasswordHash = hash
	_ = m.UpdateUser(u)

	if _, err := service.Authenticate("diana", "pw", "127.0.0.1"); err == nil {
		t.Fatal("expected disabled account to be rejected")
	}
}

func TestAuthenticate_BlacklistedUsername(t *testing.T) {
	testutil.InitTestLogger()
	m := installFreshMock(t)

	if err := service.CreateAccessListEntry(&db_models.UserAccessList{
		ListType: db_models.AccessListTypeBlacklist,
		MatchType: db_models.AccessListMatchUsername,
		Pattern:  "eve",
	}); err != nil {
		t.Fatal(err)
	}
	u := &db_models.User{Username: "eve", IsEnabled: true}
	if err := service.CreateUser(u, "Str0ng-pass!"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Authenticate("eve", "Str0ng-pass!", "127.0.0.1"); err == nil {
		t.Fatal("expected blacklist to block login")
	}
	_ = m
}

// nowForTest is a tiny helper that returns a time.Time in UTC so other tests
// in this package can reuse it without importing "time" themselves.
func nowForTest() time.Time { return time.Now().UTC() }
