package sshcli

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/term"
)

// RunConfigMode starts an interactive REPL on the given SSH session io, bound
// to the given MgtClient. It returns when the user exits or the channel closes.
func RunConfigMode(sess io.ReadWriter, client *MgtClient) error {
	t := term.NewTerminal(sess, "")
	t.SetPrompt("cli-config> ")
	t.AutoCompleteCallback = makeAutoComplete(sess)
	d := &Dispatcher{Client: client, Out: t}

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
// Candidates on each tab press. Any non-tab key resets the cycle.
//
// When a new cycle begins (first Tab after any text change) and there's more
// than one option, we print the list of candidates on the line *below* the
// prompt using DECSC save-cursor (`ESC 7`) + `\r\n` + erase + list + DECRC
// restore (`ESC 8`). The prompt itself is untouched because the cursor is
// restored right after writing the hint.
//
// When the cycle ends (user presses any non-Tab key), the hint line is erased
// with the same save/restore trick so the display doesn't keep a stale list.
//
// Cycling works by remembering the (line, pos) we returned from the previous
// Tab: if the next Tab fires with the same line/pos, the user hasn't typed
// anything and we advance the index. Otherwise we start a fresh cycle.
func makeAutoComplete(hintW io.Writer) func(line string, pos int, key rune) (string, int, bool) {
	var st struct {
		list      []string
		idx       int
		start     int
		lastLine  string
		lastPos   int
		active    bool
		hintShown bool
	}
	eraseHint := func() {
		if !st.hintShown || hintW == nil {
			return
		}
		fmt.Fprint(hintW, "\x1b7\r\n\x1b[2K\x1b8")
		st.hintShown = false
	}
	showHint := func(opts []string) {
		if hintW == nil || len(opts) <= 1 {
			return
		}
		fmt.Fprintf(hintW, "\x1b7\r\n\x1b[2K%s\x1b8", strings.Join(opts, "  "))
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
			showHint(list)
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
