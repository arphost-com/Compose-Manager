package core

// personalAgentStackTemplates: AI personal agents — self-hosted gateways that
// bridge chat apps (Discord, Slack, WhatsApp, Signal, Telegram, iMessage, etc.)
// to LLM-backed autonomous agents. OpenClaw and its "clones."
//
// Most of these ship as npm packages rather than pre-built Docker images, so
// each template uses node:22-slim as the runtime and installs the package on
// first boot. Rebuild as a proper image after you have picked a version to pin.
func personalAgentStackTemplates() []StackTemplate {
	return []StackTemplate{
		{
			ID: "openclaw", Name: "OpenClaw",
			Description: "Personal AI assistant that connects Discord, Slack, WhatsApp, Signal, Telegram, iMessage and 20+ other chat apps to an LLM.",
			Category:    "ai", Subcategory: "personal-agents",
			Source: "npm", Image: "node:22-slim",
			Tags: []string{"ai", "personal-agent", "chat-gateway"},
			ComposeContent: `services:
  openclaw:
    image: node:22-slim
    working_dir: /root
    command: >-
      sh -c "npm install -g openclaw@${OPENCLAW_VERSION:-latest} &&
             openclaw config set gateway.controlUi.dangerouslyDisableDeviceAuth true &&
             openclaw config set gateway.auth.token \"$OPENCLAW_TOKEN\" &&
             openclaw gateway --allow-unconfigured"
    environment:
      OPENCLAW_TOKEN: ${OPENCLAW_TOKEN:?set OPENCLAW_TOKEN in .env}
    ports:
      - "${OPENCLAW_PORT:-18789}:18789"
    volumes:
      - openclaw-home:/root
    restart: unless-stopped
volumes:
  openclaw-home:
`,
			EnvContent: "OPENCLAW_PORT=18789\nOPENCLAW_VERSION=latest\nOPENCLAW_TOKEN=change-me-32-chars-minimum\n",
			Notes:      "First-run installs openclaw over npm. Point https://<host>:${OPENCLAW_PORT} at the Control UI. Front with nginx/Caddy for TLS.",
		},
		{
			ID: "nanoclaw", Name: "NanoClaw",
			Description: "Security-first fork of OpenClaw that isolates each agent in its own Docker container.",
			Category:    "ai", Subcategory: "personal-agents",
			Source: "npm", Image: "node:22-slim",
			Tags: []string{"ai", "personal-agent", "chat-gateway", "sandboxed"},
			ComposeContent: `services:
  nanoclaw:
    image: node:22-slim
    working_dir: /root
    command: >-
      sh -c "npm install -g nanoclaw@${NANOCLAW_VERSION:-latest} &&
             nanoclaw gateway"
    environment:
      NANOCLAW_AUTH_TOKEN: ${NANOCLAW_TOKEN:?set NANOCLAW_TOKEN in .env}
    ports:
      - "${NANOCLAW_PORT:-18790}:18790"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - nanoclaw-home:/root
    restart: unless-stopped
volumes:
  nanoclaw-home:
`,
			EnvContent: "NANOCLAW_PORT=18790\nNANOCLAW_VERSION=latest\nNANOCLAW_TOKEN=change-me-32-chars-minimum\n",
			Notes:      "NanoClaw spawns per-agent sandbox containers through the mounted Docker socket. Requires Node 20+, pnpm 10+, Docker on the host.",
		},
		{
			ID: "zeroclaw", Name: "ZeroClaw",
			Description: "Minimal Rust-based AI agent framework for self-hosted systems (Landlock/Bubblewrap sandboxing).",
			Category:    "ai", Subcategory: "personal-agents",
			Source: "github", Image: "ghcr.io/zeroclaw-labs/zeroclaw:latest",
			Tags: []string{"ai", "personal-agent", "chat-gateway", "rust"},
			ComposeContent: `services:
  zeroclaw:
    image: ghcr.io/zeroclaw-labs/zeroclaw:latest
    environment:
      ZEROCLAW_TOKEN: ${ZEROCLAW_TOKEN:?set ZEROCLAW_TOKEN in .env}
    ports:
      - "${ZEROCLAW_PORT:-18791}:18791"
    volumes:
      - zeroclaw-data:/data
    restart: unless-stopped
volumes:
  zeroclaw-data:
`,
			EnvContent: "ZEROCLAW_PORT=18791\nZEROCLAW_TOKEN=change-me-32-chars-minimum\n",
			Notes:      "Confirm the exact image tag with the ZeroClaw docs before deploying. Uses OS-level sandboxes for command execution.",
		},
		{
			ID: "forgeai", Name: "ForgeAI",
			Description: "Self-hosted AI gateway that connects any LLM to WhatsApp, Telegram, Discord, Slack, Teams and WebChat via 17 security modules.",
			Category:    "ai", Subcategory: "personal-agents",
			Source: "github", Image: "ghcr.io/forgeai-dev/forgeai:latest",
			Tags: []string{"ai", "personal-agent", "chat-gateway"},
			ComposeContent: `services:
  forgeai:
    image: ghcr.io/forgeai-dev/forgeai:latest
    environment:
      FORGEAI_MASTER_KEY: ${FORGEAI_KEY:?set FORGEAI_KEY in .env}
      DATABASE_URL: postgres://forgeai:${FORGEAI_DB_PASSWORD:?set FORGEAI_DB_PASSWORD in .env}@db:5432/forgeai
    ports:
      - "${FORGEAI_PORT:-3050}:3000"
    depends_on:
      - db
    volumes:
      - forgeai-data:/app/data
    restart: unless-stopped
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: forgeai
      POSTGRES_PASSWORD: ${FORGEAI_DB_PASSWORD:?set FORGEAI_DB_PASSWORD in .env}
      POSTGRES_DB: forgeai
    volumes:
      - forgeai-db:/var/lib/postgresql/data
volumes:
  forgeai-data:
  forgeai-db:
`,
			EnvContent: "FORGEAI_PORT=3050\nFORGEAI_KEY=change-me-32-chars-minimum\nFORGEAI_DB_PASSWORD=change-me\n",
			Notes:      "Every secret is encrypted with AES-256-GCM; every request goes through the security modules. First-run creates the admin account through the dashboard.",
		},
		{
			ID: "nanobot", Name: "Nanobot",
			Description: "Lightweight open-source AI agent for tools, chats and workflows. Supports Feishu, Discord, Slack, Teams.",
			Category:    "ai", Subcategory: "personal-agents",
			Source: "github", Image: "ghcr.io/hkuds/nanobot:latest",
			Tags: []string{"ai", "personal-agent", "chat-gateway"},
			ComposeContent: `services:
  nanobot:
    image: ghcr.io/hkuds/nanobot:latest
    environment:
      NANOBOT_TOKEN: ${NANOBOT_TOKEN:?set NANOBOT_TOKEN in .env}
    ports:
      - "${NANOBOT_PORT:-18792}:8080"
    volumes:
      - nanobot-data:/data
    restart: unless-stopped
volumes:
  nanobot-data:
`,
			EnvContent: "NANOBOT_PORT=18792\nNANOBOT_TOKEN=change-me-32-chars-minimum\n",
			Notes:      "Configure LLM provider keys in nanobot's config file after first run.",
		},
		{
			ID: "hermes-agent", Name: "Hermes Agent",
			Description: "Nous Research's self-hosted AI agent framework (persistent autonomous agent across 16+ messaging platforms).",
			Category:    "ai", Subcategory: "personal-agents",
			Source: "github", Image: "ghcr.io/nousresearch/hermes-agent:latest",
			Tags: []string{"ai", "personal-agent", "chat-gateway"},
			ComposeContent: `services:
  hermes:
    image: ghcr.io/nousresearch/hermes-agent:latest
    environment:
      HERMES_LLM_PROVIDER: ${HERMES_PROVIDER:-openai}
      HERMES_LLM_API_KEY: ${HERMES_API_KEY:-}
    ports:
      - "${HERMES_PORT:-18793}:8080"
    volumes:
      - hermes-data:/data
    restart: unless-stopped
volumes:
  hermes-data:
`,
			EnvContent: "HERMES_PORT=18793\nHERMES_PROVIDER=openai\nHERMES_API_KEY=\n",
			Notes:      "Runs persistently on your infrastructure; connects to your chosen LLM. Verify the image tag on the Hermes repo before deploying.",
		},
		{
			ID: "qwenpaw", Name: "QwenPaw",
			Description: "Alibaba/Qwen ecosystem personal AI assistant with multi-agent collaboration across DingTalk, Feishu, WeChat, QQ, Discord, iMessage, Telegram.",
			Category:    "ai", Subcategory: "personal-agents",
			Source: "github", Image: "ghcr.io/agentscope-ai/qwenpaw:latest",
			Tags: []string{"ai", "personal-agent", "chat-gateway"},
			ComposeContent: `services:
  qwenpaw:
    image: ghcr.io/agentscope-ai/qwenpaw:latest
    environment:
      QWEN_API_KEY: ${QWEN_API_KEY:-}
    ports:
      - "${QWENPAW_PORT:-18794}:8080"
    volumes:
      - qwenpaw-data:/data
    restart: unless-stopped
volumes:
  qwenpaw-data:
`,
			EnvContent: "QWENPAW_PORT=18794\nQWEN_API_KEY=\n",
			Notes:      "Works with the built-in Qwen runtime or with Ollama / LM Studio / any OpenAI-compatible provider.",
		},
		{
			ID: "openjarvis", Name: "OpenJarvis",
			Description: "Local-first personal AI framework; on-device agent that only reaches the cloud when necessary.",
			Category:    "ai", Subcategory: "personal-agents",
			Source: "github", Image: "ghcr.io/open-jarvis/openjarvis:latest",
			Tags: []string{"ai", "personal-agent", "local-first"},
			ComposeContent: `services:
  openjarvis:
    image: ghcr.io/open-jarvis/openjarvis:latest
    environment:
      OPENJARVIS_TOKEN: ${OPENJARVIS_TOKEN:?set OPENJARVIS_TOKEN in .env}
    ports:
      - "${OPENJARVIS_PORT:-18795}:8080"
    volumes:
      - openjarvis-data:/data
    restart: unless-stopped
volumes:
  openjarvis-data:
`,
			EnvContent: "OPENJARVIS_PORT=18795\nOPENJARVIS_TOKEN=change-me-32-chars-minimum\n",
			Notes:      "Pair with Ollama or a local llama.cpp instance for a fully offline agent.",
		},
		{
			ID: "moltworker", Name: "Moltworker",
			Description: "Cloudflare's minimal self-hosted personal AI agent reference implementation.",
			Category:    "ai", Subcategory: "personal-agents",
			Source: "npm", Image: "node:22-slim",
			Tags: []string{"ai", "personal-agent", "chat-gateway"},
			ComposeContent: `services:
  moltworker:
    image: node:22-slim
    working_dir: /root
    command: >-
      sh -c "npm install -g moltworker@${MOLTWORKER_VERSION:-latest} &&
             moltworker start"
    environment:
      MOLTWORKER_TOKEN: ${MOLTWORKER_TOKEN:?set MOLTWORKER_TOKEN in .env}
    ports:
      - "${MOLTWORKER_PORT:-18796}:18796"
    volumes:
      - moltworker-home:/root
    restart: unless-stopped
volumes:
  moltworker-home:
`,
			EnvContent: "MOLTWORKER_PORT=18796\nMOLTWORKER_VERSION=latest\nMOLTWORKER_TOKEN=change-me-32-chars-minimum\n",
			Notes:      "Confirm the exact npm package name from Cloudflare's Moltworker docs before deploying.",
		},
	}
}
