package sshcli

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// EntitySpec describes the fields supported for one entity.
type EntitySpec struct {
	// Alias → canonical API field name.
	FieldAliases map[string]string
	// Canonical field names required when creating (set).
	Required []string
	// Integer-typed fields (canonical name). Validated on set/update.
	IntFields map[string]bool
	// Fields whose value is constrained to a fixed set (canonical name → allowed values).
	EnumFields map[string][]string
}

var entitySpecs = map[string]EntitySpec{
	"user": {
		FieldAliases: map[string]string{
			"name":         "account_name",
			"account_name": "account_name",
			"username":     "account_name",
			"password":     "password",
			"email":        "email",
			"full_name":    "full_name",
			"phone":        "phone_number",
			"phone_number": "phone_number",
			"address":      "address",
			"description":  "description",
			"account_type": "account_type",
			"type":         "account_type",
		},
		Required:  []string{"account_name", "password"},
		IntFields: map[string]bool{"account_type": true},
		EnumFields: map[string][]string{
			"account_type": {"1", "2"},
		},
	},
	"ne": {
		FieldAliases: map[string]string{
			"ne_name":              "ne_name",
			"name":                 "ne_name",
			"namespace":            "namespace",
			"conf_master_ip":       "conf_master_ip",
			"ip":                   "conf_master_ip",
			"conf_port_master_tcp": "conf_port_master_tcp",
			"port":                 "conf_port_master_tcp",
			"command_url":          "command_url",
			"site_name":            "site_name",
			"site":                 "site_name",
			"system_type":          "system_type",
			"description":          "description",
			"conf_mode":            "conf_mode",
			"mode":                 "conf_mode",
			"conf_slave_ip":        "conf_slave_ip",
			"conf_port_master_ssh": "conf_port_master_ssh",
			"conf_port_slave_ssh":  "conf_port_slave_ssh",
			"conf_port_slave_tcp":  "conf_port_slave_tcp",
			"conf_username":        "conf_username",
			"conf_password":        "conf_password",
		},
		Required: []string{"ne_name", "namespace", "conf_master_ip", "conf_port_master_tcp", "command_url"},
		IntFields: map[string]bool{
			"conf_port_master_tcp": true,
			"conf_port_master_ssh": true,
			"conf_port_slave_ssh":  true,
			"conf_port_slave_tcp":  true,
		},
		EnumFields: map[string][]string{
			"conf_mode": {"SSH", "TELNET", "NETCONF", "RESTCONF"},
		},
	},
	"group": {
		FieldAliases: map[string]string{
			"name":        "name",
			"description": "description",
		},
		Required: []string{"name"},
	},
}

// NormalizedFields converts a raw parsed field map into canonical API fields.
// It returns either the normalized map + canonical key order, or an error.
// If requireAll is true, all entity Required fields must be present.
func NormalizedFields(entity string, raw map[string]string, order []string, requireAll bool) (map[string]any, []string, error) {
	spec, ok := entitySpecs[entity]
	if !ok {
		return nil, nil, fmt.Errorf("unknown entity %q", entity)
	}
	out := map[string]any{}
	var canonicalOrder []string
	seen := map[string]string{} // canonical → original alias (for dup detection)
	for _, alias := range order {
		canonical, ok := spec.FieldAliases[alias]
		if !ok {
			return nil, nil, fmt.Errorf("unknown field %q for %s (valid: %s)", alias, entity, fieldListHint(spec))
		}
		if prev, dup := seen[canonical]; dup {
			return nil, nil, fmt.Errorf("field %q conflicts with earlier %q (both map to %q)", alias, prev, canonical)
		}
		seen[canonical] = alias
		val := raw[alias]
		if allowed, ok := spec.EnumFields[canonical]; ok {
			if !containsString(allowed, val) {
				return nil, nil, fmt.Errorf("field %q must be one of %v, got %q", alias, allowed, val)
			}
		}
		if spec.IntFields[canonical] {
			n, err := strconv.Atoi(val)
			if err != nil {
				return nil, nil, fmt.Errorf("field %q must be an integer, got %q", alias, val)
			}
			out[canonical] = n
		} else {
			out[canonical] = val
		}
		canonicalOrder = append(canonicalOrder, canonical)
	}
	if requireAll {
		for _, req := range spec.Required {
			if _, ok := seen[req]; !ok {
				return nil, nil, fmt.Errorf("missing required field %q for set %s (required: %s)", req, entity, strings.Join(spec.Required, ", "))
			}
		}
	}
	return out, canonicalOrder, nil
}

func fieldListHint(spec EntitySpec) string {
	uniq := map[string]bool{}
	for _, v := range spec.FieldAliases {
		uniq[v] = true
	}
	keys := make([]string, 0, len(uniq))
	for k := range uniq {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}

func containsString(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

// FieldNames returns all canonical field names for an entity, sorted.
func FieldNames(entity string) []string {
	spec, ok := entitySpecs[entity]
	if !ok {
		return nil
	}
	uniq := map[string]bool{}
	for _, v := range spec.FieldAliases {
		uniq[v] = true
	}
	out := make([]string, 0, len(uniq))
	for k := range uniq {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// FieldAliasNames returns all alias names for an entity, sorted — useful for tab completion.
func FieldAliasNames(entity string) []string {
	spec, ok := entitySpecs[entity]
	if !ok {
		return nil
	}
	out := make([]string, 0, len(spec.FieldAliases))
	for k := range spec.FieldAliases {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ShowFilterSpec describes the filter fields supported by `show <entity> <field> <value>`.
// These are separate from the full CRUD field set because the show command
// only filters on a handful of columns (name/id/email/role/site/namespace).
type ShowFilterSpec struct {
	// Alias → canonical filter name.
	FieldAliases map[string]string
	// Canonical → allowed enum values (used for Tab completion on value position).
	EnumValues map[string][]string
}

var showFilterSpecs = map[string]ShowFilterSpec{
	"user": {
		FieldAliases: map[string]string{
			"name":         "name",
			"username":     "name",
			"account_name": "name",
			"id":           "id",
			"account_id":   "id",
			"email":        "email",
			"role":         "role",
			"type":         "role",
			"account_type": "role",
		},
		EnumValues: map[string][]string{
			"role": {"SuperAdmin", "Admin", "Normal"},
		},
	},
	"ne": {
		FieldAliases: map[string]string{
			"name":      "name",
			"ne_name":   "name",
			"id":        "id",
			"site":      "site",
			"site_name": "site",
			"namespace": "namespace",
		},
	},
	"group": {
		FieldAliases: map[string]string{
			"name": "name",
			"id":   "id",
		},
	},
}

// ResolveShowFilter normalizes a user-supplied alias against an entity's
// show-filter spec and returns the canonical filter name, or false if the
// alias isn't recognized.
func ResolveShowFilter(entity, alias string) (string, bool) {
	spec, ok := showFilterSpecs[entity]
	if !ok {
		return "", false
	}
	canon, ok := spec.FieldAliases[strings.ToLower(alias)]
	return canon, ok
}

// ShowFilterAliases returns all alias names for an entity's show filters,
// used for Tab completion.
func ShowFilterAliases(entity string) []string {
	spec, ok := showFilterSpecs[entity]
	if !ok {
		return nil
	}
	out := make([]string, 0, len(spec.FieldAliases))
	for k := range spec.FieldAliases {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ShowFilterEnumValues returns the enum values allowed for a canonical
// show-filter field (if any), for Tab completion at the value position.
func ShowFilterEnumValues(entity, canonical string) []string {
	spec, ok := showFilterSpecs[entity]
	if !ok {
		return nil
	}
	v, ok := spec.EnumValues[canonical]
	if !ok {
		return nil
	}
	return append([]string(nil), v...)
}
