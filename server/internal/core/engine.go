package core

import (
	"fmt"
	"os"
	"path/filepath"
)

const inactiveMarker = ".inactive"

// ErrNotFound is returned when a project or resource is not found.
type ErrNotFound struct {
	Msg string
}

func (e *ErrNotFound) Error() string { return e.Msg }

// Engine is the core compose management engine.
type Engine struct {
	RootDir  string
	HooksDir string
}

// NewEngine creates a new Engine.
func NewEngine(rootDir, hooksDir string) *Engine {
	return &Engine{
		RootDir:  rootDir,
		HooksDir: hooksDir,
	}
}

// GetProject returns a single project by name.
func (e *Engine) GetProject(name string) (*Project, error) {
	projects, err := e.DiscoverProjects()
	if err != nil {
		return nil, err
	}
	for _, p := range projects {
		if p.Name == name {
			return &p, nil
		}
	}
	return nil, &ErrNotFound{Msg: "project not found: " + name}
}

// Pull pulls images for a project.
func (e *Engine) Pull(project *Project, timeout int) *OpResult {
	return e.ExecComposeWithTimeout(project, timeout, "pull")
}

// Up brings up containers for a project.
func (e *Engine) Up(project *Project) *OpResult {
	return e.ExecCompose(project, "up", "-d")
}

// Down stops and removes containers for a project.
func (e *Engine) Down(project *Project) *OpResult {
	return e.ExecCompose(project, "down")
}

// Restart restarts containers for a project.
func (e *Engine) Restart(project *Project) *OpResult {
	return e.ExecCompose(project, "restart")
}

// Update performs a full update: if a post-update hook exists, run only that;
// otherwise pull + up.
func (e *Engine) Update(project *Project, timeout int) []OpResult {
	if e.HasHook("post", "update", project.Name) {
		hookResult := e.RunHook("post", "update", project)
		hookResult.Action = "update (hook)"
		return []OpResult{*hookResult}
	}

	var results []OpResult
	pullResult := e.Pull(project, timeout)
	pullResult.Action = "pull"
	results = append(results, *pullResult)

	if pullResult.Success {
		upResult := e.Up(project)
		upResult.Action = "up"
		results = append(results, *upResult)
	}

	return results
}

// Status returns the compose ps output for a project.
func (e *Engine) Status(project *Project) *OpResult {
	return e.ExecCompose(project, "ps")
}

// SetInactive marks or unmarks a project as inactive.
func (e *Engine) SetInactive(name string, inactive bool) error {
	project, err := e.GetProject(name)
	if err != nil {
		return err
	}

	markerPath := filepath.Join(project.Dir, inactiveMarker)

	if inactive {
		f, err := os.Create(markerPath)
		if err != nil {
			return fmt.Errorf("failed to create inactive marker: %w", err)
		}
		f.Close()
	} else {
		if err := os.Remove(markerPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove inactive marker: %w", err)
		}
	}

	return nil
}

// Prune runs docker system prune.
func (e *Engine) Prune() *OpResult {
	result := &OpResult{
		Project: "(system)",
		Action:  "prune",
	}

	var outputs []string

	// Image prune
	imgResult, _ := DockerExec("image", "prune", "-f")
	if imgResult != nil {
		outputs = append(outputs, "=== Image Prune ===\n"+imgResult.Stdout+imgResult.Stderr)
	}

	// Network prune
	netResult, _ := DockerExec("network", "prune", "-f")
	if netResult != nil {
		outputs = append(outputs, "=== Network Prune ===\n"+netResult.Stdout+netResult.Stderr)
	}

	// Volume prune
	volResult, _ := DockerExec("volume", "prune", "-f")
	if volResult != nil {
		outputs = append(outputs, "=== Volume Prune ===\n"+volResult.Stdout+volResult.Stderr)
	}

	result.Output = joinOutputs(outputs)
	result.Success = true
	return result
}

func joinOutputs(parts []string) string {
	out := ""
	for _, p := range parts {
		out += p + "\n"
	}
	return out
}
