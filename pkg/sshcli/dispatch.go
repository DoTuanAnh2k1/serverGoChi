package sshcli

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"
)

// Dispatcher runs a parsed Command against the mgt-svc client, writing results
// and errors to out. It is factored out of the REPL so unit tests can exercise
// it directly against a fake MgtClient.
type Dispatcher struct {
	Client *MgtClient
	Out    io.Writer
}

func (d *Dispatcher) Run(c *Command) error {
	switch c.Verb {
	case "help":
		d.printHelp(c.Target)
		return nil
	case "exit", "quit":
		return io.EOF
	case "show":
		return d.handleShow(c)
	case "set":
		return d.handleSet(c)
	case "update":
		return d.handleUpdate(c)
	case "delete":
		return d.handleDelete(c)
	case "map":
		return d.handleMap(c, true)
	case "unmap":
		return d.handleMap(c, false)
	}
	return fmt.Errorf("unhandled verb %q", c.Verb)
}

func (d *Dispatcher) handleShow(c *Command) error {
	switch c.Entity {
	case "user":
		users, err := d.Client.ListUsers()
		if err != nil {
			return err
		}
		if c.Target != "" {
			for _, u := range users {
				if u.AccountName == c.Target || strconv.FormatInt(u.AccountID, 10) == c.Target {
					d.printUserDetail(u)
					return nil
				}
			}
			return fmt.Errorf("no user with name or id %q", c.Target)
		}
		d.printUserTable(users)
		return nil
	case "ne":
		nes, err := d.Client.ListNEs()
		if err != nil {
			return err
		}
		if c.Target != "" {
			for _, n := range nes {
				if n.NeName == c.Target || strconv.FormatInt(n.ID, 10) == c.Target {
					d.printNeDetail(n)
					return nil
				}
			}
			return fmt.Errorf("no NE with name or id %q", c.Target)
		}
		d.printNeTable(nes)
		return nil
	case "group":
		gs, err := d.Client.ListGroups()
		if err != nil {
			return err
		}
		if c.Target != "" {
			for _, g := range gs {
				if g.Name == c.Target || strconv.FormatInt(g.ID, 10) == c.Target {
					detail, err := d.Client.ShowGroup(g.ID)
					if err != nil {
						return err
					}
					d.printGroupDetail(detail)
					return nil
				}
			}
			return fmt.Errorf("no group with name or id %q", c.Target)
		}
		d.printGroupTable(gs)
		return nil
	}
	return fmt.Errorf("show %s not supported", c.Entity)
}

func (d *Dispatcher) handleSet(c *Command) error {
	fields, _, err := NormalizedFields(c.Entity, c.Fields, c.FieldOrder, true)
	if err != nil {
		return err
	}
	switch c.Entity {
	case "user":
		if err := d.Client.CreateUser(fields); err != nil {
			return err
		}
		fmt.Fprintln(d.Out, "OK: user created")
	case "ne":
		if err := d.Client.CreateNE(fields); err != nil {
			return err
		}
		fmt.Fprintln(d.Out, "OK: NE created")
	case "group":
		if err := d.Client.CreateGroup(fields); err != nil {
			return err
		}
		fmt.Fprintln(d.Out, "OK: group created")
	}
	return nil
}

func (d *Dispatcher) handleUpdate(c *Command) error {
	fields, _, err := NormalizedFields(c.Entity, c.Fields, c.FieldOrder, false)
	if err != nil {
		return err
	}
	switch c.Entity {
	case "user":
		fields["account_name"] = c.Target
		if err := d.Client.UpdateUser(fields); err != nil {
			return err
		}
		fmt.Fprintln(d.Out, "OK: user updated")
	case "ne":
		id, err := d.Client.ResolveNEID(c.Target)
		if err != nil {
			return err
		}
		fields["id"] = id
		if err := d.Client.UpdateNE(fields); err != nil {
			return err
		}
		fmt.Fprintln(d.Out, "OK: NE updated")
	case "group":
		id, err := d.Client.ResolveGroupID(c.Target)
		if err != nil {
			return err
		}
		fields["id"] = id
		if err := d.Client.UpdateGroup(fields); err != nil {
			return err
		}
		fmt.Fprintln(d.Out, "OK: group updated")
	}
	return nil
}

func (d *Dispatcher) handleDelete(c *Command) error {
	switch c.Entity {
	case "user":
		if err := d.Client.DeleteUser(c.Target); err != nil {
			return err
		}
		fmt.Fprintln(d.Out, "OK: user deleted")
	case "ne":
		id, err := d.Client.ResolveNEID(c.Target)
		if err != nil {
			return err
		}
		if err := d.Client.DeleteNEByID(id); err != nil {
			return err
		}
		fmt.Fprintln(d.Out, "OK: NE deleted")
	case "group":
		id, err := d.Client.ResolveGroupID(c.Target)
		if err != nil {
			return err
		}
		if err := d.Client.DeleteGroupByID(id); err != nil {
			return err
		}
		fmt.Fprintln(d.Out, "OK: group deleted")
	}
	return nil
}

func (d *Dispatcher) handleMap(c *Command, attach bool) error {
	verb := "unmap"
	if attach {
		verb = "map"
	}
	switch {
	case c.Entity == "user" && c.Relation == "ne":
		id, err := d.Client.ResolveNEID(c.Related)
		if err != nil {
			return err
		}
		if attach {
			return reportOK(d.Out, verb, "user↔NE", d.Client.AssignNEToUser(c.Target, id))
		}
		return reportOK(d.Out, verb, "user↔NE", d.Client.UnassignNEFromUser(c.Target, id))
	case c.Entity == "user" && c.Relation == "group":
		id, err := d.Client.ResolveGroupID(c.Related)
		if err != nil {
			return err
		}
		if attach {
			return reportOK(d.Out, verb, "user↔group", d.Client.AssignUserToGroup(c.Target, id))
		}
		return reportOK(d.Out, verb, "user↔group", d.Client.UnassignUserFromGroup(c.Target, id))
	case c.Entity == "group" && c.Relation == "ne":
		gid, err := d.Client.ResolveGroupID(c.Target)
		if err != nil {
			return err
		}
		nid, err := d.Client.ResolveNEID(c.Related)
		if err != nil {
			return err
		}
		if attach {
			return reportOK(d.Out, verb, "group↔NE", d.Client.AssignNEToGroup(gid, nid))
		}
		return reportOK(d.Out, verb, "group↔NE", d.Client.UnassignNEFromGroup(gid, nid))
	}
	return fmt.Errorf("unsupported mapping: %s %s %s", c.Entity, c.Relation, c.Verb)
}

func reportOK(w io.Writer, verb, rel string, err error) error {
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "OK: %s %s\n", verb, rel)
	return nil
}

// ---- printing ----

func (d *Dispatcher) printUserTable(us []UserInfo) {
	tw := tabwriter.NewWriter(d.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tNAME\tROLE\tENABLED\tEMAIL\tFULL NAME")
	for _, u := range us {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%t\t%s\t%s\n", u.AccountID, u.AccountName, roleLabel(u.AccountType), u.IsEnable, u.Email, u.FullName)
	}
	tw.Flush()
	fmt.Fprintf(d.Out, "(%d user%s)\r\n", len(us), plural(len(us)))
}

func (d *Dispatcher) printUserDetail(u UserInfo) {
	fmt.Fprintf(d.Out, "id:           %d\r\n", u.AccountID)
	fmt.Fprintf(d.Out, "name:         %s\r\n", u.AccountName)
	fmt.Fprintf(d.Out, "role:         %s\r\n", roleLabel(u.AccountType))
	fmt.Fprintf(d.Out, "enabled:      %t\r\n", u.IsEnable)
	fmt.Fprintf(d.Out, "email:        %s\r\n", u.Email)
	fmt.Fprintf(d.Out, "full_name:    %s\r\n", u.FullName)
	fmt.Fprintf(d.Out, "phone:        %s\r\n", u.PhoneNumber)
	fmt.Fprintf(d.Out, "address:      %s\r\n", u.Address)
	fmt.Fprintf(d.Out, "description:  %s\r\n", u.Description)
	fmt.Fprintf(d.Out, "created_by:   %s\r\n", u.CreatedBy)
}

func roleLabel(t int) string {
	switch t {
	case 0:
		return "SuperAdmin"
	case 1:
		return "Admin"
	case 2:
		return "Normal"
	}
	return strconv.Itoa(t)
}

func (d *Dispatcher) printNeTable(ns []NeInfo) {
	tw := tabwriter.NewWriter(d.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tNAME\tSITE\tNAMESPACE\tIP\tPORT\tMODE")
	for _, n := range ns {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\t%d\t%s\n", n.ID, n.NeName, n.SiteName, n.Namespace, n.ConfMasterIP, n.ConfPortMasterTCP, n.ConfMode)
	}
	tw.Flush()
	fmt.Fprintf(d.Out, "(%d NE%s)\r\n", len(ns), plural(len(ns)))
}

func (d *Dispatcher) printNeDetail(n NeInfo) {
	fmt.Fprintf(d.Out, "id:                    %d\r\n", n.ID)
	fmt.Fprintf(d.Out, "ne_name:               %s\r\n", n.NeName)
	fmt.Fprintf(d.Out, "site_name:             %s\r\n", n.SiteName)
	fmt.Fprintf(d.Out, "namespace:             %s\r\n", n.Namespace)
	fmt.Fprintf(d.Out, "conf_master_ip:        %s\r\n", n.ConfMasterIP)
	fmt.Fprintf(d.Out, "conf_port_master_tcp:  %d\r\n", n.ConfPortMasterTCP)
	fmt.Fprintf(d.Out, "command_url:           %s\r\n", n.CommandURL)
	fmt.Fprintf(d.Out, "conf_mode:             %s\r\n", n.ConfMode)
	fmt.Fprintf(d.Out, "system_type:           %s\r\n", n.SystemType)
	fmt.Fprintf(d.Out, "description:           %s\r\n", n.Description)
}

func (d *Dispatcher) printGroupTable(gs []GroupInfo) {
	tw := tabwriter.NewWriter(d.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tNAME\tDESCRIPTION")
	for _, g := range gs {
		fmt.Fprintf(tw, "%d\t%s\t%s\n", g.ID, g.Name, g.Description)
	}
	tw.Flush()
	fmt.Fprintf(d.Out, "(%d group%s)\r\n", len(gs), plural(len(gs)))
}

func (d *Dispatcher) printGroupDetail(g *GroupDetail) {
	fmt.Fprintf(d.Out, "id:           %d\r\n", g.ID)
	fmt.Fprintf(d.Out, "name:         %s\r\n", g.Name)
	fmt.Fprintf(d.Out, "description:  %s\r\n", g.Description)
	fmt.Fprintf(d.Out, "users:        %s\r\n", strings.Join(g.Users, ", "))
	ids := make([]string, 0, len(g.NeIDs))
	for _, id := range g.NeIDs {
		ids = append(ids, strconv.FormatInt(id, 10))
	}
	fmt.Fprintf(d.Out, "ne_ids:       %s\r\n", strings.Join(ids, ", "))
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// ---- help ----

const helpGeneral = `Available commands (type 'help <command>' for details):
  show user|ne|group [<name|id>]
  set user|ne|group <field> <value> [<field> <value> ...]
  update user|ne|group <name|id> <field> <value> [...]
  delete user|ne|group <name|id>
  map user <name> ne <ne_name|id>
  map user <name> group <group_name|id>
  map group <group_name|id> ne <ne_name|id>
  unmap ...  (same shape as map)
  help [command]
  exit | quit

Notes:
  - Field/value pairs are space-separated. Quote values that contain spaces:
      set user name alice password pw full_name "Alice Wonder"
  - Field aliases exist (e.g. 'name' → account_name for user, or → ne_name for NE).
  - account_type accepts 1 (Admin) or 2 (Normal). SuperAdmin can't be created via CLI.
  - Tab to complete verbs, entities, field names, and enum values.
`

var helpTopics = map[string]string{
	"set": `set <entity> <field> <value> [<field> <value> ...]

Create a new record. All required fields must be present.

  user   required: name, password     optional: email, full_name, phone, address,
                                                 description, account_type
  ne     required: ne_name, namespace, conf_master_ip, conf_port_master_tcp, command_url
         optional: site_name, conf_mode (SSH|TELNET|NETCONF|RESTCONF), conf_username,
                   conf_password, conf_slave_ip, conf_port_master_ssh, ...
  group  required: name                optional: description
`,
	"update": `update <entity> <name|id> <field> <value> [<field> <value> ...]

Update an existing record. Any subset of fields may be specified.

  user   target: account_name         e.g. update user alice email new@b.c
  ne     target: ne_name or id        e.g. update ne HTSMF01 site_name HN
  group  target: name or id           e.g. update group dev description "dev team"
`,
	"delete": `delete <entity> <name|id>

Delete the target record.

  user   target: account_name         (SuperAdmin cannot be deleted)
  ne     target: ne_name or id        (cascades user↔NE mappings)
  group  target: name or id
`,
	"show": `show <entity> [<name|id>]

Without a target, lists all records in a table. With a target, shows detail.
`,
	"map": `map <entity> <target> <relation> <related>
unmap <entity> <target> <relation> <related>

Supported shapes:
  map user <username> ne <ne_name|id>        attach NE directly to user
  map user <username> group <group_name|id>  attach user to group
  map group <group_name|id> ne <ne_name|id>  attach NE to group

unmap uses the same shape — removes the relation. A user gains access to NEs
transitively through groups; unmapping one does not remove access granted via
the other.
`,
	"exit": "exit | quit — end the CLI session.\n",
	"help": "help [command] — show general help, or details for one command.\n",
}

func (d *Dispatcher) printHelp(topic string) {
	if topic == "" {
		fmt.Fprint(d.Out, helpGeneral)
		return
	}
	if body, ok := helpTopics[topic]; ok {
		fmt.Fprint(d.Out, body)
		return
	}
	fmt.Fprintf(d.Out, "no help topic %q — available: set, update, delete, show, map, exit, help\n", topic)
}
