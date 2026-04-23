package sshcli

import (
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

// ─────────────────────────────────────────────────────────────────────────
// RBAC dispatch handlers: ne-profile, command-def, command-group, and the
// allow/deny/revoke verbs for cli_group_cmd_permission.
// ─────────────────────────────────────────────────────────────────────────

// ── NE Profile ──

func (d *Dispatcher) showNeProfile(c *Command) error {
	profiles, err := d.Client.ListNeProfiles()
	if err != nil {
		return err
	}
	if c.Target == "" && len(c.Fields) == 0 {
		d.printNeProfileTable(profiles)
		return nil
	}
	target := c.Target
	if target == "" && len(c.FieldOrder) > 0 {
		target = c.Fields[c.FieldOrder[0]]
	}
	for _, p := range profiles {
		if p.Name == target || strconv.FormatInt(p.ID, 10) == target {
			fmt.Fprintf(d.Out, "id:           %d\r\n", p.ID)
			fmt.Fprintf(d.Out, "name:         %s\r\n", p.Name)
			fmt.Fprintf(d.Out, "description:  %s\r\n", p.Description)
			return nil
		}
	}
	return fmt.Errorf("no ne-profile with name or id %q", target)
}

func (d *Dispatcher) setNeProfile(c *Command) error {
	fields, _, err := NormalizedFields(c.Entity, c.Fields, c.FieldOrder, true)
	if err != nil {
		return err
	}
	if err := d.Client.CreateNeProfile(fields); err != nil {
		return err
	}
	fmt.Fprintln(d.Out, "OK: ne-profile created")
	return nil
}

func (d *Dispatcher) updateNeProfile(c *Command) error {
	fields, _, err := NormalizedFields(c.Entity, c.Fields, c.FieldOrder, false)
	if err != nil {
		return err
	}
	id, err := d.Client.ResolveNeProfileID(c.Target)
	if err != nil {
		return err
	}
	fields["id"] = id
	if err := d.Client.UpdateNeProfile(fields); err != nil {
		return err
	}
	fmt.Fprintln(d.Out, "OK: ne-profile updated")
	return nil
}

func (d *Dispatcher) deleteNeProfile(c *Command) error {
	id, err := d.Client.ResolveNeProfileID(c.Target)
	if err != nil {
		return err
	}
	if !d.confirmDelete("ne-profile", c.Target) {
		fmt.Fprintln(d.Out, "aborted")
		return nil
	}
	if err := d.Client.DeleteNeProfileByID(id); err != nil {
		return err
	}
	fmt.Fprintln(d.Out, "OK: ne-profile deleted")
	return nil
}

// ── Command Def ──

func (d *Dispatcher) showCommandDef(c *Command) error {
	// Optional filter via <field> <value> form: service / ne_profile / category.
	service, neProfile, category := "", "", ""
	if len(c.FieldOrder) > 0 {
		switch c.FieldOrder[0] {
		case "service":
			service = c.Fields["service"]
		case "ne_profile", "profile":
			neProfile = c.Fields[c.FieldOrder[0]]
		case "category":
			category = c.Fields["category"]
		}
	}
	defs, err := d.Client.ListCommandDefs(service, neProfile, category)
	if err != nil {
		return err
	}
	if c.Target != "" {
		id, err := strconv.ParseInt(c.Target, 10, 64)
		if err != nil {
			return fmt.Errorf("show command-def target must be numeric id")
		}
		for _, dd := range defs {
			if dd.ID == id {
				d.printCommandDefDetail(dd)
				return nil
			}
		}
		return fmt.Errorf("no command-def with id %d", id)
	}
	d.printCommandDefTable(defs)
	return nil
}

func (d *Dispatcher) setCommandDef(c *Command) error {
	fields, _, err := NormalizedFields(c.Entity, c.Fields, c.FieldOrder, true)
	if err != nil {
		return err
	}
	if err := d.Client.CreateCommandDef(fields); err != nil {
		return err
	}
	fmt.Fprintln(d.Out, "OK: command-def created")
	return nil
}

func (d *Dispatcher) updateCommandDef(c *Command) error {
	fields, _, err := NormalizedFields(c.Entity, c.Fields, c.FieldOrder, false)
	if err != nil {
		return err
	}
	id, err := strconv.ParseInt(c.Target, 10, 64)
	if err != nil {
		return fmt.Errorf("update command-def target must be numeric id")
	}
	fields["id"] = id
	if err := d.Client.UpdateCommandDef(fields); err != nil {
		return err
	}
	fmt.Fprintln(d.Out, "OK: command-def updated")
	return nil
}

func (d *Dispatcher) deleteCommandDef(c *Command) error {
	id, err := strconv.ParseInt(c.Target, 10, 64)
	if err != nil {
		return fmt.Errorf("delete command-def target must be numeric id")
	}
	if !d.confirmDelete("command-def", c.Target) {
		fmt.Fprintln(d.Out, "aborted")
		return nil
	}
	if err := d.Client.DeleteCommandDefByID(id); err != nil {
		return err
	}
	fmt.Fprintln(d.Out, "OK: command-def deleted")
	return nil
}

// ── Command Group ──

func (d *Dispatcher) showCommandGroup(c *Command) error {
	service, neProfile := "", ""
	if len(c.FieldOrder) > 0 {
		switch c.FieldOrder[0] {
		case "service":
			service = c.Fields["service"]
		case "ne_profile", "profile":
			neProfile = c.Fields[c.FieldOrder[0]]
		}
	}
	groups, err := d.Client.ListCommandGroups(service, neProfile)
	if err != nil {
		return err
	}
	if c.Target == "" && len(c.Fields) == 0 {
		d.printCommandGroupTable(groups)
		return nil
	}
	target := c.Target
	if target == "" && len(c.FieldOrder) > 0 && c.FieldOrder[0] == "name" {
		target = c.Fields["name"]
	}
	if target == "" {
		d.printCommandGroupTable(groups)
		return nil
	}
	for _, g := range groups {
		if g.Name == target || strconv.FormatInt(g.ID, 10) == target {
			commands, _ := d.Client.ListCommandsOfGroup(g.ID)
			d.printCommandGroupDetail(g, commands)
			return nil
		}
	}
	return fmt.Errorf("no command-group with name or id %q", target)
}

func (d *Dispatcher) setCommandGroup(c *Command) error {
	fields, _, err := NormalizedFields(c.Entity, c.Fields, c.FieldOrder, true)
	if err != nil {
		return err
	}
	if err := d.Client.CreateCommandGroup(fields); err != nil {
		return err
	}
	fmt.Fprintln(d.Out, "OK: command-group created")
	return nil
}

func (d *Dispatcher) updateCommandGroup(c *Command) error {
	fields, _, err := NormalizedFields(c.Entity, c.Fields, c.FieldOrder, false)
	if err != nil {
		return err
	}
	id, err := d.Client.ResolveCommandGroupID(c.Target)
	if err != nil {
		return err
	}
	fields["id"] = id
	if err := d.Client.UpdateCommandGroup(fields); err != nil {
		return err
	}
	fmt.Fprintln(d.Out, "OK: command-group updated")
	return nil
}

func (d *Dispatcher) deleteCommandGroup(c *Command) error {
	id, err := d.Client.ResolveCommandGroupID(c.Target)
	if err != nil {
		return err
	}
	if !d.confirmDelete("command-group", c.Target) {
		fmt.Fprintln(d.Out, "aborted")
		return nil
	}
	if err := d.Client.DeleteCommandGroupByID(id); err != nil {
		return err
	}
	fmt.Fprintln(d.Out, "OK: command-group deleted")
	return nil
}

// ── map/unmap command-group <cg> command <cmd_id> ──

func (d *Dispatcher) mapCommandGroup(c *Command, attach bool) error {
	groupID, err := d.Client.ResolveCommandGroupID(c.Target)
	if err != nil {
		return err
	}
	cmdID, err := d.Client.ResolveCommandDefID(c.Related)
	if err != nil {
		return err
	}
	verb := "map"
	if !attach {
		verb = "unmap"
	}
	if attach {
		if err := d.Client.AddCommandToGroup(groupID, cmdID); err != nil {
			return err
		}
	} else {
		if err := d.Client.RemoveCommandFromGroup(groupID, cmdID); err != nil {
			return err
		}
	}
	fmt.Fprintf(d.Out, "OK: %s command-group↔command\n", verb)
	return nil
}

// ── allow / deny / revoke ──

func (d *Dispatcher) handleGrant(c *Command, effect string) error {
	groupID, err := d.Client.ResolveGroupID(c.Target)
	if err != nil {
		return err
	}
	// Normalize grant_type.
	grantType := c.Relation
	switch grantType {
	case "command-group", "commandgroup":
		grantType = db_models.GrantTypeCommandGroup
	case "command_group", "category", "pattern":
		// ok
	default:
		return fmt.Errorf("grant_type must be one of command-group | category | pattern, got %q", grantType)
	}
	body := map[string]any{
		"service":     c.Fields["service"],
		"ne_scope":    c.Fields["ne_scope"],
		"grant_type":  grantType,
		"grant_value": c.Related,
		"effect":      effect,
	}
	if body["service"] == nil || body["service"] == "" {
		body["service"] = db_models.CommandServiceAny
	}
	if body["ne_scope"] == nil || body["ne_scope"] == "" {
		body["ne_scope"] = db_models.NeScopeAny
	}
	if err := d.Client.CreateGroupCmdPermission(groupID, body); err != nil {
		return err
	}
	fmt.Fprintf(d.Out, "OK: %s %s=%s on scope=%s added to group %q\n",
		effect, grantType, c.Related, body["ne_scope"], c.Target)
	return nil
}

func (d *Dispatcher) handleRevoke(c *Command) error {
	groupID, err := d.Client.ResolveGroupID(c.Target)
	if err != nil {
		return err
	}
	permID, err := strconv.ParseInt(c.Related, 10, 64)
	if err != nil {
		return fmt.Errorf("perm_id must be numeric, got %q", c.Related)
	}
	if !d.confirmDelete("permission", c.Related) {
		fmt.Fprintln(d.Out, "aborted")
		return nil
	}
	if err := d.Client.DeleteGroupCmdPermission(groupID, permID); err != nil {
		return err
	}
	fmt.Fprintln(d.Out, "OK: permission revoked")
	return nil
}

// ── printers ──

func (d *Dispatcher) printNeProfileTable(ps []*db_models.CliNeProfile) {
	tw := tabwriter.NewWriter(d.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tNAME\tDESCRIPTION")
	for _, p := range ps {
		fmt.Fprintf(tw, "%d\t%s\t%s\n", p.ID, p.Name, p.Description)
	}
	tw.Flush()
	fmt.Fprintf(d.Out, "(%d ne-profile%s)\r\n", len(ps), plural(len(ps)))
}

func (d *Dispatcher) printCommandDefTable(defs []*db_models.CliCommandDef) {
	tw := tabwriter.NewWriter(d.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tSERVICE\tPROFILE\tPATTERN\tCATEGORY\tRISK")
	for _, c := range defs {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\t%d\n",
			c.ID, c.Service, c.NeProfile, c.Pattern, c.Category, c.RiskLevel)
	}
	tw.Flush()
	fmt.Fprintf(d.Out, "(%d command-def%s)\r\n", len(defs), plural(len(defs)))
}

func (d *Dispatcher) printCommandDefDetail(c *db_models.CliCommandDef) {
	fmt.Fprintf(d.Out, "id:          %d\r\n", c.ID)
	fmt.Fprintf(d.Out, "service:     %s\r\n", c.Service)
	fmt.Fprintf(d.Out, "ne_profile:  %s\r\n", c.NeProfile)
	fmt.Fprintf(d.Out, "pattern:     %s\r\n", c.Pattern)
	fmt.Fprintf(d.Out, "category:    %s\r\n", c.Category)
	fmt.Fprintf(d.Out, "risk_level:  %d\r\n", c.RiskLevel)
	fmt.Fprintf(d.Out, "description: %s\r\n", c.Description)
}

func (d *Dispatcher) printCommandGroupTable(gs []*db_models.CliCommandGroup) {
	tw := tabwriter.NewWriter(d.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tNAME\tSERVICE\tPROFILE\tDESCRIPTION")
	for _, g := range gs {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\n", g.ID, g.Name, g.Service, g.NeProfile, g.Description)
	}
	tw.Flush()
	fmt.Fprintf(d.Out, "(%d command-group%s)\r\n", len(gs), plural(len(gs)))
}

func (d *Dispatcher) printCommandGroupDetail(g *db_models.CliCommandGroup, members []*db_models.CliCommandDef) {
	fmt.Fprintf(d.Out, "id:          %d\r\n", g.ID)
	fmt.Fprintf(d.Out, "name:        %s\r\n", g.Name)
	fmt.Fprintf(d.Out, "service:     %s\r\n", g.Service)
	fmt.Fprintf(d.Out, "ne_profile:  %s\r\n", g.NeProfile)
	fmt.Fprintf(d.Out, "description: %s\r\n", g.Description)
	if len(members) == 0 {
		fmt.Fprintln(d.Out, "members:     (none)")
		return
	}
	parts := make([]string, 0, len(members))
	for _, m := range members {
		parts = append(parts, fmt.Sprintf("%d:%s", m.ID, m.Pattern))
	}
	fmt.Fprintf(d.Out, "members:     %s\r\n", strings.Join(parts, ", "))
}
