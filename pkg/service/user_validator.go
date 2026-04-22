package service

import (
	"errors"
	"regexp"
	"strings"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
)

// E.164-style generic international phone: optional +, 7–15 digits total.
var phoneRegex = regexp.MustCompile(`^\+?[0-9]{7,15}$`)

// emailRegex: simple but practical — user@host.tld with common chars.
var emailRegex = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)

// IsAdminAccountType reports whether account_type designates an admin (0=SuperAdmin, 1=Admin).
func IsAdminAccountType(t int32) bool { return t == 0 || t == 1 }

// ValidatePhone returns nil if empty (allowed) or well-formed; otherwise an error.
func ValidatePhone(phone string) error {
	p := strings.TrimSpace(phone)
	if p == "" {
		return nil
	}
	if !phoneRegex.MatchString(p) {
		return errors.New("invalid phone number format (expected 7–15 digits, optional leading +)")
	}
	return nil
}

// ValidateEmailFormat returns nil if empty (allowed) or well-formed; otherwise an error.
func ValidateEmailFormat(email string) error {
	e := strings.TrimSpace(email)
	if e == "" {
		return nil
	}
	if !emailRegex.MatchString(e) {
		return errors.New("invalid email format")
	}
	return nil
}

// EnsureEmailUnique checks that no other ACTIVE user owns this email.
// excludeAccountName is skipped (used on update so a user keeps their own
// email). Empty email is always allowed. Disabled accounts (is_enable=false)
// are skipped so their email can be reused — either by re-enabling the same
// account (merge path) or by creating a fresh account with that email.
func EnsureEmailUnique(email, excludeAccountName string) error {
	e := strings.TrimSpace(email)
	if e == "" {
		return nil
	}
	all, err := GetAllUser()
	if err != nil {
		return err
	}
	for _, u := range all {
		if u == nil {
			continue
		}
		if !u.IsEnable {
			continue
		}
		if strings.EqualFold(u.AccountName, excludeAccountName) {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(u.Email), e) {
			return errors.New("email already in use")
		}
	}
	return nil
}

// ValidateAdminUserFields enforces the extra requirements for admin accounts:
//   - full_name required
//   - phone_number required AND well-formed
//   - user must belong to ≥1 group (at least one of groupIds resolves)
// Caller passes accountType so we know whether to enforce.
// groupIds is the set the caller intends to assign (admin path only).
func ValidateAdminUserFields(u *db_models.TblAccount, groupIds []int64) error {
	if !IsAdminAccountType(u.AccountType) {
		return nil
	}
	if strings.TrimSpace(u.FullName) == "" {
		return errors.New("full_name is required for admin users")
	}
	if strings.TrimSpace(u.PhoneNumber) == "" {
		return errors.New("phone_number is required for admin users")
	}
	if err := ValidatePhone(u.PhoneNumber); err != nil {
		return err
	}
	if len(groupIds) == 0 {
		return errors.New("at least one group is required for admin users")
	}
	s := store.GetSingleton()
	for _, gid := range groupIds {
		g, err := s.GetGroupById(gid)
		if err != nil {
			return err
		}
		if g == nil {
			return errors.New("group not found: one or more group_ids are invalid")
		}
	}
	return nil
}

// ValidateUserCommon enforces baseline rules applied to ALL create/edit APIs
// (except self-change-password): email format + email uniqueness + phone format
// (phone is only format-checked here; required-ness for admin is in ValidateAdminUserFields).
func ValidateUserCommon(u *db_models.TblAccount, excludeAccountName string) error {
	if err := ValidateEmailFormat(u.Email); err != nil {
		return err
	}
	if err := ValidatePhone(u.PhoneNumber); err != nil {
		return err
	}
	if err := EnsureEmailUnique(u.Email, excludeAccountName); err != nil {
		return err
	}
	return nil
}

// FilterOutSuperAdmins removes account_type==0 entries, returning a fresh slice.
func FilterOutSuperAdmins(users []*db_models.TblAccount) []*db_models.TblAccount {
	out := make([]*db_models.TblAccount, 0, len(users))
	for _, u := range users {
		if u == nil || u.AccountType == 0 {
			continue
		}
		out = append(out, u)
	}
	return out
}
