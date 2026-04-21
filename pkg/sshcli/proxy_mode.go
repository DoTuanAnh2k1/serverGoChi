package sshcli

import (
	"fmt"
	"io"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

// ProxySession transparently forwards the current incoming SSH session to an
// upstream SSH server, reusing the same username/password the user just
// supplied. The upstream server is expected to accept those credentials 1:1.
type ProxySession struct {
	UpstreamAddr string
	Username     string
	Password     string
	// PTY holds details of the PTY requested on the incoming session, so we
	// can negotiate the same PTY upstream and forward window-change signals.
	Term   string
	Width  uint32
	Height uint32
	Modes  ssh.TerminalModes
	// WindowChanges receives size-change events (w,h) from the incoming
	// session; pump them into WindowChanges to have them forwarded.
	WindowChanges <-chan WindowSize
}

type WindowSize struct{ Width, Height uint32 }

// Run dials upstream, opens a shell channel, pipes both streams, and returns
// once the session ends (or on a hard error).
func (p *ProxySession) Run(incoming io.ReadWriter) error {
	cfg := &ssh.ClientConfig{
		User: p.Username,
		Auth: []ssh.AuthMethod{ssh.Password(p.Password)},
		// The upstream SSH servers here are internal (same private network),
		// so we accept any host key rather than pinning.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	conn, err := net.DialTimeout("tcp", p.UpstreamAddr, cfg.Timeout)
	if err != nil {
		return fmt.Errorf("dial %s: %w", p.UpstreamAddr, err)
	}
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, p.UpstreamAddr, cfg)
	if err != nil {
		conn.Close()
		return fmt.Errorf("ssh handshake: %w", err)
	}
	client := ssh.NewClient(sshConn, chans, reqs)
	defer client.Close()

	sess, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("new session: %w", err)
	}
	defer sess.Close()

	term := p.Term
	if term == "" {
		term = "xterm"
	}
	w, h := p.Width, p.Height
	if w == 0 {
		w = 80
	}
	if h == 0 {
		h = 24
	}
	if err := sess.RequestPty(term, int(h), int(w), p.Modes); err != nil {
		return fmt.Errorf("request pty: %w", err)
	}

	upIn, err := sess.StdinPipe()
	if err != nil {
		return err
	}
	upOut, err := sess.StdoutPipe()
	if err != nil {
		return err
	}
	upErr, err := sess.StderrPipe()
	if err != nil {
		return err
	}

	if err := sess.Shell(); err != nil {
		return fmt.Errorf("shell: %w", err)
	}

	// Window-change forwarder.
	if p.WindowChanges != nil {
		go func() {
			for ws := range p.WindowChanges {
				// "window-change" is SSH_MSG_CHANNEL_REQUEST 'window-change'
				// with width, height, widthPx, heightPx (all uint32).
				payload := sshPayloadWinchg(ws.Width, ws.Height)
				_, _ = sess.SendRequest("window-change", false, payload)
			}
		}()
	}

	// Bidirectional copy. We return when either side closes: Wait returns
	// when upstream closes; copying stdin returns when the local side closes.
	done := make(chan error, 3)
	go func() {
		_, err := io.Copy(upIn, incoming)
		_ = upIn.Close()
		done <- err
	}()
	go func() {
		_, err := io.Copy(incoming, upOut)
		done <- err
	}()
	go func() {
		_, err := io.Copy(incoming, upErr)
		done <- err
	}()

	// Wait for the upstream shell to exit; that's the authoritative signal.
	waitErr := sess.Wait()
	// Drain one copy-worker to avoid leaking the goroutine when upstream
	// closes first; the other copy goroutines will unblock once their pipes
	// are closed via defer.
	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
	}

	if waitErr != nil {
		// Exit-status non-zero is not a CLI error.
		if _, ok := waitErr.(*ssh.ExitError); ok {
			return nil
		}
		return waitErr
	}
	return nil
}

// sshPayloadWinchg marshals a window-change payload (width, height, wPx, hPx).
func sshPayloadWinchg(w, h uint32) []byte {
	b := make([]byte, 16)
	putUint32(b[0:4], w)
	putUint32(b[4:8], h)
	putUint32(b[8:12], 0)
	putUint32(b[12:16], 0)
	return b
}

func putUint32(b []byte, v uint32) {
	b[0] = byte(v >> 24)
	b[1] = byte(v >> 16)
	b[2] = byte(v >> 8)
	b[3] = byte(v)
}
