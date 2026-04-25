package service

import (
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
)

// SeedFirstBoot is idempotent — it creates a default password policy and
// initial users only if the user table is empty.
func SeedFirstBoot() {
	s := store.GetSingleton()

	if _, err := EffectivePasswordPolicy(); err == nil {
		existing, _ := s.GetPasswordPolicy()
		if existing == nil {
			if err := UpsertPasswordPolicy(&DefaultPasswordPolicy); err != nil {
				logger.Logger.Warnf("seed: upsert default policy: %v", err)
			}
		}
	}

	users, err := s.ListUsers()
	if err != nil {
		logger.Logger.Warnf("seed: list users: %v", err)
		return
	}
	if len(users) > 0 {
		return
	}
	seeds := []struct {
		username string
		password string
		fullName string
		role     string
	}{
		{"admin", "admin", "Default Admin", db_models.RoleSuperAdmin},
		{"anhdt195", "123", "Anh Do Tuan", db_models.RoleSuperAdmin},
	}
	now := time.Now().UTC()
	for _, sd := range seeds {
		hash, err := HashPassword(sd.password)
		if err != nil {
			logger.Logger.Warnf("seed: hash password for %s: %v", sd.username, err)
			continue
		}
		u := &db_models.User{
			Username:     sd.username,
			PasswordHash: hash,
			FullName:     sd.fullName,
			Role:         sd.role,
			IsEnabled:    true,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := s.CreateUser(u); err != nil {
			logger.Logger.Warnf("seed: create user %s: %v", sd.username, err)
			continue
		}
		_ = s.AppendPasswordHistory(&db_models.PasswordHistory{
			UserID: u.ID, PasswordHash: hash, ChangedAt: now,
		})
		logger.Logger.Infof("seed: created user '%s' (password '%s') — change immediately", sd.username, sd.password)
	}
}
