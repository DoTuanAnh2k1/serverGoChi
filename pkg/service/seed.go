package service

import (
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
)

// SeedFirstBoot is idempotent — it creates a default password policy and an
// initial admin-equivalent user ("admin" / "admin") only if the user table is
// empty. v2 has no role concept so the seeded user has no special power
// beyond any other account; RBAC is layered via groups.
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
	hash, err := HashPassword("admin")
	if err != nil {
		logger.Logger.Warnf("seed: hash default password: %v", err)
		return
	}
	now := time.Now().UTC()
	u := &db_models.User{
		Username:     "admin",
		PasswordHash: hash,
		FullName:     "Default Admin",
		IsEnabled:    true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.CreateUser(u); err != nil {
		logger.Logger.Warnf("seed: create default user: %v", err)
		return
	}
	_ = s.AppendPasswordHistory(&db_models.PasswordHistory{
		UserID: u.ID, PasswordHash: hash, ChangedAt: now,
	})
	logger.Logger.Info("seed: created default user 'admin' (password 'admin') — change immediately")
}
