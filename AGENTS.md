# AGENTS.md

This file provides guidance to Codex (Codex.ai/code) when working with code in this repository.

## Project Overview

Compose Manager has two modes:

1. **CLI tool** (`compose-manager.sh`) — Bash script for discovering and managing Docker Compose projects under a root directory
2. **API server + Web UI** (`server/` + `web/`) — Go REST API with modular skill system and React dashboard

## Build & Test Commands

```bash
# CLI (Bash)
bash -n compose-manager.sh
./compose-manager.sh --root /docker --dry-run update    # safe simulation

# Server (Go 1.23)
cd server && go test ./... && go build ./cmd/server
cd server && go run ./cmd/server                        # requires API_KEY env

# Cross-compile for Linux
cd server && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ../bin/compose-manager-server ./cmd/server

# Web (React 18 + Vite 6 + Tailwind)
cd web && npm ci && npm run build
cd web && npm run dev                                   # dev server on :5173

# Docker
docker compose up -d --build                            # requires .env with API_KEY

# Makefile
make build          # build server + web
make test           # go test + bash -n
make docker-build   # docker compose build
```

## Architecture

### CLI (`compose-manager.sh`)

Single-file Bash tool (~870 lines). Sequential flow: discover projects → filter → execute command → summary.

### API Server + Skills

```
┌────────────────────────────────────────────────┐
│              API Server (Go/Chi)               │
│  /api/v1/projects/*     Core compose ops       │
│  /api/v1/skills/*       Skill dispatch         │
│  /health                Public health check    │
└─────────────┬──────────────────────────────────┘
              │
┌─────────────▼──────────────────────────────────┐
│           Skill Registry                        │
│  Each skill implements Skill interface          │
│  Skills register their own routes               │
├─────────────────────────────────────────────────┤
│ security │ debug │ backup │ dbadmin │ frontend  │
└─────────────────────────────────────────────────┘
              │
┌─────────────▼──────────────────────────────────┐
│         Core Engine                             │
│  Discover, filter, compose exec, hooks          │
│  Ports bash script logic to native Go           │
└─────────────────────────────────────────────────┘
```

**Server** (`server/`):
- Entry: `cmd/server/main.go` — wires Chi router, skill registry, handlers
- Core engine: `internal/core/` — discover.go, compose.go, hooks.go, engine.go, types.go
- Handlers: `internal/handlers/` — projects.go, skills.go, helpers.go
- Middleware: `internal/middleware/auth.go` — X-API-Key constant-time comparison
- Skills: `internal/skills/` — registry.go + one package per skill

**Web** (`web/`):
- React SPA: Dashboard (project list + stats), ProjectDetail (tabs: overview, logs, stats, security, backups, databases, inspect), Settings (API key entry)
- API client: `src/api/client.js`
- Nginx reverse proxy: `/api/` → `http://server:8080`

### Docker Services

| Service | Port | Purpose |
|---------|------|---------|
| server  | 8080 (internal) | Go API server |
| web     | 3020 → 8080 | React SPA via nginx |

Required env: `API_KEY`. Docker socket mounted for compose operations.

## Skill System

### Interface

Every skill implements `internal/skills/registry.go:Skill`:
- `Name()`, `Description()`, `Version()` — metadata
- `Init(ctx, engine, cfg)` — receives core engine reference
- `RegisterRoutes(r chi.Router)` — mounts routes under `/api/v1/skills/{name}/`
- `HealthCheck(ctx)`, `Shutdown(ctx)` — lifecycle

### Adding a New Skill

1. Create package in `server/internal/skills/<name>/`
2. Implement the `Skill` interface
3. Add `registry.Register(<name>.New())` in `cmd/server/main.go`

No other files need to change — the registry handles route mounting and lifecycle.

### Built-in Skills

| Skill | Routes | Purpose |
|-------|--------|---------|
| **security** | `GET scan/{name}`, `GET audit/{name}`, `GET report` | Trivy scans, compose config audit (privileged, host net, unpinned tags) |
| **debug** | `GET logs/{name}`, `GET stats/{name}`, `GET inspect/{name}`, `GET events`, `GET top/{name}` | Container logs, resource usage, docker inspect/events |
| **backup** | `POST create/{name}`, `GET list`, `POST restore/{name}/{id}`, `DELETE {id}` | tar.gz project directories, list/restore/delete |
| **dbadmin** | `GET discover`, `GET health/{name}`, `POST dump/{name}`, `GET dumps` | Auto-detect DB containers (postgres/mysql/mariadb/redis/mongo), health checks, pg_dumpall/mysqldump |
| **frontend** | `GET /*` | Serves the React SPA |

## API Endpoints

Auth: `X-API-Key` header on all `/api/v1/` routes. Health check (`/health`) is public.

### Core Project Routes

```
GET    /api/v1/projects                         List all (filters: only, exclude, include_inactive, running_only)
GET    /api/v1/projects/{name}                   Get project detail with containers
GET    /api/v1/projects/{name}/status            docker compose ps
POST   /api/v1/projects/{name}/pull              Pull images (?timeout=N)
POST   /api/v1/projects/{name}/up                docker compose up -d
POST   /api/v1/projects/{name}/down              docker compose down
POST   /api/v1/projects/{name}/update            Hook or pull+up (?timeout=N)
POST   /api/v1/projects/{name}/restart           docker compose restart
PUT    /api/v1/projects/{name}/inactive           {"inactive": true/false}
POST   /api/v1/projects/bulk/{action}            Bulk: {"projects":[], "exclude":[], "timeout":N}
POST   /api/v1/prune                             docker system prune
```

### Response Envelope

```json
{"status": "ok", "data": {...}, "timestamp": "2026-04-18T12:00:00Z"}
{"status": "error", "error": "message", "timestamp": "..."}
```

## CLI: Critical Behavior

### Update Hook Override

The `update` command: if `post-update_<project>.sh` exists in hooks dir, **only the hook runs** — normal pull+up is skipped. Both CLI and API server implement this. **Do not change without explicit intent.**

### Adding New CLI Commands

1. Create `cmd_<name>()` function
2. Add to case statement in main dispatch
3. If mutating, add to `is_mutating_command()`
4. Update `usage()` function

### Known Quirks

- `-p` flag means `--prune` (not project)
- Flags must come before the command; flags after command are treated as project names

## Configuration

### Server (env vars)

| Variable | Default | Required |
|----------|---------|----------|
| `API_KEY` | — | Yes |
| `ROOT` | `/docker` | No |
| `PORT` | `8080` | No |
| `HOOKS_DIR` | `<ROOT>/.compose-manager/hooks` | No |
| `BACKUP_DIR` | `<ROOT>/.compose-manager/backups` | No |

### CLI (config files)

Loaded in order: `/etc/compose-manager.conf` → `~/.config/compose-manager.conf`
Variables: `ROOT`, `INACTIVE_MARKER`, `LOG_ENABLED`, `LOG_DIR`, `TIMEOUT_SECS`, `HOOKS_ENABLED`, `HOOKS_DIR`

## Commit Messages

Use `Co-Authored-By: BarryBot` in commit messages. Never use "Codex".
