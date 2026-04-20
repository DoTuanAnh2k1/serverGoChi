package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
)

// ── Group CRUD ───────────────────────────────────────────────────────────────

type groupResp struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// HandlerGroupList returns all groups.
func HandlerGroupList(w http.ResponseWriter, r *http.Request) {
	groups, err := service.GetAllGroups()
	if err != nil {
		response.InternalError(w, "failed to list groups")
		return
	}
	out := make([]groupResp, 0, len(groups))
	for _, g := range groups {
		out = append(out, groupResp{ID: g.ID, Name: g.Name, Description: g.Description})
	}
	response.Write(w, http.StatusOK, out)
}

// HandlerGroupCreate creates a new group.
// Body: { "name": string (required), "description": string }
func HandlerGroupCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		response.Write(w, http.StatusBadRequest, "name is required")
		return
	}
	existing, err := service.GetGroupByName(req.Name)
	if err != nil {
		response.InternalError(w, "failed to check group")
		return
	}
	if existing != nil {
		response.Write(w, http.StatusConflict, "group name already exists")
		return
	}
	g := &db_models.CliGroup{Name: req.Name, Description: req.Description}
	if err := service.CreateGroup(g); err != nil {
		response.InternalError(w, "failed to create group")
		return
	}
	actor := mustUser(r)
	saveHistory(opHistory("group create", req.Name, actor.Username), "success")
	response.Created(w)
}

// HandlerGroupUpdate updates a group's name/description.
// Body: { "id": int64 (required), "name": string, "description": string }
func HandlerGroupUpdate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == 0 {
		response.Write(w, http.StatusBadRequest, "id is required")
		return
	}
	g, err := service.GetGroupById(req.ID)
	if err != nil {
		response.InternalError(w, "failed to retrieve group")
		return
	}
	if g == nil {
		response.NotFound(w, "group not found")
		return
	}
	if strings.TrimSpace(req.Name) != "" && req.Name != g.Name {
		if other, _ := service.GetGroupByName(req.Name); other != nil && other.ID != g.ID {
			response.Write(w, http.StatusConflict, "group name already exists")
			return
		}
		g.Name = req.Name
	}
	g.Description = req.Description
	if err := service.UpdateGroup(g); err != nil {
		response.InternalError(w, "failed to update group")
		return
	}
	actor := mustUser(r)
	saveHistory(opHistory("group update", fmt.Sprintf("id=%d", req.ID), actor.Username), "success")
	response.Success(w, "group updated")
}

// HandlerGroupDelete cascades through mappings, then the group.
// Body: { "id": int64 }
func HandlerGroupDelete(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == 0 {
		response.Write(w, http.StatusBadRequest, "id is required")
		return
	}
	if err := service.DeleteGroupById(req.ID); err != nil {
		response.InternalError(w, "failed to delete group")
		return
	}
	actor := mustUser(r)
	saveHistory(opHistory("group delete", fmt.Sprintf("id=%d", req.ID), actor.Username), "success")
	response.Success(w, "group deleted")
}

// HandlerGroupShow returns details for a single group plus its users and NEs.
// Body: { "id": int64 }
func HandlerGroupShow(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == 0 {
		response.Write(w, http.StatusBadRequest, "id is required")
		return
	}
	g, err := service.GetGroupById(req.ID)
	if err != nil {
		response.InternalError(w, "failed to retrieve group")
		return
	}
	if g == nil {
		response.NotFound(w, "group not found")
		return
	}
	userMaps, _ := service.GetUsersOfGroup(req.ID)
	userList := make([]string, 0, len(userMaps))
	allUsers, _ := service.GetAllUser()
	userNameByID := make(map[int64]string, len(allUsers))
	for _, uu := range allUsers {
		userNameByID[uu.AccountID] = uu.AccountName
	}
	for _, m := range userMaps {
		if name, ok := userNameByID[m.UserID]; ok {
			userList = append(userList, name)
		}
	}
	neMaps, _ := service.GetNesOfGroup(req.ID)
	neIds := make([]int64, 0, len(neMaps))
	for _, m := range neMaps {
		neIds = append(neIds, m.TblNeID)
	}
	response.Write(w, http.StatusOK, map[string]interface{}{
		"id":          g.ID,
		"name":        g.Name,
		"description": g.Description,
		"users":       userList,
		"ne_ids":      neIds,
	})
}

// ── user ↔ group ─────────────────────────────────────────────────────────────

type userGroupReq struct {
	Username string `json:"username"`
	GroupID  int64  `json:"group_id"`
}

// HandlerUserGroupAssign adds a user-group mapping.
func HandlerUserGroupAssign(w http.ResponseWriter, r *http.Request) {
	var req userGroupReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" || req.GroupID == 0 {
		response.Write(w, http.StatusBadRequest, "username and group_id are required")
		return
	}
	u, err := service.GetUserByUserName(req.Username)
	if err != nil || u == nil {
		response.NotFound(w, "user not found")
		return
	}
	g, err := service.GetGroupById(req.GroupID)
	if err != nil || g == nil {
		response.NotFound(w, "group not found")
		return
	}
	existing, _ := service.GetGroupsOfUser(u.AccountID)
	for _, m := range existing {
		if m.GroupID == req.GroupID {
			response.Write(w, http.StatusNotModified, "already assigned")
			return
		}
	}
	if err := service.AssignUserToGroup(u.AccountID, req.GroupID); err != nil {
		response.InternalError(w, "failed to assign user to group")
		return
	}
	actor := mustUser(r)
	saveHistory(opHistory("user-group assign", fmt.Sprintf("%s->%d", req.Username, req.GroupID), actor.Username), "success")
	response.Success(w, "user assigned to group")
}

// HandlerUserGroupUnassign removes a user-group mapping.
func HandlerUserGroupUnassign(w http.ResponseWriter, r *http.Request) {
	var req userGroupReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" || req.GroupID == 0 {
		response.Write(w, http.StatusBadRequest, "username and group_id are required")
		return
	}
	u, err := service.GetUserByUserName(req.Username)
	if err != nil || u == nil {
		response.NotFound(w, "user not found")
		return
	}
	if err := service.UnassignUserFromGroup(u.AccountID, req.GroupID); err != nil {
		response.InternalError(w, "failed to unassign user from group")
		return
	}
	actor := mustUser(r)
	saveHistory(opHistory("user-group unassign", fmt.Sprintf("%s->%d", req.Username, req.GroupID), actor.Username), "success")
	response.Success(w, "user unassigned from group")
}

// HandlerUserGroupList returns the groups a user belongs to.
// Query: ?username=alice
func HandlerUserGroupList(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		response.Write(w, http.StatusBadRequest, "username query is required")
		return
	}
	u, err := service.GetUserByUserName(username)
	if err != nil || u == nil {
		response.NotFound(w, "user not found")
		return
	}
	maps, err := service.GetGroupsOfUser(u.AccountID)
	if err != nil {
		response.InternalError(w, "failed to list groups")
		return
	}
	out := make([]groupResp, 0, len(maps))
	for _, m := range maps {
		g, err := service.GetGroupById(m.GroupID)
		if err != nil || g == nil {
			continue
		}
		out = append(out, groupResp{ID: g.ID, Name: g.Name, Description: g.Description})
	}
	response.Write(w, http.StatusOK, out)
}

// ── group ↔ NE ───────────────────────────────────────────────────────────────

type groupNeReq struct {
	GroupID int64 `json:"group_id"`
	NeID    int64 `json:"ne_id"`
}

// HandlerGroupNeAssign adds a group-ne mapping.
func HandlerGroupNeAssign(w http.ResponseWriter, r *http.Request) {
	var req groupNeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.GroupID == 0 || req.NeID == 0 {
		response.Write(w, http.StatusBadRequest, "group_id and ne_id are required")
		return
	}
	g, err := service.GetGroupById(req.GroupID)
	if err != nil || g == nil {
		response.NotFound(w, "group not found")
		return
	}
	ne, err := service.GetNeByNeId(req.NeID)
	if err != nil || ne == nil {
		response.NotFound(w, "NE not found")
		return
	}
	existing, _ := service.GetNesOfGroup(req.GroupID)
	for _, m := range existing {
		if m.TblNeID == req.NeID {
			response.Write(w, http.StatusNotModified, "already assigned")
			return
		}
	}
	if err := service.AssignNeToGroup(req.GroupID, req.NeID); err != nil {
		response.InternalError(w, "failed to assign NE to group")
		return
	}
	actor := mustUser(r)
	saveHistory(opHistory("group-ne assign", fmt.Sprintf("%d->%d", req.GroupID, req.NeID), actor.Username), "success")
	response.Success(w, "NE assigned to group")
}

// HandlerGroupNeUnassign removes a group-ne mapping.
func HandlerGroupNeUnassign(w http.ResponseWriter, r *http.Request) {
	var req groupNeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.GroupID == 0 || req.NeID == 0 {
		response.Write(w, http.StatusBadRequest, "group_id and ne_id are required")
		return
	}
	if err := service.UnassignNeFromGroup(req.GroupID, req.NeID); err != nil {
		response.InternalError(w, "failed to unassign NE from group")
		return
	}
	actor := mustUser(r)
	saveHistory(opHistory("group-ne unassign", fmt.Sprintf("%d->%d", req.GroupID, req.NeID), actor.Username), "success")
	response.Success(w, "NE unassigned from group")
}

// HandlerGroupNeList returns the NE ids belonging to a group.
// Query: ?group_id=123
func HandlerGroupNeList(w http.ResponseWriter, r *http.Request) {
	var groupID int64
	if _, err := fmt.Sscanf(r.URL.Query().Get("group_id"), "%d", &groupID); err != nil || groupID == 0 {
		response.Write(w, http.StatusBadRequest, "group_id query is required")
		return
	}
	maps, err := service.GetNesOfGroup(groupID)
	if err != nil {
		response.InternalError(w, "failed to list NEs")
		return
	}
	ids := make([]int64, 0, len(maps))
	for _, m := range maps {
		ids = append(ids, m.TblNeID)
	}
	response.Write(w, http.StatusOK, ids)
}
