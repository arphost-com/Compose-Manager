package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/arphost-com/Compose-Manager/server/internal/core"
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	DB       *sql.DB
	Redis    *redis.Client
	CacheTTL time.Duration
}

func (s *Store) ImportLegacyFiles(ctx context.Context, stateDir string) error {
	if stateDir == "" {
		return nil
	}
	if err := s.importLegacyUsers(ctx, filepath.Join(stateDir, "users.json")); err != nil {
		return err
	}
	if err := s.importLegacyJobs(ctx, filepath.Join(stateDir, "jobs")); err != nil {
		return err
	}
	return nil
}

func New(ctx context.Context, dsn, redisAddr, redisPassword string, redisDB int, cacheTTL time.Duration) (*Store, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	var pingErr error
	for i := 0; i < 30; i++ {
		pingErr = db.PingContext(ctx)
		if pingErr == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if pingErr != nil {
		_ = db.Close()
		return nil, pingErr
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = db.Close()
		_ = rdb.Close()
		return nil, err
	}

	s := &Store{DB: db, Redis: rdb, CacheTTL: cacheTTL}
	if err := s.Migrate(ctx); err != nil {
		_ = s.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	if s.Redis != nil {
		_ = s.Redis.Close()
	}
	if s.DB != nil {
		return s.DB.Close()
	}
	return nil
}

func (s *Store) Migrate(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			username VARCHAR(64) PRIMARY KEY,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(32) NOT NULL,
			created_at DATETIME(6) NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS jobs (
			id VARCHAR(64) PRIMARY KEY,
			project VARCHAR(255) NOT NULL,
			action VARCHAR(64) NOT NULL,
			status VARCHAR(32) NOT NULL,
			success BOOLEAN NOT NULL,
			exit_code INT NOT NULL,
			output MEDIUMTEXT NOT NULL,
			started_at DATETIME(6) NOT NULL,
			ended_at DATETIME(6) NULL,
			duration VARCHAR(64) NOT NULL DEFAULT '',
			error TEXT NULL,
			INDEX idx_jobs_started_at (started_at),
			INDEX idx_jobs_project_started_at (project, started_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS project_settings (
			project_name VARCHAR(255) PRIMARY KEY,
			update_policy VARCHAR(32) NOT NULL DEFAULT 'auto',
			source_type VARCHAR(64) NOT NULL DEFAULT '',
			source_url TEXT NULL,
			no_updates_reason TEXT NULL,
			notes TEXT NULL,
			auto_detected BOOLEAN NOT NULL DEFAULT FALSE,
			created_at DATETIME(6) NOT NULL,
			updated_at DATETIME(6) NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}
	for _, stmt := range stmts {
		if _, err := s.DB.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) importLegacyUsers(ctx context.Context, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	var payload struct {
		Users []struct {
			Username     string    `json:"username"`
			PasswordHash string    `json:"password_hash"`
			Role         string    `json:"role"`
			CreatedAt    time.Time `json:"created_at"`
		} `json:"users"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}
	for _, user := range payload.Users {
		if user.Username == "" || user.PasswordHash == "" {
			continue
		}
		if user.CreatedAt.IsZero() {
			user.CreatedAt = time.Now().UTC()
		}
		if user.Role == "" {
			user.Role = "operator"
		}
		if _, err := s.DB.ExecContext(ctx, `INSERT IGNORE INTO users (username, password_hash, role, created_at) VALUES (?, ?, ?, ?)`, user.Username, user.PasswordHash, user.Role, user.CreatedAt); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) importLegacyJobs(ctx context.Context, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return err
		}
		var job core.ActionJob
		if err := json.Unmarshal(data, &job); err != nil {
			return err
		}
		if job.ID == "" {
			continue
		}
		if err := s.SaveJob(ctx, &job); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) GetJSON(ctx context.Context, key string, dest interface{}) bool {
	if s.Redis == nil {
		return false
	}
	raw, err := s.Redis.Get(ctx, key).Bytes()
	if err != nil {
		return false
	}
	return json.Unmarshal(raw, dest) == nil
}

func (s *Store) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) {
	if s.Redis == nil {
		return
	}
	if ttl <= 0 {
		ttl = s.CacheTTL
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return
	}
	_ = s.Redis.Set(ctx, key, raw, ttl).Err()
}

func (s *Store) DeleteCache(ctx context.Context, keys ...string) {
	if s.Redis == nil || len(keys) == 0 {
		return
	}
	_ = s.Redis.Del(ctx, keys...).Err()
}

func (s *Store) SaveJob(ctx context.Context, job *core.ActionJob) error {
	cp := core.JobSnapshot(job)
	var endedAt interface{}
	if !cp.EndedAt.IsZero() {
		endedAt = cp.EndedAt
	}
	_, err := s.DB.ExecContext(ctx, `INSERT INTO jobs
		(id, project, action, status, success, exit_code, output, started_at, ended_at, duration, error)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			project=VALUES(project),
			action=VALUES(action),
			status=VALUES(status),
			success=VALUES(success),
			exit_code=VALUES(exit_code),
			output=VALUES(output),
			started_at=VALUES(started_at),
			ended_at=VALUES(ended_at),
			duration=VALUES(duration),
			error=VALUES(error)`,
		cp.ID, cp.Project, cp.Action, cp.Status, cp.Success, cp.ExitCode, cp.Output, cp.StartedAt, endedAt, cp.Duration, nullableString(cp.Error))
	if err == nil {
		s.SetJSON(ctx, "job:"+cp.ID, cp, time.Hour)
		s.DeleteCache(ctx, "jobs:list")
	}
	return err
}

func (s *Store) LoadJob(ctx context.Context, id string) (*core.ActionJob, error) {
	var cached core.ActionJob
	if s.GetJSON(ctx, "job:"+id, &cached) {
		return &cached, nil
	}

	job, err := scanJob(s.DB.QueryRowContext(ctx, `SELECT id, project, action, status, success, exit_code, output, started_at, ended_at, duration, error FROM jobs WHERE id=?`, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	s.SetJSON(ctx, "job:"+id, job, time.Hour)
	return job, nil
}

func (s *Store) ListJobs(ctx context.Context) ([]core.ActionJob, error) {
	var cached []core.ActionJob
	if s.GetJSON(ctx, "jobs:list", &cached) {
		return cached, nil
	}

	rows, err := s.DB.QueryContext(ctx, `SELECT id, project, action, status, success, exit_code, output, started_at, ended_at, duration, error FROM jobs ORDER BY started_at DESC LIMIT 300`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := make([]core.ActionJob, 0)
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, *job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	s.SetJSON(ctx, "jobs:list", jobs, s.CacheTTL)
	return jobs, nil
}

type jobScanner interface {
	Scan(dest ...interface{}) error
}

func scanJob(scanner jobScanner) (*core.ActionJob, error) {
	var job core.ActionJob
	var endedAt sql.NullTime
	var errText sql.NullString
	if err := scanner.Scan(&job.ID, &job.Project, &job.Action, &job.Status, &job.Success, &job.ExitCode, &job.Output, &job.StartedAt, &endedAt, &job.Duration, &errText); err != nil {
		return nil, err
	}
	if endedAt.Valid {
		job.EndedAt = endedAt.Time
	}
	if errText.Valid {
		job.Error = errText.String
	}
	return &job, nil
}

func nullableString(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}

func (s *Store) ResolveUpdatePolicy(project core.Project) core.ProjectUpdatePolicy {
	ctx := context.Background()
	detected := core.DetectProjectUpdatePolicy(project)
	var cached core.ProjectUpdatePolicy
	if s.GetJSON(ctx, "project_policy:"+project.Name, &cached) {
		cached.DetectedPolicy = detected.DetectedPolicy
		cached.DetectedSourceType = detected.DetectedSourceType
		cached.DetectedSourceURL = detected.DetectedSourceURL
		cached.DetectedReason = detected.DetectedReason
		if cached.Mode == "auto" {
			cached.EffectivePolicy = detected.EffectivePolicy
			cached.SourceType = detected.SourceType
			cached.SourceURL = detected.SourceURL
			cached.NoUpdatesReason = detected.NoUpdatesReason
			cached.AutoDetected = detected.AutoDetected
		}
		return cached
	}

	policy, err := s.loadProjectPolicy(ctx, project.Name, detected)
	if err != nil {
		return detected
	}
	s.SetJSON(ctx, "project_policy:"+project.Name, policy, s.CacheTTL)
	return policy
}

func (s *Store) GetProjectPolicy(ctx context.Context, project core.Project) (core.ProjectUpdatePolicy, error) {
	detected := core.DetectProjectUpdatePolicy(project)
	return s.loadProjectPolicy(ctx, project.Name, detected)
}

func (s *Store) SetProjectPolicy(ctx context.Context, project core.Project, mode, notes string) (core.ProjectUpdatePolicy, error) {
	if mode == "" {
		mode = "auto"
	}
	if !core.ValidProjectUpdatePolicyMode(mode) {
		return core.ProjectUpdatePolicy{}, fmt.Errorf("invalid update policy: %s", mode)
	}
	detected := core.DetectProjectUpdatePolicy(project)
	resolved := core.ResolveProjectUpdatePolicy(mode, detected)
	resolved.Notes = notes
	now := time.Now().UTC()
	_, err := s.DB.ExecContext(ctx, `INSERT INTO project_settings
		(project_name, update_policy, source_type, source_url, no_updates_reason, notes, auto_detected, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			update_policy=VALUES(update_policy),
			source_type=VALUES(source_type),
			source_url=VALUES(source_url),
			no_updates_reason=VALUES(no_updates_reason),
			notes=VALUES(notes),
			auto_detected=VALUES(auto_detected),
			updated_at=VALUES(updated_at)`,
		project.Name, mode, resolved.SourceType, nullableString(resolved.SourceURL), nullableString(resolved.NoUpdatesReason), nullableString(notes), resolved.AutoDetected, now, now)
	if err != nil {
		return core.ProjectUpdatePolicy{}, err
	}
	s.DeleteCache(ctx, "project_policy:"+project.Name, "projects:list")
	s.SetJSON(ctx, "project_policy:"+project.Name, resolved, s.CacheTTL)
	return resolved, nil
}

func (s *Store) loadProjectPolicy(ctx context.Context, projectName string, detected core.ProjectUpdatePolicy) (core.ProjectUpdatePolicy, error) {
	var mode string
	var sourceType string
	var sourceURL sql.NullString
	var reason sql.NullString
	var notes sql.NullString
	var autoDetected bool
	err := s.DB.QueryRowContext(ctx, `SELECT update_policy, source_type, source_url, no_updates_reason, notes, auto_detected FROM project_settings WHERE project_name=?`, projectName).
		Scan(&mode, &sourceType, &sourceURL, &reason, &notes, &autoDetected)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return detected, nil
		}
		return detected, err
	}

	resolved := core.ResolveProjectUpdatePolicy(mode, detected)
	if sourceType != "" {
		resolved.SourceType = sourceType
	}
	if sourceURL.Valid {
		resolved.SourceURL = sourceURL.String
	}
	if reason.Valid {
		resolved.NoUpdatesReason = reason.String
	}
	if notes.Valid {
		resolved.Notes = notes.String
	}
	if mode != "auto" {
		resolved.AutoDetected = autoDetected
	}
	return resolved, nil
}
