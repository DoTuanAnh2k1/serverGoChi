package service

import (
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/bcrypt"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
)

const (
	seedUsername = "anhdt195"
	seedPassword = "123"
)

// SeedDefaultUser tạo user mặc định nếu chưa tồn tại.
// Được gọi một lần khi app khởi động — idempotent (không làm gì nếu user đã có).
func SeedDefaultUser() {
	existing, err := GetUserByUserName(seedUsername)
	if err != nil {
		logger.Logger.Errorf("seed: check user %q: %v", seedUsername, err)
		return
	}
	if existing != nil {
		logger.Logger.Infof("seed: user %q already exists, skip", seedUsername)
		return
	}

	now := time.Now()
	user := &db_models.TblAccount{
		AccountName:    seedUsername,
		Password:       bcrypt.Encode(seedUsername + seedPassword),
		AccountType:    2,
		IsEnable:       true,
		Status:         true,
		CreatedBy:      "system",
		CreatedDate:    now,
		UpdatedDate:    now,
		LastLoginTime:  now,
		LastChangePass: now,
		LockedTime:     now,
	}

	if err := AddUser(user); err != nil {
		logger.Logger.Errorf("seed: create user %q: %v", seedUsername, err)
		return
	}
	logger.Logger.Infof("seed: created default user %q", seedUsername)
}
