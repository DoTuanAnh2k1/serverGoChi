package service

import (
	"errors"
	"unicode"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
)

// DefaultPasswordPolicy is returned when the policy row is missing. It's also
// the seed value for first-run installs (see seed.go).
var DefaultPasswordPolicy = db_models.PasswordPolicy{
	ID:                1,
	MinLength:         8,
	MaxAgeDays:        0,
	RequireUppercase:  false,
	RequireLowercase:  false,
	RequireDigit:      false,
	RequireSpecial:    false,
	HistoryCount:      0,
	MaxLoginFailure:   0,
	LockoutMinutes:    0,
}

func EffectivePasswordPolicy() (db_models.PasswordPolicy, error) {
	p, err := store.GetSingleton().GetPasswordPolicy()
	if err != nil {
		return DefaultPasswordPolicy, err
	}
	if p == nil {
		return DefaultPasswordPolicy, nil
	}
	return *p, nil
}

func UpsertPasswordPolicy(p *db_models.PasswordPolicy) error {
	if p.MinLength < 1 {
		p.MinLength = 1
	}
	return store.GetSingleton().UpsertPasswordPolicy(p)
}

func ValidatePasswordAgainstPolicy(pw string) error {
	p, _ := EffectivePasswordPolicy()
	if int32(len(pw)) < p.MinLength {
		return errors.New("password: shorter than min length")
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range pw {
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
		return errors.New("password: must contain an uppercase letter")
	}
	if p.RequireLowercase && !hasLower {
		return errors.New("password: must contain a lowercase letter")
	}
	if p.RequireDigit && !hasDigit {
		return errors.New("password: must contain a digit")
	}
	if p.RequireSpecial && !hasSpecial {
		return errors.New("password: must contain a special character")
	}
	return nil
}
