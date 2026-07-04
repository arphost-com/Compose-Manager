package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/arphost-com/Compose-Manager/server/internal/core"
	"github.com/go-chi/chi/v5"
)

// Skill implements the backup/restore skill.
type Skill struct {
	engine    *core.Engine
	backupDir string
}

func New() *Skill { return &Skill{} }

func (s *Skill) Name() string { return "backup" }
func (s *Skill) Description() string {
	return "Backup and restore Docker Compose projects (configs, volumes, data)"
}
func (s *Skill) Version() string { return "1.0.0" }

func (s *Skill) Init(_ context.Context, engine *core.Engine, cfg map[string]interface{}) error {
	s.engine = engine
	if dir, ok := cfg["backup_dir"].(string); ok && dir != "" {
		s.backupDir = dir
	} else {
		s.backupDir = filepath.Join(engine.RootDir, ".compose-manager", "backups")
	}
	return os.MkdirAll(s.backupDir, 0755)
}

func (s *Skill) Shutdown(_ context.Context) error { return nil }
func (s *Skill) HealthCheck(_ context.Context) error {
	if _, err := os.Stat(s.backupDir); err != nil {
		return fmt.Errorf("backup directory not accessible: %w", err)
	}
	return nil
}

func (s *Skill) RegisterRoutes(r chi.Router) {
	r.Post("/create/{name}", s.Create)
	r.Get("/list", s.List)
	r.Get("/list/{name}", s.ListProject)
	r.Post("/restore/{name}/{backupId}", s.Restore)
	r.Delete("/{backupId}", s.Delete)
}

// Create creates a backup of a project's directory.
func (s *Skill) Create(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	project, err := s.engine.GetProject(name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	timestamp := time.Now().UTC().Format("20060102_150405")
	backupName := fmt.Sprintf("%s__%s.tar.gz", name, timestamp)
	backupPath := filepath.Join(s.backupDir, backupName)

	// Create tar.gz of the project directory
	cmd := exec.Command("tar", "-czf", backupPath, "-C", filepath.Dir(project.Dir), filepath.Base(project.Dir))
	output, err := cmd.CombinedOutput()
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			fmt.Sprintf("backup failed: %s — %s", err.Error(), string(output)))
		return
	}

	info, _ := os.Stat(backupPath)
	backup := core.BackupInfo{
		ID:        backupName,
		Project:   name,
		File:      backupPath,
		SizeBytes: info.Size(),
		CreatedAt: time.Now().UTC(),
	}

	writeJSON(w, http.StatusCreated, backup)
}

// List returns all backups.
func (s *Skill) List(w http.ResponseWriter, r *http.Request) {
	backups := s.listBackups("")
	writeJSON(w, http.StatusOK, backups)
}

// ListProject returns backups for a specific project.
func (s *Skill) ListProject(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	backups := s.listBackups(name)
	writeJSON(w, http.StatusOK, backups)
}

// Restore restores a project from a backup.
func (s *Skill) Restore(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	backupID := chi.URLParam(r, "backupId")

	if !validBackupID(backupID) {
		writeError(w, http.StatusBadRequest, "invalid backup ID")
		return
	}
	if !strings.HasPrefix(backupID, name+"__") {
		writeError(w, http.StatusBadRequest, "backup does not belong to project: "+name)
		return
	}

	backupPath := filepath.Join(s.backupDir, backupID)
	if _, err := os.Stat(backupPath); err != nil {
		writeError(w, http.StatusNotFound, "backup not found: "+backupID)
		return
	}

	project, err := s.engine.GetProject(name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	// Stop containers first
	if project.Running {
		s.engine.Down(project)
	}

	// Extract backup over the project directory
	cmd := exec.Command("tar", "-xzf", backupPath, "-C", filepath.Dir(project.Dir))
	output, err := cmd.CombinedOutput()
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			fmt.Sprintf("restore failed: %s — %s", err.Error(), string(output)))
		return
	}

	// Bring containers back up
	upResult := s.engine.Up(project)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"project":  name,
		"backup":   backupID,
		"restored": true,
		"up":       upResult,
	})
}

// Delete removes a backup file.
func (s *Skill) Delete(w http.ResponseWriter, r *http.Request) {
	backupID := chi.URLParam(r, "backupId")

	if !validBackupID(backupID) {
		writeError(w, http.StatusBadRequest, "invalid backup ID")
		return
	}

	backupPath := filepath.Join(s.backupDir, backupID)
	if _, err := os.Stat(backupPath); err != nil {
		writeError(w, http.StatusNotFound, "backup not found: "+backupID)
		return
	}

	if err := os.Remove(backupPath); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"deleted": backupID,
	})
}

func validBackupID(backupID string) bool {
	if backupID == "" || strings.Contains(backupID, "/") || strings.Contains(backupID, "\\") || strings.Contains(backupID, "..") {
		return false
	}
	return strings.HasSuffix(backupID, ".tar.gz")
}

func (s *Skill) listBackups(projectFilter string) []core.BackupInfo {
	var backups []core.BackupInfo

	entries, err := os.ReadDir(s.backupDir)
	if err != nil {
		return backups
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tar.gz") {
			continue
		}

		// Parse project name from filename: <project>__<timestamp>.tar.gz
		parts := strings.SplitN(entry.Name(), "__", 2)
		if len(parts) != 2 {
			continue
		}
		projectName := parts[0]

		if projectFilter != "" && projectName != projectFilter {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		backups = append(backups, core.BackupInfo{
			ID:        entry.Name(),
			Project:   projectName,
			File:      filepath.Join(s.backupDir, entry.Name()),
			SizeBytes: info.Size(),
			CreatedAt: info.ModTime(),
		})
	}

	// Sort newest first
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"data":      data,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "error",
		"error":     msg,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
