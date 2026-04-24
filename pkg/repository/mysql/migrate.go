package mysql

import (
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"gorm.io/gorm"
)

// v1Tables are every table from the pre-v2 schema. On startup against a DB
// that still has them, we drop them so the v2 AutoMigrate lands in a clean
// namespace. The user explicitly chose the "drop + rebuild" migration path
// (no v1→v2 data-preserving migration is supported).
var v1Tables = []string{
	"cli_user_ne_mapping",
	"cli_user_group_mapping",
	"cli_group_ne_mapping",
	"cli_group_cmd_permission",
	"cli_group_mgt_permission",
	"cli_command_group_mapping",
	"cli_command_def",
	"cli_command_group",
	"cli_ne_profile",
	"cli_password_history",
	"cli_password_policy",
	"cli_role_user_mapping",
	"cli_role",
	"cli_group",
	"cli_ne_monitor",
	"cli_ne_slave",
	"cli_ne_config",
	"cli_ne",
	"cli_login_history",
	"cli_operation_history",
	"cli_config_backup",
	"tbl_account",
}

// dropLegacyTables is idempotent — Migrator().DropTable silently skips tables
// that don't exist, so fresh installs pay no cost. FK constraints are
// disabled for the drop so the order between parent/child tables doesn't
// matter; MySQL/MariaDB only.
func dropLegacyTables(db *gorm.DB) error {
	// Disable FK checks — on MySQL/MariaDB this is a session variable; on
	// Postgres the SET is harmless (ignored) because this file only runs
	// from the mysql package.
	if err := db.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
		logger.Logger.Warnf("mysql: legacy drop: disable FK checks: %v", err)
	}
	for _, t := range v1Tables {
		if db.Migrator().HasTable(t) {
			if err := db.Migrator().DropTable(t); err != nil {
				logger.Logger.Warnf("mysql: legacy drop %q: %v", t, err)
			} else {
				logger.Logger.Infof("mysql: dropped legacy table %q", t)
			}
		}
	}
	if err := db.Exec("SET FOREIGN_KEY_CHECKS = 1").Error; err != nil {
		logger.Logger.Warnf("mysql: legacy drop: re-enable FK checks: %v", err)
	}
	return nil
}
