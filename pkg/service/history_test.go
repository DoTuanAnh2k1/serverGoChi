package service

import (
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/testutil"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

// csvHeader indices — must match csvHeader in history.go
const (
	colUser          = 0
	colCommand       = 1
	colResult        = 2
	colExecuteTime   = 3
	colNe            = 4
	colRemoteAddress = 5
	colID            = 6
)

var testDate = time.Date(2024, 6, 15, 10, 0, 0, 0, time.Local)

func init() {
	testutil.InitTestLogger()
}

// ── SaveHistoryCommand ────────────────────────────────────────────────────────

func TestSaveHistoryCommand_Success(t *testing.T) {
	want := db_models.CliOperationHistory{
		CmdName:      "show running-config",
		NeName:       "NE-HCM-01",
		NeIP:         "10.0.0.1",
		Scope:        "config-view",
		Result:       "success",
		Account:      "admin",
		CreatedDate:  time.Now(),
		ExecutedTime: time.Now(),
	}

	var got db_models.CliOperationHistory
	store.SetSingleton(&testutil.MockStore{
		SaveHistoryCommandFn: func(h db_models.CliOperationHistory) error {
			got = h
			return nil
		},
	})

	if err := SaveHistoryCommand(want); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.CmdName != want.CmdName {
		t.Errorf("CmdName: got %q, want %q", got.CmdName, want.CmdName)
	}
	if got.NeName != want.NeName {
		t.Errorf("NeName: got %q, want %q", got.NeName, want.NeName)
	}
	if got.Account != want.Account {
		t.Errorf("Account: got %q, want %q", got.Account, want.Account)
	}
}

func TestSaveHistoryCommand_DBError(t *testing.T) {
	dbErr := errors.New("connection refused")
	store.SetSingleton(&testutil.MockStore{
		SaveHistoryCommandFn: func(_ db_models.CliOperationHistory) error {
			return dbErr
		},
	})

	err := SaveHistoryCommand(db_models.CliOperationHistory{CmdName: "test"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}

func TestSaveHistoryCommand_MultipleRecords(t *testing.T) {
	var savedCount int
	store.SetSingleton(&testutil.MockStore{
		SaveHistoryCommandFn: func(_ db_models.CliOperationHistory) error {
			savedCount++
			return nil
		},
	})

	records := []db_models.CliOperationHistory{
		{CmdName: "cmd-1", NeName: "NE-01"},
		{CmdName: "cmd-2", NeName: "NE-02"},
		{CmdName: "cmd-3", NeName: "NE-03"},
	}
	for _, r := range records {
		if err := SaveHistoryCommand(r); err != nil {
			t.Fatalf("unexpected error on %q: %v", r.CmdName, err)
		}
	}

	if savedCount != len(records) {
		t.Errorf("saved %d records, want %d", savedCount, len(records))
	}
}

// ── ExportDailyHistoryByNe ────────────────────────────────────────────────────

func sampleRecords() []db_models.CliOperationHistory {
	return []db_models.CliOperationHistory{
		{
			ID: 1, CmdName: "show version", NeName: "NE-HCM-01", NeIP: "10.0.0.1",
			Result: "success", Account: "admin", IPAddress: "192.168.1.1",
			ExecutedTime: testDate,
		},
		{
			ID: 2, CmdName: "show ip route", NeName: "NE-HCM-01", NeIP: "10.0.0.1",
			Result: "success", Account: "admin", IPAddress: "192.168.1.1",
			ExecutedTime: testDate,
		},
		{
			ID: 3, CmdName: "show running-config", NeName: "NE-HAN-02", NeIP: "10.0.1.1",
			Result: "success", Account: "ops", IPAddress: "192.168.1.2",
			ExecutedTime: testDate,
		},
	}
}

func withExportDir(t *testing.T, dir string, fn func()) {
	t.Helper()
	t.Setenv("CLI_LOG_EXPORT_DIR", dir)
	fn()
}

func TestExportDailyHistoryByNe_CreatesCorrectFiles(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetDailyOperationHistoryFn: func(_ time.Time) ([]db_models.CliOperationHistory, error) {
			return sampleRecords(), nil
		},
	})

	dir := t.TempDir()
	withExportDir(t, dir, func() {
		if err := ExportDailyHistoryByNe(testDate); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	dateStr := "20240615"
	for _, ne := range []string{"NE-HCM-01", "NE-HAN-02"} {
		path := filepath.Join(dir, "cli_log_"+ne+"_"+dateStr+".csv")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", path)
		}
	}
}

func TestExportDailyHistoryByNe_CSVContentIsCorrect(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetDailyOperationHistoryFn: func(_ time.Time) ([]db_models.CliOperationHistory, error) {
			return sampleRecords(), nil
		},
	})

	dir := t.TempDir()
	withExportDir(t, dir, func() {
		if err := ExportDailyHistoryByNe(testDate); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	path := filepath.Join(dir, "cli_log_NE-HCM-01_20240615.csv")
	rows := mustReadCSV(t, path)

	if len(rows) != 3 {
		t.Fatalf("expected 3 rows (header + 2 data), got %d", len(rows))
	}

	// Verify header order
	wantHeader := []string{"user", "command", "result", "execute_time", "ne", "remote_address", "id"}
	for i, want := range wantHeader {
		if rows[0][i] != want {
			t.Errorf("header[%d]: got %q, want %q", i, rows[0][i], want)
		}
	}

	// Verify data row
	if rows[1][colUser] != "admin" {
		t.Errorf("user: got %q, want %q", rows[1][colUser], "admin")
	}
	if rows[1][colCommand] != "show version" {
		t.Errorf("command: got %q, want %q", rows[1][colCommand], "show version")
	}
	if rows[1][colResult] != "success" {
		t.Errorf("result: got %q, want %q", rows[1][colResult], "success")
	}
	if rows[1][colNe] != "NE-HCM-01" {
		t.Errorf("ne: got %q, want %q", rows[1][colNe], "NE-HCM-01")
	}
	if rows[1][colRemoteAddress] != "192.168.1.1" {
		t.Errorf("remote_address: got %q, want %q", rows[1][colRemoteAddress], "192.168.1.1")
	}
	if rows[1][colID] != "1" {
		t.Errorf("id: got %q, want %q", rows[1][colID], "1")
	}
}

func TestExportDailyHistoryByNe_RecordCountPerNE(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetDailyOperationHistoryFn: func(_ time.Time) ([]db_models.CliOperationHistory, error) {
			return sampleRecords(), nil
		},
	})

	dir := t.TempDir()
	withExportDir(t, dir, func() {
		if err := ExportDailyHistoryByNe(testDate); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	cases := []struct {
		file     string
		wantRows int
	}{
		{"cli_log_NE-HCM-01_20240615.csv", 3}, // header + 2 records
		{"cli_log_NE-HAN-02_20240615.csv", 2}, // header + 1 record
	}
	for _, tc := range cases {
		path := filepath.Join(dir, tc.file)
		rows := mustReadCSV(t, path)
		if len(rows) != tc.wantRows {
			t.Errorf("%s: got %d rows, want %d", tc.file, len(rows), tc.wantRows)
		}
	}
}

func TestExportDailyHistoryByNe_EmptyRecords(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetDailyOperationHistoryFn: func(_ time.Time) ([]db_models.CliOperationHistory, error) {
			return nil, nil
		},
	})

	dir := t.TempDir()
	withExportDir(t, dir, func() {
		if err := ExportDailyHistoryByNe(testDate); err != nil {
			t.Fatalf("unexpected error on empty records: %v", err)
		}
	})

	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Error("expected no files to be created for empty records")
	}
}

func TestExportDailyHistoryByNe_DBError(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetDailyOperationHistoryFn: func(_ time.Time) ([]db_models.CliOperationHistory, error) {
			return nil, errors.New("db timeout")
		},
	})

	dir := t.TempDir()
	withExportDir(t, dir, func() {
		if err := ExportDailyHistoryByNe(testDate); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestExportDailyHistoryByNe_SpecialCharNEName(t *testing.T) {
	records := []db_models.CliOperationHistory{
		{ID: 1, NeName: "NE/HCM:01*test", CmdName: "cmd", ExecutedTime: testDate},
	}
	store.SetSingleton(&testutil.MockStore{
		GetDailyOperationHistoryFn: func(_ time.Time) ([]db_models.CliOperationHistory, error) {
			return records, nil
		},
	})

	dir := t.TempDir()
	withExportDir(t, dir, func() {
		if err := ExportDailyHistoryByNe(testDate); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	sanitized := filepath.Join(dir, "cli_log_NE_HCM_01_test_20240615.csv")
	if _, err := os.Stat(sanitized); os.IsNotExist(err) {
		t.Errorf("expected sanitized file %s to exist", sanitized)
	}
}

func TestExportDailyHistoryByNe_EmptyNENameFallsToUnknown(t *testing.T) {
	records := []db_models.CliOperationHistory{
		{ID: 1, NeName: "", CmdName: "cmd1", ExecutedTime: testDate},
		{ID: 2, NeName: "   ", CmdName: "cmd2", ExecutedTime: testDate},
	}
	store.SetSingleton(&testutil.MockStore{
		GetDailyOperationHistoryFn: func(_ time.Time) ([]db_models.CliOperationHistory, error) {
			return records, nil
		},
	})

	dir := t.TempDir()
	withExportDir(t, dir, func() {
		if err := ExportDailyHistoryByNe(testDate); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	unknownFile := filepath.Join(dir, "cli_log_unknown_20240615.csv")
	if _, err := os.Stat(unknownFile); os.IsNotExist(err) {
		t.Errorf("expected %s to exist for records with empty NeName", unknownFile)
	}

	rows := mustReadCSV(t, unknownFile)
	if len(rows) != 3 {
		t.Errorf("unknown file: got %d rows, want 3 (header + 2 data)", len(rows))
	}
}

func TestExportDailyHistoryByNe_DefaultDirWhenEnvNotSet(t *testing.T) {
	records := []db_models.CliOperationHistory{
		{ID: 1, NeName: "NE-TEST", CmdName: "cmd", ExecutedTime: testDate},
	}
	store.SetSingleton(&testutil.MockStore{
		GetDailyOperationHistoryFn: func(_ time.Time) ([]db_models.CliOperationHistory, error) {
			return records, nil
		},
	})

	// Unset env var — should default to "."
	t.Setenv("CLI_LOG_EXPORT_DIR", "")

	// Use a temp dir as working dir to avoid polluting the repo
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	_ = os.Chdir(tmp)
	defer os.Chdir(orig)

	if err := ExportDailyHistoryByNe(testDate); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "cli_log_NE-TEST_20240615.csv")); os.IsNotExist(err) {
		t.Error("expected file in current dir when CLI_LOG_EXPORT_DIR is unset")
	}
}

// ── sanitizeFilename ──────────────────────────────────────────────────────────

func TestSanitizeFilename(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"NE-HCM-01", "NE-HCM-01"},
		{"NE/HCM", "NE_HCM"},
		{"NE:HCM*01", "NE_HCM_01"},
		{`NE\test`, "NE_test"},
		{`NE"test"`, "NE_test_"},
		{"NE<HCM>01", "NE_HCM_01"},
		{"NE|HCM", "NE_HCM"},
		{"NE?HCM", "NE_HCM"},
		{"plain", "plain"},
		{"", ""},
	}

	for _, tc := range cases {
		got := sanitizeFilename(tc.input)
		if got != tc.want {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// ── writeCSV ─────────────────────────────────────────────────────────────────

func TestWriteCSV_HeaderAndRows(t *testing.T) {
	records := []db_models.CliOperationHistory{
		{
			ID: 42, CmdName: "test-cmd",
			NeName: "NE-01", Result: "ok",
			Account: "user1", IPAddress: "10.1.2.3",
			ExecutedTime: testDate,
		},
	}

	path := filepath.Join(t.TempDir(), "test.csv")
	if err := writeCSV(path, records); err != nil {
		t.Fatalf("writeCSV error: %v", err)
	}

	rows := mustReadCSV(t, path)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows (header + 1), got %d", len(rows))
	}

	if len(rows[0]) != len(csvHeader) {
		t.Errorf("header columns: got %d, want %d", len(rows[0]), len(csvHeader))
	}

	data := rows[1]
	checks := map[int]string{
		colUser:          "user1",
		colCommand:       "test-cmd",
		colResult:        "ok",
		colNe:            "NE-01",
		colRemoteAddress: "10.1.2.3",
		colID:            "42",
	}
	for col, want := range checks {
		if data[col] != want {
			t.Errorf("data[%d]: got %q, want %q", col, data[col], want)
		}
	}
}

func TestWriteCSV_EmptyRecords(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.csv")
	if err := writeCSV(path, nil); err != nil {
		t.Fatalf("writeCSV with nil records: %v", err)
	}

	rows := mustReadCSV(t, path)
	if len(rows) != 1 {
		t.Errorf("expected 1 row (header only), got %d", len(rows))
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func mustReadCSV(t *testing.T, path string) [][]string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("parse csv %s: %v", path, err)
	}
	return rows
}
