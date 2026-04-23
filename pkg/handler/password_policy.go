package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
)

// ─── Password Policy ───

func HandlerListPasswordPolicies(w http.ResponseWriter, r *http.Request) {
	out, err := service.ListPasswordPolicies()
	if err != nil {
		response.InternalError(w, "failed to list password policies")
		return
	}
	if out == nil {
		out = []*db_models.CliPasswordPolicy{}
	}
	response.Write(w, http.StatusOK, out)
}

func HandlerCreatePasswordPolicy(w http.ResponseWriter, r *http.Request) {
	var p db_models.CliPasswordPolicy
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := service.CreatePasswordPolicy(&p); err != nil {
		response.Write(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Write(w, http.StatusCreated, p)
}

func HandlerUpdatePasswordPolicy(w http.ResponseWriter, r *http.Request) {
	var p db_models.CliPasswordPolicy
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if p.ID == 0 {
		response.Write(w, http.StatusBadRequest, "id is required")
		return
	}
	if err := service.UpdatePasswordPolicy(&p); err != nil {
		response.InternalError(w, "failed to update password policy")
		return
	}
	response.Write(w, http.StatusOK, p)
}

func HandlerDeletePasswordPolicy(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := service.DeletePasswordPolicyById(id); err != nil {
		response.InternalError(w, "failed to delete password policy")
		return
	}
	response.Write(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// HandlerAssignPasswordPolicyToGroup POST /aa/group/{id}/password-policy
// body { "password_policy_id": <id> | null }
func HandlerAssignPasswordPolicyToGroup(w http.ResponseWriter, r *http.Request) {
	gid, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid group id")
		return
	}
	var body struct {
		PasswordPolicyID *int64 `json:"password_policy_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := service.AssignPasswordPolicyToGroup(gid, body.PasswordPolicyID); err != nil {
		response.Write(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Write(w, http.StatusOK, map[string]string{"status": "assigned"})
}

// ─── Mgt Permissions ───

func HandlerListMgtPermissions(w http.ResponseWriter, r *http.Request) {
	gid, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid group id")
		return
	}
	out, err := service.ListMgtPermissions(gid)
	if err != nil {
		response.InternalError(w, "failed to list mgt permissions")
		return
	}
	if out == nil {
		out = []*db_models.CliGroupMgtPermission{}
	}
	response.Write(w, http.StatusOK, out)
}

func HandlerCreateMgtPermission(w http.ResponseWriter, r *http.Request) {
	gid, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid group id")
		return
	}
	var p db_models.CliGroupMgtPermission
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	p.GroupID = gid
	if err := service.CreateMgtPermission(&p); err != nil {
		response.Write(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Write(w, http.StatusCreated, p)
}

func HandlerDeleteMgtPermission(w http.ResponseWriter, r *http.Request) {
	permID, err := strconv.ParseInt(chi.URLParam(r, "perm_id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid perm_id")
		return
	}
	if err := service.DeleteMgtPermissionById(permID); err != nil {
		response.InternalError(w, "failed to delete mgt permission")
		return
	}
	response.Write(w, http.StatusOK, map[string]string{"status": "deleted"})
}
