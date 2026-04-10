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

// mNe mirrors db_models.CliNe.
type mNe struct {
	ID          int64  `bson:"id"`
	Description string `bson:"description"`
	IPAddress   string `bson:"ip_address"`
	Name        string `bson:"name"`
	Namespace   string `bson:"namespace"`
	SiteName    string `bson:"site_name"`
	SystemType  string `bson:"system_type"`
	Port        int32  `bson:"port"`
	MetaData    string `bson:"meta_data"`
}

func toMNe(n *db_models.CliNe) *mNe {
	return &mNe{
		ID: n.ID, Description: n.Description, IPAddress: n.IPAddress, Name: n.Name,
		Namespace: n.Namespace, SiteName: n.SiteName, SystemType: n.SystemType,
		Port: n.Port, MetaData: n.MetaData,
	}
}

func fromMNe(m *mNe) *db_models.CliNe {
	return &db_models.CliNe{
		ID: m.ID, Description: m.Description, IPAddress: m.IPAddress, Name: m.Name,
		Namespace: m.Namespace, SiteName: m.SiteName, SystemType: m.SystemType,
		Port: m.Port, MetaData: m.MetaData,
	}
}

// mNeMonitor mirrors db_models.CliNeMonitor.
type mNeMonitor struct {
	NeID   int64  `bson:"ne_id"`
	NeName string `bson:"ne_name"`
	NeIP   string `bson:"ne_ip"`
}

func fromMNeMonitor(m *mNeMonitor) *db_models.CliNeMonitor {
	return &db_models.CliNeMonitor{NeID: m.NeID, NeName: m.NeName, NeIP: m.NeIP}
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

// mOperationHistory mirrors db_models.CliOperationHistory.
type mOperationHistory struct {
	ID             int32     `bson:"id"`
	IsOssType      int32     `bson:"is_oss_type"`
	CmdName        string    `bson:"cmd_name"`
	FunctionName   string    `bson:"function_name"`
	CreatedDate    time.Time `bson:"created_date"`
	ExecutedTime   time.Time `bson:"executed_time"`
	NeIP           string    `bson:"ne_ip"`
	NeName         string    `bson:"ne_name"`
	Scope          string    `bson:"scope"`
	Result         string    `bson:"result"`
	Account        string    `bson:"account"`
	IPAddress      string    `bson:"ip_address"`
	InputType      string    `bson:"input_type"`
	TimeToComplete int64     `bson:"time_to_complete"`
	NeID           int32     `bson:"ne_id"`
	Session        string    `bson:"session"`
	BatchID        string    `bson:"batch_id"`
}

func toMOperationHistory(h db_models.CliOperationHistory) *mOperationHistory {
	return &mOperationHistory{
		ID: h.ID, IsOssType: h.IsOssType, CmdName: h.CmdName, FunctionName: h.FunctionName,
		CreatedDate: h.CreatedDate, ExecutedTime: h.ExecutedTime, NeIP: h.NeIP, NeName: h.NeName,
		Scope: h.Scope, Result: h.Result, Account: h.Account, IPAddress: h.IPAddress,
		InputType: h.InputType, TimeToComplete: h.TimeToComplete, NeID: h.NeID,
		Session: h.Session, BatchID: h.BatchID,
	}
}

func fromMOperationHistory(m *mOperationHistory) db_models.CliOperationHistory {
	return db_models.CliOperationHistory{
		ID: m.ID, IsOssType: m.IsOssType, CmdName: m.CmdName, FunctionName: m.FunctionName,
		CreatedDate: m.CreatedDate, ExecutedTime: m.ExecutedTime, NeIP: m.NeIP, NeName: m.NeName,
		Scope: m.Scope, Result: m.Result, Account: m.Account, IPAddress: m.IPAddress,
		InputType: m.InputType, TimeToComplete: m.TimeToComplete, NeID: m.NeID,
		Session: m.Session, BatchID: m.BatchID,
	}
}
