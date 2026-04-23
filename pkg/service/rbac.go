package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
)

// ─────────────────────────────────────────────────────────────────────────
// CRUD wrappers for the RBAC entities. Service functions log once at the
// failure point so handlers can just forward the error.
// ─────────────────────────────────────────────────────────────────────────

// ────── NE Profile ──────

func CreateNeProfile(p *db_models.CliNeProfile) error {
	if strings.TrimSpace(p.Name) == "" {
		return errors.New("ne_profile name is required")
	}
	return store.GetSingleton().CreateNeProfile(p)
}

func GetNeProfileById(id int64) (*db_models.CliNeProfile, error) {
	return store.GetSingleton().GetNeProfileById(id)
}

func GetNeProfileByName(name string) (*db_models.CliNeProfile, error) {
	return store.GetSingleton().GetNeProfileByName(name)
}

func ListNeProfiles() ([]*db_models.CliNeProfile, error) {
	return store.GetSingleton().ListNeProfiles()
}

func UpdateNeProfile(p *db_models.CliNeProfile) error {
	return store.GetSingleton().UpdateNeProfile(p)
}

func DeleteNeProfileById(id int64) error {
	return store.GetSingleton().DeleteNeProfileById(id)
}

// ────── Command Def ──────

func CreateCommandDef(d *db_models.CliCommandDef) error {
	if strings.TrimSpace(d.Service) == "" {
		return errors.New("service is required")
	}
	if strings.TrimSpace(d.Pattern) == "" {
		return errors.New("pattern is required")
	}
	if strings.TrimSpace(d.Category) == "" {
		return errors.New("category is required")
	}
	if strings.TrimSpace(d.NeProfile) == "" {
		d.NeProfile = db_models.NeScopeAny
	}
	return store.GetSingleton().CreateCommandDef(d)
}

func GetCommandDefById(id int64) (*db_models.CliCommandDef, error) {
	return store.GetSingleton().GetCommandDefById(id)
}

func ListCommandDefs(service, neProfile, category string) ([]*db_models.CliCommandDef, error) {
	return store.GetSingleton().ListCommandDefs(service, neProfile, category)
}

func UpdateCommandDef(d *db_models.CliCommandDef) error {
	return store.GetSingleton().UpdateCommandDef(d)
}

func DeleteCommandDefById(id int64) error {
	return store.GetSingleton().DeleteCommandDefById(id)
}

// ────── Command Group ──────

func CreateCommandGroup(g *db_models.CliCommandGroup) error {
	if strings.TrimSpace(g.Name) == "" {
		return errors.New("command group name is required")
	}
	if strings.TrimSpace(g.Service) == "" {
		g.Service = db_models.CommandServiceAny
	}
	if strings.TrimSpace(g.NeProfile) == "" {
		g.NeProfile = db_models.NeScopeAny
	}
	return store.GetSingleton().CreateCommandGroup(g)
}

func GetCommandGroupById(id int64) (*db_models.CliCommandGroup, error) {
	return store.GetSingleton().GetCommandGroupById(id)
}

func GetCommandGroupByName(name string) (*db_models.CliCommandGroup, error) {
	return store.GetSingleton().GetCommandGroupByName(name)
}

func ListCommandGroups(service, neProfile string) ([]*db_models.CliCommandGroup, error) {
	return store.GetSingleton().ListCommandGroups(service, neProfile)
}

func UpdateCommandGroup(g *db_models.CliCommandGroup) error {
	return store.GetSingleton().UpdateCommandGroup(g)
}

func DeleteCommandGroupById(id int64) error {
	return store.GetSingleton().DeleteCommandGroupById(id)
}

func AddCommandToGroup(groupID, cmdID int64) error {
	return store.GetSingleton().AddCommandToGroup(&db_models.CliCommandGroupMapping{
		CommandGroupID: groupID,
		CommandDefID:   cmdID,
	})
}

func RemoveCommandFromGroup(groupID, cmdID int64) error {
	return store.GetSingleton().RemoveCommandFromGroup(&db_models.CliCommandGroupMapping{
		CommandGroupID: groupID,
		CommandDefID:   cmdID,
	})
}

func ListCommandsOfGroup(groupID int64) ([]*db_models.CliCommandDef, error) {
	return store.GetSingleton().ListCommandsOfGroup(groupID)
}

// ────── Group Cmd Permission ──────

func CreateGroupCmdPermission(p *db_models.CliGroupCmdPermission) error {
	if p.GroupID == 0 {
		return errors.New("group_id is required")
	}
	if p.Effect != db_models.PermissionEffectAllow && p.Effect != db_models.PermissionEffectDeny {
		return fmt.Errorf("effect must be %q or %q", db_models.PermissionEffectAllow, db_models.PermissionEffectDeny)
	}
	switch p.GrantType {
	case db_models.GrantTypeCommandGroup, db_models.GrantTypeCategory, db_models.GrantTypePattern:
	default:
		return fmt.Errorf("grant_type must be one of command_group | category | pattern, got %q", p.GrantType)
	}
	if strings.TrimSpace(p.GrantValue) == "" {
		return errors.New("grant_value is required")
	}
	if strings.TrimSpace(p.NeScope) == "" {
		p.NeScope = db_models.NeScopeAny
	}
	if strings.TrimSpace(p.Service) == "" {
		p.Service = db_models.CommandServiceAny
	}
	return store.GetSingleton().CreateGroupCmdPermission(p)
}

func ListGroupCmdPermissions(groupID int64) ([]*db_models.CliGroupCmdPermission, error) {
	return store.GetSingleton().ListGroupCmdPermissions(groupID)
}

func DeleteGroupCmdPermissionById(id int64) error {
	return store.GetSingleton().DeleteGroupCmdPermissionById(id)
}

// ─────────────────────────────────────────────────────────────────────────
// Permission Evaluator
//
// Implements the 3-step algorithm from docs/rbac-design.md §4.7–4.9:
//   1. NE access check (union of direct user-ne + group-ne mappings)
//   2. Collect applicable rules from every group the user belongs to,
//      filtered by the ne_scope matcher
//   3. Evaluate with (scope-specificity × AWS-IAM "explicit-deny > explicit-
//      allow > implicit-deny") — at the most-specific scope level that has
//      any match, a deny beats an allow. A broader-scope rule can only win
//      when no more-specific rule matches.
// ─────────────────────────────────────────────────────────────────────────

// ScopeLevel orders ne_scope specificity. Higher number = more specific.
type ScopeLevel int

const (
	scopeLevelGlobal  ScopeLevel = 0
	scopeLevelProfile ScopeLevel = 1
	scopeLevelNe      ScopeLevel = 2
)

// EvaluateResult reports the outcome of a CheckCommand evaluation.
type EvaluateResult struct {
	Allowed     bool
	NeName      string
	NeProfile   string
	MatchedRule string
	Reason      string
	RiskLevel   int32
}

// CheckCommand answers "is user X allowed to run <command> on <ne_id> in
// <service>?". The returned reason is human-readable for CLI/API consumers.
func CheckCommand(userID, neID int64, service, command string) (*EvaluateResult, error) {
	// 1. NE access check.
	nes, err := GetNesReachableByUser(userID)
	if err != nil {
		return nil, err
	}
	var target *db_models.CliNe
	for _, n := range nes {
		if n.ID == neID {
			target = n
			break
		}
	}
	if target == nil {
		return &EvaluateResult{Allowed: false, Reason: fmt.Sprintf("no access to NE id %d", neID)}, nil
	}

	// Resolve NE profile name (blank if profile unset).
	profileName := ""
	if target.NeProfileID != nil && *target.NeProfileID > 0 {
		p, err := store.GetSingleton().GetNeProfileById(*target.NeProfileID)
		if err == nil && p != nil {
			profileName = p.Name
		}
	}

	// 2. Collect rules from all groups the user belongs to.
	rules, err := collectUserRules(userID, service, target.NeName, profileName)
	if err != nil {
		return nil, err
	}

	// 3. For each rule, determine whether the command matches. The winning
	// rule is the one at the highest ScopeLevel with a command match; ties
	// at the same level resolve with deny > allow.
	bestLevel := ScopeLevel(-1)
	bestAllow := ""
	bestDeny := ""
	var riskLevel int32
	for _, r := range rules {
		if !r.matchesScope(target.NeName, profileName) {
			continue
		}
		patterns, err := expandGrant(r.perm, service)
		if err != nil {
			logger.Logger.Errorf("rbac: expand grant for group %d: %v", r.perm.GroupID, err)
			continue
		}
		for _, pat := range patterns {
			if !matchPattern(pat.Pattern, command) {
				continue
			}
			level := r.scopeLevel()
			if level < bestLevel {
				continue
			}
			if level > bestLevel {
				bestLevel = level
				bestAllow, bestDeny = "", ""
			}
			ruleLabel := fmt.Sprintf("%s %s:%s (scope:%s, group_id:%d)",
				r.perm.Effect, r.perm.GrantType, r.perm.GrantValue, r.perm.NeScope, r.perm.GroupID)
			if r.perm.Effect == db_models.PermissionEffectDeny {
				if bestDeny == "" {
					bestDeny = ruleLabel
				}
			} else {
				if bestAllow == "" {
					bestAllow = ruleLabel
					riskLevel = pat.RiskLevel
				}
			}
		}
	}

	res := &EvaluateResult{NeName: target.NeName, NeProfile: profileName, RiskLevel: riskLevel}
	switch {
	case bestDeny != "":
		res.Allowed = false
		res.MatchedRule = bestDeny
		res.Reason = "denied by " + bestDeny
	case bestAllow != "":
		res.Allowed = true
		res.MatchedRule = bestAllow
		res.Reason = "allowed by " + bestAllow
	default:
		res.Allowed = false
		res.Reason = "no matching rule (implicit deny)"
	}
	return res, nil
}

// EffectivePermissions returns the full set of rules a user has across all
// reachable NEs. Intended for ne-config / ne-command to cache at session
// start and evaluate locally until cache invalidation.
type EffectiveEntry struct {
	Service  string           `json:"service"`
	NeScope  string           `json:"ne_scope"`
	Effect   string           `json:"effect"`
	Patterns []EffectivePat   `json:"patterns"`
	Source   EffectiveSource  `json:"source"`
}

type EffectivePat struct {
	Pattern   string `json:"pattern"`
	RiskLevel int32  `json:"risk_level,omitempty"`
}

type EffectiveSource struct {
	GroupID    int64  `json:"group_id"`
	GrantType  string `json:"grant_type"`
	GrantValue string `json:"grant_value"`
}

type EffectiveNE struct {
	NeID      int64  `json:"ne_id"`
	NeName    string `json:"ne_name"`
	NeProfile string `json:"ne_profile,omitempty"`
}

type EffectiveResponse struct {
	Username string                     `json:"username"`
	Nes      []EffectiveNE              `json:"nes"`
	Commands map[string][]EffectiveEntry `json:"commands"`
}

// GetEffectivePermissions returns all reachable NEs + expanded per-service
// rule set for the given user.
func GetEffectivePermissions(userID int64, username string) (*EffectiveResponse, error) {
	nes, err := GetNesReachableByUser(userID)
	if err != nil {
		return nil, err
	}
	resp := &EffectiveResponse{
		Username: username,
		Commands: map[string][]EffectiveEntry{},
	}
	// Resolve profile names once.
	profileCache := map[int64]string{}
	for _, n := range nes {
		profileName := ""
		if n.NeProfileID != nil && *n.NeProfileID > 0 {
			if cached, ok := profileCache[*n.NeProfileID]; ok {
				profileName = cached
			} else if p, err := store.GetSingleton().GetNeProfileById(*n.NeProfileID); err == nil && p != nil {
				profileName = p.Name
				profileCache[*n.NeProfileID] = p.Name
			}
		}
		resp.Nes = append(resp.Nes, EffectiveNE{NeID: n.ID, NeName: n.NeName, NeProfile: profileName})
	}

	// Collect every permission from every group the user belongs to, then
	// expand grant → patterns. Rules are grouped by (service, ne_scope,
	// effect, grant_value) for the response.
	gms, err := store.GetSingleton().GetAllGroupsOfUser(userID)
	if err != nil {
		return nil, err
	}
	seenGroup := map[int64]bool{}
	for _, gm := range gms {
		if gm == nil || seenGroup[gm.GroupID] {
			continue
		}
		seenGroup[gm.GroupID] = true
		perms, err := store.GetSingleton().ListGroupCmdPermissions(gm.GroupID)
		if err != nil {
			return nil, err
		}
		for _, p := range perms {
			patterns, err := expandGrant(p, p.Service)
			if err != nil {
				logger.Logger.Errorf("rbac: expand grant for group %d: %v", p.GroupID, err)
				continue
			}
			if len(patterns) == 0 {
				continue
			}
			entry := EffectiveEntry{
				Service:  p.Service,
				NeScope:  p.NeScope,
				Effect:   p.Effect,
				Patterns: make([]EffectivePat, 0, len(patterns)),
				Source: EffectiveSource{
					GroupID:    p.GroupID,
					GrantType:  p.GrantType,
					GrantValue: p.GrantValue,
				},
			}
			for _, pat := range patterns {
				entry.Patterns = append(entry.Patterns, EffectivePat{
					Pattern: pat.Pattern, RiskLevel: pat.RiskLevel,
				})
			}
			key := p.Service
			if key == "" {
				key = db_models.CommandServiceAny
			}
			resp.Commands[key] = append(resp.Commands[key], entry)
		}
	}
	return resp, nil
}

// ─── helpers ───

// applicableRule carries a raw permission + a cached scope level.
type applicableRule struct {
	perm     *db_models.CliGroupCmdPermission
	level    ScopeLevel
	levelSet bool
}

func (r *applicableRule) matchesScope(neName, profile string) bool {
	switch {
	case r.perm.NeScope == db_models.NeScopeAny:
		return true
	case strings.HasPrefix(r.perm.NeScope, db_models.NeScopePrefixProfile):
		want := strings.TrimPrefix(r.perm.NeScope, db_models.NeScopePrefixProfile)
		return profile != "" && want == profile
	case strings.HasPrefix(r.perm.NeScope, db_models.NeScopePrefixSpecific):
		want := strings.TrimPrefix(r.perm.NeScope, db_models.NeScopePrefixSpecific)
		return want == neName
	}
	return false
}

func (r *applicableRule) scopeLevel() ScopeLevel {
	if r.levelSet {
		return r.level
	}
	switch {
	case strings.HasPrefix(r.perm.NeScope, db_models.NeScopePrefixSpecific):
		r.level = scopeLevelNe
	case strings.HasPrefix(r.perm.NeScope, db_models.NeScopePrefixProfile):
		r.level = scopeLevelProfile
	default:
		r.level = scopeLevelGlobal
	}
	r.levelSet = true
	return r.level
}

// collectUserRules gathers every permission across every group the user
// belongs to, filtered to the target service (with "*" always matching).
func collectUserRules(userID int64, service, neName, profile string) ([]applicableRule, error) {
	gms, err := store.GetSingleton().GetAllGroupsOfUser(userID)
	if err != nil {
		return nil, err
	}
	seenGroup := map[int64]bool{}
	var out []applicableRule
	for _, gm := range gms {
		if gm == nil || seenGroup[gm.GroupID] {
			continue
		}
		seenGroup[gm.GroupID] = true
		perms, err := store.GetSingleton().ListGroupCmdPermissions(gm.GroupID)
		if err != nil {
			return nil, err
		}
		for _, p := range perms {
			if !serviceMatches(p.Service, service) {
				continue
			}
			out = append(out, applicableRule{perm: p})
		}
	}
	return out, nil
}

func serviceMatches(ruleService, wantService string) bool {
	if ruleService == "" || ruleService == db_models.CommandServiceAny {
		return true
	}
	return ruleService == wantService
}

// patternDef bundles a pattern string with its source risk level so the
// evaluator can surface risk when it picks the allowing rule.
type patternDef struct {
	Pattern   string
	RiskLevel int32
}

// expandGrant resolves a permission's grant into the concrete pattern list
// it covers. For command_group / category grants we look up the underlying
// CliCommandDef rows; for pattern grants we return the pattern directly.
func expandGrant(p *db_models.CliGroupCmdPermission, service string) ([]patternDef, error) {
	switch p.GrantType {
	case db_models.GrantTypePattern:
		return []patternDef{{Pattern: p.GrantValue}}, nil
	case db_models.GrantTypeCategory:
		defs, err := store.GetSingleton().ListCommandDefs(service, "", p.GrantValue)
		if err != nil {
			return nil, err
		}
		out := make([]patternDef, 0, len(defs))
		for _, d := range defs {
			out = append(out, patternDef{Pattern: d.Pattern, RiskLevel: d.RiskLevel})
		}
		return out, nil
	case db_models.GrantTypeCommandGroup:
		grp, err := store.GetSingleton().GetCommandGroupByName(p.GrantValue)
		if err != nil {
			return nil, err
		}
		if grp == nil {
			return nil, nil
		}
		defs, err := store.GetSingleton().ListCommandsOfGroup(grp.ID)
		if err != nil {
			return nil, err
		}
		out := make([]patternDef, 0, len(defs))
		for _, d := range defs {
			out = append(out, patternDef{Pattern: d.Pattern, RiskLevel: d.RiskLevel})
		}
		return out, nil
	}
	return nil, fmt.Errorf("unknown grant_type %q", p.GrantType)
}

// matchPattern is prefix-aware glob matching. The only wildcard is "*" and
// it only appears at the end ("show *") or stands alone ("*"). Pattern
// without "*" matches by exact equality OR by being a full prefix of the
// input followed by whitespace (so "get subscriber" matches "get subscriber
// 12345" but not "get subscriber-detail").
func matchPattern(pattern, command string) bool {
	if pattern == db_models.NeScopeAny {
		return true
	}
	p := strings.TrimSpace(pattern)
	c := strings.TrimSpace(command)
	if strings.HasSuffix(p, "*") {
		prefix := strings.TrimSuffix(p, "*")
		prefix = strings.TrimRight(prefix, " ")
		if prefix == "" {
			return true
		}
		if c == prefix {
			return true
		}
		return strings.HasPrefix(c, prefix+" ") || strings.HasPrefix(c, prefix+"\t")
	}
	if c == p {
		return true
	}
	if strings.HasPrefix(c, p+" ") || strings.HasPrefix(c, p+"\t") {
		return true
	}
	return false
}

// GetNesReachableByUser returns the union of NEs available to a user via
// direct user↔ne mapping AND group↔ne mapping, deduped by id.
func GetNesReachableByUser(userID int64) ([]*db_models.CliNe, error) {
	sto := store.GetSingleton()
	seen := map[int64]bool{}
	var out []*db_models.CliNe

	direct, err := sto.GetAllNeOfUserByUserId(userID)
	if err != nil {
		return nil, err
	}
	for _, m := range direct {
		if m == nil || seen[m.TblNeID] {
			continue
		}
		n, err := sto.GetCliNeByNeId(m.TblNeID)
		if err != nil {
			return nil, err
		}
		if n == nil {
			continue
		}
		seen[n.ID] = true
		out = append(out, n)
	}

	groups, err := sto.GetAllGroupsOfUser(userID)
	if err != nil {
		return nil, err
	}
	for _, gm := range groups {
		if gm == nil {
			continue
		}
		nes, err := sto.GetAllNesOfGroup(gm.GroupID)
		if err != nil {
			return nil, err
		}
		for _, m := range nes {
			if m == nil || seen[m.TblNeID] {
				continue
			}
			n, err := sto.GetCliNeByNeId(m.TblNeID)
			if err != nil {
				return nil, err
			}
			if n == nil {
				continue
			}
			seen[n.ID] = true
			out = append(out, n)
		}
	}
	return out, nil
}
