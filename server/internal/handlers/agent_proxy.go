package handlers

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/arphost-com/Stack-Manager/server/internal/middleware"
	"github.com/arphost-com/Stack-Manager/server/internal/storage"
	"github.com/go-chi/chi/v5"
)

// AgentProxyHandler forwards project actions from the controller
// dashboard to a registered inbound agent. The agent's /agent/v1/
// API handles the actual compose operations; the controller just
// proxies the request and returns the result.
type AgentProxyHandler struct {
	Store  *storage.Store
	client *http.Client
}

func NewAgentProxyHandler(store *storage.Store) *AgentProxyHandler {
	return &AgentProxyHandler{
		Store: store,
		client: &http.Client{
			Timeout: 5 * time.Minute,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{ // nosemgrep: problem-based-packs.insecure-transport.go-stdlib.bypass-tls-verification.bypass-tls-verification
					InsecureSkipVerify: true,
					MinVersion:         tls.VersionTLS12,
				},
			},
		},
	}
}

// Proxy forwards a request to an agent's API. The route pattern is:
//   /api/v1/agent-proxy/{agentId}/{path...}
// The agentId identifies the registered agent; the rest of the path
// maps to the agent's /agent/v1/ endpoint.
func (h *AgentProxyHandler) Proxy(w http.ResponseWriter, r *http.Request) {
	if !middleware.RequireAdmin(w, r) {
		return
	}

	agentIDStr := chi.URLParam(r, "agentId")
	agentID, err := strconv.ParseInt(agentIDStr, 10, 64)
	if err != nil || agentID < 1 {
		writeError(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	agent, err := h.Store.GetAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "agent not found")
		return
	}

	if agent.BaseURL == "" {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("agent %s is outbound-only (no base URL) — direct actions require an inbound or combined agent", agent.Name))
		return
	}

	// Build the target URL: agent's base_url + /agent/v1/ + remaining path.
	remaining := chi.URLParam(r, "*")
	targetURL := strings.TrimRight(agent.BaseURL, "/") + "/agent/v1/" + remaining
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	proxyReq, err := http.NewRequestWithContext(ctx, r.Method, targetURL, r.Body)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create proxy request: "+err.Error())
		return
	}

	// Forward auth + content type.
	proxyReq.Header.Set("Authorization", "Bearer "+agent.Token)
	if ct := r.Header.Get("Content-Type"); ct != "" {
		proxyReq.Header.Set("Content-Type", ct)
	}

	resp, err := h.client.Do(proxyReq)
	if err != nil {
		writeError(w, http.StatusBadGateway, fmt.Sprintf("agent %s unreachable: %s", agent.Name, err.Error()))
		return
	}
	defer resp.Body.Close()

	// Pass through the agent's response.
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}
