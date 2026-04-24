package service

import (
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/token"
	"golang.org/x/crypto/bcrypt"
)

// Authenticate runs the full login gate and returns a signed JWT on success:
//  1. access-list blacklist/whitelist (username + client IP + email)
//  2. user must exist, be enabled, not locked, not past password expiry
//  3. bcrypt password check — on miss, increment failure count and lock if
//     the policy threshold was hit
//  4. reset failure count, bump last_login_at, write a login_history row,
//     mint the token
func Authenticate(username, password, clientIP string) (string, error) {
	u, err := store.GetSingleton().GetUserByUsername(username)
	if err != nil {
		return "", err
	}
	email := ""
	if u != nil {
		email = u.Email
	}
	if ok, reason := EvaluateAccessList(username, clientIP, email); !ok {
		logger.Logger.WithField("user", username).Warnf("auth: access-list denied: %s", reason)
		return "", ErrAccessDenied
	}
	if u == nil {
		return "", ErrInvalidPassword
	}
	if !u.IsEnabled {
		return "", ErrAccountDisabled
	}

	policy, _ := EffectivePasswordPolicy()
	if u.LockedAt != nil && policy.LockoutMinutes > 0 {
		unlock := u.LockedAt.Add(time.Duration(policy.LockoutMinutes) * time.Minute)
		if time.Now().UTC().Before(unlock) {
			return "", ErrAccountLocked
		}
		u.LockedAt = nil
		u.LoginFailureCount = 0
	}
	if u.PasswordExpiresAt != nil && time.Now().UTC().After(*u.PasswordExpiresAt) {
		return "", ErrPasswordExpired
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		u.LoginFailureCount++
		if policy.MaxLoginFailure > 0 && u.LoginFailureCount >= policy.MaxLoginFailure {
			now := time.Now().UTC()
			u.LockedAt = &now
		}
		u.UpdatedAt = time.Now().UTC()
		_ = store.GetSingleton().UpdateUser(u)
		return "", ErrInvalidPassword
	}

	now := time.Now().UTC()
	u.LoginFailureCount = 0
	u.LockedAt = nil
	u.LastLoginAt = &now
	u.UpdatedAt = now
	_ = store.GetSingleton().UpdateUser(u)
	_ = store.GetSingleton().UpdateLoginHistory(u.Username, clientIP, now)
	return token.CreateToken(u.Username)
}
