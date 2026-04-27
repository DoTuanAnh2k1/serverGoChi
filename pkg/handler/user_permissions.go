package handler

import (
	"net/http"
	"strconv"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/go-chi/chi"
)

// HandlerUserExecutableCommands returns all commands a user can execute,
// derived from the union of all cmd_exec_groups the user belongs to.
//
//	GET /aa/users/{id}/executable-commands → []*Command
func HandlerUserExecutableCommands(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	s := store.GetSingleton()

	groupIDs, err := s.ListCmdExecGroupsOfUser(userID)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	seen := make(map[int64]bool)
	var result []*db_models.Command
	for _, gid := range groupIDs {
		cmdIDs, _ := s.ListCommandsInCmdExecGroup(gid)
		for _, cid := range cmdIDs {
			if seen[cid] {
				continue
			}
			seen[cid] = true
			if cmd, err := s.GetCommandByID(cid); err == nil && cmd != nil {
				result = append(result, cmd)
			}
		}
	}
	if result == nil {
		result = []*db_models.Command{}
	}
	response.Write(w, http.StatusOK, result)
}

// HandlerUserAccessibleNEs returns all NEs a user can reach,
// derived from the union of all ne_access_groups the user belongs to.
//
//	GET /aa/users/{id}/accessible-nes → []*NE
func HandlerUserAccessibleNEs(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	s := store.GetSingleton()

	groupIDs, err := s.ListNeAccessGroupsOfUser(userID)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	seen := make(map[int64]bool)
	var result []*db_models.NE
	for _, gid := range groupIDs {
		neIDs, _ := s.ListNEsInNeAccessGroup(gid)
		for _, nid := range neIDs {
			if seen[nid] {
				continue
			}
			seen[nid] = true
			if ne, err := s.GetNEByID(nid); err == nil && ne != nil {
				result = append(result, ne)
			}
		}
	}
	if result == nil {
		result = []*db_models.NE{}
	}
	response.Write(w, http.StatusOK, result)
}

// HandlerCommandAuthorizedUsers returns all users authorized to execute a command.
//
//	GET /aa/commands/{id}/authorized-users → []*User
func HandlerCommandAuthorizedUsers(w http.ResponseWriter, r *http.Request) {
	cmdID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	s := store.GetSingleton()

	groupIDs, err := s.ListCmdExecGroupsOfCommand(cmdID)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	seen := make(map[int64]bool)
	var result []*db_models.User
	for _, gid := range groupIDs {
		userIDs, _ := s.ListUsersInCmdExecGroup(gid)
		for _, uid := range userIDs {
			if seen[uid] {
				continue
			}
			seen[uid] = true
			if u, err := s.GetUserByID(uid); err == nil && u != nil {
				result = append(result, u)
			}
		}
	}
	if result == nil {
		result = []*db_models.User{}
	}
	response.Write(w, http.StatusOK, result)
}

// HandlerNEAuthorizedUsers returns all users who can access an NE.
//
//	GET /aa/nes/{id}/authorized-users → []*User
func HandlerNEAuthorizedUsers(w http.ResponseWriter, r *http.Request) {
	neID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	s := store.GetSingleton()

	groupIDs, err := s.ListNeAccessGroupsOfNE(neID)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	seen := make(map[int64]bool)
	var result []*db_models.User
	for _, gid := range groupIDs {
		userIDs, _ := s.ListUsersInNeAccessGroup(gid)
		for _, uid := range userIDs {
			if seen[uid] {
				continue
			}
			seen[uid] = true
			if u, err := s.GetUserByID(uid); err == nil && u != nil {
				result = append(result, u)
			}
		}
	}
	if result == nil {
		result = []*db_models.User{}
	}
	response.Write(w, http.StatusOK, result)
}
