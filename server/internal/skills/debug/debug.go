package debug

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/arphost-com/Compose-Manager/server/internal/core"
	"github.com/go-chi/chi/v5"
)

// Skill implements the debug/troubleshooting skill.
type Skill struct {
	engine *core.Engine
}

func New() *Skill { return &Skill{} }

func (s *Skill) Name() string        { return "debug" }
func (s *Skill) Description() string { return "Container logs, inspection, events, and resource usage" }
func (s *Skill) Version() string     { return "1.0.0" }

func (s *Skill) Init(_ context.Context, engine *core.Engine, _ map[string]interface{}) error {
	s.engine = engine
	return nil
}

func (s *Skill) Shutdown(_ context.Context) error    { return nil }
func (s *Skill) HealthCheck(_ context.Context) error { return nil }

func (s *Skill) RegisterRoutes(r chi.Router) {
	r.Get("/logs/{name}", s.Logs)
	r.Get("/inspect/{name}", s.Inspect)
	r.Get("/stats/{name}", s.Stats)
	r.Get("/events", s.Events)
	r.Get("/top/{name}", s.Top)
}

// Logs returns container logs for a project.
func (s *Skill) Logs(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	project, err := s.engine.GetProject(name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	tail := r.URL.Query().Get("tail")
	if tail == "" {
		tail = "100"
	}
	container := r.URL.Query().Get("container")

	var logs []map[string]string

	if container != "" {
		if !projectHasContainer(project, container) {
			writeError(w, http.StatusBadRequest, "container does not belong to project: "+container)
			return
		}
		// Logs for specific container
		result, _ := core.DockerExec("logs", "--tail", tail, "--timestamps", container)
		if result != nil {
			logs = append(logs, map[string]string{
				"container": container,
				"output":    result.Stdout + result.Stderr,
			})
		}
	} else {
		// Logs for all containers in the project
		result := s.engine.ExecCompose(project, "logs", "--tail", tail, "--timestamps")
		logs = append(logs, map[string]string{
			"project": name,
			"output":  result.Output,
		})
	}

	writeJSON(w, http.StatusOK, logs)
}

// Inspect returns docker inspect for all containers in a project.
func (s *Skill) Inspect(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	project, err := s.engine.GetProject(name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	var inspections []json.RawMessage
	for _, c := range project.Containers {
		result, err := core.DockerExec("inspect", c.Name)
		if err != nil {
			continue
		}
		var parsed json.RawMessage
		if json.Unmarshal([]byte(result.Stdout), &parsed) == nil {
			inspections = append(inspections, parsed)
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"project":     name,
		"inspections": inspections,
	})
}

// Stats returns resource usage for containers in a project.
func (s *Skill) Stats(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	project, err := s.engine.GetProject(name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	var containerIDs []string
	for _, c := range project.Containers {
		containerIDs = append(containerIDs, c.Name)
	}

	if len(containerIDs) == 0 {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"project": name,
			"stats":   []interface{}{},
		})
		return
	}

	args := append([]string{"stats", "--no-stream", "--format",
		`{"container":"{{.Name}}","cpu":"{{.CPUPerc}}","memory":"{{.MemUsage}}","mem_percent":"{{.MemPerc}}","net_io":"{{.NetIO}}","block_io":"{{.BlockIO}}","pids":"{{.PIDs}}"}`},
		containerIDs...)

	result, _ := core.DockerExec(args...)
	if result == nil {
		writeError(w, http.StatusInternalServerError, "failed to get stats")
		return
	}

	var stats []json.RawMessage
	for _, line := range strings.Split(strings.TrimSpace(result.Stdout), "\n") {
		if line == "" {
			continue
		}
		var parsed json.RawMessage
		if json.Unmarshal([]byte(line), &parsed) == nil {
			stats = append(stats, parsed)
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"project": name,
		"stats":   stats,
	})
}

// Events returns recent Docker events.
func (s *Skill) Events(w http.ResponseWriter, r *http.Request) {
	since := r.URL.Query().Get("since")
	if since == "" {
		since = "1h"
	}
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "50"
	}

	limitN, _ := strconv.Atoi(limit)
	if limitN <= 0 {
		limitN = 50
	}

	result, _ := core.DockerExec("events", "--since", since, "--until", "0s",
		"--format", `{"time":"{{.Time}}","type":"{{.Type}}","action":"{{.Action}}","actor":"{{.Actor.Attributes.name}}","image":"{{.Actor.Attributes.image}}"}`)

	if result == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"events": []interface{}{}})
		return
	}

	var events []json.RawMessage
	for _, line := range strings.Split(strings.TrimSpace(result.Stdout), "\n") {
		if line == "" {
			continue
		}
		var parsed json.RawMessage
		if json.Unmarshal([]byte(line), &parsed) == nil {
			events = append(events, parsed)
		}
	}

	// Limit results
	if len(events) > limitN {
		events = events[len(events)-limitN:]
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"events": events,
		"count":  len(events),
	})
}

// Top returns running processes in containers.
func (s *Skill) Top(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	project, err := s.engine.GetProject(name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	var processes []map[string]string
	for _, c := range project.Containers {
		result, err := core.DockerExec("top", c.Name)
		if err != nil {
			continue
		}
		processes = append(processes, map[string]string{
			"container": c.Name,
			"output":    result.Stdout,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"project":   name,
		"processes": processes,
	})
}

func projectHasContainer(project *core.Project, name string) bool {
	for _, c := range project.Containers {
		if c.Name == name || c.ID == name {
			return true
		}
	}
	return false
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
		"timestamp": fmt.Sprintf("%s", time.Now().UTC().Format(time.RFC3339)),
	})
}
