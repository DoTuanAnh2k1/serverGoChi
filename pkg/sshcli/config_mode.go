package sshcli

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync/atomic"

	"golang.org/x/term"
)

// RunConfigMode starts an interactive REPL on the given SSH session io, bound
// to the given MgtClient. It returns when the user exits or the channel closes.
// The session runner is used to register a resize sink so the cli-config
// terminal's word-wrap tracks the client's real PTY size instead of the
// x/term default of 80x24.
func RunConfigMode(sess io.ReadWriter, client *MgtClient, sr *SessionRunner) error {
	const mainPrompt = "cli-config> "
	t := term.NewTerminal(sess, "")
	t.SetPrompt(mainPrompt)

	// termWidth is read by the autocomplete callback to skip the hint line
	// when the current input would wrap to a second screen row — the DECSC
	// trick that draws the hint assumes the cursor shares the prompt row,
	// which isn't true once x/term wraps long lines across multiple rows.
	var termWidth atomic.Int32
	termWidth.Store(80)
	widthFn := func() int {
		v := termWidth.Load()
		if v < 1 {
			return 80
		}
		return int(v)
	}

	t.AutoCompleteCallback = makeAutoComplete(sess, len(mainPrompt), widthFn)
	if sr != nil {
		sr.SetResizeSink(func(w, h uint32) {
			applyTermSize(t, w, h)
			if w >= 1 {
				termWidth.Store(int32(w))
			}
		})
		defer sr.SetResizeSink(nil)
	}
	// Confirm prompt for destructive ops (delete). Temporarily disables the
	// autocomplete callback so Tab doesn't try to complete the y/N answer,
	// swaps the prompt, reads a line, then restores both. Anything other
	// than "y"/"yes" (case-insensitive, trimmed) — including an empty line,
	// an embedded Ctrl+C byte, or a read error — counts as abort.
	confirm := func(msg string) bool {
		prevAC := t.AutoCompleteCallback
		t.AutoCompleteCallback = nil
		t.SetPrompt(msg + " [y/N]: ")
		defer func() {
			t.SetPrompt(mainPrompt)
			t.AutoCompleteCallback = prevAC
		}()
		line, err := t.ReadLine()
		if err != nil {
			return false
		}
		ans := strings.ToLower(strings.TrimSpace(line))
		return ans == "y" || ans == "yes"
	}
	d := &Dispatcher{Client: client, Out: t, Confirm: confirm}

	fmt.Fprint(t, "\r\n== cli-config mode ==\r\n")
	fmt.Fprint(t, "Type 'help' for commands. Type 'exit' to return to the mode menu.\r\n\r\n")

	for {
		line, err := t.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		cmd, perr := Parse(line)
		if perr != nil {
			fmt.Fprintf(t, "error: %s\r\n", perr)
			continue
		}
		if cmd == nil {
			continue
		}
		if rerr := d.Run(cmd); rerr != nil {
			if errors.Is(rerr, io.EOF) {
				return nil
			}
			fmt.Fprintf(t, "error: %s\r\n", rerr)
		}
	}
}

// makeAutoComplete returns a term.AutoCompleteCallback that rotates through
// Candidates on each tab press. Any non-tab key resets the cycle. Cycling
// works by remembering the (line, pos) we returned from the previous Tab:
// if the next Tab fires with the same line/pos, the user hasn't typed
// anything and we advance the index. Otherwise we start a fresh cycle.
//
// The candidate list is drawn on the line below the prompt via DECSC/DECRC.
// The trick assumes the cursor shares the prompt row — which breaks once
// the input wraps to a second screen row. To keep the terminal sane on
// long commands we pass the prompt length + a width probe, and skip the
// hint whenever `prompt + line + 2` would reach the terminal width (the
// "+2" leaves room for x/term's own cursor marker and the deferred char
// it sometimes reserves at the right margin). Tab cycling still works —
// only the visual hint is suppressed.
//
// promptLen = number of printable columns the prompt consumes (e.g. 12 for
// "cli-config> "); widthFn returns the latest known terminal width (>=1).
func makeAutoComplete(hintW io.Writer, promptLen int, widthFn func() int) func(line string, pos int, key rune) (string, int, bool) {
	var st struct {
		list      []string
		idx       int
		start     int
		lastLine  string
		lastPos   int
		active    bool
		hintShown bool
	}
	safeWidth := func() int {
		if widthFn == nil {
			return 80
		}
		w := widthFn()
		if w < 1 {
			return 80
		}
		return w
	}
	// wouldWrap reports whether the current (prompt + line) reaches or
	// exceeds the terminal width — which would force x/term to wrap the
	// input display across multiple rows, invalidating our hint geometry.
	wouldWrap := func(line string) bool {
		return promptLen+len(line)+2 >= safeWidth()
	}
	eraseHint := func() {
		if !st.hintShown || hintW == nil {
			return
		}
		fmt.Fprint(hintW, "\x1bD\x1b[A\x1b7\r\n\x1b[2K\x1b8")
		st.hintShown = false
	}
	showHint := func(opts []string, line string) {
		if hintW == nil || len(opts) <= 1 {
			return
		}
		if wouldWrap(line) {
			// Skip the hint entirely on a wrapping line — attempting the
			// DECSC/DECRC dance with a multi-row cursor overwrites the
			// wrapped input. Tab cycling still works; user just won't see
			// the candidate list preview.
			return
		}
		fmt.Fprintf(hintW, "\x1bD\x1b[A\x1b7\r\n\x1b[2K%s\x1b8", strings.Join(opts, "  "))
		st.hintShown = true
	}
	return func(line string, pos int, key rune) (string, int, bool) {
		if key != '\t' {
			eraseHint()
			st.active = false
			return "", 0, false
		}
		if st.active && line == st.lastLine && pos == st.lastPos && len(st.list) > 0 {
			st.idx = (st.idx + 1) % len(st.list)
		} else {
			start := tokenStart(line, pos)
			list := Candidates(line, pos)
			if len(list) == 0 {
				eraseHint()
				st.active = false
				return "", 0, false
			}
			eraseHint()
			st.list = list
			st.idx = 0
			st.start = start
			st.active = true
			showHint(list, line)
		}
		cand := st.list[st.idx]
		newLine := line[:st.start] + cand + line[pos:]
		newPos := st.start + len(cand)
		st.lastLine = newLine
		st.lastPos = newPos
		return newLine, newPos, true
	}
}

// tokenStart returns the byte index of the start of the token ending at pos.
func tokenStart(line string, pos int) int {
	if pos > len(line) {
		pos = len(line)
	}
	i := pos
	inQuote := false
	// Scan the line once to know our quote state at pos.
	for _, r := range line[:pos] {
		if r == '"' {
			inQuote = !inQuote
		}
	}
	for i > 0 {
		r := rune(line[i-1])
		if inQuote {
			if r == '"' {
				break
			}
		} else {
			if r == ' ' || r == '\t' {
				break
			}
		}
		i--
	}
	return i
}
