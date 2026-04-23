package service_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/testutil"
)

// rbacFixture builds a store mock covering every RBAC read path the
// evaluator touches. Tests only need to set the slices they care about;
// the rest default to empty. GetXById lookups derive from the seeded
// slices so indices stay in sync without hand-coded maps.
type rbacFixture struct {
	nes         []*db_models.CliNe
	profiles    []*db_models.CliNeProfile
	userGroups  map[int64][]*db_models.CliUserGroupMapping
	userNes     map[int64][]*db_models.CliUserNeMapping
	groupNes    map[int64][]*db_models.CliGroupNeMapping
	perms       map[int64][]*db_models.CliGroupCmdPermission
	defs        []*db_models.CliCommandDef
	cmdGroups   []*db_models.CliCommandGroup
	groupDefs   map[int64][]*db_models.CliCommandDef
}

func (f *rbacFixture) install() {
	store.SetSingleton(&testutil.MockStore{
		GetCliNeByNeIdFn: func(id int64) (*db_models.CliNe, error) {
			for _, n := range f.nes {
				if n.ID == id {
					return n, nil
				}
			}
			return nil, nil
		},
		GetNeProfileByIdFn: func(id int64) (*db_models.CliNeProfile, error) {
			for _, p := range f.profiles {
				if p.ID == id {
					return p, nil
				}
			}
			return nil, nil
		},
		GetAllGroupsOfUserFn: func(uid int64) ([]*db_models.CliUserGroupMapping, error) {
			return f.userGroups[uid], nil
		},
		GetAllNeOfUserByUserIdFn: func(uid int64) ([]*db_models.CliUserNeMapping, error) {
			return f.userNes[uid], nil
		},
		GetAllNesOfGroupFn: func(gid int64) ([]*db_models.CliGroupNeMapping, error) {
			return f.groupNes[gid], nil
		},
		ListGroupCmdPermissionsFn: func(gid int64) ([]*db_models.CliGroupCmdPermission, error) {
			return f.perms[gid], nil
		},
		ListCommandDefsFn: func(svc, prof, cat string) ([]*db_models.CliCommandDef, error) {
			out := []*db_models.CliCommandDef{}
			for _, d := range f.defs {
				if svc != "" && d.Service != svc {
					continue
				}
				if prof != "" && d.NeProfile != prof {
					continue
				}
				if cat != "" && d.Category != cat {
					continue
				}
				out = append(out, d)
			}
			return out, nil
		},
		GetCommandGroupByNameFn: func(name string) (*db_models.CliCommandGroup, error) {
			for _, g := range f.cmdGroups {
				if g.Name == name {
					return g, nil
				}
			}
			return nil, nil
		},
		ListCommandsOfGroupFn: func(gid int64) ([]*db_models.CliCommandDef, error) {
			return f.groupDefs[gid], nil
		},
	})
}

// helper constructors — keep tests concise.

func p(id int64, name string) *db_models.CliNeProfile {
	return &db_models.CliNeProfile{ID: id, Name: name}
}

func ne(id int64, name string, profileID *int64) *db_models.CliNe {
	return &db_models.CliNe{ID: id, NeName: name, NeProfileID: profileID}
}

func i64p(v int64) *int64 { return &v }

func perm(gid int64, service, scope, grantType, grantValue, effect string) *db_models.CliGroupCmdPermission {
	return &db_models.CliGroupCmdPermission{
		GroupID: gid, Service: service, NeScope: scope,
		GrantType: grantType, GrantValue: grantValue, Effect: effect,
	}
}

// ─── NE access ────────────────────────────────────────────────────────────

func TestCheckCommand_NoNeAccess(t *testing.T) {
	f := &rbacFixture{
		nes:      []*db_models.CliNe{ne(1, "SMF-01", nil)},
		userNes:  map[int64][]*db_models.CliUserNeMapping{},
		userGroups: map[int64][]*db_models.CliUserGroupMapping{},
	}
	f.install()
	res, err := service.CheckCommand(42, 1, "ne-command", "show version")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if res.Allowed {
		t.Errorf("expected deny, got allow; reason=%q", res.Reason)
	}
	if !strings.Contains(res.Reason, "no access") {
		t.Errorf("reason should mention access, got %q", res.Reason)
	}
}

func TestCheckCommand_DirectUserNeMapping(t *testing.T) {
	f := &rbacFixture{
		nes: []*db_models.CliNe{ne(1, "SMF-01", nil)},
		userNes: map[int64][]*db_models.CliUserNeMapping{
			42: {{UserID: 42, TblNeID: 1}},
		},
		userGroups: map[int64][]*db_models.CliUserGroupMapping{
			42: {{UserID: 42, GroupID: 5}},
		},
		perms: map[int64][]*db_models.CliGroupCmdPermission{
			5: {perm(5, "*", "*", "pattern", "show version", "allow")},
		},
	}
	f.install()
	res, _ := service.CheckCommand(42, 1, "ne-command", "show version")
	if !res.Allowed {
		t.Errorf("expected allow, got deny; reason=%q", res.Reason)
	}
}

// ─── Implicit deny ────────────────────────────────────────────────────────

func TestCheckCommand_ImplicitDeny(t *testing.T) {
	f := &rbacFixture{
		nes: []*db_models.CliNe{ne(1, "SMF-01", nil)},
		userGroups: map[int64][]*db_models.CliUserGroupMapping{
			42: {{UserID: 42, GroupID: 5}},
		},
		groupNes: map[int64][]*db_models.CliGroupNeMapping{
			5: {{GroupID: 5, TblNeID: 1}},
		},
		perms: map[int64][]*db_models.CliGroupCmdPermission{
			5: {}, // no rules at all
		},
	}
	f.install()
	res, _ := service.CheckCommand(42, 1, "ne-command", "show version")
	if res.Allowed {
		t.Errorf("expected implicit deny, got allow")
	}
	if !strings.Contains(res.Reason, "implicit deny") {
		t.Errorf("reason should be implicit deny, got %q", res.Reason)
	}
}

// ─── Scope specificity: ne:X > profile:Y > * ────────────────────────────

func TestCheckCommand_SpecificScope_BeatsBroader(t *testing.T) {
	smfID := i64p(1)
	f := &rbacFixture{
		profiles: []*db_models.CliNeProfile{p(1, "SMF")},
		nes:      []*db_models.CliNe{ne(1, "SMF-01", smfID)},
		userGroups: map[int64][]*db_models.CliUserGroupMapping{
			42: {{UserID: 42, GroupID: 5}},
		},
		groupNes: map[int64][]*db_models.CliGroupNeMapping{
			5: {{GroupID: 5, TblNeID: 1}},
		},
		perms: map[int64][]*db_models.CliGroupCmdPermission{
			5: {
				// Broad allow, but a specific deny on this NE should win.
				perm(5, "*", "*", "pattern", "delete *", "allow"),
				perm(5, "*", "ne:SMF-01", "pattern", "delete *", "deny"),
			},
		},
	}
	f.install()
	res, _ := service.CheckCommand(42, 1, "ne-command", "delete session 123")
	if res.Allowed {
		t.Errorf("ne:X deny should beat * allow; reason=%q", res.Reason)
	}
	if !strings.Contains(res.Reason, "denied by") {
		t.Errorf("reason %q", res.Reason)
	}
}

func TestCheckCommand_SpecificAllow_OverridesBroaderDeny(t *testing.T) {
	smfID := i64p(1)
	f := &rbacFixture{
		profiles: []*db_models.CliNeProfile{p(1, "SMF")},
		nes:      []*db_models.CliNe{ne(1, "SMF-01", smfID)},
		userGroups: map[int64][]*db_models.CliUserGroupMapping{
			42: {{UserID: 42, GroupID: 5}},
		},
		groupNes: map[int64][]*db_models.CliGroupNeMapping{
			5: {{GroupID: 5, TblNeID: 1}},
		},
		perms: map[int64][]*db_models.CliGroupCmdPermission{
			5: {
				// Broad deny at *, but a specific allow at ne:SMF-01 wins.
				perm(5, "*", "*", "pattern", "get subscriber", "deny"),
				perm(5, "*", "ne:SMF-01", "pattern", "get subscriber", "allow"),
			},
		},
	}
	f.install()
	res, _ := service.CheckCommand(42, 1, "ne-command", "get subscriber")
	if !res.Allowed {
		t.Errorf("ne:X allow should override * deny; reason=%q", res.Reason)
	}
}

// Same-scope deny beats allow (AWS-IAM).
func TestCheckCommand_SameScope_DenyBeatsAllow(t *testing.T) {
	f := &rbacFixture{
		nes: []*db_models.CliNe{ne(1, "SMF-01", nil)},
		userGroups: map[int64][]*db_models.CliUserGroupMapping{
			42: {{UserID: 42, GroupID: 5}},
		},
		groupNes: map[int64][]*db_models.CliGroupNeMapping{
			5: {{GroupID: 5, TblNeID: 1}},
		},
		perms: map[int64][]*db_models.CliGroupCmdPermission{
			5: {
				perm(5, "*", "*", "pattern", "show version", "allow"),
				perm(5, "*", "*", "pattern", "show version", "deny"),
			},
		},
	}
	f.install()
	res, _ := service.CheckCommand(42, 1, "ne-command", "show version")
	if res.Allowed {
		t.Errorf("same-scope deny must beat allow; reason=%q", res.Reason)
	}
}

// Profile-scope rule applies only to NEs with the matching profile.
func TestCheckCommand_ProfileScope_Match(t *testing.T) {
	smfID := i64p(1)
	f := &rbacFixture{
		profiles: []*db_models.CliNeProfile{p(1, "SMF")},
		nes:      []*db_models.CliNe{ne(1, "SMF-01", smfID)},
		userGroups: map[int64][]*db_models.CliUserGroupMapping{
			42: {{UserID: 42, GroupID: 5}},
		},
		groupNes: map[int64][]*db_models.CliGroupNeMapping{
			5: {{GroupID: 5, TblNeID: 1}},
		},
		perms: map[int64][]*db_models.CliGroupCmdPermission{
			5: {perm(5, "*", "profile:SMF", "pattern", "get subscriber", "allow")},
		},
	}
	f.install()
	res, _ := service.CheckCommand(42, 1, "ne-command", "get subscriber")
	if !res.Allowed {
		t.Errorf("profile:SMF rule should match SMF NE; reason=%q", res.Reason)
	}
}

func TestCheckCommand_ProfileScope_NoMatch(t *testing.T) {
	amfID := i64p(2)
	f := &rbacFixture{
		profiles: []*db_models.CliNeProfile{p(1, "SMF"), p(2, "AMF")},
		nes:      []*db_models.CliNe{ne(1, "AMF-01", amfID)},
		userGroups: map[int64][]*db_models.CliUserGroupMapping{
			42: {{UserID: 42, GroupID: 5}},
		},
		groupNes: map[int64][]*db_models.CliGroupNeMapping{
			5: {{GroupID: 5, TblNeID: 1}},
		},
		perms: map[int64][]*db_models.CliGroupCmdPermission{
			5: {perm(5, "*", "profile:SMF", "pattern", "get subscriber", "allow")},
		},
	}
	f.install()
	res, _ := service.CheckCommand(42, 1, "ne-command", "get subscriber")
	if res.Allowed {
		t.Errorf("profile:SMF rule must not match AMF NE; reason=%q", res.Reason)
	}
}

// ─── Pattern matching ─────────────────────────────────────────────────────

func TestCheckCommand_ExactMatch(t *testing.T) {
	f := rbacSimpleFixture(t, "pattern", "show version", "allow")
	res, _ := service.CheckCommand(42, 1, "ne-command", "show version")
	if !res.Allowed {
		t.Errorf("exact match must allow; reason=%q", res.Reason)
	}
	_ = f
}

func TestCheckCommand_WildcardSuffix(t *testing.T) {
	_ = rbacSimpleFixture(t, "pattern", "show *", "allow")
	inputs := map[string]bool{
		"show version":          true,
		"show running-config":   true,
		"show interface Gi0/0":  true,
		"configure terminal":    false,
		"show":                  true, // exact prefix match (no trailing space → still matches)
	}
	for cmd, want := range inputs {
		res, _ := service.CheckCommand(42, 1, "ne-command", cmd)
		if res.Allowed != want {
			t.Errorf("cmd=%q: got allowed=%v want %v; reason=%q", cmd, res.Allowed, want, res.Reason)
		}
	}
}

func TestCheckCommand_PrefixWithoutWildcard(t *testing.T) {
	_ = rbacSimpleFixture(t, "pattern", "get subscriber", "allow")
	inputs := map[string]bool{
		"get subscriber":                true,
		"get subscriber 12345":          true,  // prefix match
		"get subscriber detail imsi=1": true,  // prefix match
		"get subscriber-extended":       false, // boundary not whitespace
		"get session":                   false,
	}
	for cmd, want := range inputs {
		res, _ := service.CheckCommand(42, 1, "ne-command", cmd)
		if res.Allowed != want {
			t.Errorf("cmd=%q: got allowed=%v want %v", cmd, res.Allowed, want)
		}
	}
}

// ─── Service filter ───────────────────────────────────────────────────────

func TestCheckCommand_ServiceMismatch_Ignored(t *testing.T) {
	// Rule for ne-config should not match when evaluating ne-command.
	f := &rbacFixture{
		nes: []*db_models.CliNe{ne(1, "R-01", nil)},
		userGroups: map[int64][]*db_models.CliUserGroupMapping{
			42: {{UserID: 42, GroupID: 5}},
		},
		groupNes: map[int64][]*db_models.CliGroupNeMapping{
			5: {{GroupID: 5, TblNeID: 1}},
		},
		perms: map[int64][]*db_models.CliGroupCmdPermission{
			5: {perm(5, "ne-config", "*", "pattern", "show version", "allow")},
		},
	}
	f.install()
	res, _ := service.CheckCommand(42, 1, "ne-command", "show version")
	if res.Allowed {
		t.Errorf("ne-config rule must not apply to ne-command; reason=%q", res.Reason)
	}
}

// ─── Grant-type expansion ────────────────────────────────────────────────

func TestCheckCommand_CategoryGrant_ExpandsDefs(t *testing.T) {
	f := &rbacFixture{
		nes: []*db_models.CliNe{ne(1, "R-01", nil)},
		userGroups: map[int64][]*db_models.CliUserGroupMapping{
			42: {{UserID: 42, GroupID: 5}},
		},
		groupNes: map[int64][]*db_models.CliGroupNeMapping{
			5: {{GroupID: 5, TblNeID: 1}},
		},
		perms: map[int64][]*db_models.CliGroupCmdPermission{
			5: {perm(5, "ne-command", "*", "category", "monitoring", "allow")},
		},
		defs: []*db_models.CliCommandDef{
			{ID: 1, Service: "ne-command", NeProfile: "*", Pattern: "show version", Category: "monitoring"},
			{ID: 2, Service: "ne-command", NeProfile: "*", Pattern: "ping *", Category: "monitoring"},
			{ID: 3, Service: "ne-command", NeProfile: "*", Pattern: "reload", Category: "admin"},
		},
	}
	f.install()

	res, _ := service.CheckCommand(42, 1, "ne-command", "show version")
	if !res.Allowed {
		t.Errorf("monitoring category should cover 'show version'; reason=%q", res.Reason)
	}
	res, _ = service.CheckCommand(42, 1, "ne-command", "reload")
	if res.Allowed {
		t.Errorf("monitoring category must NOT cover 'reload' (admin category); reason=%q", res.Reason)
	}
}

func TestCheckCommand_CommandGroupGrant_ExpandsMembers(t *testing.T) {
	f := &rbacFixture{
		nes: []*db_models.CliNe{ne(1, "R-01", nil)},
		userGroups: map[int64][]*db_models.CliUserGroupMapping{
			42: {{UserID: 42, GroupID: 5}},
		},
		groupNes: map[int64][]*db_models.CliGroupNeMapping{
			5: {{GroupID: 5, TblNeID: 1}},
		},
		perms: map[int64][]*db_models.CliGroupCmdPermission{
			5: {perm(5, "*", "*", "command_group", "monitoring-bundle", "allow")},
		},
		cmdGroups: []*db_models.CliCommandGroup{
			{ID: 10, Name: "monitoring-bundle"},
		},
		groupDefs: map[int64][]*db_models.CliCommandDef{
			10: {
				{ID: 1, Pattern: "show version"},
				{ID: 2, Pattern: "ping *"},
			},
		},
	}
	f.install()

	if res, _ := service.CheckCommand(42, 1, "ne-command", "ping 8.8.8.8"); !res.Allowed {
		t.Errorf("command_group member should allow; reason=%q", res.Reason)
	}
	if res, _ := service.CheckCommand(42, 1, "ne-command", "reload"); res.Allowed {
		t.Errorf("non-member command must not be allowed; reason=%q", res.Reason)
	}
}

// ─── CRUD validator tests (quick, no store needed) ───────────────────────

func TestCreateNeProfile_NameRequired(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{})
	if err := service.CreateNeProfile(&db_models.CliNeProfile{}); err == nil {
		t.Errorf("empty name should error")
	}
	if err := service.CreateNeProfile(&db_models.CliNeProfile{Name: "SMF"}); err != nil {
		t.Errorf("valid create: %v", err)
	}
}

func TestCreateCommandDef_RequiredFields(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{})
	// Missing service.
	if err := service.CreateCommandDef(&db_models.CliCommandDef{Pattern: "x", Category: "monitoring"}); err == nil {
		t.Errorf("missing service should error")
	}
	// Missing pattern.
	if err := service.CreateCommandDef(&db_models.CliCommandDef{Service: "ne-command", Category: "monitoring"}); err == nil {
		t.Errorf("missing pattern should error")
	}
	// Missing category.
	if err := service.CreateCommandDef(&db_models.CliCommandDef{Service: "ne-command", Pattern: "x"}); err == nil {
		t.Errorf("missing category should error")
	}
	// Happy path defaults ne_profile to "*".
	d := &db_models.CliCommandDef{Service: "ne-command", Pattern: "show version", Category: "monitoring"}
	if err := service.CreateCommandDef(d); err != nil {
		t.Fatalf("happy path: %v", err)
	}
	if d.NeProfile != "*" {
		t.Errorf("ne_profile default: got %q want *", d.NeProfile)
	}
}

func TestCreateGroupCmdPermission_Validates(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{})
	// Missing group_id.
	if err := service.CreateGroupCmdPermission(&db_models.CliGroupCmdPermission{Effect: "allow", GrantType: "pattern", GrantValue: "x"}); err == nil {
		t.Errorf("missing group_id should error")
	}
	// Bad effect.
	if err := service.CreateGroupCmdPermission(&db_models.CliGroupCmdPermission{GroupID: 1, Effect: "maybe", GrantType: "pattern", GrantValue: "x"}); err == nil {
		t.Errorf("bad effect should error")
	}
	// Bad grant_type.
	if err := service.CreateGroupCmdPermission(&db_models.CliGroupCmdPermission{GroupID: 1, Effect: "allow", GrantType: "whatever", GrantValue: "x"}); err == nil {
		t.Errorf("bad grant_type should error")
	}
	// Missing grant_value.
	if err := service.CreateGroupCmdPermission(&db_models.CliGroupCmdPermission{GroupID: 1, Effect: "allow", GrantType: "pattern"}); err == nil {
		t.Errorf("missing grant_value should error")
	}
	// Happy path defaults scope and service.
	p := &db_models.CliGroupCmdPermission{GroupID: 1, Effect: "allow", GrantType: "pattern", GrantValue: "show version"}
	if err := service.CreateGroupCmdPermission(p); err != nil {
		t.Fatalf("happy: %v", err)
	}
	if p.NeScope != "*" {
		t.Errorf("ne_scope default: got %q want *", p.NeScope)
	}
	if p.Service != "*" {
		t.Errorf("service default: got %q want *", p.Service)
	}
}

// ─── GetEffectivePermissions ──────────────────────────────────────────────

func TestGetEffectivePermissions_UnionsAcrossGroups(t *testing.T) {
	smfID := i64p(1)
	f := &rbacFixture{
		profiles: []*db_models.CliNeProfile{p(1, "SMF")},
		nes:      []*db_models.CliNe{ne(1, "SMF-01", smfID), ne(2, "SMF-02", smfID)},
		userGroups: map[int64][]*db_models.CliUserGroupMapping{
			42: {
				{UserID: 42, GroupID: 5},
				{UserID: 42, GroupID: 6},
			},
		},
		groupNes: map[int64][]*db_models.CliGroupNeMapping{
			5: {{GroupID: 5, TblNeID: 1}},
			6: {{GroupID: 6, TblNeID: 2}},
		},
		perms: map[int64][]*db_models.CliGroupCmdPermission{
			5: {perm(5, "ne-command", "*", "pattern", "show version", "allow")},
			6: {perm(6, "ne-config", "profile:SMF", "pattern", "get-config *", "allow")},
		},
	}
	f.install()
	resp, err := service.GetEffectivePermissions(42, "alice")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(resp.Nes) != 2 {
		t.Errorf("expected 2 NEs (union of groups), got %d", len(resp.Nes))
	}
	for _, n := range resp.Nes {
		if n.NeProfile != "SMF" {
			t.Errorf("NE %s: expected profile SMF, got %q", n.NeName, n.NeProfile)
		}
	}
	if len(resp.Commands["ne-command"]) != 1 || len(resp.Commands["ne-config"]) != 1 {
		t.Errorf("expected 1 entry per service, got %+v", resp.Commands)
	}
}

// ─── shared helpers ──────────────────────────────────────────────────────

// rbacSimpleFixture builds the minimal single-rule fixture used by pattern
// tests — one NE, one group, one permission.
func rbacSimpleFixture(t *testing.T, grantType, grantValue, effect string) *rbacFixture {
	t.Helper()
	f := &rbacFixture{
		nes: []*db_models.CliNe{ne(1, "R-01", nil)},
		userGroups: map[int64][]*db_models.CliUserGroupMapping{
			42: {{UserID: 42, GroupID: 5}},
		},
		groupNes: map[int64][]*db_models.CliGroupNeMapping{
			5: {{GroupID: 5, TblNeID: 1}},
		},
		perms: map[int64][]*db_models.CliGroupCmdPermission{
			5: {perm(5, "*", "*", grantType, grantValue, effect)},
		},
	}
	f.install()
	return f
}

// Keep fmt import live without a direct test use above — allows future
// error-format assertions without adding another `errors.As`-style guard.
var _ = fmt.Errorf
