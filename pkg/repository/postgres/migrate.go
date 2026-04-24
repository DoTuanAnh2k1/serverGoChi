package postgres

import (
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"gorm.io/gorm"
)

// v1Tables — same list as the mysql driver. Kept in sync manually; the two
// SQL dialects share the pre-v2 schema identically.
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

// dropLegacyTables is idempotent. Postgres doesn't need the FK-check toggle
// MySQL uses because `DROP TABLE ... CASCADE` (what GORM issues) already
// handles dependent constraints.
func dropLegacyTables(db *gorm.DB) error {
	for _, t := range v1Tables {
		if db.Migrator().HasTable(t) {
			if err := db.Migrator().DropTable(t); err != nil {
				logger.Logger.Warnf("postgres: legacy drop %q: %v", t, err)
			} else {
				logger.Logger.Infof("postgres: dropped legacy table %q", t)
			}
		}
	}
	return nil
}
