package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/arphost-com/Stack-Manager/server/internal/core"
	"github.com/go-chi/chi/v5"
)

// WatchHandler exposes the Up + Watch session lifecycle. Persistent live-tail
// logs land in STATE_DIR/logs/{project}/{session}.log so a browser refresh
// mid-stream replays cleanly from the on-disk record.
type WatchHandler struct {
	Manager *core.WatchManager
}

func NewWatchHandler(mgr *core.WatchManager) *WatchHandler {
	return &WatchHandler{Manager: mgr}
}

type watchStartResponse struct {
	Session    *core.WatchSession `json:"session"`
	StreamPath string             `json:"stream_path"`
	UpOutput   string             `json:"up_output"`
}

// Start runs `docker compose up -d` and hands back a session id the client
// can subscribe to for live-tailed startup logs.
func (h *WatchHandler) Start(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	session, upOutput, err := h.Manager.Start(name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, watchStartResponse{
		Session:    session,
		StreamPath: fmt.Sprintf("/api/v1/projects/%s/watch/%s/stream", name, session.ID),
		UpOutput:   upOutput,
	})
}

// List returns every persisted session for the project newest first, for the
// "Reopen a past run" dropdown.
func (h *WatchHandler) List(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	sessions, err := h.Manager.Sessions(name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sessions)
}

// Get returns metadata for a single session (size, exit code, still running).
func (h *WatchHandler) Get(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	sessionID := chi.URLParam(r, "sessionID")
	session, err := h.Manager.LoadSession(name, sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, session)
}

// Stream is the Server-Sent Events endpoint. Replay first, then live tail.
// Every log line becomes one SSE `data:` event so the browser can render
// it directly. Closes on client disconnect, session end, or idle timeout.
func (h *WatchHandler) Stream(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	sessionID := chi.URLParam(r, "sessionID")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no") // hint to nginx not to buffer

	writer := &sseWriter{w: w}
	err := h.Manager.Stream(r.Context(), name, sessionID, writer, func() { flusher.Flush() })
	if err != nil {
		writer.event("error", err.Error())
	}
	writer.event("end", "session stream closed")
	flusher.Flush()
}

// Stop cancels a running session's log subprocess. The log file stays on
// disk so past runs are still browseable.
func (h *WatchHandler) Stop(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	sessionID := chi.URLParam(r, "sessionID")
	if err := h.Manager.Stop(name, sessionID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"session_id": sessionID, "status": "stopped"})
}

// sseWriter buffers a partial line so we always emit whole `data:` events
// even when the subprocess writes half a line at a time.
type sseWriter struct {
	w      http.ResponseWriter
	buffer strings.Builder
}

func (s *sseWriter) Write(p []byte) (int, error) {
	n := len(p)
	s.buffer.Write(p)
	str := s.buffer.String()
	for {
		idx := strings.IndexByte(str, '\n')
		if idx < 0 {
			break
		}
		line := str[:idx]
		str = str[idx+1:]
		if _, err := fmt.Fprintf(s.w, "data: %s\n\n", strings.TrimRight(line, "\r")); err != nil {
			return n, err
		}
	}
	s.buffer.Reset()
	s.buffer.WriteString(str)
	return n, nil
}

func (s *sseWriter) event(name, payload string) {
	fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", name, strings.ReplaceAll(payload, "\n", " "))
}
