package core

import "time"

// Project represents a discovered Docker Compose project.
type Project struct {
	Name         string          `json:"name"`
	Dir          string          `json:"dir"`
	ComposeFile  string          `json:"compose_file"`
	Inactive     bool            `json:"inactive"`
	Running      bool            `json:"running"`
	Containers   []Container     `json:"containers,omitempty"`
	HasHook      map[string]bool `json:"has_hook,omitempty"`
	ImageSources []ImageSource   `json:"image_sources,omitempty"`
}

// Container represents a running Docker container.
type Container struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
	State string `json:"state"`
	Ports string `json:"ports,omitempty"`
}

// ImageSource describes how a compose service gets its image.
type ImageSource struct {
	Service      string `json:"service"`
	Image        string `json:"image,omitempty"`
	Build        bool   `json:"build"`
	BuildContext string `json:"build_context,omitempty"`
	SourceType   string `json:"source_type"`
	Registry     string `json:"registry,omitempty"`
	Repository   string `json:"repository,omitempty"`
	Tag          string `json:"tag,omitempty"`
	Access       string `json:"access,omitempty"`
	Message      string `json:"message,omitempty"`
}

// CreateProjectRequest creates a compose project folder under the configured root.
type CreateProjectRequest struct {
	Name           string `json:"name"`
	ComposeContent string `json:"compose_content"`
	EnvContent     string `json:"env_content,omitempty"`
	Inactive       bool   `json:"inactive,omitempty"`
	Overwrite      bool   `json:"overwrite,omitempty"`
}

// RegistryLoginRequest logs Docker into a registry using password-stdin.
type RegistryLoginRequest struct {
	Registry string `json:"registry,omitempty"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// OpResult is the result of a compose operation (pull, up, down, etc.).
type OpResult struct {
	Project  string `json:"project"`
	Action   string `json:"action"`
	Success  bool   `json:"success"`
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code"`
	Duration string `json:"duration,omitempty"`
}

// ExecResult is the raw result of running a command.
type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// BulkRequest is used for bulk operations on multiple projects.
type BulkRequest struct {
	Projects []string `json:"projects,omitempty"`
	Exclude  []string `json:"exclude,omitempty"`
	Timeout  int      `json:"timeout,omitempty"`
}

// BulkResult collects results from a bulk operation.
type BulkResult struct {
	Results  []OpResult `json:"results"`
	Total    int        `json:"total"`
	Success  int        `json:"success"`
	Failed   int        `json:"failed"`
	Duration string     `json:"duration,omitempty"`
}

// BackupInfo describes a stored backup.
type BackupInfo struct {
	ID        string    `json:"id"`
	Project   string    `json:"project"`
	File      string    `json:"file"`
	SizeBytes int64     `json:"size_bytes"`
	CreatedAt time.Time `json:"created_at"`
}

// DatabaseInfo describes a database found inside a container.
type DatabaseInfo struct {
	Container string   `json:"container"`
	Engine    string   `json:"engine"`
	Host      string   `json:"host,omitempty"`
	Databases []string `json:"databases,omitempty"`
}

// SecurityFinding represents a single security issue found during a scan.
type SecurityFinding struct {
	Severity    string `json:"severity"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Project     string `json:"project,omitempty"`
	Container   string `json:"container,omitempty"`
}

// SecurityReport is the result of a security scan.
type SecurityReport struct {
	Project   string            `json:"project"`
	Findings  []SecurityFinding `json:"findings"`
	ScannedAt time.Time         `json:"scanned_at"`
	Summary   map[string]int    `json:"summary"`
}
