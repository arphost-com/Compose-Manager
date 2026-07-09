package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"

	"github.com/arphost-com/Stack-Manager/server/internal/core"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type ShellHandler struct {
	engine *core.Engine
}

func NewShellHandler(engine *core.Engine) *ShellHandler {
	return &ShellHandler{engine: engine}
}

func (h *ShellHandler) ListContainers(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	project, err := h.engine.GetProject(name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	type container struct {
		Name  string `json:"name"`
		Image string `json:"image"`
		State string `json:"state"`
	}
	var containers []container
	for _, c := range project.Containers {
		containers = append(containers, container{
			Name:  c.Name,
			Image: c.Image,
			State: c.State,
		})
	}
	writeJSON(w, http.StatusOK, containers)
}

// ExecWebSocket upgrades to a WebSocket and spawns an interactive
// docker exec -it session with a real PTY. xterm.js on the frontend
// gets full terminal behavior: prompt, colors, tab completion, arrow
// keys, Ctrl-C, and resize.
func (h *ShellHandler) ExecWebSocket(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	containerName := r.URL.Query().Get("container")
	if containerName == "" {
		writeError(w, http.StatusBadRequest, "container query parameter is required")
		return
	}

	project, err := h.engine.GetProject(name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	found := false
	for _, c := range project.Containers {
		if c.Name == containerName {
			found = true
			break
		}
	}
	if !found {
		writeError(w, http.StatusBadRequest, "container does not belong to this project")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("shell: websocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Minute)
	defer cancel()

	shell := detectShell(containerName)
	cmd := exec.CommandContext(ctx, "docker", "exec", "-it", containerName, shell)
	cmd.Env = os.Environ()

	// Start the command with a real PTY so the shell gets full terminal
	// support (prompts, line editing, tab completion, signal handling).
	ptmx, err := pty.Start(cmd)
	if err != nil {
		sendWSError(conn, "failed to start PTY: "+err.Error())
		return
	}
	defer ptmx.Close()

	// Set an initial reasonable size.
	_ = pty.Setsize(ptmx, &pty.Winsize{Rows: 24, Cols: 80})

	var wg sync.WaitGroup

	// PTY → WebSocket (stdout + stderr come through the PTY master)
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 4096)
		for {
			n, err := ptmx.Read(buf)
			if n > 0 {
				if writeErr := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); writeErr != nil {
					return
				}
			}
			if err != nil {
				return
			}
		}
	}()

	// WebSocket → PTY (keystrokes + resize events)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				cancel()
				return
			}
			if msgType == websocket.TextMessage {
				var ctrl struct {
					Type string `json:"type"`
					Cols int    `json:"cols"`
					Rows int    `json:"rows"`
				}
				if json.Unmarshal(msg, &ctrl) == nil && ctrl.Type == "resize" {
					if ctrl.Cols > 0 && ctrl.Rows > 0 {
						_ = pty.Setsize(ptmx, &pty.Winsize{
							Rows: uint16(ctrl.Rows),
							Cols: uint16(ctrl.Cols),
						})
					}
					continue
				}
				// nosemgrep: go.lang.security.audit.dangerous-command-write.dangerous-command-write
				if _, err := ptmx.Write(msg); err != nil {
					return
				}
				continue
			}
			// nosemgrep: go.lang.security.audit.dangerous-command-write.dangerous-command-write
			if _, err := ptmx.Write(msg); err != nil {
				return
			}
		}
	}()

	_ = cmd.Wait()
	wg.Wait()
	_ = conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "shell exited"))
}

func detectShell(containerName string) string {
	for _, sh := range []string{"/bin/bash", "/bin/sh"} {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := exec.CommandContext(ctx, "docker", "exec", containerName, "test", "-x", sh).CombinedOutput()
		cancel()
		if err == nil {
			return sh
		}
	}
	return "/bin/sh"
}

func sendWSError(conn *websocket.Conn, msg string) {
	_ = conn.WriteMessage(websocket.TextMessage, []byte("error: "+msg+"\r\n"))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

var _ = (*ShellHandler)(nil)

func ShellAuthFromQuery(r *http.Request) string {
	if t := r.URL.Query().Get("token"); t != "" {
		return t
	}
	return strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
}
