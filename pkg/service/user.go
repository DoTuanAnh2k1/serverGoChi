package service

import (
	"errors"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("user: not found")
	ErrUserExists        = errors.New("user: username already taken")
	ErrInvalidPassword   = errors.New("user: invalid password")
	ErrAccountLocked     = errors.New("user: account locked")
	ErrAccountDisabled   = errors.New("user: account disabled")
	ErrAccessDenied      = errors.New("user: access denied by policy")
	ErrPasswordReuse     = errors.New("user: cannot reuse a recent password")
	ErrPasswordExpired   = errors.New("user: password expired — must reset")
)

func HashPassword(pw string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

func CreateUser(u *db_models.User, plainPassword string) error {
	existing, err := store.GetSingleton().GetUserByUsername(u.Username)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrUserExists
	}
	if err := ValidatePasswordAgainstPolicy(plainPassword); err != nil {
		return err
	}
	hash, err := HashPassword(plainPassword)
	if err != nil {
		return err
	}
	u.PasswordHash = hash
	u.IsEnabled = true
	u.CreatedAt = time.Now().UTC()
	u.UpdatedAt = u.CreatedAt
	if err := store.GetSingleton().CreateUser(u); err != nil {
		return err
	}
	_ = store.GetSingleton().AppendPasswordHistory(&db_models.PasswordHistory{
		UserID: u.ID, PasswordHash: hash, ChangedAt: u.CreatedAt,
	})
	return nil
}

func GetUser(id int64) (*db_models.User, error) {
	u, err := store.GetSingleton().GetUserByID(id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func GetUserByUsername(username string) (*db_models.User, error) {
	u, err := store.GetSingleton().GetUserByUsername(username)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func ListUsers() ([]*db_models.User, error) {
	return store.GetSingleton().ListUsers()
}

// UpdateUserProfile patches the mutable profile fields. Password changes go
// through ChangePassword / AdminResetPassword which also run the policy.
func UpdateUserProfile(id int64, email, fullName, phone string, isEnabled *bool, role string) error {
	u, err := GetUser(id)
	if err != nil {
		return err
	}
	u.Email = email
	u.FullName = fullName
	u.Phone = phone
	if isEnabled != nil {
		u.IsEnabled = *isEnabled
	}
	if role != "" {
		u.Role = role
	}
	u.UpdatedAt = time.Now().UTC()
	return store.GetSingleton().UpdateUser(u)
}

func DeleteUser(id int64) error {
	return store.GetSingleton().DeleteUserByID(id)
}

// ChangePassword enforces the policy and history count. Admin resets bypass
// the old-password check via AdminResetPassword.
func ChangePassword(userID int64, oldPassword, newPassword string) error {
	u, err := GetUser(userID)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(oldPassword)) != nil {
		return ErrInvalidPassword
	}
	return replaceUserPassword(u, newPassword)
}

// AdminResetPassword is used by administrators; it skips the old-password
// check but still enforces the policy and appends to history.
func AdminResetPassword(userID int64, newPassword string) error {
	u, err := GetUser(userID)
	if err != nil {
		return err
	}
	return replaceUserPassword(u, newPassword)
}

func replaceUserPassword(u *db_models.User, newPassword string) error {
	if err := ValidatePasswordAgainstPolicy(newPassword); err != nil {
		return err
	}
	policy, _ := EffectivePasswordPolicy()
	if policy.HistoryCount > 0 {
		recent, _ := store.GetSingleton().GetRecentPasswordHistory(u.ID, int(policy.HistoryCount))
		for _, ph := range recent {
			if bcrypt.CompareHashAndPassword([]byte(ph.PasswordHash), []byte(newPassword)) == nil {
				return ErrPasswordReuse
			}
		}
	}
	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	u.PasswordHash = hash
	u.LoginFailureCount = 0
	u.LockedAt = nil
	u.UpdatedAt = time.Now().UTC()
	if policy.MaxAgeDays > 0 {
		exp := u.UpdatedAt.Add(time.Duration(policy.MaxAgeDays) * 24 * time.Hour)
		u.PasswordExpiresAt = &exp
	} else {
		u.PasswordExpiresAt = nil
	}
	if err := store.GetSingleton().UpdateUser(u); err != nil {
		return err
	}
	_ = store.GetSingleton().AppendPasswordHistory(&db_models.PasswordHistory{
		UserID: u.ID, PasswordHash: hash, ChangedAt: u.UpdatedAt,
	})
	if policy.HistoryCount > 0 {
		_ = store.GetSingleton().PrunePasswordHistory(u.ID, int(policy.HistoryCount))
	}
	return nil
}
