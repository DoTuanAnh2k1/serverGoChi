package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/middleware"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

// saveHistoryReq is the body for POST /aa/history/save. The endpoint is
// unauthenticated, so the caller supplies `account` (the actor that ran the
// command) in the body. If omitted we fall back to the JWT actor when one is
// available, and finally to "unknown".
type saveHistoryReq struct {
	CmdName string `json:"cmd_name"`
	NeName  string `json:"ne_name"`
	NeIP    string `json:"ne_ip"`
	Scope   string `json:"scope"`
	Result  string `json:"result"`
	Account string `json:"account"`
}

// HandlerListHistory trả về lịch sử thao tác gần đây, có hỗ trợ lọc.
//
// Input : GET query params:
//         ?limit=<int>   — số bản ghi tối đa (1–500, mặc định 100)
//         ?scope=<string> — lọc theo scope (tuỳ chọn)
//         ?ne_name=<string> — lọc theo tên NE (tuỳ chọn)
//         ?account=<string> — lọc theo username (tuỳ chọn)
// Output: 200 [ ...CliOperationHistory ] (mảng rỗng nếu không có bản ghi)
//         500 nếu lỗi DB
// Flow  : parse limit từ query → nếu có scope/ne_name/account dùng GetRecentHistoryFiltered,
//         ngược lại dùng GetRecentHistory → trả danh sách
func HandlerListHistory(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}

	scope := r.URL.Query().Get("scope")
	neName := r.URL.Query().Get("ne_name")
	account := r.URL.Query().Get("account")

	var records []db_models.CliOperationHistory
	var err error
	if scope != "" || neName != "" || account != "" {
		records, err = service.GetRecentHistoryFiltered(limit, scope, neName, account)
	} else {
		records, err = service.GetRecentHistory(limit)
	}
	if err != nil {
		logger.Logger.Error("list history: db error: ", err)
		response.InternalError(w, "failed to retrieve history")
		return
	}

	if len(records) == 0 {
		response.Write(w, http.StatusOK, []struct{}{})
		return
	}

	response.Write(w, http.StatusOK, records)
}

// HandlerSaveHistory lưu một bản ghi lịch sử thao tác CLI vào database.
// Endpoint này KHÔNG yêu cầu JWT — caller tự cung cấp `account` trong body.
// Nếu có JWT trong context vẫn ưu tiên lấy từ JWT (thao tác từ UI đã đăng nhập).
//
// Input : POST body JSON { "cmd_name": string (bắt buộc), "ne_name": string (bắt buộc),
//         "ne_ip", "scope", "result", "account" }
// Output: 201 { ...CliOperationHistory } nếu lưu thành công
//         400 nếu thiếu cmd_name/ne_name hoặc body không hợp lệ
//         500 nếu lỗi DB
func HandlerSaveHistory(w http.ResponseWriter, r *http.Request) {
	var req saveHistoryReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Error("save history: invalid request body: ", err)
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.CmdName) == "" {
		response.Write(w, http.StatusBadRequest, "cmd_name is required")
		return
	}
	if strings.TrimSpace(req.NeName) == "" {
		response.Write(w, http.StatusBadRequest, "ne_name is required")
		return
	}

	// Resolve the actor: JWT (if present) > explicit account in body > "unknown".
	account := strings.TrimSpace(req.Account)
	if u, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User); ok && u != nil {
		account = u.Username
	}
	if account == "" {
		account = "unknown"
	}

	now := time.Now()
	record := db_models.CliOperationHistory{
		CmdName:      req.CmdName,
		NeName:       req.NeName,
		NeIP:         req.NeIP,
		Scope:        req.Scope,
		Result:       req.Result,
		Account:      account,
		CreatedDate:  now,
		ExecutedTime: now,
	}

	if err := service.SaveHistoryCommand(record); err != nil {
		logger.Logger.Error("save history: db error: ", err)
		response.InternalError(w, "failed to save history")
		return
	}

	logger.Logger.Infof("save history: saved cmd=%q ne=%q by=%q", req.CmdName, req.NeName, account)
	response.Write(w, http.StatusCreated, record)
}
