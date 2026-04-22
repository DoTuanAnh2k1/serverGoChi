package sshcli

import (
	"fmt"
	"strings"
	"unicode"
)

// Command is the parsed form of one REPL line.
type Command struct {
	Verb     string
	Entity   string
	Target   string
	Relation string
	Related  string
	Fields   map[string]string
	FieldOrder []string
	Raw      string
}

// Tokenize splits a line into tokens, honoring double-quoted strings so that
// `full_name "John Doe"` yields two tokens.
func Tokenize(line string) ([]string, error) {
	var tokens []string
	var buf strings.Builder
	inQuote := false
	escaped := false
	for _, r := range line {
		switch {
		case escaped:
			buf.WriteRune(r)
			escaped = false
		case r == '\\' && inQuote:
			escaped = true
		case r == '"':
			inQuote = !inQuote
		case !inQuote && unicode.IsSpace(r):
			if buf.Len() > 0 {
				tokens = append(tokens, buf.String())
				buf.Reset()
			}
		default:
			buf.WriteRune(r)
		}
	}
	if inQuote {
		return nil, fmt.Errorf("unterminated quoted string")
	}
	if buf.Len() > 0 {
		tokens = append(tokens, buf.String())
	}
	return tokens, nil
}

// Parse converts a raw line into a Command or an error with a user-facing message.
func Parse(line string) (*Command, error) {
	tokens, err := Tokenize(line)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, nil
	}
	// --help / -h anywhere in the line short-circuits to context-specific help.
	// The topic is derived from the tokens preceding the flag: "set user",
	// "show ne", etc. — or just the verb if no entity follows.
	if topic, ok := extractHelpTopic(tokens); ok {
		return &Command{Verb: "help", Target: topic, Raw: line, Fields: map[string]string{}}, nil
	}
	verb := strings.ToLower(tokens[0])
	c := &Command{Verb: verb, Raw: line, Fields: map[string]string{}}

	switch verb {
	case "exit", "quit":
		if len(tokens) != 1 {
			return nil, fmt.Errorf("%s takes no arguments", verb)
		}
		return c, nil
	case "help":
		if len(tokens) > 3 {
			return nil, fmt.Errorf("help takes at most two arguments (verb [entity])")
		}
		parts := make([]string, 0, 2)
		for _, t := range tokens[1:] {
			parts = append(parts, strings.ToLower(t))
		}
		c.Target = strings.Join(parts, " ")
		return c, nil
	case "show", "set", "update", "delete", "map", "unmap":
	default:
		return nil, fmt.Errorf("unknown command %q", verb)
	}

	if len(tokens) < 2 {
		return nil, fmt.Errorf("%s requires an entity (user, ne, group)", verb)
	}
	c.Entity = strings.ToLower(tokens[1])
	switch c.Entity {
	case "user", "ne", "group":
	default:
		return nil, fmt.Errorf("unknown entity %q (expected user, ne, group)", c.Entity)
	}

	rest := tokens[2:]
	switch verb {
	case "show":
		switch len(rest) {
		case 0:
			// list all
		case 1:
			// legacy single-target: interpreted as name or id
			c.Target = rest[0]
		case 2:
			// <field> <value> filter form
			field := strings.ToLower(rest[0])
			c.Fields[field] = rest[1]
			c.FieldOrder = append(c.FieldOrder, field)
		default:
			return nil, fmt.Errorf("show %s takes at most <field> <value> or one target", c.Entity)
		}
	case "delete":
		if len(rest) != 1 {
			return nil, fmt.Errorf("delete %s requires exactly one target (name or id)", c.Entity)
		}
		c.Target = rest[0]
	case "set":
		if len(rest) == 0 {
			return nil, fmt.Errorf("set %s requires field/value pairs", c.Entity)
		}
		if err := parsePairs(rest, c); err != nil {
			return nil, err
		}
	case "update":
		if len(rest) < 3 {
			return nil, fmt.Errorf("update %s requires <target> <field> <value> ...", c.Entity)
		}
		c.Target = rest[0]
		if err := parsePairs(rest[1:], c); err != nil {
			return nil, err
		}
	case "map", "unmap":
		if len(rest) != 3 {
			return nil, fmt.Errorf("%s %s requires <target> <relation> <related> (e.g. %s user alice ne HTSMF01)", verb, c.Entity, verb)
		}
		c.Target = rest[0]
		c.Relation = strings.ToLower(rest[1])
		c.Related = rest[2]
		if err := validateMapShape(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// extractHelpTopic returns the help topic and true if any token is --help or
// -h. The topic is "verb entity" when an entity is present after the verb,
// otherwise just the verb; empty when the flag is alone on the line. The
// flag itself is stripped before deriving the topic so `set --help user`
// and `set user --help` both produce "set user".
func extractHelpTopic(tokens []string) (string, bool) {
	clean := make([]string, 0, len(tokens))
	hasHelp := false
	for _, t := range tokens {
		if t == "--help" || t == "-h" {
			hasHelp = true
			continue
		}
		clean = append(clean, t)
	}
	if !hasHelp {
		return "", false
	}
	if len(clean) == 0 {
		return "", true
	}
	verb := strings.ToLower(clean[0])
	if len(clean) >= 2 {
		ent := strings.ToLower(clean[1])
		if ent == "user" || ent == "ne" || ent == "group" {
			return verb + " " + ent, true
		}
	}
	return verb, true
}

func parsePairs(tokens []string, c *Command) error {
	if len(tokens)%2 != 0 {
		return fmt.Errorf("expected field/value pairs, got %d tokens (odd count)", len(tokens))
	}
	for i := 0; i < len(tokens); i += 2 {
		key := strings.ToLower(tokens[i])
		val := tokens[i+1]
		if _, dup := c.Fields[key]; dup {
			return fmt.Errorf("field %q specified more than once", key)
		}
		c.Fields[key] = val
		c.FieldOrder = append(c.FieldOrder, key)
	}
	return nil
}

func validateMapShape(c *Command) error {
	switch c.Entity {
	case "user":
		if c.Relation != "ne" && c.Relation != "group" {
			return fmt.Errorf("map user supports relations ne|group, got %q", c.Relation)
		}
	case "group":
		if c.Relation != "ne" {
			return fmt.Errorf("map group supports relation ne, got %q", c.Relation)
		}
	case "ne":
		return fmt.Errorf("map ne is not supported; use map user <u> ne <ne> or map group <g> ne <ne>")
	}
	return nil
}
