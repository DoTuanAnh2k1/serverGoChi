package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/middleware"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
)

// saveHistoryReq is the body for POST /aa/history/save. The endpoint is
// unauthenticated so the caller supplies `account` in the body; if a JWT is
// present we prefer that over the body value.
type saveHistoryReq struct {
	CmdText     string `json:"cmd_text"`
	NeNamespace string `json:"ne_namespace"`
	NeIP        string `json:"ne_ip"`
	Scope       string `json:"scope"`
	Result      string `json:"result"`
	Account     string `json:"account"`
}

func HandlerListHistory(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	scope := r.URL.Query().Get("scope")
	neNamespace := r.URL.Query().Get("ne_namespace")
	account := r.URL.Query().Get("account")

	var records []db_models.OperationHistory
	var err error
	if scope != "" || neNamespace != "" || account != "" {
		records, err = service.GetRecentHistoryFiltered(limit, scope, neNamespace, account)
	} else {
		records, err = service.GetRecentHistory(limit)
	}
	if err != nil {
		logger.Logger.Error("list history: db error: ", err)
		response.InternalError(w, "failed to retrieve history")
		return
	}
	if records == nil {
		records = []db_models.OperationHistory{}
	}
	response.Write(w, http.StatusOK, records)
}

func HandlerSaveHistory(w http.ResponseWriter, r *http.Request) {
	var req saveHistoryReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Error("save history: invalid request body: ", err)
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(req.CmdText) == "" {
		response.Write(w, http.StatusBadRequest, "cmd_text is required")
		return
	}

	account := strings.TrimSpace(req.Account)
	if u, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User); ok && u != nil {
		account = u.Username
	}
	if account == "" {
		account = "unknown"
	}

	now := time.Now().UTC()
	record := db_models.OperationHistory{
		CmdText:      req.CmdText,
		NeNamespace:  req.NeNamespace,
		NeIP:         req.NeIP,
		Scope:        req.Scope,
		Result:       req.Result,
		Account:      account,
		CreatedDate:  now,
		ExecutedTime: now,
	}
	if err := service.SaveHistory(record); err != nil {
		response.InternalError(w, "failed to save history")
		return
	}
	response.Write(w, http.StatusCreated, record)
}
