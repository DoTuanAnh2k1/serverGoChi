package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/internal/handler/middleware"
	"github.com/DoTuanAnh2k1/serverGoChi/internal/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/internal/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/internal/service"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

// saveHistoryReq là body của POST /aa/history/save.
type saveHistoryReq struct {
	CmdName        string `json:"cmd_name"`
	NeName         string `json:"ne_name"`
	NeIP           string `json:"ne_ip"`
	NeID           int32  `json:"ne_id"`
	Scope          string `json:"scope"`
	Result         string `json:"result"`
	InputType      string `json:"input_type"`
	Session        string `json:"session"`
	BatchID        string `json:"batch_id"`
	TimeToComplete int64  `json:"time_to_complete"`
}

// HandlerSaveHistory lưu một bản ghi lịch sử lệnh vào database.
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

	userCtx, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User)
	if !ok {
		response.InternalError(w, "cannot retrieve user from context")
		return
	}

	now := time.Now()
	record := db_models.CliOperationHistory{
		CmdName:        req.CmdName,
		NeName:         req.NeName,
		NeIP:           req.NeIP,
		NeID:           req.NeID,
		Scope:          req.Scope,
		Result:         req.Result,
		InputType:      req.InputType,
		Session:        req.Session,
		BatchID:        req.BatchID,
		TimeToComplete: req.TimeToComplete,
		Account:        userCtx.Username,
		CreatedDate:    now,
		ExecutedTime:   now,
	}

	if err := service.SaveHistoryCommand(record); err != nil {
		logger.Logger.Error("save history: db error: ", err)
		response.InternalError(w, "failed to save history")
		return
	}

	logger.Logger.Infof("save history: saved cmd=%q ne=%q by=%q", req.CmdName, req.NeName, userCtx.Username)
	response.Write(w, http.StatusCreated, record)
}
