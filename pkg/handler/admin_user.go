package handler

import (
	"net/http"

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
