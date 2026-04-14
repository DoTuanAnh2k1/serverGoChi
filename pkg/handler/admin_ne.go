package handler

import (
	"encoding/json"
	"net/http"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
)

// HandlerAdminNeList returns all NEs with full fields (for frontend admin panel).
func HandlerAdminNeList(w http.ResponseWriter, r *http.Request) {
	list, err := service.GetNeListBySystemType("5GC")
	if err != nil {
		logger.Logger.Error("admin/ne/list: ", err)
		response.InternalError(w, "failed to list NEs")
		return
	}
	if len(list) == 0 {
		response.Write(w, http.StatusOK, []struct{}{})
		return
	}
	response.Write(w, http.StatusOK, list)
}

// HandlerAdminNeCreate creates a NE with all fields.
func HandlerAdminNeCreate(w http.ResponseWriter, r *http.Request) {
	var req db_models.CliNe
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.NeName == "" {
		response.Write(w, http.StatusBadRequest, "ne_name is required")
		return
	}
	if req.SystemType == "" {
		req.SystemType = "5GC"
	}
	if err := service.CreateNe(&req); err != nil {
		logger.Logger.Error("admin/ne/create: ", err)
		response.InternalError(w, "failed to create NE")
		return
	}
	user := mustUser(r)
	saveHistory(opHistory("admin ne create", req.NeName, user.Username), "success")
	response.Write(w, http.StatusCreated, req)
}

// HandlerAdminNeUpdate updates a NE with all fields.
func HandlerAdminNeUpdate(w http.ResponseWriter, r *http.Request) {
	var req db_models.CliNe
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == 0 {
		response.Write(w, http.StatusBadRequest, "id is required")
		return
	}
	if err := service.UpdateNe(&req); err != nil {
		logger.Logger.Error("admin/ne/update: ", err)
		response.InternalError(w, "failed to update NE")
		return
	}
	user := mustUser(r)
	saveHistory(opHistory("admin ne update", req.NeName, user.Username), "success")
	response.Success(w, "NE updated")
}
