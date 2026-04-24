package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
	"github.com/go-chi/chi"
)

// ── Password Policy ─────────────────────────────────────────────────────

func HandlerGetPasswordPolicy(w http.ResponseWriter, r *http.Request) {
	p, err := service.EffectivePasswordPolicy()
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Write(w, http.StatusOK, p)
}

func HandlerUpsertPasswordPolicy(w http.ResponseWriter, r *http.Request) {
	var p db_models.PasswordPolicy
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := service.UpsertPasswordPolicy(&p); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Write(w, http.StatusOK, p)
}

// ── Access List ─────────────────────────────────────────────────────────

func HandlerListAccessList(w http.ResponseWriter, r *http.Request) {
	out, err := service.ListAccessListEntries(r.URL.Query().Get("list_type"))
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	if out == nil {
		out = []*db_models.UserAccessList{}
	}
	response.Write(w, http.StatusOK, out)
}

func HandlerCreateAccessList(w http.ResponseWriter, r *http.Request) {
	var e db_models.UserAccessList
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := service.CreateAccessListEntry(&e); err != nil {
		response.Write(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Write(w, http.StatusCreated, e)
}

func HandlerDeleteAccessList(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := service.DeleteAccessListEntry(id); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "deleted")
}

// ── Authorize Check ─────────────────────────────────────────────────────

type authorizeCheckReq struct {
	Username  string `json:"username"`
	NeID      int64  `json:"ne_id"`
	CommandID int64  `json:"command_id"`
}

func HandlerAuthorizeCheck(w http.ResponseWriter, r *http.Request) {
	var req authorizeCheckReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid body")
		return
	}
	d, err := service.Authorize(req.Username, req.NeID, req.CommandID)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Write(w, http.StatusOK, d)
}
