// Package testutil — in-memory DatabaseStore for unit tests. Not threadsafe
// for concurrent writers, which is fine because individual tests run
// sequentially. Keep this file in lockstep with store.DatabaseStore; the
// var-assertion at the bottom breaks the build if a method drifts.
package testutil

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/config_models"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
)

type MockStore struct {
	users        map[int64]*db_models.User
	nes          map[int64]*db_models.NE
	commands     map[int64]*db_models.Command
	nags         map[int64]*db_models.NeAccessGroup
	nagUser      map[int64]map[int64]struct{} // groupID → userIDs
	nagNe        map[int64]map[int64]struct{} // groupID → neIDs
	cegs         map[int64]*db_models.CmdExecGroup
	cegUser      map[int64]map[int64]struct{}
	cegCmd       map[int64]map[int64]struct{}
	policy       *db_models.PasswordPolicy
	pwHistory    []*db_models.PasswordHistory
	accessList   []*db_models.UserAccessList
	history      []db_models.OperationHistory
	logins       []db_models.LoginHistory
	backups      map[int64]*db_models.ConfigBackup
	userSeq      int64
	neSeq        int64
	cmdSeq       int64
	nagSeq       int64
	cegSeq       int64
	pwhSeq       int64
	aclSeq       int64
	opHistorySeq int32
	backupSeq    int64
}

func NewMockStore() *MockStore {
	return &MockStore{
		users:    map[int64]*db_models.User{},
		nes:      map[int64]*db_models.NE{},
		commands: map[int64]*db_models.Command{},
		nags:     map[int64]*db_models.NeAccessGroup{},
		nagUser:  map[int64]map[int64]struct{}{},
		nagNe:    map[int64]map[int64]struct{}{},
		cegs:     map[int64]*db_models.CmdExecGroup{},
		cegUser:  map[int64]map[int64]struct{}{},
		cegCmd:   map[int64]map[int64]struct{}{},
		backups:  map[int64]*db_models.ConfigBackup{},
	}
}

// InstallMockStore swaps the singleton to the given mock and returns a
// cleanup function that restores the previous singleton on defer.
func InstallMockStore(m store.DatabaseStore) func() {
	prev := store.GetSingleton()
	store.SetSingleton(m)
	return func() { store.SetSingleton(prev) }
}

func (m *MockStore) Init(_ config_models.DatabaseConfig) error { return nil }
func (m *MockStore) Ping() error                               { return nil }

// ── User ───────────────────────────────────────────────────────────────

func (m *MockStore) CreateUser(u *db_models.User) error {
	if u.ID == 0 {
		m.userSeq++
		u.ID = m.userSeq
	}
	m.users[u.ID] = u
	return nil
}
func (m *MockStore) GetUserByID(id int64) (*db_models.User, error) {
	return m.users[id], nil
}
func (m *MockStore) GetUserByUsername(name string) (*db_models.User, error) {
	for _, u := range m.users {
		if u.Username == name {
			return u, nil
		}
	}
	return nil, nil
}
func (m *MockStore) ListUsers() ([]*db_models.User, error) {
	out := make([]*db_models.User, 0, len(m.users))
	for _, u := range m.users {
		out = append(out, u)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
func (m *MockStore) UpdateUser(u *db_models.User) error {
	if _, ok := m.users[u.ID]; !ok {
		return errors.New("not found")
	}
	m.users[u.ID] = u
	return nil
}
func (m *MockStore) DeleteUserByID(id int64) error {
	delete(m.users, id)
	return nil
}

// ── NE ─────────────────────────────────────────────────────────────────

func (m *MockStore) CreateNE(n *db_models.NE) error {
	if n.ID == 0 {
		m.neSeq++
		n.ID = m.neSeq
	}
	m.nes[n.ID] = n
	return nil
}
func (m *MockStore) GetNEByID(id int64) (*db_models.NE, error) { return m.nes[id], nil }
func (m *MockStore) GetNEByNamespace(ns string) (*db_models.NE, error) {
	for _, n := range m.nes {
		if n.Namespace == ns {
			return n, nil
		}
	}
	return nil, nil
}
func (m *MockStore) ListNEs() ([]*db_models.NE, error) {
	out := make([]*db_models.NE, 0, len(m.nes))
	for _, n := range m.nes {
		out = append(out, n)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
func (m *MockStore) UpdateNE(n *db_models.NE) error {
	if _, ok := m.nes[n.ID]; !ok {
		return errors.New("not found")
	}
	m.nes[n.ID] = n
	return nil
}
func (m *MockStore) DeleteNEByID(id int64) error {
	delete(m.nes, id)
	return nil
}

// ── Command ────────────────────────────────────────────────────────────

func (m *MockStore) CreateCommand(c *db_models.Command) error {
	if c.ID == 0 {
		m.cmdSeq++
		c.ID = m.cmdSeq
	}
	m.commands[c.ID] = c
	return nil
}
func (m *MockStore) GetCommandByID(id int64) (*db_models.Command, error) {
	return m.commands[id], nil
}
func (m *MockStore) GetCommandByTriple(neID int64, service, cmdText string) (*db_models.Command, error) {
	for _, c := range m.commands {
		if c.NeID == neID && c.Service == service && c.CmdText == cmdText {
			return c, nil
		}
	}
	return nil, nil
}
func (m *MockStore) ListCommands(neID int64, service string) ([]*db_models.Command, error) {
	out := make([]*db_models.Command, 0)
	for _, c := range m.commands {
		if neID > 0 && c.NeID != neID {
			continue
		}
		if service != "" && c.Service != service {
			continue
		}
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
func (m *MockStore) UpdateCommand(c *db_models.Command) error {
	if _, ok := m.commands[c.ID]; !ok {
		return errors.New("not found")
	}
	m.commands[c.ID] = c
	return nil
}
func (m *MockStore) DeleteCommandByID(id int64) error {
	delete(m.commands, id)
	return nil
}

// ── NE Access Group ───────────────────────────────────────────────────

func (m *MockStore) CreateNeAccessGroup(g *db_models.NeAccessGroup) error {
	if g.ID == 0 {
		m.nagSeq++
		g.ID = m.nagSeq
	}
	m.nags[g.ID] = g
	return nil
}
func (m *MockStore) GetNeAccessGroupByID(id int64) (*db_models.NeAccessGroup, error) {
	return m.nags[id], nil
}
func (m *MockStore) GetNeAccessGroupByName(name string) (*db_models.NeAccessGroup, error) {
	for _, g := range m.nags {
		if g.Name == name {
			return g, nil
		}
	}
	return nil, nil
}
func (m *MockStore) ListNeAccessGroups() ([]*db_models.NeAccessGroup, error) {
	out := make([]*db_models.NeAccessGroup, 0, len(m.nags))
	for _, g := range m.nags {
		out = append(out, g)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
func (m *MockStore) UpdateNeAccessGroup(g *db_models.NeAccessGroup) error {
	if _, ok := m.nags[g.ID]; !ok {
		return errors.New("not found")
	}
	m.nags[g.ID] = g
	return nil
}
func (m *MockStore) DeleteNeAccessGroupByID(id int64) error {
	delete(m.nags, id)
	delete(m.nagUser, id)
	delete(m.nagNe, id)
	return nil
}
func (m *MockStore) AddUserToNeAccessGroup(gid, uid int64) error {
	addPivot(m.nagUser, gid, uid)
	return nil
}
func (m *MockStore) RemoveUserFromNeAccessGroup(gid, uid int64) error {
	removePivot(m.nagUser, gid, uid)
	return nil
}
func (m *MockStore) ListUsersInNeAccessGroup(gid int64) ([]int64, error) {
	return keysOf(m.nagUser[gid]), nil
}
func (m *MockStore) ListNeAccessGroupsOfUser(uid int64) ([]int64, error) {
	return rightLookup(m.nagUser, uid), nil
}
func (m *MockStore) AddNeToNeAccessGroup(gid, nid int64) error {
	addPivot(m.nagNe, gid, nid)
	return nil
}
func (m *MockStore) RemoveNeFromNeAccessGroup(gid, nid int64) error {
	removePivot(m.nagNe, gid, nid)
	return nil
}
func (m *MockStore) ListNEsInNeAccessGroup(gid int64) ([]int64, error) {
	return keysOf(m.nagNe[gid]), nil
}
func (m *MockStore) ListNeAccessGroupsOfNE(nid int64) ([]int64, error) {
	return rightLookup(m.nagNe, nid), nil
}

// ── Cmd Exec Group ────────────────────────────────────────────────────

func (m *MockStore) CreateCmdExecGroup(g *db_models.CmdExecGroup) error {
	if g.ID == 0 {
		m.cegSeq++
		g.ID = m.cegSeq
	}
	m.cegs[g.ID] = g
	return nil
}
func (m *MockStore) GetCmdExecGroupByID(id int64) (*db_models.CmdExecGroup, error) {
	return m.cegs[id], nil
}
func (m *MockStore) GetCmdExecGroupByName(name string) (*db_models.CmdExecGroup, error) {
	for _, g := range m.cegs {
		if g.Name == name {
			return g, nil
		}
	}
	return nil, nil
}
func (m *MockStore) ListCmdExecGroups() ([]*db_models.CmdExecGroup, error) {
	out := make([]*db_models.CmdExecGroup, 0, len(m.cegs))
	for _, g := range m.cegs {
		out = append(out, g)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
func (m *MockStore) UpdateCmdExecGroup(g *db_models.CmdExecGroup) error {
	if _, ok := m.cegs[g.ID]; !ok {
		return errors.New("not found")
	}
	m.cegs[g.ID] = g
	return nil
}
func (m *MockStore) DeleteCmdExecGroupByID(id int64) error {
	delete(m.cegs, id)
	delete(m.cegUser, id)
	delete(m.cegCmd, id)
	return nil
}
func (m *MockStore) AddUserToCmdExecGroup(gid, uid int64) error {
	addPivot(m.cegUser, gid, uid)
	return nil
}
func (m *MockStore) RemoveUserFromCmdExecGroup(gid, uid int64) error {
	removePivot(m.cegUser, gid, uid)
	return nil
}
func (m *MockStore) ListUsersInCmdExecGroup(gid int64) ([]int64, error) {
	return keysOf(m.cegUser[gid]), nil
}
func (m *MockStore) ListCmdExecGroupsOfUser(uid int64) ([]int64, error) {
	return rightLookup(m.cegUser, uid), nil
}
func (m *MockStore) AddCommandToCmdExecGroup(gid, cid int64) error {
	addPivot(m.cegCmd, gid, cid)
	return nil
}
func (m *MockStore) RemoveCommandFromCmdExecGroup(gid, cid int64) error {
	removePivot(m.cegCmd, gid, cid)
	return nil
}
func (m *MockStore) ListCommandsInCmdExecGroup(gid int64) ([]int64, error) {
	return keysOf(m.cegCmd[gid]), nil
}
func (m *MockStore) ListCmdExecGroupsOfCommand(cid int64) ([]int64, error) {
	return rightLookup(m.cegCmd, cid), nil
}

// ── Password Policy / History ────────────────────────────────────────

func (m *MockStore) GetPasswordPolicy() (*db_models.PasswordPolicy, error) {
	if m.policy == nil {
		return nil, nil
	}
	cp := *m.policy
	return &cp, nil
}
func (m *MockStore) UpsertPasswordPolicy(p *db_models.PasswordPolicy) error {
	p.ID = 1
	cp := *p
	m.policy = &cp
	return nil
}
func (m *MockStore) AppendPasswordHistory(h *db_models.PasswordHistory) error {
	m.pwhSeq++
	h.ID = m.pwhSeq
	if h.ChangedAt.IsZero() {
		h.ChangedAt = time.Now().UTC()
	}
	m.pwHistory = append(m.pwHistory, h)
	return nil
}
func (m *MockStore) GetRecentPasswordHistory(userID int64, limit int) ([]*db_models.PasswordHistory, error) {
	out := make([]*db_models.PasswordHistory, 0)
	for _, h := range m.pwHistory {
		if h.UserID == userID {
			out = append(out, h)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ChangedAt.After(out[j].ChangedAt) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
func (m *MockStore) PrunePasswordHistory(userID int64, keep int) error {
	kept := make([]*db_models.PasswordHistory, 0)
	userEntries := make([]*db_models.PasswordHistory, 0)
	for _, h := range m.pwHistory {
		if h.UserID != userID {
			kept = append(kept, h)
			continue
		}
		userEntries = append(userEntries, h)
	}
	sort.Slice(userEntries, func(i, j int) bool { return userEntries[i].ChangedAt.After(userEntries[j].ChangedAt) })
	if keep > 0 && len(userEntries) > keep {
		userEntries = userEntries[:keep]
	}
	if keep <= 0 {
		userEntries = nil
	}
	m.pwHistory = append(kept, userEntries...)
	return nil
}

// ── User Access List ─────────────────────────────────────────────────

func (m *MockStore) CreateAccessListEntry(e *db_models.UserAccessList) error {
	m.aclSeq++
	e.ID = m.aclSeq
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now().UTC()
	}
	m.accessList = append(m.accessList, e)
	return nil
}
func (m *MockStore) ListAccessListEntries(listType string) ([]*db_models.UserAccessList, error) {
	out := make([]*db_models.UserAccessList, 0)
	for _, e := range m.accessList {
		if listType != "" && e.ListType != listType {
			continue
		}
		out = append(out, e)
	}
	return out, nil
}
func (m *MockStore) DeleteAccessListEntryByID(id int64) error {
	kept := make([]*db_models.UserAccessList, 0, len(m.accessList))
	for _, e := range m.accessList {
		if e.ID == id {
			continue
		}
		kept = append(kept, e)
	}
	m.accessList = kept
	return nil
}

// ── History ──────────────────────────────────────────────────────────

func (m *MockStore) SaveOperationHistory(h db_models.OperationHistory) error {
	m.opHistorySeq++
	if h.ID == 0 {
		h.ID = m.opHistorySeq
	}
	if h.CreatedDate.IsZero() {
		h.CreatedDate = time.Now().UTC()
	}
	m.history = append(m.history, h)
	return nil
}
func (m *MockStore) GetRecentHistory(limit int) ([]db_models.OperationHistory, error) {
	return m.GetRecentHistoryFiltered(limit, "", "", "")
}
func (m *MockStore) GetRecentHistoryFiltered(limit int, scope, neNamespace, account string) ([]db_models.OperationHistory, error) {
	out := make([]db_models.OperationHistory, 0, len(m.history))
	for _, h := range m.history {
		if scope != "" && h.Scope != scope {
			continue
		}
		if neNamespace != "" && h.NeNamespace != neNamespace {
			continue
		}
		if account != "" && !strings.EqualFold(h.Account, account) {
			continue
		}
		out = append(out, h)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedDate.After(out[j].CreatedDate) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
func (m *MockStore) GetDailyOperationHistory(date time.Time) ([]db_models.OperationHistory, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)
	out := make([]db_models.OperationHistory, 0)
	for _, h := range m.history {
		if (h.CreatedDate.Equal(start) || h.CreatedDate.After(start)) && h.CreatedDate.Before(end) {
			out = append(out, h)
		}
	}
	return out, nil
}
func (m *MockStore) DeleteHistoryBefore(cutoff time.Time) (int64, error) {
	kept := make([]db_models.OperationHistory, 0, len(m.history))
	removed := int64(0)
	for _, h := range m.history {
		if h.CreatedDate.Before(cutoff) {
			removed++
			continue
		}
		kept = append(kept, h)
	}
	m.history = kept
	return removed, nil
}
func (m *MockStore) UpdateLoginHistory(username, ip string, t time.Time) error {
	m.logins = append(m.logins, db_models.LoginHistory{Username: username, IPAddress: ip, TimeLogin: t})
	return nil
}

// ── Config Backup ────────────────────────────────────────────────────

func (m *MockStore) SaveConfigBackup(b *db_models.ConfigBackup) error {
	if b.ID == 0 {
		m.backupSeq++
		b.ID = m.backupSeq
	}
	m.backups[b.ID] = b
	return nil
}
func (m *MockStore) ListConfigBackups(neName string) ([]*db_models.ConfigBackup, error) {
	out := make([]*db_models.ConfigBackup, 0, len(m.backups))
	for _, b := range m.backups {
		if neName != "" && b.NeName != neName {
			continue
		}
		out = append(out, b)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}
func (m *MockStore) GetConfigBackupByID(id int64) (*db_models.ConfigBackup, error) {
	return m.backups[id], nil
}

// ── helpers ──────────────────────────────────────────────────────────

func addPivot(p map[int64]map[int64]struct{}, left, right int64) {
	if p[left] == nil {
		p[left] = map[int64]struct{}{}
	}
	p[left][right] = struct{}{}
}

func removePivot(p map[int64]map[int64]struct{}, left, right int64) {
	if p[left] != nil {
		delete(p[left], right)
	}
}

func keysOf(s map[int64]struct{}) []int64 {
	out := make([]int64, 0, len(s))
	for k := range s {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// rightLookup returns every left-key whose right-set contains `right`.
func rightLookup(p map[int64]map[int64]struct{}, right int64) []int64 {
	out := make([]int64, 0)
	for left, rights := range p {
		if _, ok := rights[right]; ok {
			out = append(out, left)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Compile-time check: MockStore implements the full DatabaseStore surface.
var _ store.DatabaseStore = (*MockStore)(nil)
