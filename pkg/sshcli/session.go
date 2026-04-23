package sshcli

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// Mode represents the three top-level destinations available from the CLI menu.
type Mode string

const (
	ModeCliConfig  Mode = "cli-config"
	ModeNeConfig   Mode = "ne-config"
	ModeNeCommand  Mode = "ne-command"
)

var menuModes = []string{string(ModeCliConfig), string(ModeNeConfig), string(ModeNeCommand)}

// availableModes returns the list of mode names visible to a user of the
// given role. Only SuperAdmin / Admin (role=="admin") see cli-config;
// Normal users (role=="user") see ne-config and ne-command only.
func availableModes(role string) []string {
	if role == "admin" {
		return menuModes
	}
	return []string{string(ModeNeConfig), string(ModeNeCommand)}
}

// isAdminRole reports whether the role tag grants access to cli-config.
func isAdminRole(role string) bool { return role == "admin" }

// SessionRunner glues together the three modes for one SSH login.
type SessionRunner struct {
	Client        *MgtClient
	Username      string
	Password      string
	NeConfigAddr  string
	NeCommandAddr string
	PTYTerm       string
	PTYWidth      uint32
	PTYHeight     uint32
	PTYModes      map[uint8]uint32
	WindowChanges <-chan WindowSize

	// Window-change fan-out. Set by pumpResizes at Run start. Only one mode
	// is active at a time, so a single sink slot is sufficient; the active
	// mode registers via SetResizeSink and clears it on exit.
	sizeMu   sync.Mutex
	sizeSink func(w, h uint32)
	curW     uint32
	curH     uint32
}

// Run loops through the menu until the user exits. It writes a banner and
// prompts for a mode; on return from a mode, it re-prompts.
func (s *SessionRunner) Run(sess io.ReadWriter) error {
	s.sizeMu.Lock()
	s.curW, s.curH = s.PTYWidth, s.PTYHeight
	s.sizeMu.Unlock()
	go s.pumpResizes()

	t := term.NewTerminal(sess, "")
	t.SetPrompt("mode> ")
	modes := availableModes(s.Client.Role)
	t.AutoCompleteCallback = makeMenuAutoComplete(sess, modes)
	menuSink := func(w, h uint32) { applyTermSize(t, w, h) }
	s.SetResizeSink(menuSink)

	banner := fmt.Sprintf("\r\nWelcome %s — management CLI.\r\n", s.Username)
	banner += "Available modes: " + strings.Join(modes, ", ") + " (Tab to cycle / autocomplete, 'exit' to quit).\r\n\r\n"
	fmt.Fprint(sess, banner)

	for {
		line, err := t.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		choice := strings.TrimSpace(strings.ToLower(line))
		switch choice {
		case "":
			continue
		case "exit", "quit":
			fmt.Fprint(sess, "bye.\r\n")
			return nil
		case string(ModeCliConfig):
			if !isAdminRole(s.Client.Role) {
				fmt.Fprint(sess, "cli-config is restricted to admin / superadmin accounts.\r\n")
				continue
			}
			s.SetResizeSink(nil)
			if err := RunConfigMode(sess, s.Client, s); err != nil {
				fmt.Fprintf(sess, "cli-config ended with error: %s\r\n", err)
			}
			s.SetResizeSink(menuSink)
		case string(ModeNeConfig):
			if s.NeConfigAddr == "" {
				fmt.Fprint(sess, "ne-config address not configured on this server.\r\n")
				continue
			}
			s.SetResizeSink(nil)
			s.runProxy(sess, s.NeConfigAddr)
			s.SetResizeSink(menuSink)
		case string(ModeNeCommand):
			if s.NeCommandAddr == "" {
				fmt.Fprint(sess, "ne-command address not configured on this server.\r\n")
				continue
			}
			s.SetResizeSink(nil)
			s.runProxy(sess, s.NeCommandAddr)
			s.SetResizeSink(menuSink)
		default:
			fmt.Fprintf(sess, "unknown mode %q — choose one of: %s\r\n", choice, strings.Join(menuModes, ", "))
		}
	}
}

// pumpResizes consumes WindowChanges forever. Each event updates the cached
// size and fans out to the currently registered sink (if any).
func (s *SessionRunner) pumpResizes() {
	for ws := range s.WindowChanges {
		s.sizeMu.Lock()
		s.curW, s.curH = ws.Width, ws.Height
		fn := s.sizeSink
		s.sizeMu.Unlock()
		if fn != nil {
			fn(ws.Width, ws.Height)
		}
	}
}

// SetResizeSink installs fn as the active window-change handler. fn is
// invoked immediately with the cached size so the new owner sees the current
// dimensions without waiting for the next resize event. Pass nil to clear.
func (s *SessionRunner) SetResizeSink(fn func(w, h uint32)) {
	s.sizeMu.Lock()
	s.sizeSink = fn
	w, h := s.curW, s.curH
	s.sizeMu.Unlock()
	if fn != nil {
		fn(w, h)
	}
}

// CurrentSize returns the last-known PTY size. Used by proxy mode for the
// initial upstream RequestPty call.
func (s *SessionRunner) CurrentSize() (uint32, uint32) {
	s.sizeMu.Lock()
	defer s.sizeMu.Unlock()
	return s.curW, s.curH
}

// applyTermSize applies a PTY size to a term.Terminal, enforcing sane
// minimums so the x/term word-wrap code doesn't divide by zero on early
// events before the client has announced a size.
func applyTermSize(t *term.Terminal, w, h uint32) {
	ww := int(w)
	hh := int(h)
	if ww < 1 {
		ww = 80
	}
	if hh < 1 {
		hh = 24
	}
	_ = t.SetSize(ww, hh)
}

func (s *SessionRunner) runProxy(sess io.ReadWriter, addr string) {
	// Forwarding channel: the sink pushes resize events here for ProxySession
	// to consume and relay upstream. Buffered so a burst of resizes doesn't
	// block the pumpResizes goroutine.
	ch := make(chan WindowSize, 8)
	s.SetResizeSink(func(w, h uint32) {
		select {
		case ch <- WindowSize{Width: w, Height: h}:
		default:
		}
	})
	defer func() {
		s.SetResizeSink(nil)
		close(ch)
	}()

	w, h := s.CurrentSize()
	p := &ProxySession{
		UpstreamAddr:  addr,
		Username:      s.Username,
		Password:      s.Password,
		Term:          s.PTYTerm,
		Width:         w,
		Height:        h,
		Modes:         sshTerminalModes(s.PTYModes),
		WindowChanges: ch,
	}
	fmt.Fprintf(sess, "Connecting to %s ...\r\n", addr)
	if err := p.Run(sess); err != nil {
		fmt.Fprintf(sess, "proxy error: %s\r\n", err)
	}
	fmt.Fprint(sess, "\r\n-- session ended, back to mode menu --\r\n")
}

// makeMenuAutoComplete rotates through the menu modes on repeated Tab presses.
// It detects a "same Tab again" by remembering the (line, pos) it returned
// last: if the terminal fires Tab again with an unchanged line, we advance;
// otherwise we start a fresh cycle using the current line as prefix.
//
// The candidate list is drawn on the line below the prompt. To avoid the
// known bug where the prompt sits at the terminal's bottom row and `\r\n`
// scrolls the screen — which would invalidate the DECSC-saved position and
// cause the prompt to be overwritten by the hint — we first emit IND+CUU
// (`\x1bD\x1b[A`). IND scrolls once if we're at the bottom, and CUU puts
// the cursor back on the prompt row (row N-1 after a scroll, or unchanged
// otherwise). From that point, DECSC/`\r\n`/DECRC is safe because the row
// below the prompt is guaranteed to exist.
func makeMenuAutoComplete(hintW io.Writer, modes []string) func(line string, pos int, key rune) (string, int, bool) {
	if modes == nil {
		modes = menuModes
	}
	var st struct {
		list      []string
		idx       int
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
			prefix := strings.ToLower(strings.TrimLeft(line[:pos], " "))
			list := st.list[:0]
			for _, mode := range modes {
				if strings.HasPrefix(mode, prefix) {
					list = append(list, mode)
				}
			}
			if len(list) == 0 {
				eraseHint()
				st.active = false
				return "", 0, false
			}
			eraseHint()
			st.list = list
			st.idx = 0
			st.active = true
			showHint(list)
		}
		cand := st.list[st.idx]
		st.lastLine = cand
		st.lastPos = len(cand)
		return cand, len(cand), true
	}
}

// sshTerminalModes converts a map of raw mode codes to the typed ssh.TerminalModes.
func sshTerminalModes(raw map[uint8]uint32) ssh.TerminalModes {
	out := ssh.TerminalModes{}
	for k, v := range raw {
		out[k] = v
	}
	return out
}
