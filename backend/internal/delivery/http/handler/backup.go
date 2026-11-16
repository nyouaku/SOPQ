// ===== FILE: backend/internal/delivery/http/handler/backup.go =====
package handler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nyouaku/SOPQ/backend/internal/config"
	"github.com/nyouaku/SOPQ/backend/pkg/utils"
	"gorm.io/gorm"
)

type BackupHandler struct {
	cfg *config.Config
	db  *gorm.DB
}

func NewBackupHandler(cfg *config.Config, db *gorm.DB) *BackupHandler {
	return &BackupHandler{cfg: cfg, db: db}
}

// TriggerBackup godoc
// @Summary Trigger database backup (admin only)
// @Description Runs pg_dump and returns the backup file for download
// @Tags backup
// @Produce application/octet-stream
// @Security BearerAuth
// @Success 200 {file} binary
// @Router /api/v1/backup/download [get]
func (h *BackupHandler) TriggerBackup(c *gin.Context) {
	tmpDir := os.TempDir()
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("stock_opname_backup_%s.sql", timestamp)
	target := filepath.Join(tmpDir, filename)

	dbCfg := h.cfg.Database
	cmd := exec.Command(
		"pg_dump",
		"--host", dbCfg.Host,
		"--port", dbCfg.Port,
		"--username", dbCfg.User,
		"--dbname", dbCfg.DBName,
		"--no-password",
		"--format=plain",
		"--file", target,
	)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", dbCfg.Password))

	output, err := cmd.CombinedOutput()
	if err != nil {
		utils.RespondInternalError(c, fmt.Sprintf("Backup failed: %s\n%s", err.Error(), string(output)))
		return
	}

	fi, err := os.Stat(target)
	if err != nil || fi.Size() == 0 {
		utils.RespondInternalError(c, "Backup file is empty or not created")
		return
	}

	c.Header("Content-Type", "application/sql")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.File(target)

	defer os.Remove(target)
}

// GetBackupStatus godoc
// @Summary Get backup status
// @Description Returns last backup time and database size
// @Tags backup
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse
// @Router /api/v1/backup/status [get]
func (h *BackupHandler) GetBackupStatus(c *gin.Context) {
	tmpDir := os.TempDir()
	files, _ := filepath.Glob(filepath.Join(tmpDir, "stock_opname_backup_*.sql"))

	var lastBackup time.Time
	var lastBackupSize int64
	var lastBackupFile string

	for _, f := range files {
		fi, err := os.Stat(f)
		if err != nil {
			continue
		}
		if fi.ModTime().After(lastBackup) {
			lastBackup = fi.ModTime()
			lastBackupSize = fi.Size()
			lastBackupFile = fi.Name()
		}
	}

	dbSize := "unknown"
	var sizeResult struct {
		Size string `gorm:"column:size"`
	}
	h.db.Raw("SELECT pg_size_pretty(pg_database_size(current_database())) AS size").Scan(&sizeResult)
	if sizeResult.Size != "" {
		dbSize = sizeResult.Size
	}

	utils.RespondOK(c, gin.H{
		"last_backup_time": lastBackup,
		"last_backup_size": lastBackupSize,
		"last_backup_file": lastBackupFile,
		"database_size":    dbSize,
	})
}
