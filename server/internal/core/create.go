package core

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var validProjectDirName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*$`)

// CreateProject creates a compose project under the engine root.
func (e *Engine) CreateProject(req CreateProjectRequest) (*Project, error) {
	name := strings.TrimSpace(req.Name)
	if !validProjectDirName.MatchString(name) {
		return nil, fmt.Errorf("project name must start with a letter or number and contain only letters, numbers, dots, underscores, or hyphens")
	}
	if strings.TrimSpace(req.ComposeContent) == "" {
		return nil, fmt.Errorf("compose content is required")
	}

	rootAbs, err := filepath.Abs(e.RootDir)
	if err != nil {
		return nil, err
	}
	projectDir := filepath.Join(rootAbs, name)
	projectAbs, err := filepath.Abs(projectDir)
	if err != nil {
		return nil, err
	}
	if projectAbs != filepath.Join(rootAbs, name) {
		return nil, fmt.Errorf("invalid project path")
	}

	if _, err := os.Stat(projectDir); err == nil && !req.Overwrite {
		return nil, fmt.Errorf("project already exists: %s", name)
	} else if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if err := os.MkdirAll(projectDir, 0750); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(projectDir, "compose.yml"), []byte(req.ComposeContent), 0640); err != nil {
		return nil, err
	}
	if req.EnvContent != "" {
		if err := os.WriteFile(filepath.Join(projectDir, ".env"), []byte(req.EnvContent), 0600); err != nil {
			return nil, err
		}
	}
	if req.Inactive {
		if err := os.WriteFile(filepath.Join(projectDir, inactiveMarker), []byte{}, 0640); err != nil {
			return nil, err
		}
	}

	cf := composeFileForDir(projectDir)
	if cf == "" {
		return nil, fmt.Errorf("project was created but no compose file could be found")
	}
	project := e.buildProject(projectDir, cf)
	return &project, nil
}
