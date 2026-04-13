package mongodb

import (
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

// mAccount mirrors db_models.TblAccount with bson tags.
type mAccount struct {
	AccountID         int64     `bson:"account_id"`
	AccountName       string    `bson:"account_name"`
	Password          string    `bson:"password"`
	FullName          string    `bson:"full_name"`
	Email             string    `bson:"email"`
	Address           string    `bson:"address"`
	PhoneNumber       string    `bson:"phone_number"`
	LoginFailureCount int32     `bson:"login_failure_count"`
	ForceChangePass   bool      `bson:"force_change_pass"`
	CreatedDate       time.Time `bson:"created_date"`
	UpdatedDate       time.Time `bson:"updated_date"`
	LastLoginTime     time.Time `bson:"last_login_time"`
	LastChangePass    time.Time `bson:"last_change_pass"`
	Avatar            string    `bson:"avatar"`
	Status            bool      `bson:"status"`
	DefaultDashboard  int32     `bson:"default_dashboard"`
	AccountType       int32     `bson:"account_type"`
	AutoPassword      bool      `bson:"auto_password"`
	Description       string    `bson:"description"`
	IsEnable          bool      `bson:"is_enable"`
	CreatedBy         string    `bson:"created_by"`
	UpdatedBy         string    `bson:"updated_by"`
	LockedTime        time.Time `bson:"locked_time"`
	OnlyAD            bool      `bson:"onlyAD"`
}

func toMAccount(a *db_models.TblAccount) *mAccount {
	return &mAccount{
		AccountID: a.AccountID, AccountName: a.AccountName, Password: a.Password,
		FullName: a.FullName, Email: a.Email, Address: a.Address, PhoneNumber: a.PhoneNumber,
		LoginFailureCount: a.LoginFailureCount, ForceChangePass: a.ForceChangePass,
		CreatedDate: a.CreatedDate, UpdatedDate: a.UpdatedDate, LastLoginTime: a.LastLoginTime,
		LastChangePass: a.LastChangePass, Avatar: a.Avatar, Status: a.Status,
		DefaultDashboard: a.DefaultDashboard, AccountType: a.AccountType, AutoPassword: a.AutoPassword,
		Description: a.Description, IsEnable: a.IsEnable, CreatedBy: a.CreatedBy, UpdatedBy: a.UpdatedBy,
		LockedTime: a.LockedTime, OnlyAD: a.OnlyAD,
	}
}

func fromMAccount(m *mAccount) *db_models.TblAccount {
	return &db_models.TblAccount{
		AccountID: m.AccountID, AccountName: m.AccountName, Password: m.Password,
		FullName: m.FullName, Email: m.Email, Address: m.Address, PhoneNumber: m.PhoneNumber,
		LoginFailureCount: m.LoginFailureCount, ForceChangePass: m.ForceChangePass,
		CreatedDate: m.CreatedDate, UpdatedDate: m.UpdatedDate, LastLoginTime: m.LastLoginTime,
		LastChangePass: m.LastChangePass, Avatar: m.Avatar, Status: m.Status,
		DefaultDashboard: m.DefaultDashboard, AccountType: m.AccountType, AutoPassword: m.AutoPassword,
		Description: m.Description, IsEnable: m.IsEnable, CreatedBy: m.CreatedBy, UpdatedBy: m.UpdatedBy,
		LockedTime: m.LockedTime, OnlyAD: m.OnlyAD,
	}
}

// mNe mirrors db_models.CliNe with new conf_* fields.
type mNe struct {
	ID                int64  `bson:"id"`
	NeName            string `bson:"ne_name"`
	Namespace         string `bson:"namespace"`
	SiteName          string `bson:"site_name"`
	SystemType        string `bson:"system_type"`
	Description       string `bson:"description"`
	CommandURL        string `bson:"command_url"`
	ConfMode          string `bson:"conf_mode"`
	ConfMasterIP      string `bson:"conf_master_ip"`
	ConfSlaveIP       string `bson:"conf_slave_ip"`
	ConfPortMasterSSH int32  `bson:"conf_port_master_ssh"`
	ConfPortSlaveSSH  int32  `bson:"conf_port_slave_ssh"`
	ConfPortMasterTCP int32  `bson:"conf_port_master_tcp"`
	ConfPortSlaveTCP  int32  `bson:"conf_port_slave_tcp"`
	ConfUsername      string `bson:"conf_username"`
	ConfPassword      string `bson:"conf_password"`
}

func toMNe(n *db_models.CliNe) *mNe {
	return &mNe{
		ID: n.ID, NeName: n.NeName, Namespace: n.Namespace, SiteName: n.SiteName,
		SystemType: n.SystemType, Description: n.Description, CommandURL: n.CommandURL,
		ConfMode: n.ConfMode, ConfMasterIP: n.ConfMasterIP, ConfSlaveIP: n.ConfSlaveIP,
		ConfPortMasterSSH: n.ConfPortMasterSSH, ConfPortSlaveSSH: n.ConfPortSlaveSSH,
		ConfPortMasterTCP: n.ConfPortMasterTCP, ConfPortSlaveTCP: n.ConfPortSlaveTCP,
		ConfUsername: n.ConfUsername, ConfPassword: n.ConfPassword,
	}
}

func fromMNe(m *mNe) *db_models.CliNe {
	return &db_models.CliNe{
		ID: m.ID, NeName: m.NeName, Namespace: m.Namespace, SiteName: m.SiteName,
		SystemType: m.SystemType, Description: m.Description, CommandURL: m.CommandURL,
		ConfMode: m.ConfMode, ConfMasterIP: m.ConfMasterIP, ConfSlaveIP: m.ConfSlaveIP,
		ConfPortMasterSSH: m.ConfPortMasterSSH, ConfPortSlaveSSH: m.ConfPortSlaveSSH,
		ConfPortMasterTCP: m.ConfPortMasterTCP, ConfPortSlaveTCP: m.ConfPortSlaveTCP,
		ConfUsername: m.ConfUsername, ConfPassword: m.ConfPassword,
	}
}

// mUserNeMapping mirrors db_models.CliUserNeMapping.
type mUserNeMapping struct {
	UserID  int64 `bson:"user_id"`
	TblNeID int64 `bson:"tbl_ne_id"`
}

func fromMUserNeMapping(m *mUserNeMapping) *db_models.CliUserNeMapping {
	return &db_models.CliUserNeMapping{UserID: m.UserID, TblNeID: m.TblNeID}
}

// mRole mirrors db_models.CliRole.
type mRole struct {
	RoleID      int64  `bson:"role_id"`
	IncludeType string `bson:"include_type"`
	NeType      string `bson:"ne_type"`
	Scope       string `bson:"scope"`
	Permission  string `bson:"permission"`
	Path        string `bson:"path"`
}

func toMRole(r *db_models.CliRole) *mRole {
	return &mRole{
		RoleID: r.RoleID, IncludeType: r.IncludeType, NeType: r.NeType,
		Scope: r.Scope, Permission: r.Permission, Path: r.Path,
	}
}

func fromMRole(m *mRole) *db_models.CliRole {
	return &db_models.CliRole{
		RoleID: m.RoleID, IncludeType: m.IncludeType, NeType: m.NeType,
		Scope: m.Scope, Permission: m.Permission, Path: m.Path,
	}
}

// mRoleUserMapping mirrors db_models.CliRoleUserMapping.
type mRoleUserMapping struct {
	UserID     int64  `bson:"user_id"`
	Permission string `bson:"permission"`
}

func fromMRoleUserMapping(m *mRoleUserMapping) *db_models.CliRoleUserMapping {
	return &db_models.CliRoleUserMapping{UserID: m.UserID, Permission: m.Permission}
}

// mConfigBackup mirrors db_models.CliConfigBackup.
type mConfigBackup struct {
	ID        int64     `bson:"id"`
	NeName    string    `bson:"ne_name"`
	NeIP      string    `bson:"ne_ip"`
	FilePath  string    `bson:"file_path"`
	Size      int64     `bson:"size"`
	CreatedAt time.Time `bson:"created_at"`
}

func toMConfigBackup(b *db_models.CliConfigBackup) *mConfigBackup {
	return &mConfigBackup{
		ID: b.ID, NeName: b.NeName, NeIP: b.NeIP,
		FilePath: b.FilePath, Size: b.Size, CreatedAt: b.CreatedAt,
	}
}

func fromMConfigBackup(m *mConfigBackup) *db_models.CliConfigBackup {
	return &db_models.CliConfigBackup{
		ID: m.ID, NeName: m.NeName, NeIP: m.NeIP,
		FilePath: m.FilePath, Size: m.Size, CreatedAt: m.CreatedAt,
	}
}

// mOperationHistory mirrors db_models.CliOperationHistory (simplified fields).
type mOperationHistory struct {
	ID           int32     `bson:"id"`
	Account      string    `bson:"account"`
	CmdName      string    `bson:"cmd_name"`
	NeName       string    `bson:"ne_name"`
	NeIP         string    `bson:"ne_ip"`
	IPAddress    string    `bson:"ip_address"`
	Scope        string    `bson:"scope"`
	Result       string    `bson:"result"`
	CreatedDate  time.Time `bson:"created_date"`
	ExecutedTime time.Time `bson:"executed_time"`
}

func toMOperationHistory(h db_models.CliOperationHistory) *mOperationHistory {
	return &mOperationHistory{
		ID: h.ID, Account: h.Account, CmdName: h.CmdName,
		NeName: h.NeName, NeIP: h.NeIP, IPAddress: h.IPAddress,
		Scope: h.Scope, Result: h.Result,
		CreatedDate: h.CreatedDate, ExecutedTime: h.ExecutedTime,
	}
}

func fromMOperationHistory(m *mOperationHistory) db_models.CliOperationHistory {
	return db_models.CliOperationHistory{
		ID: m.ID, Account: m.Account, CmdName: m.CmdName,
		NeName: m.NeName, NeIP: m.NeIP, IPAddress: m.IPAddress,
		Scope: m.Scope, Result: m.Result,
		CreatedDate: m.CreatedDate, ExecutedTime: m.ExecutedTime,
	}
}
