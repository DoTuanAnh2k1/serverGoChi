package sshcli

import (
	"bytes"
	"sort"
	"strings"
	"testing"
)

func TestCandidates_Verbs(t *testing.T) {
	got := Candidates("", 0)
	sort.Strings(got)
	want := append([]string(nil), topVerbs...)
	sort.Strings(want)
	if !equalSlice(got, want) {
		t.Errorf("empty: got %v want %v", got, want)
	}

	got = Candidates("sh", 2)
	if len(got) != 1 || got[0] != "show" {
		t.Errorf("sh: got %v want [show]", got)
	}

	got = Candidates("se", 2)
	sort.Strings(got)
	if !equalSlice(got, []string{"set"}) {
		t.Errorf("se: got %v", got)
	}
}

func TestCandidates_Entities(t *testing.T) {
	got := Candidates("show ", 5)
	sort.Strings(got)
	if !equalSlice(got, []string{"command-def", "command-group", "group", "ne", "ne-profile", "user"}) {
		t.Errorf("show _: %v", got)
	}

	got = Candidates("show u", 6)
	if len(got) != 1 || got[0] != "user" {
		t.Errorf("show u: %v", got)
	}

	// "set n" should match both 'ne' and 'ne-profile' now.
	got = Candidates("set n", 5)
	sort.Strings(got)
	if !equalSlice(got, []string{"ne", "ne-profile"}) {
		t.Errorf("set n: %v", got)
	}
}

func TestCandidates_SetUserFields(t *testing.T) {
	// After "set user ", the next token is a field.
	got := Candidates("set user ", 9)
	if !containsString(got, "name") || !containsString(got, "password") {
		t.Errorf("set user _: expected name/password in %v", got)
	}

	// After "set user name alice ", the next token is a field again,
	// and should not include "name" or its alias.
	got = Candidates("set user name alice ", 20)
	if containsString(got, "name") {
		t.Errorf("set user name alice _: should skip used field, got %v", got)
	}
	if !containsString(got, "password") {
		t.Errorf("set user name alice _: expected password, got %v", got)
	}
}

func TestCandidates_ShowUserFilterField(t *testing.T) {
	got := Candidates("show user ", 10)
	// Should include filter aliases: name, id, email, role, ...
	for _, want := range []string{"name", "id", "email", "role"} {
		if !containsString(got, want) {
			t.Errorf("show user _: missing %q in %v", want, got)
		}
	}
}

func TestCandidates_ShowUserRoleEnum(t *testing.T) {
	got := Candidates("show user role ", 15)
	sort.Strings(got)
	want := []string{"Admin", "Normal", "SuperAdmin"}
	sort.Strings(want)
	if !equalSlice(got, want) {
		t.Errorf("show user role _: got %v want %v", got, want)
	}
}

func TestCandidates_ShowNeFilterField(t *testing.T) {
	got := Candidates("show ne ", 8)
	for _, want := range []string{"name", "id", "site", "namespace"} {
		if !containsString(got, want) {
			t.Errorf("show ne _: missing %q in %v", want, got)
		}
	}
}

func TestCandidates_SetEnumValue(t *testing.T) {
	got := Candidates("set user name a password b account_type ", 40)
	sort.Strings(got)
	if !equalSlice(got, []string{"1", "2"}) {
		t.Errorf("account_type value: got %v", got)
	}
}

func TestCandidates_SetNeConfMode(t *testing.T) {
	got := Candidates("set ne conf_mode ", 17)
	sort.Strings(got)
	want := []string{"NETCONF", "RESTCONF", "SSH", "TELNET"}
	sort.Strings(want)
	if !equalSlice(got, want) {
		t.Errorf("conf_mode: got %v want %v", got, want)
	}
}

func TestCandidates_UpdatePairs(t *testing.T) {
	// Target position — no candidates.
	got := Candidates("update user ", 12)
	if len(got) != 0 {
		t.Errorf("update user _ (target): expected none, got %v", got)
	}
	// After target, field name.
	got = Candidates("update user alice ", 18)
	if !containsString(got, "email") {
		t.Errorf("update user alice _: expected email in %v", got)
	}
}

func TestCandidates_MapRelation(t *testing.T) {
	got := Candidates("map user alice ", 15)
	sort.Strings(got)
	if !equalSlice(got, []string{"group", "ne"}) {
		t.Errorf("map user alice _: got %v", got)
	}
	got = Candidates("map group dev ", 14)
	if !equalSlice(got, []string{"ne"}) {
		t.Errorf("map group dev _: got %v", got)
	}
}

func TestCycleState(t *testing.T) {
	var s CycleState
	// Six entities are exposed: command-def, command-group, group, ne,
	// ne-profile, user. Cycling 7 times should return to the first.
	line := "show "
	pos := len(line)
	first, ok := s.Next(line, pos)
	if !ok {
		t.Fatal("expected candidates")
	}
	seen := map[string]bool{first: true}
	for i := 0; i < 5; i++ {
		next, _ := s.Next(line, pos)
		seen[next] = true
	}
	if len(seen) != 6 {
		t.Errorf("rotation produced duplicates before full cycle: seen=%v", seen)
	}
	seventh, _ := s.Next(line, pos)
	if seventh != first {
		t.Errorf("rotation didn't wrap at 7th: first=%q, seventh=%q", first, seventh)
	}

	// New prefix → reset.
	line = "show u"
	pos = len(line)
	next, ok := s.Next(line, pos)
	if !ok || next != "user" {
		t.Errorf("new prefix: got %q (ok=%v), want user", next, ok)
	}
}

func TestCycleState_NoCandidates(t *testing.T) {
	var s CycleState
	// "update user " → target position, no candidates.
	_, ok := s.Next("update user ", 12)
	if ok {
		t.Errorf("should have no candidates")
	}
}

// TestAutoCompleteCallback_CyclesThroughEntities simulates the x/term loop:
// after the callback returns (newLine, newPos) the terminal updates its line,
// and the next Tab is fired with that updated line. The previous implementation
// forgot about this and only ever returned the first candidate.
func TestAutoCompleteCallback_CyclesThroughEntities(t *testing.T) {
	cb := makeAutoComplete(nil, 12, nil)
	line, pos := "show ", 5
	seen := map[string]bool{}
	var seq []string
	// Six entities → seven tabs to confirm wrap.
	for i := 0; i < 7; i++ {
		nl, np, ok := cb(line, pos, '\t')
		if !ok {
			t.Fatalf("tab %d: callback returned !ok", i)
		}
		line, pos = nl, np
		seq = append(seq, line[5:])
		seen[line[5:]] = true
	}
	if len(seen) != 6 {
		t.Errorf("expected 6 distinct entities across cycle, got seq=%v", seq)
	}
	if seq[0] != seq[6] {
		t.Errorf("expected wrap at 7th: seq[0]=%q seq[6]=%q", seq[0], seq[6])
	}
}

// When the combined prompt + input would wrap past the terminal's width, the
// DECSC/DECRC dance used for the hint breaks (it assumes the cursor shares
// the prompt row). The callback must keep cycling through candidates but
// SKIP the hint write so the terminal display doesn't get corrupted.
func TestAutoCompleteCallback_SkipsHintWhenLineWraps(t *testing.T) {
	var hint bytes.Buffer
	// Tight terminal: width 40, prompt width 12. A 60-char input plus the
	// prompt is well past wrap.
	cb := makeAutoComplete(&hint, 12, func() int { return 40 })
	line := "show " + strings.Repeat("x", 60)
	pos := len(line)
	_, _, ok := cb(line, pos, '\t')
	// No candidates for a gibberish 2nd token, so the callback returns !ok —
	// but that's not what we're testing. Force Tab on "show " then append.
	_ = ok
	shortLine := "show "
	if _, _, ok := cb(shortLine, len(shortLine), '\t'); !ok {
		t.Fatal("prep: expected candidates for 'show '")
	}
	hint.Reset()
	// Retry on a long line that wraps — still 'show <long>' at position 5.
	cb2 := makeAutoComplete(&hint, 12, func() int { return 40 })
	wrapLine := "show " + strings.Repeat("x", 30) // 35 chars; 12+35+2=49 > 40
	if _, _, ok := cb2(wrapLine, 5, '\t'); !ok {
		t.Fatal("expected candidates on 'show <wrap>'")
	}
	if hint.Len() != 0 {
		t.Errorf("hint must be suppressed when line wraps, got %q", hint.String())
	}
}

func TestAutoCompleteCallback_EmitsHintWhenRoomAvailable(t *testing.T) {
	var hint bytes.Buffer
	cb := makeAutoComplete(&hint, 12, func() int { return 200 })
	if _, _, ok := cb("show ", 5, '\t'); !ok {
		t.Fatal("expected candidates")
	}
	if hint.Len() == 0 {
		t.Errorf("hint must be emitted when the line fits in the terminal")
	}
}

func TestAutoCompleteCallback_ResetOnNonTab(t *testing.T) {
	cb := makeAutoComplete(nil, 12, nil)
	line, pos := "show ", 5
	nl, np, ok := cb(line, pos, '\t')
	if !ok {
		t.Fatal("first tab should return a candidate")
	}
	line, pos = nl, np
	// Non-tab key resets; subsequent tab starts fresh.
	if _, _, rok := cb(line, pos, 'x'); rok {
		t.Errorf("non-tab key should return !ok")
	}
}

func TestMenuAutoComplete_CyclesThroughModes(t *testing.T) {
	cb := makeMenuAutoComplete(nil, nil)
	line, pos := "", 0
	seen := map[string]bool{}
	var seq []string
	for i := 0; i < 4; i++ {
		nl, np, ok := cb(line, pos, '\t')
		if !ok {
			t.Fatalf("tab %d: !ok", i)
		}
		line, pos = nl, np
		seq = append(seq, line)
		seen[line] = true
	}
	if len(seen) != 3 {
		t.Errorf("expected 3 distinct modes, got seq=%v", seq)
	}
	if seq[0] != seq[3] {
		t.Errorf("expected wrap on 4th tab: %v", seq)
	}
	for _, mode := range []string{"cli-config", "ne-config", "ne-command"} {
		if !seen[mode] {
			t.Errorf("missing mode %q in cycle; got %v", mode, seq)
		}
	}
}

func TestMenuAutoComplete_PrefixFilter(t *testing.T) {
	cb := makeMenuAutoComplete(nil, nil)
	// "ne" should cycle between ne-config and ne-command only.
	line, pos := "ne", 2
	first, _, ok := cb(line, pos, '\t')
	if !ok {
		t.Fatal("tab: !ok")
	}
	line, pos = first, len(first)
	second, _, _ := cb(line, pos, '\t')
	line, pos = second, len(second)
	third, _, _ := cb(line, pos, '\t')
	if first != third {
		t.Errorf("2-item cycle should wrap: first=%q third=%q", first, third)
	}
	if first == second {
		t.Errorf("expected two distinct ne-* modes, got %q twice", first)
	}
}

func equalSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
