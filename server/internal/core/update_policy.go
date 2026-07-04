package core

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"
)

const (
	UpdatePolicyAuto      = "auto"
	UpdatePolicyAllow     = "allow"
	UpdatePolicyNoUpdates = "no_updates"
)

func ValidProjectUpdatePolicyMode(mode string) bool {
	switch mode {
	case UpdatePolicyAuto, UpdatePolicyAllow, UpdatePolicyNoUpdates:
		return true
	default:
		return false
	}
}

func DetectProjectUpdatePolicy(project Project) ProjectUpdatePolicy {
	sourceURL := gitRemoteURL(project.Dir)
	sourceType := gitSourceType(sourceURL)
	hasRegistryImage := false
	hasBuildOnly := false
	for _, source := range project.ImageSources {
		if source.SourceType == "registry" && source.Image != "" {
			hasRegistryImage = true
		}
		if source.Build && source.Image == "" {
			hasBuildOnly = true
		}
	}

	policy := ProjectUpdatePolicy{
		Mode:               UpdatePolicyAuto,
		EffectivePolicy:    UpdatePolicyAllow,
		SourceType:         sourceType,
		SourceURL:          sourceURL,
		DetectedPolicy:     UpdatePolicyAllow,
		DetectedSourceType: sourceType,
		DetectedSourceURL:  sourceURL,
	}

	if sourceType != "" && hasBuildOnly && !hasRegistryImage {
		reason := "build-only project from " + sourceType + " has no registry image to pull"
		policy.EffectivePolicy = UpdatePolicyNoUpdates
		policy.NoUpdatesReason = reason
		policy.AutoDetected = true
		policy.DetectedPolicy = UpdatePolicyNoUpdates
		policy.DetectedReason = reason
	}

	return policy
}

func ResolveProjectUpdatePolicy(mode string, detected ProjectUpdatePolicy) ProjectUpdatePolicy {
	if mode == "" {
		mode = UpdatePolicyAuto
	}
	resolved := detected
	resolved.Mode = mode
	resolved.DetectedPolicy = detected.DetectedPolicy
	resolved.DetectedSourceType = detected.DetectedSourceType
	resolved.DetectedSourceURL = detected.DetectedSourceURL
	resolved.DetectedReason = detected.DetectedReason
	switch mode {
	case UpdatePolicyAllow:
		resolved.EffectivePolicy = UpdatePolicyAllow
		resolved.NoUpdatesReason = ""
		resolved.AutoDetected = false
	case UpdatePolicyNoUpdates:
		resolved.EffectivePolicy = UpdatePolicyNoUpdates
		if resolved.NoUpdatesReason == "" {
			resolved.NoUpdatesReason = "updates disabled manually"
		}
		resolved.AutoDetected = false
	default:
		resolved.Mode = UpdatePolicyAuto
	}
	return resolved
}

func gitRemoteURL(dir string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "config", "--get", "remote.origin.url")
	cmd.Dir = dir
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return ""
	}
	return strings.TrimSpace(stdout.String())
}

func gitSourceType(url string) string {
	lower := strings.ToLower(url)
	switch {
	case strings.Contains(lower, "github.com"):
		return "github"
	case strings.Contains(lower, "gitlab.com"), strings.Contains(lower, "gitlab"):
		return "gitlab"
	default:
		return ""
	}
}
