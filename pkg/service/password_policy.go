package service

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
)

// ─────────────────────────────────────────────────────────────────────────
// Password policy, password history, mgt permission — remaining RBAC items
// from docs/rbac-design.md §4.8, §4.11, §6.3.
//
// Policy is per-group (cli_group.password_policy_id). A user may belong to
// multiple groups; the EFFECTIVE policy for that user takes the strict-est
// value per field: min(max_age_days>0), max(min_length), any required-*
// flag true wins, max(history_count), min(max_login_failure>0),
// max(lockout_minutes). Zero values mean "unset" for age/failure.
// ─────────────────────────────────────────────────────────────────────────

// ── CRUD wrappers ──

func CreatePasswordPolicy(p *db_models.CliPasswordPolicy) error {
	if strings.TrimSpace(p.Name) == "" {
		return errors.New("password_policy name is required")
	}
	if p.MinLength < 0 || p.MaxAgeDays < 0 || p.HistoryCount < 0 || p.MaxLoginFailure < 0 || p.LockoutMinutes < 0 {
		return errors.New("numeric fields must be >= 0")
	}
	return store.GetSingleton().CreatePasswordPolicy(p)
}

func GetPasswordPolicyById(id int64) (*db_models.CliPasswordPolicy, error) {
	return store.GetSingleton().GetPasswordPolicyById(id)
}

func GetPasswordPolicyByName(name string) (*db_models.CliPasswordPolicy, error) {
	return store.GetSingleton().GetPasswordPolicyByName(name)
}

func ListPasswordPolicies() ([]*db_models.CliPasswordPolicy, error) {
	return store.GetSingleton().ListPasswordPolicies()
}

func UpdatePasswordPolicy(p *db_models.CliPasswordPolicy) error {
	return store.GetSingleton().UpdatePasswordPolicy(p)
}

func DeletePasswordPolicyById(id int64) error {
	return store.GetSingleton().DeletePasswordPolicyById(id)
}

// AssignPasswordPolicyToGroup sets cli_group.password_policy_id.
func AssignPasswordPolicyToGroup(groupID int64, policyID *int64) error {
	sto := store.GetSingleton()
	g, err := sto.GetGroupById(groupID)
	if err != nil {
		return err
	}
	if g == nil {
		return errors.New("group not found")
	}
	g.PasswordPolicyID = policyID
	return sto.UpdateGroup(g)
}

// ── Effective policy resolution ──

// GetEffectivePasswordPolicy returns the strict-est combination of every
// policy attached to every group the user belongs to. Returns nil when no
// group has a policy — the caller should treat that as "no constraints".
func GetEffectivePasswordPolicy(userID int64) (*db_models.CliPasswordPolicy, error) {
	sto := store.GetSingleton()
	gms, err := sto.GetAllGroupsOfUser(userID)
	if err != nil {
		return nil, err
	}
	var effective *db_models.CliPasswordPolicy
	seen := map[int64]bool{}
	for _, gm := range gms {
		if gm == nil || seen[gm.GroupID] {
			continue
		}
		seen[gm.GroupID] = true
		g, err := sto.GetGroupById(gm.GroupID)
		if err != nil || g == nil || g.PasswordPolicyID == nil {
			continue
		}
		p, err := sto.GetPasswordPolicyById(*g.PasswordPolicyID)
		if err != nil || p == nil {
			continue
		}
		if effective == nil {
			c := *p
			effective = &c
			continue
		}
		effective = mergeStrictest(effective, p)
	}
	return effective, nil
}

// mergeStrictest returns a new policy holding the strict-est value for each
// field. "Strict" varies: for lengths/history/lockout_minutes → larger wins;
// for max_age_days / max_login_failure → smaller non-zero wins; any require_*
// true wins.
func mergeStrictest(a, b *db_models.CliPasswordPolicy) *db_models.CliPasswordPolicy {
	out := *a
	out.MinLength = maxI32(a.MinLength, b.MinLength)
	out.RequireUppercase = a.RequireUppercase || b.RequireUppercase
	out.RequireLowercase = a.RequireLowercase || b.RequireLowercase
	out.RequireDigit = a.RequireDigit || b.RequireDigit
	out.RequireSpecial = a.RequireSpecial || b.RequireSpecial
	out.HistoryCount = maxI32(a.HistoryCount, b.HistoryCount)
	out.LockoutMinutes = maxI32(a.LockoutMinutes, b.LockoutMinutes)
	out.MaxAgeDays = minNonZeroI32(a.MaxAgeDays, b.MaxAgeDays)
	out.MaxLoginFailure = minNonZeroI32(a.MaxLoginFailure, b.MaxLoginFailure)
	return &out
}

func maxI32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}
func minNonZeroI32(a, b int32) int32 {
	if a == 0 {
		return b
	}
	if b == 0 {
		return a
	}
	if a < b {
		return a
	}
	return b
}

// ── Password validation ──

// ValidatePasswordAgainstPolicy checks a new plaintext password against the
// complexity rules in p. Returns the first failure in a user-facing form.
func ValidatePasswordAgainstPolicy(p *db_models.CliPasswordPolicy, plaintext string) error {
	if p == nil {
		return nil
	}
	if p.MinLength > 0 && int32(len(plaintext)) < p.MinLength {
		return fmt.Errorf("password must be at least %d characters", p.MinLength)
	}
	hasUpper, hasLower, hasDigit, hasSpecial := false, false, false, false
	for _, r := range plaintext {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}
	if p.RequireUppercase && !hasUpper {
		return errors.New("password must contain an uppercase letter")
	}
	if p.RequireLowercase && !hasLower {
		return errors.New("password must contain a lowercase letter")
	}
	if p.RequireDigit && !hasDigit {
		return errors.New("password must contain a digit")
	}
	if p.RequireSpecial && !hasSpecial {
		return errors.New("password must contain a special character")
	}
	return nil
}

// ── Password history ──

// IsPasswordReused checks if the hash of newHash matches any of the last
// historyCount entries for the user.
func IsPasswordReused(userID int64, newHash string, historyCount int32) (bool, error) {
	if historyCount <= 0 {
		return false, nil
	}
	past, err := store.GetSingleton().GetRecentPasswordHistory(userID, int(historyCount))
	if err != nil {
		return false, err
	}
	for _, h := range past {
		if h.PasswordHash == newHash {
			return true, nil
		}
	}
	return false, nil
}

// RecordPasswordChange appends the old hash to history and prunes anything
// older than historyCount. Call BEFORE overwriting tbl_account.password with
// the new value. historyCount=0 means "don't keep any history" — we still
// prune to keep storage bounded.
func RecordPasswordChange(userID int64, oldHash string, historyCount int32) error {
	sto := store.GetSingleton()
	if err := sto.AppendPasswordHistory(&db_models.CliPasswordHistory{
		UserID: userID, PasswordHash: oldHash, ChangedAt: time.Now().UTC(),
	}); err != nil {
		return err
	}
	return sto.PrunePasswordHistory(userID, int(historyCount))
}

// ── Account lockout ──

// LockoutStatus describes whether the user is currently locked out and, if
// so, when the lockout expires.
type LockoutStatus struct {
	Locked     bool
	Until      time.Time
	RemainingS int64
}

// IsAccountLocked checks whether the user's login_failure_count has reached
// the effective max_login_failure AND we're still within lockout_minutes of
// locked_time. Returns (Locked=false) if no policy, no policy lockout, or
// the lockout has expired.
func IsAccountLocked(user *db_models.TblAccount, policy *db_models.CliPasswordPolicy) LockoutStatus {
	if user == nil || policy == nil || policy.MaxLoginFailure <= 0 || policy.LockoutMinutes <= 0 {
		return LockoutStatus{}
	}
	if user.LoginFailureCount < policy.MaxLoginFailure {
		return LockoutStatus{}
	}
	if user.LockedTime.IsZero() {
		return LockoutStatus{}
	}
	until := user.LockedTime.Add(time.Duration(policy.LockoutMinutes) * time.Minute)
	now := time.Now()
	if now.After(until) {
		return LockoutStatus{}
	}
	return LockoutStatus{Locked: true, Until: until, RemainingS: int64(until.Sub(now).Seconds())}
}

// ComputePasswordExpiry returns the PasswordExpiresAt for a fresh password
// given the effective policy. nil means "never expires".
func ComputePasswordExpiry(policy *db_models.CliPasswordPolicy, now time.Time) *time.Time {
	if policy == nil || policy.MaxAgeDays <= 0 {
		return nil
	}
	t := now.Add(time.Duration(policy.MaxAgeDays) * 24 * time.Hour)
	return &t
}

// IsPasswordExpired reports whether the stored PasswordExpiresAt is in the past.
func IsPasswordExpired(user *db_models.TblAccount) bool {
	if user == nil || user.PasswordExpiresAt == nil {
		return false
	}
	return time.Now().After(*user.PasswordExpiresAt)
}

// ─────────────────────────────────────────────────────────────────────────
// Mgt permission
// ─────────────────────────────────────────────────────────────────────────

func CreateMgtPermission(p *db_models.CliGroupMgtPermission) error {
	if p.GroupID == 0 {
		return errors.New("group_id is required")
	}
	if strings.TrimSpace(p.Resource) == "" || strings.TrimSpace(p.Action) == "" {
		return errors.New("resource and action are required")
	}
	return store.GetSingleton().CreateMgtPermission(p)
}

func ListMgtPermissions(groupID int64) ([]*db_models.CliGroupMgtPermission, error) {
	return store.GetSingleton().ListMgtPermissions(groupID)
}

func DeleteMgtPermissionById(id int64) error {
	return store.GetSingleton().DeleteMgtPermissionById(id)
}

// UserHasMgtPermission checks whether any group the user belongs to grants
// (resource, action). "*" in either slot is a wildcard.
func UserHasMgtPermission(userID int64, resource, action string) (bool, error) {
	sto := store.GetSingleton()
	gms, err := sto.GetAllGroupsOfUser(userID)
	if err != nil {
		return false, err
	}
	for _, gm := range gms {
		if gm == nil {
			continue
		}
		perms, err := sto.ListMgtPermissions(gm.GroupID)
		if err != nil {
			return false, err
		}
		for _, p := range perms {
			if mgtResourceMatches(p.Resource, resource) && mgtActionMatches(p.Action, action) {
				return true, nil
			}
		}
	}
	return false, nil
}

func mgtResourceMatches(ruleResource, want string) bool {
	return ruleResource == db_models.MgtResourceAny || ruleResource == want
}
func mgtActionMatches(ruleAction, want string) bool {
	return ruleAction == db_models.MgtActionAny || ruleAction == want
}
