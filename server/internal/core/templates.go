package core

import (
	"fmt"
	"sort"
	"strings"
)

type StackTemplate struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Category       string   `json:"category"`
	Source         string   `json:"source"`
	Image          string   `json:"image,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	ComposeContent string   `json:"compose_content"`
	EnvContent     string   `json:"env_content,omitempty"`
	Notes          string   `json:"notes,omitempty"`
}

func BuiltinStackTemplates() []StackTemplate {
	templates := []StackTemplate{
		{
			ID:          "wordpress-mariadb",
			Name:        "WordPress + MariaDB",
			Description: "Blog/CMS stack with WordPress and MariaDB.",
			Category:    "cms",
			Source:      "docker-docs-portainer-style",
			Image:       "wordpress:latest",
			Tags:        []string{"cms", "database", "website"},
			ComposeContent: `services:
  wordpress:
    image: wordpress:latest
    restart: unless-stopped
    ports:
      - "${WORDPRESS_PORT:-8080}:80"
    environment:
      WORDPRESS_DB_HOST: db
      WORDPRESS_DB_USER: ${WORDPRESS_DB_USER:-wordpress}
      WORDPRESS_DB_PASSWORD: ${WORDPRESS_DB_PASSWORD:-change-me}
      WORDPRESS_DB_NAME: ${WORDPRESS_DB_NAME:-wordpress}
    volumes:
      - wordpress-data:/var/www/html
    depends_on:
      - db
  db:
    image: mariadb:11.4
    restart: unless-stopped
    environment:
      MARIADB_DATABASE: ${WORDPRESS_DB_NAME:-wordpress}
      MARIADB_USER: ${WORDPRESS_DB_USER:-wordpress}
      MARIADB_PASSWORD: ${WORDPRESS_DB_PASSWORD:-change-me}
      MARIADB_ROOT_PASSWORD: ${MARIADB_ROOT_PASSWORD:-change-me-root}
    volumes:
      - db-data:/var/lib/mysql
volumes:
  wordpress-data:
  db-data:
`,
			EnvContent: `WORDPRESS_PORT=8080
WORDPRESS_DB_NAME=wordpress
WORDPRESS_DB_USER=wordpress
WORDPRESS_DB_PASSWORD=change-me
MARIADB_ROOT_PASSWORD=change-me-root
`,
			Notes: "Change the database passwords before starting this stack.",
		},
		{
			ID:          "nginx-static",
			Name:        "Nginx Static Site",
			Description: "Small static web server with a bind-mounted html directory.",
			Category:    "web",
			Source:      "kitematic-style",
			Image:       "nginx:stable-alpine",
			Tags:        []string{"web", "static", "nginx"},
			ComposeContent: `services:
  web:
    image: nginx:stable-alpine
    restart: unless-stopped
    ports:
      - "${WEB_PORT:-8080}:80"
    volumes:
      - ./html:/usr/share/nginx/html:ro
`,
			EnvContent: "WEB_PORT=8080\n",
			Notes:      "Create an html directory beside compose.yml before starting, or edit the bind mount.",
		},
		{
			ID:          "postgres",
			Name:        "PostgreSQL",
			Description: "PostgreSQL database with persistent volume.",
			Category:    "database",
			Source:      "kitematic-style",
			Image:       "postgres:16-alpine",
			Tags:        []string{"database", "postgres"},
			ComposeContent: `services:
  postgres:
    image: postgres:16-alpine
    restart: unless-stopped
    ports:
      - "${POSTGRES_PORT:-5432}:5432"
    environment:
      POSTGRES_DB: ${POSTGRES_DB:-app}
      POSTGRES_USER: ${POSTGRES_USER:-app}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-change-me}
    volumes:
      - postgres-data:/var/lib/postgresql/data
volumes:
  postgres-data:
`,
			EnvContent: `POSTGRES_PORT=5432
POSTGRES_DB=app
POSTGRES_USER=app
POSTGRES_PASSWORD=change-me
`,
		},
		{
			ID:          "redis",
			Name:        "Redis",
			Description: "Redis cache with append-only persistence and password auth.",
			Category:    "database",
			Source:      "kitematic-style",
			Image:       "redis:7.4-alpine",
			Tags:        []string{"cache", "redis"},
			ComposeContent: `services:
  redis:
    image: redis:7.4-alpine
    restart: unless-stopped
    command: ["redis-server", "--appendonly", "yes", "--requirepass", "${REDIS_PASSWORD:-change-me}"]
    ports:
      - "${REDIS_PORT:-6379}:6379"
    volumes:
      - redis-data:/data
volumes:
  redis-data:
`,
			EnvContent: `REDIS_PORT=6379
REDIS_PASSWORD=change-me
`,
		},
		{
			ID:          "gitea",
			Name:        "Gitea",
			Description: "Lightweight Git service with SSH and web ports.",
			Category:    "devtools",
			Source:      "portainer-style",
			Image:       "gitea/gitea:latest",
			Tags:        []string{"git", "devtools"},
			ComposeContent: `services:
  gitea:
    image: gitea/gitea:latest
    restart: unless-stopped
    environment:
      USER_UID: ${USER_UID:-1000}
      USER_GID: ${USER_GID:-1000}
    ports:
      - "${GITEA_WEB_PORT:-3000}:3000"
      - "${GITEA_SSH_PORT:-2222}:22"
    volumes:
      - gitea-data:/data
volumes:
  gitea-data:
`,
			EnvContent: `USER_UID=1000
USER_GID=1000
GITEA_WEB_PORT=3000
GITEA_SSH_PORT=2222
`,
		},
		{
			ID:          "uptime-kuma",
			Name:        "Uptime Kuma",
			Description: "Self-hosted uptime monitor.",
			Category:    "monitoring",
			Source:      "portainer-style",
			Image:       "louislam/uptime-kuma:1",
			Tags:        []string{"monitoring", "status"},
			ComposeContent: `services:
  uptime-kuma:
    image: louislam/uptime-kuma:1
    restart: unless-stopped
    ports:
      - "${UPTIME_KUMA_PORT:-3001}:3001"
    volumes:
      - uptime-kuma-data:/app/data
volumes:
  uptime-kuma-data:
`,
			EnvContent: "UPTIME_KUMA_PORT=3001\n",
		},
		{
			ID:          "portainer-agent",
			Name:        "Portainer Agent",
			Description: "Portainer agent for hosts that also need Portainer compatibility.",
			Category:    "management",
			Source:      "portainer-style",
			Image:       "portainer/agent:latest",
			Tags:        []string{"management", "agent"},
			ComposeContent: `services:
  portainer-agent:
    image: portainer/agent:latest
    restart: unless-stopped
    ports:
      - "${PORTAINER_AGENT_PORT:-9001}:9001"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - /var/lib/docker/volumes:/var/lib/docker/volumes
`,
			EnvContent: "PORTAINER_AGENT_PORT=9001\n",
		},
		{
			ID:          "prometheus-grafana",
			Name:        "Prometheus + Grafana",
			Description: "Monitoring starter stack with Prometheus and Grafana.",
			Category:    "monitoring",
			Source:      "rancher-style",
			Image:       "grafana/grafana-oss:latest",
			Tags:        []string{"monitoring", "metrics", "grafana"},
			ComposeContent: `services:
  prometheus:
    image: prom/prometheus:latest
    restart: unless-stopped
    ports:
      - "${PROMETHEUS_PORT:-9090}:9090"
    volumes:
      - prometheus-data:/prometheus
  grafana:
    image: grafana/grafana-oss:latest
    restart: unless-stopped
    ports:
      - "${GRAFANA_PORT:-3000}:3000"
    volumes:
      - grafana-data:/var/lib/grafana
volumes:
  prometheus-data:
  grafana-data:
`,
			EnvContent: `PROMETHEUS_PORT=9090
GRAFANA_PORT=3000
`,
			Notes: "Add a prometheus.yml bind mount before production use.",
		},
	}
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Name < templates[j].Name
	})
	return templates
}

func GetBuiltinStackTemplate(id string) (StackTemplate, bool) {
	id = strings.TrimSpace(id)
	for _, template := range BuiltinStackTemplates() {
		if template.ID == id {
			return template, true
		}
	}
	return StackTemplate{}, false
}

func RenderStackTemplate(id string) (CreateProjectRequest, error) {
	template, ok := GetBuiltinStackTemplate(id)
	if !ok {
		return CreateProjectRequest{}, fmt.Errorf("template not found: %s", id)
	}
	return CreateProjectRequest{
		Name:           template.ID,
		ComposeContent: template.ComposeContent,
		EnvContent:     template.EnvContent,
	}, nil
}
