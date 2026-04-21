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
// The session runner is used to register a resize sink so the cli-config
// terminal's word-wrap tracks the client's real PTY size instead of the
// x/term default of 80x24.
func RunConfigMode(sess io.ReadWriter, client *MgtClient, sr *SessionRunner) error {
	t := term.NewTerminal(sess, "")
	t.SetPrompt("cli-config> ")
	t.AutoCompleteCallback = makeAutoComplete(sess)
	if sr != nil {
		sr.SetResizeSink(func(w, h uint32) { applyTermSize(t, w, h) })
		defer sr.SetResizeSink(nil)
	}
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
// Candidates on each tab press. Any non-tab key resets the cycle. Cycling
// works by remembering the (line, pos) we returned from the previous Tab:
// if the next Tab fires with the same line/pos, the user hasn't typed
// anything and we advance the index. Otherwise we start a fresh cycle.
//
// The candidate list is drawn on the line below the prompt. To avoid the
// known bug where the prompt sits at the terminal's bottom row and `\r\n`
// scrolls the screen — which would invalidate the DECSC-saved position and
// cause the prompt to be overwritten by the hint — we first emit IND+CUU
// (`\x1bD\x1b[A`). IND scrolls once if we're at the bottom, and CUU puts
// the cursor back on the prompt row (row N-1 after a scroll, or unchanged
// otherwise). From that point, DECSC/`\r\n`/DECRC is safe because the row
// below the prompt is guaranteed to exist.
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
		fmt.Fprint(hintW, "\x1bD\x1b[A\x1b7\r\n\x1b[2K\x1b8")
		st.hintShown = false
	}
	showHint := func(opts []string) {
		if hintW == nil || len(opts) <= 1 {
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
