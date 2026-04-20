package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
)

// adminUserResp is TblAccount without password for frontend display.
type adminUserResp struct {
	AccountID         int64  `json:"account_id"`
	AccountName       string `json:"account_name"`
	FullName          string `json:"full_name"`
	Email             string `json:"email"`
	Address           string `json:"address"`
	PhoneNumber       string `json:"phone_number"`
	AccountType       int32  `json:"account_type"`
	Description       string `json:"description"`
	IsEnable          bool   `json:"is_enable"`
	Status            bool   `json:"status"`
	CreatedBy         string `json:"created_by"`
	LoginFailureCount int32  `json:"login_failure_count"`
}

// HandlerAdminUserList returns all users with full fields (no password) for frontend.
func HandlerAdminUserList(w http.ResponseWriter, r *http.Request) {
	users, err := service.GetAllUser()
	if err != nil {
		logger.Logger.Error("admin/user/list: ", err)
		response.InternalError(w, "failed to list users")
		return
	}
	users = service.FilterOutSuperAdmins(users)
	var result []adminUserResp
	for _, u := range users {
		result = append(result, adminUserResp{
			AccountID:         u.AccountID,
			AccountName:       u.AccountName,
			FullName:          u.FullName,
			Email:             u.Email,
			Address:           u.Address,
			PhoneNumber:       u.PhoneNumber,
			AccountType:       u.AccountType,
			Description:       u.Description,
			IsEnable:          u.IsEnable,
			Status:            u.Status,
			CreatedBy:         u.CreatedBy,
			LoginFailureCount: u.LoginFailureCount,
		})
	}
	if result == nil {
		result = []adminUserResp{}
	}
	response.Write(w, http.StatusOK, result)
}

// adminUserFullResp extends adminUserResp with the caller-facing role and the
// effective list of NEs the user may connect to (direct + via groups, deduped).
type adminUserFullResp struct {
	adminUserResp
	Role string      `json:"role"`
	Nes  []userNeRef `json:"nes"`
}

type userNeRef struct {
	ID        int64  `json:"id"`
	NeName    string `json:"ne_name"`
	SiteName  string `json:"site_name"`
	Namespace string `json:"namespace"`
}

// HandlerAdminUserFullList returns every non-SuperAdmin user together with
// their role and the union of NEs reachable directly or via group membership.
func HandlerAdminUserFullList(w http.ResponseWriter, r *http.Request) {
	users, err := service.GetAllUser()
	if err != nil {
		logger.Logger.Error("admin/user/full: ", err)
		response.InternalError(w, "failed to list users")
		return
	}
	users = service.FilterOutSuperAdmins(users)
	result := make([]adminUserFullResp, 0, len(users))
	for _, u := range users {
		neIds, err := service.GetAllNeIdsOfUser(u.AccountID)
		if err != nil {
			logger.Logger.Errorf("admin/user/full: ne ids for %s: %v", u.AccountName, err)
		}
		nes := make([]userNeRef, 0, len(neIds))
		for _, id := range neIds {
			ne, err := service.GetNeByNeId(id)
			if err != nil || ne == nil {
				continue
			}
			nes = append(nes, userNeRef{ID: ne.ID, NeName: ne.NeName, SiteName: ne.SiteName, Namespace: ne.Namespace})
		}
		result = append(result, adminUserFullResp{
			adminUserResp: adminUserResp{
				AccountID:         u.AccountID,
				AccountName:       u.AccountName,
				FullName:          u.FullName,
				Email:             u.Email,
				Address:           u.Address,
				PhoneNumber:       u.PhoneNumber,
				AccountType:       u.AccountType,
				Description:       u.Description,
				IsEnable:          u.IsEnable,
				Status:            u.Status,
				CreatedBy:         u.CreatedBy,
				LoginFailureCount: u.LoginFailureCount,
			},
			Role: service.GetPermissionByUser(u),
			Nes:  nes,
		})
	}
	response.Write(w, http.StatusOK, result)
}

// HandlerAdminUserUpdate updates a user's metadata fields (no password change via this endpoint).
//
// Authorization:
//   - Admins (account_type 0/1, Permission "admin") may edit any non-SuperAdmin user.
//   - Normal users may edit ONLY their own account, and cannot change account_type.
func HandlerAdminUserUpdate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AccountName string  `json:"account_name"`
		FullName    string  `json:"full_name"`
		Email       string  `json:"email"`
		PhoneNumber string  `json:"phone_number"`
		Address     string  `json:"address"`
		Description string  `json:"description"`
		AccountType int32   `json:"account_type"`
		GroupIDs    []int64 `json:"group_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.AccountName == "" {
		response.Write(w, http.StatusBadRequest, "account_name is required")
		return
	}
	actor := mustUser(r)
	isAdmin := strings.EqualFold(actor.Permission, "admin")
	if !isAdmin && actor.Username != req.AccountName {
		response.Write(w, http.StatusForbidden, "cannot modify another user")
		return
	}
	u, err := service.GetUserByUserName(req.AccountName)
	if err != nil {
		logger.Logger.Errorf("admin/user/update: get user: %v", err)
		response.InternalError(w, "failed to retrieve user")
		return
	}
	if u == nil {
		response.NotFound(w, "user not found")
		return
	}
	// Guard: SuperAdmin accounts (account_type=0) cannot be edited by anyone.
	if u.AccountType == 0 {
		response.Write(w, http.StatusForbidden, "cannot modify SuperAdmin account")
		return
	}
	// Non-admin self-edit: preserve account_type (don't let users promote/demote themselves).
	if !isAdmin {
		req.AccountType = u.AccountType
	}
	// Guard: cannot elevate a user to SuperAdmin via this endpoint.
	if req.AccountType == 0 {
		response.Write(w, http.StatusForbidden, "cannot set account_type to 0 (SuperAdmin)")
		return
	}
	// Build a shadow account for validation so we check the POST-UPDATE state.
	candidate := *u
	candidate.FullName = req.FullName
	candidate.Email = req.Email
	candidate.PhoneNumber = req.PhoneNumber
	candidate.Address = req.Address
	candidate.Description = req.Description
	candidate.AccountType = req.AccountType
	if err := service.ValidateUserCommon(&candidate, u.AccountName); err != nil {
		response.Write(w, http.StatusBadRequest, err.Error())
		return
	}
	// For admin targets enforce full_name + phone + ≥1 group.
	// Determine the groups the user will end up with: provided list, or current list if none provided.
	effectiveGroupIds := req.GroupIDs
	if effectiveGroupIds == nil {
		existingGroups, _ := service.GetGroupsOfUser(u.AccountID)
		for _, m := range existingGroups {
			effectiveGroupIds = append(effectiveGroupIds, m.GroupID)
		}
	}
	if err := service.ValidateAdminUserFields(&candidate, effectiveGroupIds); err != nil {
		response.Write(w, http.StatusBadRequest, err.Error())
		return
	}

	u.FullName = req.FullName
	u.Email = req.Email
	u.PhoneNumber = req.PhoneNumber
	u.Address = req.Address
	u.Description = req.Description
	u.AccountType = req.AccountType
	u.UpdatedDate = time.Now()
	if err := service.UpdateUser(u); err != nil {
		logger.Logger.Errorf("admin/user/update: update: %v", err)
		response.InternalError(w, "failed to update user")
		return
	}
	// If caller sent group_ids, replace the user's group membership with that set.
	if req.GroupIDs != nil {
		if err := replaceUserGroups(u.AccountID, req.GroupIDs); err != nil {
			logger.Logger.Errorf("admin/user/update: replace groups: %v", err)
			response.InternalError(w, "failed to update user groups")
			return
		}
	}
	saveHistory(opHistory("admin user update", req.AccountName, actor.Username), "success")
	response.Success(w, "user updated")
}

// replaceUserGroups sets the user's group membership to exactly the given ids.
func replaceUserGroups(userID int64, groupIDs []int64) error {
	current, err := service.GetGroupsOfUser(userID)
	if err != nil {
		return err
	}
	wanted := make(map[int64]struct{}, len(groupIDs))
	for _, id := range groupIDs {
		wanted[id] = struct{}{}
	}
	have := make(map[int64]struct{}, len(current))
	for _, m := range current {
		have[m.GroupID] = struct{}{}
	}
	for gid := range have {
		if _, ok := wanted[gid]; !ok {
			if err := service.UnassignUserFromGroup(userID, gid); err != nil {
				return err
			}
		}
	}
	for gid := range wanted {
		if _, ok := have[gid]; !ok {
			if err := service.AssignUserToGroup(userID, gid); err != nil {
				return err
			}
		}
	}
	return nil
}
