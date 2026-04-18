package service

import (
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/bcrypt"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
)

const (
	// SeedUsername is the system default user that cannot be disabled or have its
	// permission changed via API. Exported so handlers can enforce the guard.
	SeedUsername = "anhdt195"
	seedPassword = "123"
)

// SeedDefaultUser tạo user mặc định nếu chưa tồn tại.
// Được gọi một lần khi app khởi động — idempotent (không làm gì nếu user đã có).
func SeedDefaultUser() {
	existing, err := GetUserByUserName(SeedUsername)
	if err != nil {
		logger.Logger.Errorf("seed: check user %q: %v — will retry", SeedUsername, err)
	}
	if existing != nil {
		logger.Logger.Infof("seed: user %q already exists, skip", SeedUsername)
		return
	}

	now := time.Now()
	user := &db_models.TblAccount{
		AccountName:    SeedUsername,
		Password:       bcrypt.Encode(SeedUsername + seedPassword),
		AccountType:    0,
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
		// Seed user is critical — panic so container restarts and retries
		logger.Logger.Fatalf("seed: cannot create default user %q: %v", SeedUsername, err)
	}
	logger.Logger.Infof("seed: created default user %q", SeedUsername)
}

// seedNes lists the basic NEs that are created on first boot.
// Identified by system_type "5GC" and these exact names — idempotent.
var seedNes = []db_models.CliNe{
	{NeName: "AMF-01", SiteName: "HN-DC-01", SystemType: "5GC", Description: "Access and Mobility Management Function"},
	{NeName: "SMF-01", SiteName: "HN-DC-01", SystemType: "5GC", Description: "Session Management Function"},
	{NeName: "UPF-01", SiteName: "HN-DC-02", SystemType: "5GC", Description: "User Plane Function"},
}

// SeedDefaultNes creates the default NE set only when the 5GC NE table is empty.
// If any NE already exists (from prior seed or user imports), seeding is skipped
// entirely to avoid re-introducing default NEs an operator may have deleted.
func SeedDefaultNes() {
	existing, err := GetNeListBySystemType("5GC")
	if err != nil {
		logger.Logger.Errorf("seed: list existing NEs: %v", err)
		return
	}
	if len(existing) > 0 {
		logger.Logger.Infof("seed: %d NEs already exist, skip default NE seeding", len(existing))
		return
	}

	for i := range seedNes {
		ne := seedNes[i]
		if err := CreateNe(&ne); err != nil {
			logger.Logger.Errorf("seed: create NE %q: %v", ne.NeName, err)
			continue
		}
		logger.Logger.Infof("seed: created NE %q", ne.NeName)
	}
}
