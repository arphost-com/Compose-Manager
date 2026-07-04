package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/arphost-com/Compose-Manager/server/internal/core"
	"github.com/go-chi/chi/v5"
)

// ProjectHandler handles all project-related API endpoints.
type ProjectHandler struct {
	Engine *core.Engine
}

// NewProjectHandler creates a new ProjectHandler.
func NewProjectHandler(engine *core.Engine) *ProjectHandler {
	return &ProjectHandler{Engine: engine}
}

// Create creates a new compose project under the configured root.
func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req core.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	project, err := h.Engine.CreateProject(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, project)
}

// List returns all discovered projects.
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	projects, err := h.Engine.DiscoverProjects()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Apply query filters
	only := r.URL.Query()["only"]
	exclude := r.URL.Query()["exclude"]
	includeInactive := r.URL.Query().Get("include_inactive") == "true"
	onlyInactive := r.URL.Query().Get("only_inactive") == "true"
	runningOnly := r.URL.Query().Get("running_only") == "true"

	filtered := core.FilterProjects(projects, only, exclude, includeInactive, onlyInactive, runningOnly)
	writeJSON(w, http.StatusOK, filtered)
}

// Images returns image source metadata and registry accessibility for a project.
func (h *ProjectHandler) Images(w http.ResponseWriter, r *http.Request) {
	project, err := h.getProject(w, r)
	if err != nil {
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"project": project.Name,
		"images":  h.Engine.CheckImageSources(project),
	})
}

// Get returns a single project by name.
func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	project, err := h.Engine.GetProject(name)
	if err != nil {
		if _, ok := err.(*core.ErrNotFound); ok {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, project)
}

// Status returns docker compose ps for a project.
func (h *ProjectHandler) Status(w http.ResponseWriter, r *http.Request) {
	project, err := h.getProject(w, r)
	if err != nil {
		return
	}
	result := h.Engine.Status(project)
	writeJSON(w, http.StatusOK, result)
}

// Pull pulls images for a project.
func (h *ProjectHandler) Pull(w http.ResponseWriter, r *http.Request) {
	project, err := h.getProject(w, r)
	if err != nil {
		return
	}
	timeout := h.getTimeout(r)
	result := h.Engine.Pull(project, timeout)
	writeJSON(w, http.StatusOK, result)
}

// Up brings up containers.
func (h *ProjectHandler) Up(w http.ResponseWriter, r *http.Request) {
	project, err := h.getProject(w, r)
	if err != nil {
		return
	}
	result := h.Engine.Up(project)
	writeJSON(w, http.StatusOK, result)
}

// Down stops and removes containers.
func (h *ProjectHandler) Down(w http.ResponseWriter, r *http.Request) {
	project, err := h.getProject(w, r)
	if err != nil {
		return
	}
	result := h.Engine.Down(project)
	writeJSON(w, http.StatusOK, result)
}

// Update performs a full update (hook or pull+up).
func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	project, err := h.getProject(w, r)
	if err != nil {
		return
	}
	timeout := h.getTimeout(r)
	results := h.Engine.Update(project, timeout)
	writeJSON(w, http.StatusOK, results)
}

// Restart restarts containers.
func (h *ProjectHandler) Restart(w http.ResponseWriter, r *http.Request) {
	project, err := h.getProject(w, r)
	if err != nil {
		return
	}
	result := h.Engine.Restart(project)
	writeJSON(w, http.StatusOK, result)
}

// SetInactive toggles the inactive marker.
func (h *ProjectHandler) SetInactive(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	var body struct {
		Inactive bool `json:"inactive"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.Engine.SetInactive(name, body.Inactive); err != nil {
		if _, ok := err.(*core.ErrNotFound); ok {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"project":  name,
		"inactive": body.Inactive,
	})
}

// BulkAction performs an action on multiple projects.
func (h *ProjectHandler) BulkAction(w http.ResponseWriter, r *http.Request) {
	action := chi.URLParam(r, "action")

	var req core.BulkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	projects, err := h.Engine.DiscoverProjects()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	filtered := core.FilterProjects(projects, req.Projects, req.Exclude, false, false, false)

	var results []core.OpResult
	for _, p := range filtered {
		p := p
		var result *core.OpResult
		switch action {
		case "pull":
			result = h.Engine.Pull(&p, req.Timeout)
		case "up":
			result = h.Engine.Up(&p)
		case "down":
			result = h.Engine.Down(&p)
		case "restart":
			result = h.Engine.Restart(&p)
		case "update":
			subResults := h.Engine.Update(&p, req.Timeout)
			results = append(results, subResults...)
			continue
		default:
			writeError(w, http.StatusBadRequest, "invalid action: "+action)
			return
		}
		results = append(results, *result)
	}

	successes := 0
	failures := 0
	for _, r := range results {
		if r.Success {
			successes++
		} else {
			failures++
		}
	}

	writeJSON(w, http.StatusOK, core.BulkResult{
		Results: results,
		Total:   len(results),
		Success: successes,
		Failed:  failures,
	})
}

// Prune runs docker system prune.
func (h *ProjectHandler) Prune(w http.ResponseWriter, r *http.Request) {
	result := h.Engine.Prune()
	writeJSON(w, http.StatusOK, result)
}

// RegistryLogin logs Docker into a private registry for future pulls.
func (h *ProjectHandler) RegistryLogin(w http.ResponseWriter, r *http.Request) {
	var req core.RegistryLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	result := core.DockerLogin(req)
	if !result.Success {
		writeJSON(w, http.StatusBadRequest, result)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// getProject is a helper that extracts and validates the project name from URL.
func (h *ProjectHandler) getProject(w http.ResponseWriter, r *http.Request) (*core.Project, error) {
	name := chi.URLParam(r, "name")
	project, err := h.Engine.GetProject(name)
	if err != nil {
		if _, ok := err.(*core.ErrNotFound); ok {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return nil, err
	}
	return project, nil
}

// getTimeout reads timeout from query string.
func (h *ProjectHandler) getTimeout(r *http.Request) int {
	if t := r.URL.Query().Get("timeout"); t != "" {
		if v, err := strconv.Atoi(t); err == nil {
			return v
		}
	}
	return 0
}
