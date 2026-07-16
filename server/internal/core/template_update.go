package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TemplateUpdatePreview reports whether a running project can be updated from
// its matching catalog template, and what would change.
type TemplateUpdatePreview struct {
	HasTemplate    bool     `json:"has_template"`
	TemplateID     string   `json:"template_id,omitempty"`
	TemplateName   string   `json:"template_name,omitempty"`
	ComposeChanged bool     `json:"compose_changed"`
	NewEnvKeys     []string `json:"new_env_keys,omitempty"`
	GPUApplied     bool     `json:"gpu_applied"`
}

// matchTemplateForProject finds the catalog template a project came from.
// Projects deployed from the catalog use the template ID as the project name,
// so we match on that.
func matchTemplateForProject(projectName string) (StackTemplate, bool) {
	return GetBuiltinStackTemplate(projectName)
}

func normalizeComposeText(s string) string {
	// Compare ignoring trailing whitespace and blank-line noise so that a mere
	// reformat doesn't read as a change.
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		out = append(out, strings.TrimRight(l, " \t\r"))
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

// envKey returns the KEY of a "KEY=value" line, or "" for comments/blank lines.
func envKey(line string) string {
	t := strings.TrimSpace(line)
	if t == "" || strings.HasPrefix(t, "#") {
		return ""
	}
	if i := strings.IndexByte(t, '='); i > 0 {
		return strings.TrimSpace(t[:i])
	}
	return ""
}

func parseEnvMap(s string) map[string]string {
	m := map[string]string{}
	for _, line := range strings.Split(s, "\n") {
		k := envKey(line)
		if k == "" {
			continue
		}
		t := strings.TrimSpace(line)
		m[k] = t[strings.IndexByte(t, '=')+1:]
	}
	return m
}

// newEnvKeys lists keys present in the template's env but not in the project's
// current .env — the settings a template update would introduce.
func newEnvKeys(existing, template string) []string {
	have := parseEnvMap(existing)
	var keys []string
	seen := map[string]bool{}
	for _, line := range strings.Split(template, "\n") {
		k := envKey(line)
		if k == "" || seen[k] {
			continue
		}
		seen[k] = true
		if _, ok := have[k]; !ok {
			keys = append(keys, k)
		}
	}
	return keys
}

// mergeEnvKeepingValues rebuilds the .env from the template (so new keys and
// comments come in) but preserves every value the user already set. Existing
// keys the template dropped are appended at the end so nothing is lost.
func mergeEnvKeepingValues(existing, template string) string {
	have := parseEnvMap(existing)
	used := map[string]bool{}
	var out []string
	for _, line := range strings.Split(strings.TrimRight(template, "\n"), "\n") {
		k := envKey(line)
		if k != "" {
			if v, ok := have[k]; ok {
				out = append(out, k+"="+v)
				used[k] = true
				continue
			}
		}
		out = append(out, line)
	}
	// Preserve existing keys the template no longer defines.
	var extra []string
	for _, line := range strings.Split(existing, "\n") {
		k := envKey(line)
		if k != "" && !used[k] {
			extra = append(extra, line)
		}
	}
	if len(extra) > 0 {
		out = append(out, "", "# --- preserved from your previous .env ---")
		out = append(out, extra...)
	}
	return strings.Join(out, "\n") + "\n"
}

// PreviewTemplateUpdate reports whether the project's compose differs from its
// matching catalog template (an update is available) and which env keys are new.
func (e *Engine) PreviewTemplateUpdate(project *Project) *TemplateUpdatePreview {
	tmpl, ok := matchTemplateForProject(project.Name)
	if !ok {
		return &TemplateUpdatePreview{HasTemplate: false}
	}
	curCompose, _ := os.ReadFile(project.ComposeFile)
	curEnv, _ := os.ReadFile(filepath.Join(project.Dir, ".env"))
	return &TemplateUpdatePreview{
		HasTemplate:    true,
		TemplateID:     tmpl.ID,
		TemplateName:   tmpl.Name,
		ComposeChanged: normalizeComposeText(string(curCompose)) != normalizeComposeText(tmpl.ComposeContent),
		NewEnvKeys:     newEnvKeys(string(curEnv), tmpl.EnvContent),
		GPUApplied:     composeHasGPU(string(curCompose)),
	}
}

// composeHasGPU is a light check so the UI can warn that re-applying the
// template overwrites a project's GPU passthrough (it lives in the compose).
func composeHasGPU(compose string) bool {
	c := strings.ToLower(compose)
	return strings.Contains(c, "driver: nvidia") || strings.Contains(c, "[gpu]") || strings.Contains(c, "- gpu")
}

// ApplyTemplateUpdate rewrites the project's compose.yml from the catalog
// template and migrates its .env — every existing key keeps its value, new
// template keys are added with defaults. The old compose.yml and .env are
// backed up (.bak-<timestamp>) first. The caller recreates the stack after.
func (e *Engine) ApplyTemplateUpdate(project *Project) (*TemplateUpdatePreview, error) {
	tmpl, ok := matchTemplateForProject(project.Name)
	if !ok {
		return nil, fmt.Errorf("no catalog template matches project %q", project.Name)
	}
	ts := time.Now().UTC().Format("20060102-150405")

	if project.ComposeFile == "" {
		return nil, fmt.Errorf("project has no compose file to update")
	}
	curCompose, _ := os.ReadFile(project.ComposeFile)
	if len(curCompose) > 0 {
		if err := os.WriteFile(project.ComposeFile+".bak-"+ts, curCompose, 0640); err != nil {
			return nil, fmt.Errorf("back up compose: %w", err)
		}
	}
	if err := os.WriteFile(project.ComposeFile, []byte(tmpl.ComposeContent), 0640); err != nil {
		return nil, fmt.Errorf("write compose: %w", err)
	}

	envPath := filepath.Join(project.Dir, ".env")
	curEnv, _ := os.ReadFile(envPath)
	if len(curEnv) > 0 {
		if err := os.WriteFile(envPath+".bak-"+ts, curEnv, 0600); err != nil {
			return nil, fmt.Errorf("back up .env: %w", err)
		}
	}
	merged := mergeEnvKeepingValues(string(curEnv), tmpl.EnvContent)
	if err := os.WriteFile(envPath, []byte(merged), 0600); err != nil {
		return nil, fmt.Errorf("write .env: %w", err)
	}

	return &TemplateUpdatePreview{
		HasTemplate:    true,
		TemplateID:     tmpl.ID,
		TemplateName:   tmpl.Name,
		ComposeChanged: true,
		NewEnvKeys:     newEnvKeys(string(curEnv), tmpl.EnvContent),
		GPUApplied:     composeHasGPU(string(curCompose)),
	}, nil
}
