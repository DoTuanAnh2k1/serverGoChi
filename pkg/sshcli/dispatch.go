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
	// Confirm, when set, is invoked before destructive operations (delete).
	// It returns true only when the user explicitly agrees. When nil, the
	// dispatcher auto-confirms — this keeps unit tests that don't care about
	// the prompt working; production wires it to a real terminal prompt.
	Confirm func(msg string) bool
}

// confirmDelete asks the user to confirm a deletion, returning true only on
// explicit agreement. Any other input (including empty, non-"y" characters,
// Ctrl+C byte embedded in the line, or a read error / EOF from the session)
// counts as abort.
func (d *Dispatcher) confirmDelete(kind, target string) bool {
	if d.Confirm == nil {
		return true
	}
	msg := fmt.Sprintf("Delete %s %q?", kind, target)
	return d.Confirm(msg)
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
		return d.showUser(c)
	case "ne":
		return d.showNe(c)
	case "group":
		return d.showGroup(c)
	}
	return fmt.Errorf("show %s not supported", c.Entity)
}

// showUser handles `show user`, `show user <name|id>` (legacy), and the new
// `show user <field> <value>` form. Role filter always prints a table since
// multiple users can share a role.
func (d *Dispatcher) showUser(c *Command) error {
	users, err := d.Client.ListUsers()
	if err != nil {
		return err
	}
	if c.Target == "" && len(c.Fields) == 0 {
		d.printUserTable(users)
		return nil
	}
	field, value := showFilterPair(c)
	if field == "" {
		// Legacy c.Target: match name or id.
		for _, u := range users {
			if u.AccountName == c.Target || strconv.FormatInt(u.AccountID, 10) == c.Target {
				d.printUserDetail(u)
				return nil
			}
		}
		return fmt.Errorf("no user with name or id %q", c.Target)
	}
	canon, ok := ResolveShowFilter("user", field)
	if !ok {
		return fmt.Errorf("unknown filter field %q for user (valid: %s)", field, strings.Join(ShowFilterAliases("user"), ", "))
	}
	matched := filterUsers(users, canon, value)
	if canon == "role" {
		// Always table, even if only one — caller asked for a role bucket.
		if len(matched) == 0 {
			return fmt.Errorf("no user with role %q", value)
		}
		d.printUserTable(matched)
		return nil
	}
	if len(matched) == 0 {
		return fmt.Errorf("no user with %s %q", canon, value)
	}
	if len(matched) == 1 {
		d.printUserDetail(matched[0])
		return nil
	}
	d.printUserTable(matched)
	return nil
}

// showNe handles `show ne`, `show ne <name|id>` (legacy — returns a table
// when the name is ambiguous across namespaces), and `show ne <field> <value>`.
func (d *Dispatcher) showNe(c *Command) error {
	nes, err := d.Client.ListNEs()
	if err != nil {
		return err
	}
	if c.Target == "" && len(c.Fields) == 0 {
		d.printNeTable(nes)
		return nil
	}
	field, value := showFilterPair(c)
	if field == "" {
		// Legacy single-target: match name or id; name may appear in multiple
		// namespaces, so return the full set.
		matched := []NeInfo{}
		for _, n := range nes {
			if n.NeName == c.Target || strconv.FormatInt(n.ID, 10) == c.Target {
				matched = append(matched, n)
			}
		}
		if len(matched) == 0 {
			return fmt.Errorf("no NE with name or id %q", c.Target)
		}
		if len(matched) == 1 {
			d.printNeDetail(matched[0])
			return nil
		}
		d.printNeTable(matched)
		return nil
	}
	canon, ok := ResolveShowFilter("ne", field)
	if !ok {
		return fmt.Errorf("unknown filter field %q for ne (valid: %s)", field, strings.Join(ShowFilterAliases("ne"), ", "))
	}
	matched := filterNes(nes, canon, value)
	if len(matched) == 0 {
		return fmt.Errorf("no NE with %s %q", canon, value)
	}
	// For name/id filters on a single match, show the detail view; otherwise
	// always table (site/namespace are bucket filters).
	if (canon == "name" || canon == "id") && len(matched) == 1 {
		d.printNeDetail(matched[0])
		return nil
	}
	d.printNeTable(matched)
	return nil
}

// showGroup supports `show group`, `show group <name|id>` (legacy), and
// `show group <field> <value>` where field is name or id.
func (d *Dispatcher) showGroup(c *Command) error {
	gs, err := d.Client.ListGroups()
	if err != nil {
		return err
	}
	if c.Target == "" && len(c.Fields) == 0 {
		d.printGroupTable(gs)
		return nil
	}
	field, value := showFilterPair(c)
	var target string
	var byID bool
	if field == "" {
		target = c.Target
	} else {
		canon, ok := ResolveShowFilter("group", field)
		if !ok {
			return fmt.Errorf("unknown filter field %q for group (valid: %s)", field, strings.Join(ShowFilterAliases("group"), ", "))
		}
		target = value
		byID = canon == "id"
	}
	for _, g := range gs {
		if byID {
			if strconv.FormatInt(g.ID, 10) != target {
				continue
			}
		} else {
			if g.Name != target && strconv.FormatInt(g.ID, 10) != target {
				continue
			}
		}
		detail, err := d.Client.ShowGroup(g.ID)
		if err != nil {
			return err
		}
		d.printGroupDetail(detail)
		return nil
	}
	if field == "" {
		return fmt.Errorf("no group with name or id %q", target)
	}
	return fmt.Errorf("no group with %s %q", field, target)
}

// showFilterPair extracts the sole (field, value) pair from c.Fields, or
// returns empty strings when only legacy c.Target is set.
func showFilterPair(c *Command) (field, value string) {
	if len(c.FieldOrder) == 0 {
		return "", ""
	}
	f := c.FieldOrder[0]
	return f, c.Fields[f]
}

func filterUsers(users []UserInfo, canon, value string) []UserInfo {
	out := []UserInfo{}
	for _, u := range users {
		switch canon {
		case "name":
			if u.AccountName == value {
				out = append(out, u)
			}
		case "id":
			if strconv.FormatInt(u.AccountID, 10) == value {
				out = append(out, u)
			}
		case "email":
			if u.Email == value {
				out = append(out, u)
			}
		case "role":
			if matchRole(u.AccountType, value) {
				out = append(out, u)
			}
		}
	}
	return out
}

// matchRole compares a user's account_type code against either a numeric
// string ("0"/"1"/"2") or a case-insensitive label (SuperAdmin/Admin/Normal).
func matchRole(code int, input string) bool {
	input = strings.ToLower(strings.TrimSpace(input))
	if n, err := strconv.Atoi(input); err == nil {
		return n == code
	}
	switch input {
	case "superadmin", "super_admin":
		return code == 0
	case "admin":
		return code == 1
	case "normal", "user":
		return code == 2
	}
	return false
}

func filterNes(nes []NeInfo, canon, value string) []NeInfo {
	out := []NeInfo{}
	for _, n := range nes {
		switch canon {
		case "name":
			if n.NeName == value {
				out = append(out, n)
			}
		case "id":
			if strconv.FormatInt(n.ID, 10) == value {
				out = append(out, n)
			}
		case "site":
			if n.SiteName == value {
				out = append(out, n)
			}
		case "namespace":
			if n.Namespace == value {
				out = append(out, n)
			}
		}
	}
	return out
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
		if !d.confirmDelete("user", c.Target) {
			fmt.Fprintln(d.Out, "aborted")
			return nil
		}
		if err := d.Client.DeleteUser(c.Target); err != nil {
			return err
		}
		fmt.Fprintln(d.Out, "OK: user deleted")
	case "ne":
		id, err := d.Client.ResolveNEID(c.Target)
		if err != nil {
			return err
		}
		if !d.confirmDelete("NE", c.Target) {
			fmt.Fprintln(d.Out, "aborted")
			return nil
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
		if !d.confirmDelete("group", c.Target) {
			fmt.Fprintln(d.Out, "aborted")
			return nil
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

const helpGeneral = `Available commands (type 'help <command>' or append '--help' to any command):
  show user|ne|group [<field> <value> | <name|id>]
  set user|ne|group <field> <value> [<field> <value> ...]
  update user|ne|group <name|id> <field> <value> [...]
  delete user|ne|group <name|id>
  map user <name> ne <ne_name|id>
  map user <name> group <group_name|id>
  map group <group_name|id> ne <ne_name|id>
  unmap ...  (same shape as map)
  help [command [entity]]
  exit | quit

Notes:
  - Append '--help' (or '-h') to any command to see per-command help:
      set user --help         → help for 'set user'
      show ne --help          → help for 'show ne'
  - Field/value pairs are space-separated. Quote values that contain spaces:
      set user name alice password pw full_name "Alice Wonder"
  - Field aliases exist (e.g. 'name' → account_name for user, or → ne_name for NE).
  - show supports <field> <value> filters (e.g. show user role Admin,
    show ne site HN, show ne namespace default).
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
	"set user": `set user name <name> password <password> [<field> <value> ...]

Required fields:
  name (alias: account_name, username)
  password

Optional fields:
  email
  full_name                 (quote if it contains spaces)
  phone (alias: phone_number)
  address
  description
  type (alias: account_type)  1 = Admin, 2 = Normal  (SuperAdmin via CLI is not allowed)

Example:
  set user name alice password secret email alice@example.com \
      full_name "Alice Wonder" phone 0900000000 type 2
`,
	"set ne": `set ne ne_name <name> namespace <ns> conf_master_ip <ip> \
       conf_port_master_tcp <port> command_url <url> [<field> <value> ...]

Required fields:
  ne_name (alias: name)
  namespace                    (name+namespace must be unique)
  conf_master_ip (alias: ip)
  conf_port_master_tcp (alias: port)                int
  command_url

Optional fields:
  site_name (alias: site)
  system_type
  description
  conf_mode (alias: mode)      enum: SSH | TELNET | NETCONF | RESTCONF
  conf_slave_ip
  conf_port_master_ssh         int
  conf_port_slave_ssh          int
  conf_port_slave_tcp          int
  conf_username
  conf_password

Example:
  set ne name HTSMF01 namespace default ip 10.0.0.10 port 830 \
      command_url http://10.0.0.10/restconf mode NETCONF site HN \
      conf_username admin conf_password netadmin description "primary SMF"
`,
	"set group": `set group name <name> [description <text>]

Required fields:
  name

Optional fields:
  description

Example:
  set group name dev description "dev team"
`,
	"update": `update <entity> <name|id> <field> <value> [<field> <value> ...]

Update an existing record. Any subset of fields may be specified.

  user   target: account_name         e.g. update user alice email new@b.c
  ne     target: ne_name or id        e.g. update ne HTSMF01 site_name HN
  group  target: name or id           e.g. update group dev description "dev team"
`,
	"update user": `update user <account_name> <field> <value> [<field> <value> ...]

Target is the user's account_name. Any subset of the optional fields from
'set user --help' may be supplied. You cannot rename an account via update.

Example:
  update user alice email new@b.c phone 0911111111
`,
	"update ne": `update ne <ne_name|id> <field> <value> [<field> <value> ...]

Target is either ne_name or numeric id. When ne_name is duplicated across
namespaces the id form avoids ambiguity.

Example:
  update ne HTSMF01 site_name HN2 description "moved site"
  update ne 42 conf_mode NETCONF
`,
	"update group": `update group <name|id> <field> <value> [<field> <value> ...]

Target is the group name or numeric id. Allowed fields: name, description.

Example:
  update group dev description "dev team v2"
  update group 3 name dev-internal
`,
	"delete": `delete <entity> <name|id>

Delete the target record.

  user   target: account_name         (SuperAdmin cannot be deleted)
  ne     target: ne_name or id        (cascades user↔NE mappings)
  group  target: name or id
`,
	"delete user":  "delete user <account_name>\n\nDeletes the user. SuperAdmin cannot be deleted.\n",
	"delete ne":    "delete ne <ne_name|id>\n\nDeletes the NE and cascades user↔NE and group↔NE mappings.\n",
	"delete group": "delete group <name|id>\n\nDeletes the group and its user↔group and group↔NE mappings.\n",
	"show": `show <entity> [<field> <value> | <name|id>]

Without a target, lists all records in a table.

  show user                       list all users
  show user <name|id>             detail for one user (legacy single-target)
  show user <field> <value>       filter — see per-entity help
  show ne                         list all NEs
  show ne <name|id>               all NEs matching name (ambiguous across
                                  namespaces) or a single NE by id
  show ne <field> <value>         filter — see per-entity help
  show group                      list all groups
  show group <name|id>            group detail (users + NE ids)
`,
	"show user": `show user [<field> <value>]

Filter fields:
  name (alias: username, account_name)   exact match, prints detail
  id   (alias: account_id)               exact match, prints detail
  email                                  exact match, prints detail
  role (alias: type, account_type)       always prints a table.
        accepts label (SuperAdmin|Admin|Normal) or numeric (0|1|2).

Examples:
  show user
  show user alice                 (legacy — same as 'show user name alice')
  show user name alice
  show user email alice@example.com
  show user role Admin
`,
	"show ne": `show ne [<field> <value>]

Filter fields:
  name (alias: ne_name)        returns all NEs with that name — useful when
                               ne_name is duplicated across namespaces.
  id                           exact match, prints detail.
  site (alias: site_name)      returns all NEs at that site.
  namespace                    returns all NEs in that namespace.

Examples:
  show ne
  show ne HTSMF01               (legacy — matches name or id)
  show ne name HTSMF01
  show ne site HN
  show ne namespace default
  show ne id 42
`,
	"show group": `show group [<field> <value>]

Filter fields:
  name    match group name, prints detail (users + NE ids).
  id      match numeric id, prints detail.

Examples:
  show group
  show group dev                (legacy — matches name or id)
  show group name dev
  show group id 3
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
	"unmap":        "unmap <entity> <target> <relation> <related>\n\nSame shape as 'map' — see 'help map'.\n",
	"exit":         "exit | quit — end the CLI session.\n",
	"help":         "help [command [entity]] — show general help or topic-specific help.\nYou can also append '--help' (or '-h') to any command to see its help.\n",
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
	// Fall back from "verb entity" to "verb" if no specific entry.
	if i := strings.Index(topic, " "); i > 0 {
		if body, ok := helpTopics[topic[:i]]; ok {
			fmt.Fprint(d.Out, body)
			return
		}
	}
	fmt.Fprintf(d.Out, "no help topic %q — available: show, set, update, delete, map, unmap, exit, help\n", topic)
}
