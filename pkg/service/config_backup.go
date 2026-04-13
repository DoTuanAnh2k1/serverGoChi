package service

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
)

// configBackupDir returns the root directory for storing backup XML files.
// Controlled by the CLI_CONFIG_BACKUP_DIR env var (default /data/config-backups).
func configBackupDir() string {
	if d := os.Getenv("CLI_CONFIG_BACKUP_DIR"); d != "" {
		return d
	}
	return "/data/config-backups"
}

// SaveConfigBackup writes the XML to disk then records metadata in the database.
// On DB failure the file is removed so disk and DB stay in sync.
// Returns the saved record (ID is populated after DB insert).
func SaveConfigBackup(neName, neIP, configXML string) (*db_models.CliConfigBackup, error) {
	now := time.Now().UTC()

	dir := filepath.Join(configBackupDir(), neName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("config-backup: create dir %s: %w", dir, err)
	}

	// Filename = Unix nanoseconds — unique and sortable.
	filePath := filepath.Join(dir, fmt.Sprintf("%d.xml", now.UnixNano()))
	if err := os.WriteFile(filePath, []byte(configXML), 0o644); err != nil {
		return nil, fmt.Errorf("config-backup: write file: %w", err)
	}

	b := &db_models.CliConfigBackup{
		NeName:    neName,
		NeIP:      neIP,
		FilePath:  filePath,
		Size:      int64(len(configXML)),
		CreatedAt: now,
	}

	if err := store.GetSingleton().SaveConfigBackup(b); err != nil {
		// Keep disk and DB consistent — remove orphaned file.
		_ = os.Remove(filePath)
		return nil, fmt.Errorf("config-backup: save metadata: %w", err)
	}

	return b, nil
}

// ListConfigBackups returns backup metadata for a given NE (all NEs if neName is empty).
func ListConfigBackups(neName string) ([]*db_models.CliConfigBackup, error) {
	return store.GetSingleton().ListConfigBackups(neName)
}

// GetConfigBackupById retrieves backup metadata and reads the XML from disk.
// Returns (nil, "", nil) when the record does not exist.
func GetConfigBackupById(id int64) (*db_models.CliConfigBackup, string, error) {
	b, err := store.GetSingleton().GetConfigBackupById(id)
	if err != nil {
		return nil, "", err
	}
	if b == nil {
		return nil, "", nil
	}

	data, err := os.ReadFile(b.FilePath)
	if err != nil {
		return nil, "", fmt.Errorf("config-backup: read file %s: %w", b.FilePath, err)
	}

	return b, string(data), nil
}
