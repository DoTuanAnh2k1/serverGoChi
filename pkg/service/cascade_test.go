package service_test

import (
	"testing"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/testutil"
)

// seedChain sets up one user, one NE, one command, one group of each kind,
// with the user+NE in the ne_access_group and the user+command in the
// cmd_exec_group — the minimum wiring that satisfies the authorize rule.
func seedChain(t *testing.T, m *testutil.MockStore) (uID, neID, cmdID, nagID, cegID int64) {
	t.Helper()
	u := &db_models.User{Username: "alice", IsEnabled: true}
	if err := m.CreateUser(u); err != nil {
		t.Fatal(err)
	}
	n := &db_models.NE{Namespace: "htsmf01", NeType: "SMF"}
	if err := m.CreateNE(n); err != nil {
		t.Fatal(err)
	}
	c := &db_models.Command{NeID: n.ID, Service: db_models.CommandServiceNeConfig, CmdText: "show ver"}
	if err := m.CreateCommand(c); err != nil {
		t.Fatal(err)
	}
	nag := &db_models.NeAccessGroup{Name: "g-ne"}
	_ = m.CreateNeAccessGroup(nag)
	_ = m.AddUserToNeAccessGroup(nag.ID, u.ID)
	_ = m.AddNeToNeAccessGroup(nag.ID, n.ID)
	ceg := &db_models.CmdExecGroup{Name: "g-cmd"}
	_ = m.CreateCmdExecGroup(ceg)
	_ = m.AddUserToCmdExecGroup(ceg.ID, u.ID)
	_ = m.AddCommandToCmdExecGroup(ceg.ID, c.ID)
	return u.ID, n.ID, c.ID, nag.ID, ceg.ID
}

func TestDeleteUser_CleansMemberships(t *testing.T) {
	m := testutil.NewMockStore()
	store.SetSingleton(m)
	uID, neID, cmdID, nagID, cegID := seedChain(t, m)

	// Sanity: the authorize rule should hold before we delete.
	d, _ := service.Authorize("alice", neID, cmdID)
	if !d.Allowed {
		t.Fatalf("precondition: expected allow, got %+v", d)
	}

	if err := service.DeleteUser(uID); err != nil {
		t.Fatal(err)
	}
	users, _ := m.ListUsersInNeAccessGroup(nagID)
	if len(users) != 0 {
		t.Errorf("ne_access_group still contains user after delete: %v", users)
	}
	users, _ = m.ListUsersInCmdExecGroup(cegID)
	if len(users) != 0 {
		t.Errorf("cmd_exec_group still contains user after delete: %v", users)
	}
}

func TestDeleteNE_RemovesCommandsAndMemberships(t *testing.T) {
	m := testutil.NewMockStore()
	store.SetSingleton(m)
	_, neID, cmdID, nagID, cegID := seedChain(t, m)

	if err := service.DeleteNE(neID); err != nil {
		t.Fatal(err)
	}
	if c, _ := m.GetCommandByID(cmdID); c != nil {
		t.Errorf("command on deleted NE should be gone, got %+v", c)
	}
	nes, _ := m.ListNEsInNeAccessGroup(nagID)
	if len(nes) != 0 {
		t.Errorf("ne_access_group still contains deleted NE: %v", nes)
	}
	cmds, _ := m.ListCommandsInCmdExecGroup(cegID)
	if len(cmds) != 0 {
		t.Errorf("cmd_exec_group still contains deleted command: %v", cmds)
	}
}

func TestDeleteGroup_EmptiesMemberships(t *testing.T) {
	m := testutil.NewMockStore()
	store.SetSingleton(m)
	_, _, _, nagID, cegID := seedChain(t, m)

	if err := service.DeleteNeAccessGroup(nagID); err != nil {
		t.Fatal(err)
	}
	if g, _ := m.GetNeAccessGroupByID(nagID); g != nil {
		t.Errorf("ne_access_group row survived delete: %+v", g)
	}
	users, _ := m.ListUsersInNeAccessGroup(nagID)
	if len(users) != 0 {
		t.Errorf("ne_access_group user pivot not cleaned: %v", users)
	}

	if err := service.DeleteCmdExecGroup(cegID); err != nil {
		t.Fatal(err)
	}
	if g, _ := m.GetCmdExecGroupByID(cegID); g != nil {
		t.Errorf("cmd_exec_group row survived delete: %+v", g)
	}
}
