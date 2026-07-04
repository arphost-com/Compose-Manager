package core

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// hookPath returns the filesystem path for a hook.
func (e *Engine) hookPath(phase, command, project string) string {
	return filepath.Join(e.HooksDir, fmt.Sprintf("%s-%s_%s.sh", phase, command, project))
}

// HasHook checks if a hook exists and is executable.
func (e *Engine) HasHook(phase, command, project string) bool {
	path := e.hookPath(phase, command, project)
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir() && info.Mode()&0111 != 0
}

// RunHook executes a hook script if it exists and is executable.
func (e *Engine) RunHook(phase, command string, project *Project) *OpResult {
	path := e.hookPath(phase, command, project.Name)

	result := &OpResult{
		Project: project.Name,
		Action:  fmt.Sprintf("hook:%s-%s", phase, command),
	}

	info, err := os.Stat(path)
	if err != nil || info.IsDir() || info.Mode()&0111 == 0 {
		result.Success = false
		result.Output = "hook not found or not executable"
		result.ExitCode = 1
		return result
	}

	start := time.Now()
	cmd := exec.Command(path, project.Name, project.Dir)
	cmd.Dir = project.Dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	duration := time.Since(start)
	result.Duration = duration.Round(time.Millisecond).String()
	result.Output = stdout.String() + stderr.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
		result.Success = false
	} else {
		result.ExitCode = 0
		result.Success = true
	}

	return result
}
