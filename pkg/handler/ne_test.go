package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/testutil"
)

// HandlerListNe must return the union of direct NE mappings and NEs reachable
// via the user's group memberships, deduplicated.
func TestHandlerListNe_UnionDirectAndGroup(t *testing.T) {
	neRows := map[int64]*db_models.CliNe{
		10: {ID: 10, NeName: "direct-ne", SiteName: "HCM", ConfMasterIP: "10.0.0.10"},
		20: {ID: 20, NeName: "group-ne-1", SiteName: "HN", ConfMasterIP: "10.0.0.20"},
		30: {ID: 30, NeName: "group-ne-2", SiteName: "DN", ConfMasterIP: "10.0.0.30"},
	}

	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(_ string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 1, AccountName: "alice"}, nil
		},
		GetAllNeOfUserByUserIdFn: func(uid int64) ([]*db_models.CliUserNeMapping, error) {
			return []*db_models.CliUserNeMapping{{UserID: uid, TblNeID: 10}}, nil
		},
		GetAllGroupsOfUserFn: func(uid int64) ([]*db_models.CliUserGroupMapping, error) {
			return []*db_models.CliUserGroupMapping{{UserID: uid, GroupID: 7}}, nil
		},
		GetAllNesOfGroupFn: func(gid int64) ([]*db_models.CliGroupNeMapping, error) {
			if gid == 7 {
				return []*db_models.CliGroupNeMapping{
					{GroupID: 7, TblNeID: 20},
					{GroupID: 7, TblNeID: 30},
					{GroupID: 7, TblNeID: 10}, // overlap with direct — should dedupe
				}, nil
			}
			return nil, nil
		},
		GetCliNeByNeIdFn: func(id int64) (*db_models.CliNe, error) {
			return neRows[id], nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/aa/list/ne", nil)
	req = reqWithUser(req, adminUser())
	w := httptest.NewRecorder()

	handler.HandlerListNe(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("status: got %d, want 302; body=%s", w.Code, w.Body.String())
	}

	var body struct {
		NeDataList []struct {
			Ne   string `json:"ne"`
			Site string `json:"site"`
			Ip   string `json:"ip"`
		} `json:"neDataList"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.NeDataList) != 3 {
		t.Fatalf("ne count: got %d, want 3 (direct 10 + group 20,30; 10 deduped) — %+v", len(body.NeDataList), body.NeDataList)
	}
	got := make([]string, 0, len(body.NeDataList))
	for _, n := range body.NeDataList {
		got = append(got, n.Ne)
	}
	sort.Strings(got)
	want := []string{"direct-ne", "group-ne-1", "group-ne-2"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ne[%d]: got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestHandlerListNe_NoNes(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(_ string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 1}, nil
		},
		GetAllNeOfUserByUserIdFn: func(_ int64) ([]*db_models.CliUserNeMapping, error) {
			return nil, nil
		},
		GetAllGroupsOfUserFn: func(_ int64) ([]*db_models.CliUserGroupMapping, error) {
			return nil, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/aa/list/ne", nil)
	req = reqWithUser(req, adminUser())
	w := httptest.NewRecorder()

	handler.HandlerListNe(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", w.Code)
	}
}

func TestHandlerListNe_GroupOnly(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(_ string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 2}, nil
		},
		GetAllNeOfUserByUserIdFn: func(_ int64) ([]*db_models.CliUserNeMapping, error) {
			return nil, nil
		},
		GetAllGroupsOfUserFn: func(uid int64) ([]*db_models.CliUserGroupMapping, error) {
			return []*db_models.CliUserGroupMapping{{UserID: uid, GroupID: 42}}, nil
		},
		GetAllNesOfGroupFn: func(gid int64) ([]*db_models.CliGroupNeMapping, error) {
			return []*db_models.CliGroupNeMapping{{GroupID: gid, TblNeID: 99}}, nil
		},
		GetCliNeByNeIdFn: func(id int64) (*db_models.CliNe, error) {
			return &db_models.CliNe{ID: id, NeName: "only-via-group", SiteName: "X"}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/aa/list/ne", nil)
	req = reqWithUser(req, adminUser())
	w := httptest.NewRecorder()

	handler.HandlerListNe(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("status: got %d, want 302; body=%s", w.Code, w.Body.String())
	}
}
