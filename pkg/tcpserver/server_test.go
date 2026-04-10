package tcpserver

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/testutil"
)

func TestMain(m *testing.M) {
	testutil.InitTestLogger()
	os.Exit(m.Run())
}

// startTestServer khởi động server trên port ngẫu nhiên và trả về addr + cleanup fn.
func startTestServer(t *testing.T) (addr string, dataDir string) {
	t.Helper()
	dataDir = t.TempDir()
	srv := New("127.0.0.1:0", dataDir) // port 0 = OS tự chọn
	if err := srv.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	addr = srv.listener.Addr().String()
	t.Cleanup(srv.Stop)
	return addr, dataDir
}

func sendLines(t *testing.T, addr string, lines []string) {
	t.Helper()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial %s: %v", addr, err)
	}
	for _, l := range lines {
		fmt.Fprintln(conn, l)
	}
	conn.Close()
}

func readFile(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines
}

// waitForFile polls until the file exists or timeout.
func waitForFile(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("file %s did not appear within 2s", path)
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestServer_WritesFileOnDisconnect(t *testing.T) {
	addr, dataDir := startTestServer(t)

	sendLines(t, addr, []string{"line one", "line two", "line three"})

	file := filepath.Join(dataDir, "list_subscribers_results.0")
	waitForFile(t, file)

	got := readFile(t, file)
	want := []string{"line one", "line two", "line three"}
	if len(got) != len(want) {
		t.Fatalf("lines: got %d, want %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("line[%d]: got %q, want %q", i, got[i], w)
		}
	}
}

func TestServer_IndexIncrementsOnSecondConnection(t *testing.T) {
	addr, dataDir := startTestServer(t)

	sendLines(t, addr, []string{"session-1"})
	waitForFile(t, filepath.Join(dataDir, "list_subscribers_results.0"))

	sendLines(t, addr, []string{"session-2"})
	waitForFile(t, filepath.Join(dataDir, "list_subscribers_results.1"))

	got := readFile(t, filepath.Join(dataDir, "list_subscribers_results.1"))
	if len(got) != 1 || got[0] != "session-2" {
		t.Errorf("second file: got %v, want [session-2]", got)
	}
}

func TestServer_SkipsExistingIndexes(t *testing.T) {
	addr, dataDir := startTestServer(t)

	// Pre-create .0 and .1 manually to simulate existing files
	for _, idx := range []int{0, 1} {
		p := filepath.Join(dataDir, fmt.Sprintf("list_subscribers_results.%d", idx))
		if err := os.WriteFile(p, []byte("existing\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	sendLines(t, addr, []string{"new data"})
	waitForFile(t, filepath.Join(dataDir, "list_subscribers_results.2"))

	got := readFile(t, filepath.Join(dataDir, "list_subscribers_results.2"))
	if len(got) != 1 || got[0] != "new data" {
		t.Errorf("file .2: got %v, want [new data]", got)
	}
}

func TestServer_EmptyConnectionWritesNoFile(t *testing.T) {
	addr, dataDir := startTestServer(t)

	// Connect and disconnect immediately without sending anything
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	conn.Close()

	time.Sleep(200 * time.Millisecond)

	entries, _ := os.ReadDir(dataDir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "list_subscribers_results.") {
			t.Errorf("no file should be written for empty connection, found %s", e.Name())
		}
	}
}

func TestServer_MultipleLinesSingleConnection(t *testing.T) {
	addr, dataDir := startTestServer(t)

	want := make([]string, 100)
	for i := range want {
		want[i] = fmt.Sprintf("data-line-%03d", i)
	}
	sendLines(t, addr, want)

	waitForFile(t, filepath.Join(dataDir, "list_subscribers_results.0"))
	got := readFile(t, filepath.Join(dataDir, "list_subscribers_results.0"))
	if len(got) != len(want) {
		t.Fatalf("lines: got %d, want %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("line[%d]: got %q, want %q", i, got[i], w)
		}
	}
}

// ── nextAvailablePath ─────────────────────────────────────────────────────────

func TestNextAvailablePath_StartsAtZero(t *testing.T) {
	s := &Server{dataDir: t.TempDir()}
	path, err := s.nextAvailablePath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(path, "list_subscribers_results.0") {
		t.Errorf("got %q, want suffix list_subscribers_results.0", path)
	}
}

func TestNextAvailablePath_SkipsExisting(t *testing.T) {
	dir := t.TempDir()
	s := &Server{dataDir: dir}

	for _, idx := range []int{0, 1, 2} {
		p := filepath.Join(dir, fmt.Sprintf("list_subscribers_results.%d", idx))
		_ = os.WriteFile(p, nil, 0644)
	}

	path, err := s.nextAvailablePath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(path, "list_subscribers_results.3") {
		t.Errorf("got %q, want suffix list_subscribers_results.3", path)
	}
}
