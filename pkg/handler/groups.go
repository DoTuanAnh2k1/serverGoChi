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

type pivotReq struct {
	UserID    int64 `json:"user_id,omitempty"`
	NeID      int64 `json:"ne_id,omitempty"`
	CommandID int64 `json:"command_id,omitempty"`
}

func parseGroupID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
}

// ── NE Access Group ─────────────────────────────────────────────────────

func HandlerListNeAccessGroups(w http.ResponseWriter, r *http.Request) {
	out, err := service.ListNeAccessGroups()
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	if out == nil {
		out = []*db_models.NeAccessGroup{}
	}
	response.Write(w, http.StatusOK, out)
}

func HandlerCreateNeAccessGroup(w http.ResponseWriter, r *http.Request) {
	var g db_models.NeAccessGroup
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := service.CreateNeAccessGroup(&g); err != nil {
		response.Write(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Write(w, http.StatusCreated, g)
}

func HandlerUpdateNeAccessGroup(w http.ResponseWriter, r *http.Request) {
	id, err := parseGroupID(r)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	var g db_models.NeAccessGroup
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid body")
		return
	}
	g.ID = id
	if err := service.UpdateNeAccessGroup(&g); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "updated")
}

func HandlerDeleteNeAccessGroup(w http.ResponseWriter, r *http.Request) {
	id, err := parseGroupID(r)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := service.DeleteNeAccessGroup(id); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "deleted")
}

func HandlerNeAccessGroupUsers(w http.ResponseWriter, r *http.Request) {
	id, err := parseGroupID(r)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	ids, err := service.ListUsersInNeAccessGroup(id)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Write(w, http.StatusOK, ids)
}

func HandlerNeAccessGroupAddUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseGroupID(r)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req pivotReq
	_ = json.NewDecoder(r.Body).Decode(&req)
	if err := service.AssignUserToNeAccessGroup(id, req.UserID); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "assigned")
}

func HandlerNeAccessGroupRemoveUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseGroupID(r)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	uid, err := strconv.ParseInt(chi.URLParam(r, "user_id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid user_id")
		return
	}
	if err := service.UnassignUserFromNeAccessGroup(id, uid); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "removed")
}

func HandlerNeAccessGroupNEs(w http.ResponseWriter, r *http.Request) {
	id, err := parseGroupID(r)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	ids, err := service.ListNEsInNeAccessGroup(id)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Write(w, http.StatusOK, ids)
}

func HandlerNeAccessGroupAddNE(w http.ResponseWriter, r *http.Request) {
	id, err := parseGroupID(r)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req pivotReq
	_ = json.NewDecoder(r.Body).Decode(&req)
	if err := service.AssignNeToNeAccessGroup(id, req.NeID); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "assigned")
}

func HandlerNeAccessGroupRemoveNE(w http.ResponseWriter, r *http.Request) {
	id, err := parseGroupID(r)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	nid, err := strconv.ParseInt(chi.URLParam(r, "ne_id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid ne_id")
		return
	}
	if err := service.UnassignNeFromNeAccessGroup(id, nid); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "removed")
}

// ── Cmd Exec Group ──────────────────────────────────────────────────────

func HandlerListCmdExecGroups(w http.ResponseWriter, r *http.Request) {
	out, err := service.ListCmdExecGroups()
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	if out == nil {
		out = []*db_models.CmdExecGroup{}
	}
	response.Write(w, http.StatusOK, out)
}

func HandlerCreateCmdExecGroup(w http.ResponseWriter, r *http.Request) {
	var g db_models.CmdExecGroup
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := service.CreateCmdExecGroup(&g); err != nil {
		response.Write(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Write(w, http.StatusCreated, g)
}

func HandlerUpdateCmdExecGroup(w http.ResponseWriter, r *http.Request) {
	id, err := parseGroupID(r)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	var g db_models.CmdExecGroup
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid body")
		return
	}
	g.ID = id
	if err := service.UpdateCmdExecGroup(&g); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "updated")
}

func HandlerDeleteCmdExecGroup(w http.ResponseWriter, r *http.Request) {
	id, err := parseGroupID(r)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := service.DeleteCmdExecGroup(id); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "deleted")
}

func HandlerCmdExecGroupUsers(w http.ResponseWriter, r *http.Request) {
	id, err := parseGroupID(r)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	ids, err := service.ListUsersInCmdExecGroup(id)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Write(w, http.StatusOK, ids)
}

func HandlerCmdExecGroupAddUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseGroupID(r)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req pivotReq
	_ = json.NewDecoder(r.Body).Decode(&req)
	if err := service.AssignUserToCmdExecGroup(id, req.UserID); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "assigned")
}

func HandlerCmdExecGroupRemoveUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseGroupID(r)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	uid, err := strconv.ParseInt(chi.URLParam(r, "user_id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid user_id")
		return
	}
	if err := service.UnassignUserFromCmdExecGroup(id, uid); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "removed")
}

func HandlerCmdExecGroupCommands(w http.ResponseWriter, r *http.Request) {
	id, err := parseGroupID(r)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	ids, err := service.ListCommandsInCmdExecGroup(id)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Write(w, http.StatusOK, ids)
}

func HandlerCmdExecGroupAddCommand(w http.ResponseWriter, r *http.Request) {
	id, err := parseGroupID(r)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req pivotReq
	_ = json.NewDecoder(r.Body).Decode(&req)
	if err := service.AssignCommandToCmdExecGroup(id, req.CommandID); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "assigned")
}

func HandlerCmdExecGroupRemoveCommand(w http.ResponseWriter, r *http.Request) {
	id, err := parseGroupID(r)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	cid, err := strconv.ParseInt(chi.URLParam(r, "command_id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid command_id")
		return
	}
	if err := service.UnassignCommandFromCmdExecGroup(id, cid); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "removed")
}
