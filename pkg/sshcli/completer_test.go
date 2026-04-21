package sshcli

import (
	"sort"
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
	if !equalSlice(got, []string{"group", "ne", "user"}) {
		t.Errorf("show _: %v", got)
	}

	got = Candidates("show u", 6)
	if len(got) != 1 || got[0] != "user" {
		t.Errorf("show u: %v", got)
	}

	got = Candidates("set n", 5)
	if len(got) != 1 || got[0] != "ne" {
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
	// "show " → 3 entities
	line := "show "
	pos := len(line)
	first, ok := s.Next(line, pos)
	if !ok {
		t.Fatal("expected candidates")
	}
	second, _ := s.Next(line, pos)
	third, _ := s.Next(line, pos)
	fourth, _ := s.Next(line, pos)
	seen := map[string]bool{first: true, second: true, third: true}
	if len(seen) != 3 {
		t.Errorf("rotation produced duplicates before full cycle: %q %q %q", first, second, third)
	}
	if fourth != first {
		t.Errorf("rotation didn't wrap: first=%q, fourth=%q", first, fourth)
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
	cb := makeAutoComplete(nil)
	line, pos := "show ", 5
	seen := map[string]bool{}
	var seq []string
	for i := 0; i < 4; i++ {
		nl, np, ok := cb(line, pos, '\t')
		if !ok {
			t.Fatalf("tab %d: callback returned !ok", i)
		}
		line, pos = nl, np
		seq = append(seq, line[5:])
		seen[line[5:]] = true
	}
	if len(seen) != 3 {
		t.Errorf("expected 3 distinct entities across cycle, got seq=%v", seq)
	}
	if seq[0] != seq[3] {
		t.Errorf("expected wrap: seq[0]=%q seq[3]=%q", seq[0], seq[3])
	}
}

func TestAutoCompleteCallback_ResetOnNonTab(t *testing.T) {
	cb := makeAutoComplete(nil)
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
	cb := makeMenuAutoComplete(nil)
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
	cb := makeMenuAutoComplete(nil)
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
