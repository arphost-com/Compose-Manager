package core

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
	"unicode"
)

// DockerLogin authenticates Docker against a registry using password-stdin.
func DockerLogin(req RegistryLoginRequest) *OpResult {
	registry := strings.TrimSpace(req.Registry)
	username := strings.TrimSpace(req.Username)

	result := &OpResult{
		Project: "(registry)",
		Action:  "login",
	}

	if username == "" || req.Password == "" {
		result.Success = false
		result.ExitCode = 1
		result.Output = "username and password are required"
		return result
	}
	if !validRegistryName(registry) {
		result.Success = false
		result.ExitCode = 1
		result.Output = "registry must be a Docker registry host without a URL scheme or path"
		return result
	}

	args := []string{"login"}
	if registry != "" {
		args = append(args, registry)
	}
	args = append(args, "-u", username, "--password-stdin")

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdin = strings.NewReader(req.Password)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	result.Duration = time.Since(start).Round(time.Millisecond).String()
	result.Output = sanitizeLoginOutput(stdout.String() + stderr.String())
	if registry != "" {
		result.Action = fmt.Sprintf("login %s", registry)
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
		result.Success = false
		return result
	}

	result.ExitCode = 0
	result.Success = true
	return result
}

func validRegistryName(registry string) bool {
	if registry == "" {
		return true
	}
	if strings.Contains(registry, "://") || strings.Contains(registry, "/") {
		return false
	}
	for _, r := range registry {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return false
		}
	}
	return true
}

func sanitizeLoginOutput(output string) string {
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if strings.Contains(strings.ToLower(line), "password") {
			lines[i] = strings.ReplaceAll(line, "\r", "")
		}
	}
	return strings.Join(lines, "\n")
}
