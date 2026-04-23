package service_test

import (
	"strings"
	"testing"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/testutil"
)

// ─── Policy validation ─────────────────────────────────────────────────

func TestValidatePasswordAgainstPolicy_NilPolicy(t *testing.T) {
	if err := service.ValidatePasswordAgainstPolicy(nil, "anything"); err != nil {
		t.Errorf("nil policy should never error, got %v", err)
	}
}

func TestValidatePasswordAgainstPolicy_MinLength(t *testing.T) {
	p := &db_models.CliPasswordPolicy{MinLength: 8}
	if err := service.ValidatePasswordAgainstPolicy(p, "short"); err == nil {
		t.Errorf("short password must fail")
	}
	if err := service.ValidatePasswordAgainstPolicy(p, "longenough"); err != nil {
		t.Errorf("long password should pass, got %v", err)
	}
}

func TestValidatePasswordAgainstPolicy_ComplexityFlags(t *testing.T) {
	p := &db_models.CliPasswordPolicy{
		MinLength: 8, RequireUppercase: true, RequireLowercase: true,
		RequireDigit: true, RequireSpecial: true,
	}
	cases := []struct {
		pass string
		ok   bool
	}{
		{"alllowercase1!", false}, // no uppercase
		{"ALLUPPERCASE1!", false}, // no lowercase
		{"MissingDigits!", false}, // no digit
		{"MissingSpecial1", false},
		{"Strong1Pass!", true},
	}
	for _, c := range cases {
		err := service.ValidatePasswordAgainstPolicy(p, c.pass)
		if c.ok && err != nil {
			t.Errorf("pass=%q: expected OK, got %v", c.pass, err)
		}
		if !c.ok && err == nil {
			t.Errorf("pass=%q: expected error", c.pass)
		}
	}
}

// ─── Effective policy (strict-est merge) ───────────────────────────────

func TestGetEffectivePasswordPolicy_UnionsAcrossGroups(t *testing.T) {
	strict := &db_models.CliPasswordPolicy{
		ID: 1, Name: "strict", MinLength: 12, MaxAgeDays: 60,
		RequireUppercase: true, HistoryCount: 12,
		MaxLoginFailure: 3, LockoutMinutes: 30,
	}
	standard := &db_models.CliPasswordPolicy{
		ID: 2, Name: "standard", MinLength: 8, MaxAgeDays: 90,
		RequireDigit: true, HistoryCount: 6,
		MaxLoginFailure: 5, LockoutMinutes: 15,
	}
	pid1 := int64(1)
	pid2 := int64(2)
	store.SetSingleton(&testutil.MockStore{
		GetAllGroupsOfUserFn: func(_ int64) ([]*db_models.CliUserGroupMapping, error) {
			return []*db_models.CliUserGroupMapping{{UserID: 1, GroupID: 10}, {UserID: 1, GroupID: 11}}, nil
		},
		GetGroupByIdFn: func(id int64) (*db_models.CliGroup, error) {
			switch id {
			case 10:
				return &db_models.CliGroup{ID: 10, PasswordPolicyID: &pid1}, nil
			case 11:
				return &db_models.CliGroup{ID: 11, PasswordPolicyID: &pid2}, nil
			}
			return nil, nil
		},
		GetPasswordPolicyByIdFn: func(id int64) (*db_models.CliPasswordPolicy, error) {
			switch id {
			case 1:
				return strict, nil
			case 2:
				return standard, nil
			}
			return nil, nil
		},
	})
	eff, err := service.GetEffectivePasswordPolicy(1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if eff == nil {
		t.Fatal("effective policy nil")
	}
	// Strict-est picks:
	//   MinLength      → max = 12
	//   MaxAgeDays     → min-non-zero = 60
	//   RequireUppercase → union = true
	//   RequireDigit   → union = true
	//   HistoryCount   → max = 12
	//   MaxLoginFailure → min-non-zero = 3
	//   LockoutMinutes → max = 30
	if eff.MinLength != 12 {
		t.Errorf("MinLength: got %d want 12", eff.MinLength)
	}
	if eff.MaxAgeDays != 60 {
		t.Errorf("MaxAgeDays: got %d want 60", eff.MaxAgeDays)
	}
	if !eff.RequireUppercase || !eff.RequireDigit {
		t.Errorf("require_*: expected both set, got %+v", eff)
	}
	if eff.HistoryCount != 12 {
		t.Errorf("HistoryCount: got %d want 12", eff.HistoryCount)
	}
	if eff.MaxLoginFailure != 3 {
		t.Errorf("MaxLoginFailure: got %d want 3", eff.MaxLoginFailure)
	}
	if eff.LockoutMinutes != 30 {
		t.Errorf("LockoutMinutes: got %d want 30", eff.LockoutMinutes)
	}
}

func TestGetEffectivePasswordPolicy_NoPolicy(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetAllGroupsOfUserFn: func(_ int64) ([]*db_models.CliUserGroupMapping, error) {
			return []*db_models.CliUserGroupMapping{{UserID: 1, GroupID: 10}}, nil
		},
		GetGroupByIdFn: func(_ int64) (*db_models.CliGroup, error) {
			return &db_models.CliGroup{ID: 10}, nil // no policy
		},
	})
	eff, err := service.GetEffectivePasswordPolicy(1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if eff != nil {
		t.Errorf("expected nil policy, got %+v", eff)
	}
}

// ─── Password history ──────────────────────────────────────────────────

func TestIsPasswordReused_DetectsOld(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetRecentPasswordHistoryFn: func(userID int64, limit int) ([]*db_models.CliPasswordHistory, error) {
			return []*db_models.CliPasswordHistory{
				{PasswordHash: "hash_old_1"},
				{PasswordHash: "hash_old_2"},
			}, nil
		},
	})
	reused, _ := service.IsPasswordReused(1, "hash_old_2", 5)
	if !reused {
		t.Errorf("expected reuse detected")
	}
	reused, _ = service.IsPasswordReused(1, "hash_fresh", 5)
	if reused {
		t.Errorf("fresh hash must not be flagged")
	}
}

func TestIsPasswordReused_ZeroHistoryCountSkips(t *testing.T) {
	// Even if the repo has rows, historyCount=0 must short-circuit true.
	store.SetSingleton(&testutil.MockStore{
		GetRecentPasswordHistoryFn: func(_ int64, _ int) ([]*db_models.CliPasswordHistory, error) {
			t.Error("should not be called when historyCount=0")
			return nil, nil
		},
	})
	reused, err := service.IsPasswordReused(1, "any", 0)
	if err != nil || reused {
		t.Errorf("historyCount=0 should return (false, nil), got (%v, %v)", reused, err)
	}
}

// ─── Lockout ──────────────────────────────────────────────────────────

func TestIsAccountLocked_NoPolicy(t *testing.T) {
	u := &db_models.TblAccount{LoginFailureCount: 99}
	if s := service.IsAccountLocked(u, nil); s.Locked {
		t.Errorf("nil policy should never lock, got %+v", s)
	}
}

func TestIsAccountLocked_BelowThreshold(t *testing.T) {
	u := &db_models.TblAccount{LoginFailureCount: 2, LockedTime: time.Now()}
	p := &db_models.CliPasswordPolicy{MaxLoginFailure: 5, LockoutMinutes: 30}
	if s := service.IsAccountLocked(u, p); s.Locked {
		t.Errorf("below threshold must not lock, got %+v", s)
	}
}

func TestIsAccountLocked_ActiveLockout(t *testing.T) {
	u := &db_models.TblAccount{LoginFailureCount: 5, LockedTime: time.Now().Add(-5 * time.Minute)}
	p := &db_models.CliPasswordPolicy{MaxLoginFailure: 5, LockoutMinutes: 30}
	s := service.IsAccountLocked(u, p)
	if !s.Locked {
		t.Errorf("expected lock")
	}
	if s.RemainingS <= 0 || s.RemainingS > 30*60 {
		t.Errorf("remaining out of expected band, got %ds", s.RemainingS)
	}
}

func TestIsAccountLocked_Expired(t *testing.T) {
	u := &db_models.TblAccount{LoginFailureCount: 5, LockedTime: time.Now().Add(-31 * time.Minute)}
	p := &db_models.CliPasswordPolicy{MaxLoginFailure: 5, LockoutMinutes: 30}
	if s := service.IsAccountLocked(u, p); s.Locked {
		t.Errorf("lockout window should have expired, got %+v", s)
	}
}

// ─── ComputePasswordExpiry / IsPasswordExpired ─────────────────────────

func TestComputePasswordExpiry_NilWhenUnlimited(t *testing.T) {
	if exp := service.ComputePasswordExpiry(nil, time.Now()); exp != nil {
		t.Errorf("nil policy should return nil")
	}
	p := &db_models.CliPasswordPolicy{MaxAgeDays: 0}
	if exp := service.ComputePasswordExpiry(p, time.Now()); exp != nil {
		t.Errorf("MaxAgeDays=0 should return nil")
	}
}

func TestComputePasswordExpiry_ShiftsByDays(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	p := &db_models.CliPasswordPolicy{MaxAgeDays: 30}
	exp := service.ComputePasswordExpiry(p, now)
	if exp == nil {
		t.Fatal("expected non-nil")
	}
	if !exp.Equal(now.Add(30 * 24 * time.Hour)) {
		t.Errorf("expected +30d, got %v", exp)
	}
}

func TestIsPasswordExpired(t *testing.T) {
	past := time.Now().Add(-time.Hour)
	if !service.IsPasswordExpired(&db_models.TblAccount{PasswordExpiresAt: &past}) {
		t.Errorf("expired time in past should be expired")
	}
	future := time.Now().Add(time.Hour)
	if service.IsPasswordExpired(&db_models.TblAccount{PasswordExpiresAt: &future}) {
		t.Errorf("future time must not be expired")
	}
	if service.IsPasswordExpired(&db_models.TblAccount{}) {
		t.Errorf("nil expiry must not be expired")
	}
}

// ─── Mgt Permission ───────────────────────────────────────────────────

func TestCreateMgtPermission_Validates(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{})
	if err := service.CreateMgtPermission(&db_models.CliGroupMgtPermission{Resource: "user", Action: "read"}); err == nil {
		t.Errorf("missing group_id should error")
	}
	if err := service.CreateMgtPermission(&db_models.CliGroupMgtPermission{GroupID: 1, Action: "read"}); err == nil {
		t.Errorf("missing resource should error")
	}
	if err := service.CreateMgtPermission(&db_models.CliGroupMgtPermission{GroupID: 1, Resource: "user"}); err == nil {
		t.Errorf("missing action should error")
	}
	if err := service.CreateMgtPermission(&db_models.CliGroupMgtPermission{GroupID: 1, Resource: "user", Action: "read"}); err != nil {
		t.Errorf("happy: %v", err)
	}
}

func TestUserHasMgtPermission_Wildcards(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetAllGroupsOfUserFn: func(_ int64) ([]*db_models.CliUserGroupMapping, error) {
			return []*db_models.CliUserGroupMapping{{UserID: 1, GroupID: 10}}, nil
		},
		ListMgtPermissionsFn: func(gid int64) ([]*db_models.CliGroupMgtPermission, error) {
			return []*db_models.CliGroupMgtPermission{
				{ID: 1, GroupID: 10, Resource: "user", Action: "read"},
				{ID: 2, GroupID: 10, Resource: "ne", Action: "*"}, // wildcard action
				{ID: 3, GroupID: 10, Resource: "*", Action: "read"}, // wildcard resource
			}, nil
		},
	})
	cases := []struct {
		resource, action string
		want             bool
	}{
		{"user", "read", true},    // exact
		{"user", "update", false}, // user.update not granted (only user.read)
		{"ne", "create", true},    // ne.* wildcard
		{"ne", "delete", true},    // ne.* wildcard
		{"command", "read", true}, // *.read wildcard
		{"command", "update", false},
	}
	for _, c := range cases {
		ok, err := service.UserHasMgtPermission(1, c.resource, c.action)
		if err != nil {
			t.Errorf("(%s, %s): err=%v", c.resource, c.action, err)
			continue
		}
		if ok != c.want {
			t.Errorf("(%s, %s): got %v want %v", c.resource, c.action, ok, c.want)
		}
	}
}

// Ensures CreatePasswordPolicy validator rejects bad input.
func TestCreatePasswordPolicy_Validates(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{})
	if err := service.CreatePasswordPolicy(&db_models.CliPasswordPolicy{}); err == nil {
		t.Errorf("empty name should error")
	}
	if err := service.CreatePasswordPolicy(&db_models.CliPasswordPolicy{Name: "x", MinLength: -1}); err == nil {
		t.Errorf("negative MinLength should error")
	}
	if err := service.CreatePasswordPolicy(&db_models.CliPasswordPolicy{Name: "strict", MinLength: 12}); err != nil {
		t.Errorf("happy: %v", err)
	}
}

// Silences unused import warnings if test helpers shift.
var _ = strings.Contains
