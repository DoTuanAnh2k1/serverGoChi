package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/middleware"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
)

// ─────────────────────────────────────────────────────────────────────────
// RBAC endpoints (docs/rbac-design.md §7).
// All mutation endpoints require admin role (wire up via CheckRole in the
// router). Read endpoints use Authenticate only.
// ─────────────────────────────────────────────────────────────────────────

// ─── NE Profile ───

func HandlerListNeProfiles(w http.ResponseWriter, r *http.Request) {
	out, err := service.ListNeProfiles()
	if err != nil {
		response.InternalError(w, "failed to list ne_profiles")
		return
	}
	if out == nil {
		out = []*db_models.CliNeProfile{}
	}
	response.Write(w, http.StatusOK, out)
}

func HandlerCreateNeProfile(w http.ResponseWriter, r *http.Request) {
	var p db_models.CliNeProfile
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := service.CreateNeProfile(&p); err != nil {
		response.Write(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Write(w, http.StatusCreated, p)
}

func HandlerUpdateNeProfile(w http.ResponseWriter, r *http.Request) {
	var p db_models.CliNeProfile
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if p.ID == 0 {
		response.Write(w, http.StatusBadRequest, "id is required")
		return
	}
	if err := service.UpdateNeProfile(&p); err != nil {
		response.InternalError(w, "failed to update ne_profile")
		return
	}
	response.Write(w, http.StatusOK, p)
}

func HandlerDeleteNeProfile(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := service.DeleteNeProfileById(id); err != nil {
		response.InternalError(w, "failed to delete ne_profile")
		return
	}
	response.Write(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// HandlerAssignNeProfile sets cli_ne.ne_profile_id for a given NE.
// Input: POST /aa/ne/{ne_id}/profile  body { "ne_profile_id": <id> | null }
func HandlerAssignNeProfile(w http.ResponseWriter, r *http.Request) {
	neID, err := strconv.ParseInt(chi.URLParam(r, "ne_id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid ne_id")
		return
	}
	var body struct {
		NeProfileID *int64 `json:"ne_profile_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	sto := store.GetSingleton()
	ne, err := sto.GetCliNeByNeId(neID)
	if err != nil {
		response.InternalError(w, "failed to load ne")
		return
	}
	if ne == nil {
		response.Write(w, http.StatusNotFound, "ne not found")
		return
	}
	ne.NeProfileID = body.NeProfileID
	if err := sto.UpdateCliNe(ne); err != nil {
		response.InternalError(w, "failed to update ne")
		return
	}
	response.Write(w, http.StatusOK, ne)
}

// ─── Command Def ───

func HandlerListCommandDefs(w http.ResponseWriter, r *http.Request) {
	out, err := service.ListCommandDefs(
		r.URL.Query().Get("service"),
		r.URL.Query().Get("ne_profile"),
		r.URL.Query().Get("category"),
	)
	if err != nil {
		response.InternalError(w, "failed to list command_defs")
		return
	}
	if out == nil {
		out = []*db_models.CliCommandDef{}
	}
	response.Write(w, http.StatusOK, out)
}

func HandlerCreateCommandDef(w http.ResponseWriter, r *http.Request) {
	var d db_models.CliCommandDef
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if actor, ok := actorFromCtx(r); ok {
		d.CreatedBy = actor
	}
	if err := service.CreateCommandDef(&d); err != nil {
		response.Write(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Write(w, http.StatusCreated, d)
}

func HandlerUpdateCommandDef(w http.ResponseWriter, r *http.Request) {
	var d db_models.CliCommandDef
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if d.ID == 0 {
		response.Write(w, http.StatusBadRequest, "id is required")
		return
	}
	if err := service.UpdateCommandDef(&d); err != nil {
		response.InternalError(w, "failed to update command_def")
		return
	}
	response.Write(w, http.StatusOK, d)
}

func HandlerDeleteCommandDef(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := service.DeleteCommandDefById(id); err != nil {
		response.InternalError(w, "failed to delete command_def")
		return
	}
	response.Write(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// HandlerImportCommandDefs accepts an array of CliCommandDef and inserts one
// by one, short-circuiting on the first failure. Used by CLI bulk import.
func HandlerImportCommandDefs(w http.ResponseWriter, r *http.Request) {
	var defs []*db_models.CliCommandDef
	if err := json.NewDecoder(r.Body).Decode(&defs); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body (expected JSON array)")
		return
	}
	actor, _ := actorFromCtx(r)
	for i, d := range defs {
		if d == nil {
			continue
		}
		if d.CreatedBy == "" {
			d.CreatedBy = actor
		}
		if err := service.CreateCommandDef(d); err != nil {
			response.Write(w, http.StatusBadRequest, "at index "+strconv.Itoa(i)+": "+err.Error())
			return
		}
	}
	response.Write(w, http.StatusCreated, map[string]int{"inserted": len(defs)})
}

// ─── Command Group ───

func HandlerListCommandGroups(w http.ResponseWriter, r *http.Request) {
	out, err := service.ListCommandGroups(r.URL.Query().Get("service"), r.URL.Query().Get("ne_profile"))
	if err != nil {
		response.InternalError(w, "failed to list command_groups")
		return
	}
	if out == nil {
		out = []*db_models.CliCommandGroup{}
	}
	response.Write(w, http.StatusOK, out)
}

func HandlerCreateCommandGroup(w http.ResponseWriter, r *http.Request) {
	var g db_models.CliCommandGroup
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if actor, ok := actorFromCtx(r); ok {
		g.CreatedBy = actor
	}
	if err := service.CreateCommandGroup(&g); err != nil {
		response.Write(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Write(w, http.StatusCreated, g)
}

func HandlerUpdateCommandGroup(w http.ResponseWriter, r *http.Request) {
	var g db_models.CliCommandGroup
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if g.ID == 0 {
		response.Write(w, http.StatusBadRequest, "id is required")
		return
	}
	if err := service.UpdateCommandGroup(&g); err != nil {
		response.InternalError(w, "failed to update command_group")
		return
	}
	response.Write(w, http.StatusOK, g)
}

func HandlerDeleteCommandGroup(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := service.DeleteCommandGroupById(id); err != nil {
		response.InternalError(w, "failed to delete command_group")
		return
	}
	response.Write(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func HandlerListCommandsOfGroup(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	defs, err := service.ListCommandsOfGroup(id)
	if err != nil {
		response.InternalError(w, "failed to list commands of group")
		return
	}
	if defs == nil {
		defs = []*db_models.CliCommandDef{}
	}
	response.Write(w, http.StatusOK, defs)
}

// HandlerAddCommandToGroup POST body { "command_def_id": <id> } OR { "command_def_ids": [<id>,...] }
func HandlerAddCommandToGroup(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid group id")
		return
	}
	var body struct {
		CommandDefID  int64   `json:"command_def_id"`
		CommandDefIDs []int64 `json:"command_def_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	ids := body.CommandDefIDs
	if body.CommandDefID != 0 {
		ids = append(ids, body.CommandDefID)
	}
	if len(ids) == 0 {
		response.Write(w, http.StatusBadRequest, "command_def_id(s) required")
		return
	}
	for _, id := range ids {
		if err := service.AddCommandToGroup(groupID, id); err != nil {
			response.Write(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	response.Write(w, http.StatusCreated, map[string]int{"mapped": len(ids)})
}

func HandlerRemoveCommandFromGroup(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid group id")
		return
	}
	cmdID, err := strconv.ParseInt(chi.URLParam(r, "cmd_id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid cmd_id")
		return
	}
	if err := service.RemoveCommandFromGroup(groupID, cmdID); err != nil {
		response.InternalError(w, "failed to remove command from group")
		return
	}
	response.Write(w, http.StatusOK, map[string]string{"status": "removed"})
}

// ─── Group Cmd Permission ───

func HandlerListGroupCmdPermissions(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid group id")
		return
	}
	out, err := service.ListGroupCmdPermissions(groupID)
	if err != nil {
		response.InternalError(w, "failed to list group permissions")
		return
	}
	if out == nil {
		out = []*db_models.CliGroupCmdPermission{}
	}
	response.Write(w, http.StatusOK, out)
}

func HandlerCreateGroupCmdPermission(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid group id")
		return
	}
	var p db_models.CliGroupCmdPermission
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	p.GroupID = groupID
	if err := service.CreateGroupCmdPermission(&p); err != nil {
		response.Write(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Write(w, http.StatusCreated, p)
}

func HandlerDeleteGroupCmdPermission(w http.ResponseWriter, r *http.Request) {
	permID, err := strconv.ParseInt(chi.URLParam(r, "perm_id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid perm_id")
		return
	}
	if err := service.DeleteGroupCmdPermissionById(permID); err != nil {
		response.InternalError(w, "failed to delete permission")
		return
	}
	response.Write(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ─── Authorize ───

// HandlerAuthorizeEffective returns the full effective permission set for
// the authenticated caller. Intended for ne-config / ne-command to cache
// at session start.
func HandlerAuthorizeEffective(w http.ResponseWriter, r *http.Request) {
	user, ok := middlewareUser(r)
	if !ok {
		response.InternalError(w, "missing user context")
		return
	}
	u, err := service.GetUserByUserName(user.Username)
	if err != nil || u == nil {
		response.Write(w, http.StatusNotFound, "user not found")
		return
	}
	resp, err := service.GetEffectivePermissions(u.AccountID, u.AccountName)
	if err != nil {
		logger.Logger.Errorf("authorize effective: %v", err)
		response.InternalError(w, "failed to compute effective permissions")
		return
	}
	response.Write(w, http.StatusOK, resp)
}

// HandlerAuthorizeCheckCommand answers "is the caller allowed to run the
// requested command on the requested NE?".
func HandlerAuthorizeCheckCommand(w http.ResponseWriter, r *http.Request) {
	user, ok := middlewareUser(r)
	if !ok {
		response.InternalError(w, "missing user context")
		return
	}
	var body struct {
		Service string `json:"service"`
		Command string `json:"command"`
		NeID    int64  `json:"ne_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(body.Command) == "" || body.NeID == 0 {
		response.Write(w, http.StatusBadRequest, "command and ne_id are required")
		return
	}
	u, err := service.GetUserByUserName(user.Username)
	if err != nil || u == nil {
		response.Write(w, http.StatusNotFound, "user not found")
		return
	}
	res, err := service.CheckCommand(u.AccountID, body.NeID, body.Service, body.Command)
	if err != nil {
		logger.Logger.Errorf("authorize check-command: %v", err)
		response.InternalError(w, "failed to evaluate permission")
		return
	}
	response.Write(w, http.StatusOK, res)
}

// ─── helpers ───

func actorFromCtx(r *http.Request) (string, bool) {
	u, ok := middlewareUser(r)
	if !ok {
		return "", false
	}
	return u.Username, true
}

func middlewareUser(r *http.Request) (*middleware.User, bool) {
	u, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User)
	if !ok || u == nil {
		return nil, false
	}
	return u, true
}

// Keep the errors import live — the handler file intentionally uses
// errors.As in future guards; remove this if the reference is dropped.
var _ = errors.New
