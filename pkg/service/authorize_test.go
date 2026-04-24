package service_test

import (
	"testing"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/testutil"
)

// setupAuthorize installs a fresh MockStore with a single user, NE and
// command already wired up. Returns the mock plus the IDs the test will need.
func setupAuthorize(t *testing.T) (*testutil.MockStore, int64, int64, int64) {
	t.Helper()
	m := testutil.NewMockStore()
	store.SetSingleton(m)

	u := &db_models.User{Username: "alice", IsEnabled: true}
	if err := m.CreateUser(u); err != nil {
		t.Fatalf("create user: %v", err)
	}
	n := &db_models.NE{Namespace: "htsmf01", NeType: "SMF"}
	if err := m.CreateNE(n); err != nil {
		t.Fatalf("create ne: %v", err)
	}
	c := &db_models.Command{NeID: n.ID, Service: db_models.CommandServiceNeConfig, CmdText: "show version"}
	if err := m.CreateCommand(c); err != nil {
		t.Fatalf("create cmd: %v", err)
	}
	return m, u.ID, n.ID, c.ID
}

func TestAuthorize_DeniesWhenNoGroupMemberships(t *testing.T) {
	_, _, neID, cmdID := setupAuthorize(t)

	d, err := service.Authorize("alice", neID, cmdID)
	if err != nil {
		t.Fatalf("Authorize error: %v", err)
	}
	if d.Allowed {
		t.Fatalf("expected deny when user is in no groups, got allow")
	}
	if !d.UserExists || !d.UserEnabled || !d.CommandOnNe {
		t.Errorf("trace should confirm user + command-on-NE, got %+v", d)
	}
	if d.NeReachable {
		t.Errorf("NeReachable should be false without a matching ne_access_group")
	}
}

func TestAuthorize_AllowsWhenBothGroupsMatch(t *testing.T) {
	m, uID, neID, cmdID := setupAuthorize(t)

	nag := &db_models.NeAccessGroup{Name: "all-ne"}
	if err := m.CreateNeAccessGroup(nag); err != nil {
		t.Fatal(err)
	}
	if err := m.AddUserToNeAccessGroup(nag.ID, uID); err != nil {
		t.Fatal(err)
	}
	if err := m.AddNeToNeAccessGroup(nag.ID, neID); err != nil {
		t.Fatal(err)
	}

	ceg := &db_models.CmdExecGroup{Name: "show-only"}
	if err := m.CreateCmdExecGroup(ceg); err != nil {
		t.Fatal(err)
	}
	if err := m.AddUserToCmdExecGroup(ceg.ID, uID); err != nil {
		t.Fatal(err)
	}
	if err := m.AddCommandToCmdExecGroup(ceg.ID, cmdID); err != nil {
		t.Fatal(err)
	}

	d, err := service.Authorize("alice", neID, cmdID)
	if err != nil {
		t.Fatalf("Authorize: %v", err)
	}
	if !d.Allowed {
		t.Fatalf("expected allow when both layers grant, got deny (%s)", d.Reason)
	}
}

func TestAuthorize_DeniesWhenUserLocked(t *testing.T) {
	m, uID, neID, cmdID := setupAuthorize(t)
	u, _ := m.GetUserByID(uID)
	now := nowForTest()
	u.LockedAt = &now
	_ = m.UpdateUser(u)

	d, err := service.Authorize("alice", neID, cmdID)
	if err != nil {
		t.Fatalf("Authorize: %v", err)
	}
	if d.Allowed {
		t.Fatalf("expected deny when user is locked")
	}
}

func TestAuthorize_DeniesWhenCommandRegisteredOnDifferentNe(t *testing.T) {
	m, uID, neID, _ := setupAuthorize(t)

	otherNe := &db_models.NE{Namespace: "htamf01", NeType: "AMF"}
	_ = m.CreateNE(otherNe)
	foreignCmd := &db_models.Command{NeID: otherNe.ID, Service: db_models.CommandServiceNeConfig, CmdText: "reload"}
	_ = m.CreateCommand(foreignCmd)

	// Wire user into groups that would normally allow everything.
	nag := &db_models.NeAccessGroup{Name: "g"}
	_ = m.CreateNeAccessGroup(nag)
	_ = m.AddUserToNeAccessGroup(nag.ID, uID)
	_ = m.AddNeToNeAccessGroup(nag.ID, neID)
	ceg := &db_models.CmdExecGroup{Name: "g"}
	_ = m.CreateCmdExecGroup(ceg)
	_ = m.AddUserToCmdExecGroup(ceg.ID, uID)
	_ = m.AddCommandToCmdExecGroup(ceg.ID, foreignCmd.ID)

	// Targeting (alice, neID, foreignCmd.ID) must fail because the command is
	// bound to otherNe, not neID.
	d, err := service.Authorize("alice", neID, foreignCmd.ID)
	if err != nil {
		t.Fatalf("Authorize: %v", err)
	}
	if d.Allowed {
		t.Fatalf("expected deny when command belongs to a different NE")
	}
}
