package sshcli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

// HTTP client methods for the RBAC endpoints added by docs/rbac-design.md.

// ── NE Profile ──

func (c *MgtClient) ListNeProfiles() ([]*db_models.CliNeProfile, error) {
	raw, _, err := c.do(http.MethodGet, "/aa/ne-profile/list", nil)
	if err != nil {
		return nil, err
	}
	var out []*db_models.CliNeProfile
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode ne_profiles: %w", err)
	}
	return out, nil
}

func (c *MgtClient) CreateNeProfile(fields map[string]any) error {
	_, _, err := c.do(http.MethodPost, "/aa/ne-profile/create", fields)
	return err
}

func (c *MgtClient) UpdateNeProfile(fields map[string]any) error {
	_, _, err := c.do(http.MethodPost, "/aa/ne-profile/update", fields)
	return err
}

func (c *MgtClient) DeleteNeProfileByID(id int64) error {
	_, _, err := c.do(http.MethodDelete, "/aa/ne-profile/"+strconv.FormatInt(id, 10), nil)
	return err
}

// ResolveNeProfileID returns a profile id given either a numeric id or a name.
func (c *MgtClient) ResolveNeProfileID(target string) (int64, error) {
	if id, err := strconv.ParseInt(target, 10, 64); err == nil {
		return id, nil
	}
	profiles, err := c.ListNeProfiles()
	if err != nil {
		return 0, err
	}
	for _, p := range profiles {
		if p.Name == target {
			return p.ID, nil
		}
	}
	return 0, fmt.Errorf("no ne_profile with name or id %q", target)
}

// AssignNeProfile sets cli_ne.ne_profile_id for the given NE.
func (c *MgtClient) AssignNeProfile(neID int64, profileID *int64) error {
	body := map[string]any{}
	if profileID != nil {
		body["ne_profile_id"] = *profileID
	} else {
		body["ne_profile_id"] = nil
	}
	_, _, err := c.do(http.MethodPost, fmt.Sprintf("/aa/ne/%d/profile", neID), body)
	return err
}

// ── Command Def ──

func (c *MgtClient) ListCommandDefs(service, neProfile, category string) ([]*db_models.CliCommandDef, error) {
	q := url.Values{}
	if service != "" {
		q.Set("service", service)
	}
	if neProfile != "" {
		q.Set("ne_profile", neProfile)
	}
	if category != "" {
		q.Set("category", category)
	}
	path := "/aa/command-def/list"
	if enc := q.Encode(); enc != "" {
		path += "?" + enc
	}
	raw, _, err := c.do(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var out []*db_models.CliCommandDef
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode command_defs: %w", err)
	}
	return out, nil
}

func (c *MgtClient) CreateCommandDef(fields map[string]any) error {
	_, _, err := c.do(http.MethodPost, "/aa/command-def/create", fields)
	return err
}

func (c *MgtClient) UpdateCommandDef(fields map[string]any) error {
	_, _, err := c.do(http.MethodPost, "/aa/command-def/update", fields)
	return err
}

func (c *MgtClient) DeleteCommandDefByID(id int64) error {
	_, _, err := c.do(http.MethodDelete, "/aa/command-def/"+strconv.FormatInt(id, 10), nil)
	return err
}

// ── Command Group ──

func (c *MgtClient) ListCommandGroups(service, neProfile string) ([]*db_models.CliCommandGroup, error) {
	q := url.Values{}
	if service != "" {
		q.Set("service", service)
	}
	if neProfile != "" {
		q.Set("ne_profile", neProfile)
	}
	path := "/aa/command-group/list"
	if enc := q.Encode(); enc != "" {
		path += "?" + enc
	}
	raw, _, err := c.do(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var out []*db_models.CliCommandGroup
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode command_groups: %w", err)
	}
	return out, nil
}

func (c *MgtClient) CreateCommandGroup(fields map[string]any) error {
	_, _, err := c.do(http.MethodPost, "/aa/command-group/create", fields)
	return err
}

func (c *MgtClient) UpdateCommandGroup(fields map[string]any) error {
	_, _, err := c.do(http.MethodPost, "/aa/command-group/update", fields)
	return err
}

func (c *MgtClient) DeleteCommandGroupByID(id int64) error {
	_, _, err := c.do(http.MethodDelete, "/aa/command-group/"+strconv.FormatInt(id, 10), nil)
	return err
}

func (c *MgtClient) ListCommandsOfGroup(groupID int64) ([]*db_models.CliCommandDef, error) {
	raw, _, err := c.do(http.MethodGet, fmt.Sprintf("/aa/command-group/%d/commands", groupID), nil)
	if err != nil {
		return nil, err
	}
	var out []*db_models.CliCommandDef
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode commands of group: %w", err)
	}
	return out, nil
}

func (c *MgtClient) AddCommandToGroup(groupID, commandID int64) error {
	_, _, err := c.do(http.MethodPost, fmt.Sprintf("/aa/command-group/%d/commands", groupID),
		map[string]int64{"command_def_id": commandID})
	return err
}

func (c *MgtClient) RemoveCommandFromGroup(groupID, commandID int64) error {
	_, _, err := c.do(http.MethodDelete, fmt.Sprintf("/aa/command-group/%d/commands/%d", groupID, commandID), nil)
	return err
}

// ResolveCommandGroupID resolves a group name or id to its numeric id.
func (c *MgtClient) ResolveCommandGroupID(target string) (int64, error) {
	if id, err := strconv.ParseInt(target, 10, 64); err == nil {
		return id, nil
	}
	gs, err := c.ListCommandGroups("", "")
	if err != nil {
		return 0, err
	}
	for _, g := range gs {
		if g.Name == target {
			return g.ID, nil
		}
	}
	return 0, fmt.Errorf("no command-group with name or id %q", target)
}

// ResolveCommandDefID resolves by numeric id (pattern-based lookup isn't
// meaningful because patterns can repeat across profiles).
func (c *MgtClient) ResolveCommandDefID(target string) (int64, error) {
	if id, err := strconv.ParseInt(target, 10, 64); err == nil {
		return id, nil
	}
	return 0, fmt.Errorf("command-def target must be numeric id, got %q", target)
}

// ── Group Cmd Permission ──

func (c *MgtClient) ListGroupCmdPermissions(groupID int64) ([]*db_models.CliGroupCmdPermission, error) {
	raw, _, err := c.do(http.MethodGet, fmt.Sprintf("/aa/group/%d/cmd-permissions", groupID), nil)
	if err != nil {
		return nil, err
	}
	var out []*db_models.CliGroupCmdPermission
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode cmd_permissions: %w", err)
	}
	return out, nil
}

func (c *MgtClient) CreateGroupCmdPermission(groupID int64, fields map[string]any) error {
	_, _, err := c.do(http.MethodPost, fmt.Sprintf("/aa/group/%d/cmd-permissions", groupID), fields)
	return err
}

func (c *MgtClient) DeleteGroupCmdPermission(groupID, permID int64) error {
	_, _, err := c.do(http.MethodDelete, fmt.Sprintf("/aa/group/%d/cmd-permissions/%d", groupID, permID), nil)
	return err
}
