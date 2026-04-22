package sshcli

import (
	"sort"
	"strings"
	"unicode"
)

var (
	topVerbs = []string{"show", "set", "update", "delete", "map", "unmap", "help", "exit"}
	entities = []string{"user", "ne", "group"}
)

// Candidates returns the list of completion candidates that extend the current
// line up to cursor position `pos`. It's the pure-logic half of tab completion;
// the Terminal integration is in config_mode.go.
//
// The returned list is sorted, contains whole tokens (not differences), and
// all candidates have the same prefix as the token currently being typed.
func Candidates(line string, pos int) []string {
	if pos > len(line) {
		pos = len(line)
	}
	head := line[:pos]
	// Split head into tokens. We care about which token we are in — if the
	// cursor is on whitespace, we're starting a new token (empty prefix at
	// index = len(prevTokens)).
	tokens, cur, prefix := splitForCompletion(head)
	all := candidatesAt(tokens, cur)
	var out []string
	for _, c := range all {
		if strings.HasPrefix(c, prefix) {
			out = append(out, c)
		}
	}
	sort.Strings(out)
	return out
}

// splitForCompletion returns the tokens already typed, the index of the token
// currently being edited, and the prefix of that token. If the cursor is on
// whitespace just after a completed token, the "current" token is empty and
// cur = len(tokens).
func splitForCompletion(head string) (tokens []string, cur int, prefix string) {
	var buf strings.Builder
	inQuote := false
	for _, r := range head {
		switch {
		case r == '"':
			inQuote = !inQuote
			buf.WriteRune(r)
		case !inQuote && unicode.IsSpace(r):
			if buf.Len() > 0 {
				tokens = append(tokens, buf.String())
				buf.Reset()
			}
		default:
			buf.WriteRune(r)
		}
	}
	if buf.Len() > 0 {
		prefix = buf.String()
		cur = len(tokens)
		return
	}
	cur = len(tokens)
	prefix = ""
	return
}

func candidatesAt(prev []string, index int) []string {
	switch index {
	case 0:
		return topVerbs
	case 1:
		verb := strings.ToLower(prev[0])
		switch verb {
		case "show", "set", "update", "delete", "map", "unmap":
			return entities
		case "help":
			return topVerbs
		default:
			return nil
		}
	}
	verb := strings.ToLower(prev[0])
	entity := strings.ToLower(prev[1])
	switch verb {
	case "set":
		return fieldOrEnumCandidate(entity, prev, 2)
	case "update":
		// prev[2] is target (free input); pairs start at index 3.
		if index <= 2 {
			return nil
		}
		return fieldOrEnumCandidate(entity, prev, 3)
	case "map", "unmap":
		// prev[2] target (free), prev[3] relation, prev[4] related.
		switch index {
		case 3:
			return relationCandidates(entity)
		default:
			return nil
		}
	case "show":
		return showFilterCandidate(entity, prev, 2)
	}
	return nil
}

// showFilterCandidate suggests filter field aliases at index 2 of a show
// command, and enum values at index 3 when the preceding token is a known
// filter field with an enum (e.g. role → SuperAdmin/Admin/Normal).
func showFilterCandidate(entity string, prev []string, pairsStart int) []string {
	relOffset := len(prev) - pairsStart
	if relOffset%2 == 0 {
		return ShowFilterAliases(entity)
	}
	fieldAlias := strings.ToLower(prev[len(prev)-1])
	canon, ok := ResolveShowFilter(entity, fieldAlias)
	if !ok {
		return nil
	}
	return ShowFilterEnumValues(entity, canon)
}

// fieldOrEnumCandidate decides, given that pairs start at `pairsStart`, whether
// we're about to type a field name or an enum-constrained value.
func fieldOrEnumCandidate(entity string, prev []string, pairsStart int) []string {
	spec, ok := entitySpecs[entity]
	if !ok {
		return nil
	}
	if len(prev) < pairsStart {
		return nil
	}
	relOffset := len(prev) - pairsStart
	// Even offset → next token is a field name; odd → it's a value.
	if relOffset%2 == 0 {
		// Field name. Skip already-used canonical fields.
		used := map[string]bool{}
		for i := pairsStart; i+1 < len(prev); i += 2 {
			if canon, ok := spec.FieldAliases[strings.ToLower(prev[i])]; ok {
				used[canon] = true
			}
		}
		out := []string{}
		for alias, canon := range spec.FieldAliases {
			if !used[canon] {
				out = append(out, alias)
			}
		}
		return out
	}
	// Value position. Offer enum values if the preceding field is an enum field.
	fieldAlias := strings.ToLower(prev[len(prev)-1])
	canon, ok := spec.FieldAliases[fieldAlias]
	if !ok {
		return nil
	}
	if allowed, ok := spec.EnumFields[canon]; ok {
		return append([]string(nil), allowed...)
	}
	return nil
}

func relationCandidates(entity string) []string {
	switch entity {
	case "user":
		return []string{"ne", "group"}
	case "group":
		return []string{"ne"}
	}
	return nil
}

// CycleState tracks a tab-cycle session so repeated tab presses rotate through
// candidates for the same prefix. It is reset whenever the prefix or cursor
// position changes.
type CycleState struct {
	lastLine string
	lastPos  int
	list     []string
	idx      int
}

// Next returns the next candidate to insert given the current line/pos, or
// ("", false) if there are no candidates. It rotates on consecutive calls with
// an unchanged (line, pos) pair.
func (s *CycleState) Next(line string, pos int) (string, bool) {
	if line == s.lastLine && pos == s.lastPos && len(s.list) > 0 {
		s.idx = (s.idx + 1) % len(s.list)
		return s.list[s.idx], true
	}
	s.list = Candidates(line, pos)
	if len(s.list) == 0 {
		s.lastLine, s.lastPos, s.idx = line, pos, 0
		return "", false
	}
	s.idx = 0
	s.lastLine, s.lastPos = line, pos
	return s.list[0], true
}

// Reset clears the cycle state.
func (s *CycleState) Reset() {
	s.lastLine = ""
	s.lastPos = -1
	s.list = nil
	s.idx = 0
}
