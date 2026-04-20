package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/testutil"
)

// ── HandlerGroupList ──────────────────────────────────────────────────────────

func TestHandlerGroupList_ReturnsGroups(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetAllGroupsFn: func() ([]*db_models.CliGroup, error) {
			return []*db_models.CliGroup{
				{ID: 1, Name: "ops", Description: "ops team"},
				{ID: 2, Name: "ro", Description: "read-only"},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerGroupList(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	var out []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out) != 2 {
		t.Errorf("groups: got %d, want 2", len(out))
	}
}

// ── HandlerGroupCreate ────────────────────────────────────────────────────────

func TestHandlerGroupCreate_Success(t *testing.T) {
	var saved *db_models.CliGroup
	store.SetSingleton(&testutil.MockStore{
		GetGroupByNameFn: func(name string) (*db_models.CliGroup, error) { return nil, nil },
		CreateGroupFn: func(g *db_models.CliGroup) error {
			saved = g
			g.ID = 7
			return nil
		},
		SaveHistoryCommandFn: func(h db_models.CliOperationHistory) error { return nil },
	})

	body, _ := json.Marshal(map[string]any{"name": "ops", "description": "ops team"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerGroupCreate(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status: got %d, want 201", w.Code)
	}
	if saved == nil || saved.Name != "ops" || saved.Description != "ops team" {
		t.Errorf("saved group wrong: %+v", saved)
	}
}

func TestHandlerGroupCreate_MissingName(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{})

	body, _ := json.Marshal(map[string]any{"description": "no name"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerGroupCreate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestHandlerGroupCreate_DuplicateName(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetGroupByNameFn: func(name string) (*db_models.CliGroup, error) {
			return &db_models.CliGroup{ID: 3, Name: name}, nil
		},
	})

	body, _ := json.Marshal(map[string]any{"name": "ops"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerGroupCreate(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("status: got %d, want 409", w.Code)
	}
}

// ── HandlerGroupUpdate ────────────────────────────────────────────────────────

func TestHandlerGroupUpdate_Success(t *testing.T) {
	var saved *db_models.CliGroup
	store.SetSingleton(&testutil.MockStore{
		GetGroupByIdFn: func(id int64) (*db_models.CliGroup, error) {
			return &db_models.CliGroup{ID: id, Name: "ops"}, nil
		},
		GetGroupByNameFn: func(name string) (*db_models.CliGroup, error) { return nil, nil },
		UpdateGroupFn: func(g *db_models.CliGroup) error {
			saved = g
			return nil
		},
		SaveHistoryCommandFn: func(h db_models.CliOperationHistory) error { return nil },
	})

	body, _ := json.Marshal(map[string]any{"id": 5, "name": "ops-v2", "description": "renamed"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerGroupUpdate(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	if saved == nil || saved.Name != "ops-v2" || saved.Description != "renamed" {
		t.Errorf("updated group wrong: %+v", saved)
	}
}

func TestHandlerGroupUpdate_NotFound(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetGroupByIdFn: func(id int64) (*db_models.CliGroup, error) { return nil, nil },
	})

	body, _ := json.Marshal(map[string]any{"id": 99, "name": "x"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerGroupUpdate(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", w.Code)
	}
}

// ── HandlerGroupDelete ────────────────────────────────────────────────────────

func TestHandlerGroupDelete_CascadesAndSucceeds(t *testing.T) {
	var deletedGroup int64
	var delUserMaps, delNeMaps bool
	store.SetSingleton(&testutil.MockStore{
		DeleteAllUserGroupMappingByGroupIdFn: func(id int64) error {
			delUserMaps = true
			return nil
		},
		DeleteAllGroupNeMappingByGroupIdFn: func(id int64) error {
			delNeMaps = true
			return nil
		},
		DeleteGroupByIdFn: func(id int64) error {
			deletedGroup = id
			return nil
		},
		SaveHistoryCommandFn: func(h db_models.CliOperationHistory) error { return nil },
	})

	body, _ := json.Marshal(map[string]any{"id": 5})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerGroupDelete(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	if deletedGroup != 5 || !delUserMaps || !delNeMaps {
		t.Errorf("cascade incomplete: group=%d userMaps=%v neMaps=%v", deletedGroup, delUserMaps, delNeMaps)
	}
}

// ── HandlerGroupShow ──────────────────────────────────────────────────────────

func TestHandlerGroupShow_ReturnsUsersAndNeIds(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetGroupByIdFn: func(id int64) (*db_models.CliGroup, error) {
			return &db_models.CliGroup{ID: id, Name: "ops", Description: "ops team"}, nil
		},
		GetAllUsersOfGroupFn: func(groupId int64) ([]*db_models.CliUserGroupMapping, error) {
			return []*db_models.CliUserGroupMapping{{UserID: 5, GroupID: groupId}}, nil
		},
		GetAllUserFn: func() ([]*db_models.TblAccount, error) {
			return []*db_models.TblAccount{
				{AccountID: 5, AccountName: "bob"},
			}, nil
		},
		GetAllNesOfGroupFn: func(groupId int64) ([]*db_models.CliGroupNeMapping, error) {
			return []*db_models.CliGroupNeMapping{{GroupID: groupId, TblNeID: 10}}, nil
		},
	})

	body, _ := json.Marshal(map[string]any{"id": 1})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerGroupShow(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	users, _ := out["users"].([]any)
	if len(users) != 1 || users[0] != "bob" {
		t.Errorf("users: got %v, want [bob]", users)
	}
	neIds, _ := out["ne_ids"].([]any)
	if len(neIds) != 1 {
		t.Errorf("ne_ids: got %v, want [10]", neIds)
	}
}

// ── HandlerUserGroupAssign / Unassign ─────────────────────────────────────────

func TestHandlerUserGroupAssign_Success(t *testing.T) {
	var saved *db_models.CliUserGroupMapping
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 5, AccountName: name}, nil
		},
		GetGroupByIdFn: func(id int64) (*db_models.CliGroup, error) {
			return &db_models.CliGroup{ID: id, Name: "ops"}, nil
		},
		GetAllGroupsOfUserFn: func(userId int64) ([]*db_models.CliUserGroupMapping, error) {
			return nil, nil
		},
		CreateUserGroupMappingFn: func(m *db_models.CliUserGroupMapping) error {
			saved = m
			return nil
		},
		SaveHistoryCommandFn: func(h db_models.CliOperationHistory) error { return nil },
	})

	body, _ := json.Marshal(map[string]any{"username": "bob", "group_id": 3})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerUserGroupAssign(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	if saved == nil || saved.UserID != 5 || saved.GroupID != 3 {
		t.Errorf("mapping: got %+v, want {UserID:5, GroupID:3}", saved)
	}
}

func TestHandlerUserGroupAssign_UserNotFound(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) { return nil, nil },
	})
	body, _ := json.Marshal(map[string]any{"username": "ghost", "group_id": 1})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerUserGroupAssign(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", w.Code)
	}
}

func TestHandlerUserGroupAssign_AlreadyAssigned(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 5, AccountName: name}, nil
		},
		GetGroupByIdFn: func(id int64) (*db_models.CliGroup, error) {
			return &db_models.CliGroup{ID: id, Name: "ops"}, nil
		},
		GetAllGroupsOfUserFn: func(userId int64) ([]*db_models.CliUserGroupMapping, error) {
			return []*db_models.CliUserGroupMapping{{UserID: userId, GroupID: 3}}, nil
		},
	})

	body, _ := json.Marshal(map[string]any{"username": "bob", "group_id": 3})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerUserGroupAssign(w, req)

	if w.Code != http.StatusNotModified {
		t.Errorf("status: got %d, want 304 (already assigned)", w.Code)
	}
}

func TestHandlerUserGroupUnassign_Success(t *testing.T) {
	var deleted *db_models.CliUserGroupMapping
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 5, AccountName: name}, nil
		},
		DeleteUserGroupMappingFn: func(m *db_models.CliUserGroupMapping) error {
			deleted = m
			return nil
		},
		SaveHistoryCommandFn: func(h db_models.CliOperationHistory) error { return nil },
	})

	body, _ := json.Marshal(map[string]any{"username": "bob", "group_id": 3})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerUserGroupUnassign(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	if deleted == nil || deleted.UserID != 5 || deleted.GroupID != 3 {
		t.Errorf("deleted mapping: got %+v", deleted)
	}
}

// ── HandlerUserGroupList ──────────────────────────────────────────────────────

func TestHandlerUserGroupList_ReturnsGroups(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 5, AccountName: name}, nil
		},
		GetAllGroupsOfUserFn: func(userId int64) ([]*db_models.CliUserGroupMapping, error) {
			return []*db_models.CliUserGroupMapping{{UserID: userId, GroupID: 1}, {UserID: userId, GroupID: 2}}, nil
		},
		GetGroupByIdFn: func(id int64) (*db_models.CliGroup, error) {
			return &db_models.CliGroup{ID: id, Name: "g"}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/?username=bob", nil)
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerUserGroupList(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	var out []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out) != 2 {
		t.Errorf("groups: got %d, want 2", len(out))
	}
}

func TestHandlerUserGroupList_MissingUsernameQuery(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerUserGroupList(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

// ── HandlerGroupNeAssign / Unassign / List ────────────────────────────────────

func TestHandlerGroupNeAssign_Success(t *testing.T) {
	var saved *db_models.CliGroupNeMapping
	store.SetSingleton(&testutil.MockStore{
		GetGroupByIdFn: func(id int64) (*db_models.CliGroup, error) {
			return &db_models.CliGroup{ID: id, Name: "ops"}, nil
		},
		GetCliNeByNeIdFn: func(id int64) (*db_models.CliNe, error) {
			return &db_models.CliNe{ID: id, NeName: "HTSMF01"}, nil
		},
		GetAllNesOfGroupFn: func(groupId int64) ([]*db_models.CliGroupNeMapping, error) {
			return nil, nil
		},
		CreateGroupNeMappingFn: func(m *db_models.CliGroupNeMapping) error {
			saved = m
			return nil
		},
		SaveHistoryCommandFn: func(h db_models.CliOperationHistory) error { return nil },
	})

	body, _ := json.Marshal(map[string]any{"group_id": 3, "ne_id": 10})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerGroupNeAssign(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	if saved == nil || saved.GroupID != 3 || saved.TblNeID != 10 {
		t.Errorf("mapping: got %+v", saved)
	}
}

func TestHandlerGroupNeAssign_GroupNotFound(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetGroupByIdFn: func(id int64) (*db_models.CliGroup, error) { return nil, nil },
	})

	body, _ := json.Marshal(map[string]any{"group_id": 99, "ne_id": 10})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerGroupNeAssign(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", w.Code)
	}
}

func TestHandlerGroupNeUnassign_DBError(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		DeleteGroupNeMappingFn: func(m *db_models.CliGroupNeMapping) error { return errors.New("db down") },
		SaveHistoryCommandFn:   func(h db_models.CliOperationHistory) error { return nil },
	})

	body, _ := json.Marshal(map[string]any{"group_id": 3, "ne_id": 10})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerGroupNeUnassign(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", w.Code)
	}
}

func TestHandlerGroupNeList_ReturnsIds(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetAllNesOfGroupFn: func(groupId int64) ([]*db_models.CliGroupNeMapping, error) {
			return []*db_models.CliGroupNeMapping{
				{GroupID: groupId, TblNeID: 10},
				{GroupID: groupId, TblNeID: 11},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/?group_id=3", nil)
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerGroupNeList(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	var ids []int64
	if err := json.Unmarshal(w.Body.Bytes(), &ids); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(ids) != 2 || ids[0] != 10 || ids[1] != 11 {
		t.Errorf("ids: got %v, want [10 11]", ids)
	}
}

// ── HandlerAdminUserFullList ──────────────────────────────────────────────────

func TestHandlerAdminUserFullList_UnionDirectAndGroup(t *testing.T) {
	// bob has direct NE 10 and is in group 3 which maps to NE 20.
	// Expected: bob.nes == [10, 20] (dedup-safe), role == "user".
	store.SetSingleton(&testutil.MockStore{
		GetAllUserFn: func() ([]*db_models.TblAccount, error) {
			return []*db_models.TblAccount{
				{AccountID: 5, AccountName: "bob", AccountType: 2},
			}, nil
		},
		GetAllNeOfUserByUserIdFn: func(userID int64) ([]*db_models.CliUserNeMapping, error) {
			return []*db_models.CliUserNeMapping{{UserID: userID, TblNeID: 10}}, nil
		},
		GetAllGroupsOfUserFn: func(userId int64) ([]*db_models.CliUserGroupMapping, error) {
			return []*db_models.CliUserGroupMapping{{UserID: userId, GroupID: 3}}, nil
		},
		GetAllNesOfGroupFn: func(groupId int64) ([]*db_models.CliGroupNeMapping, error) {
			return []*db_models.CliGroupNeMapping{{GroupID: groupId, TblNeID: 20}}, nil
		},
		GetCliNeByNeIdFn: func(id int64) (*db_models.CliNe, error) {
			return &db_models.CliNe{ID: id, NeName: "NE-" + string(rune('0'+id)), SiteName: "HCM", Namespace: "ns"}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAdminUserFullList(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	var out []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("users: got %d, want 1", len(out))
	}
	entry := out[0]
	if entry["account_name"] != "bob" {
		t.Errorf("account_name: got %v", entry["account_name"])
	}
	if entry["role"] != "user" {
		t.Errorf("role: got %v, want user", entry["role"])
	}
	nes, _ := entry["nes"].([]any)
	if len(nes) != 2 {
		t.Errorf("nes: got %d, want 2 (union of direct + group)", len(nes))
	}
}

func TestHandlerAdminUserFullList_ExcludesSuperAdmins(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetAllUserFn: func() ([]*db_models.TblAccount, error) {
			return []*db_models.TblAccount{
				{AccountID: 1, AccountName: "root", AccountType: 0},
				{AccountID: 5, AccountName: "bob", AccountType: 2},
			}, nil
		},
		GetAllNeOfUserByUserIdFn: func(userID int64) ([]*db_models.CliUserNeMapping, error) { return nil, nil },
		GetAllGroupsOfUserFn:     func(userId int64) ([]*db_models.CliUserGroupMapping, error) { return nil, nil },
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAdminUserFullList(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	var out []map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	if len(out) != 1 || out[0]["account_name"] != "bob" {
		t.Errorf("expected only bob; got %v", out)
	}
}

func TestHandlerAdminUserFullList_AdminRole(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetAllUserFn: func() ([]*db_models.TblAccount, error) {
			return []*db_models.TblAccount{
				{AccountID: 5, AccountName: "carol", AccountType: 1},
			}, nil
		},
		GetAllNeOfUserByUserIdFn: func(userID int64) ([]*db_models.CliUserNeMapping, error) { return nil, nil },
		GetAllGroupsOfUserFn:     func(userId int64) ([]*db_models.CliUserGroupMapping, error) { return nil, nil },
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAdminUserFullList(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	var out []map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	if len(out) != 1 || out[0]["role"] != "admin" {
		t.Errorf("role: got %v, want admin", out)
	}
}
