package core

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var composeUserEnvKeys = []string{
	"STACK_UID",
	"STACK_GID",
	"PUID",
	"PGID",
	"UID",
	"GID",
	"USER_UID",
	"USER_GID",
}

func stackManagerUserEnv(project *Project) []string {
	if project == nil {
		return nil
	}
	uid := strconv.Itoa(os.Getuid())
	gid := strconv.Itoa(os.Getgid())
	values := map[string]string{
		"STACK_UID": uid,
		"STACK_GID": gid,
		"PUID":      uid,
		"PGID":      gid,
		"UID":       uid,
		"GID":       gid,
		"USER_UID":  uid,
		"USER_GID":  gid,
	}
	projectEnv := readDotEnvKeys(filepath.Join(project.Dir, ".env"))
	out := make([]string, 0, len(composeUserEnvKeys))
	for _, key := range composeUserEnvKeys {
		if _, ok := projectEnv[key]; ok {
			continue
		}
		out = append(out, key+"="+values[key])
	}
	return out
}

func composeFileArgs(project *Project) []string {
	args := []string{"-f", project.ComposeFile}
	override := filepath.Join(project.Dir, "compose.override.yml")
	if info, err := os.Stat(override); err == nil && !info.IsDir() {
		args = append(args, "-f", override)
	}
	return args
}

func ComposeCommandArgs(project *Project, args ...string) []string {
	composeArgs := []string{"compose"}
	composeArgs = append(composeArgs, composeFileArgs(project)...)
	composeArgs = append(composeArgs, args...)
	return composeArgs
}

func ComposeUserEnv(project *Project) []string {
	return stackManagerUserEnv(project)
}

func readDotEnvKeys(path string) map[string]struct{} {
	keys := map[string]struct{}{}
	raw, err := os.ReadFile(path)
	if err != nil {
		return keys
	}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		if key != "" {
			keys[key] = struct{}{}
		}
	}
	return keys
}
