package core

// vettedAIStackTemplates adds well-known, actively maintained AI projects with
// credible upstream docs and self-hosting paths. These are separated from the
// catalog fill templates so quality-vetted additions stay easy to review.
func vettedAIStackTemplates() []StackTemplate {
	return []StackTemplate{
		{
			ID:          "librechat",
			Name:        "LibreChat",
			Description: "Self-hosted multi-provider ChatGPT-style platform with agents, MCP, files, search, artifacts, and multi-user auth.",
			Category:    "ai",
			Subcategory: "workflow-rag",
			Source:      "official-github",
			Image:       "registry.librechat.ai/danny-avila/librechat-dev:latest",
			Tags:        []string{"ai", "chat", "agents", "mcp", "multi-user"},
			ComposeContent: `services:
  api:
    image: registry.librechat.ai/danny-avila/librechat-dev:latest
    restart: unless-stopped
    user: "${UID:-1000}:${GID:-1000}"
    ports:
      - "${LIBRECHAT_PORT:-3080}:3080"
    environment:
      HOST: 0.0.0.0
      PORT: 3080
      MONGO_URI: mongodb://mongodb:27017/LibreChat
      MEILI_HOST: http://meilisearch:7700
      MEILI_MASTER_KEY: ${MEILI_MASTER_KEY:?set MEILI_MASTER_KEY in .env}
      RAG_PORT: 8000
      RAG_API_URL: http://rag_api:8000
      JWT_SECRET: ${JWT_SECRET:?set JWT_SECRET in .env}
      JWT_REFRESH_SECRET: ${JWT_REFRESH_SECRET:?set JWT_REFRESH_SECRET in .env}
      CREDS_KEY: ${CREDS_KEY:?set CREDS_KEY in .env}
      CREDS_IV: ${CREDS_IV:?set CREDS_IV in .env}
      OPENAI_API_KEY: ${OPENAI_API_KEY:-}
      ANTHROPIC_API_KEY: ${ANTHROPIC_API_KEY:-}
    volumes:
      - librechat-images:/app/client/public/images
      - librechat-uploads:/app/uploads
      - librechat-logs:/app/logs
    depends_on:
      - mongodb
      - meilisearch
      - rag_api
  admin-panel:
    image: registry.librechat.ai/clickhouse/librechat-admin-panel:latest
    restart: unless-stopped
    ports:
      - "${LIBRECHAT_ADMIN_PORT:-3001}:3000"
    environment:
      SESSION_SECRET: ${ADMIN_PANEL_SESSION_SECRET:?set ADMIN_PANEL_SESSION_SECRET in .env}
      API_SERVER_URL: http://api:3080
      VITE_API_BASE_URL: ${LIBRECHAT_PUBLIC_URL:-http://localhost:3080}
    depends_on:
      - api
  mongodb:
    image: mongo:8.0.20
    restart: unless-stopped
    user: "${UID:-1000}:${GID:-1000}"
    command: mongod --noauth
    volumes:
      - librechat-mongodb:/data/db
  meilisearch:
    image: getmeili/meilisearch:v1.35.1
    restart: unless-stopped
    user: "${UID:-1000}:${GID:-1000}"
    environment:
      MEILI_NO_ANALYTICS: "true"
      MEILI_MASTER_KEY: ${MEILI_MASTER_KEY:?set MEILI_MASTER_KEY in .env}
    volumes:
      - librechat-meili:/meili_data
  vectordb:
    image: pgvector/pgvector:0.8.0-pg15-trixie
    restart: unless-stopped
    environment:
      POSTGRES_DB: librechat
      POSTGRES_USER: librechat
      POSTGRES_PASSWORD: ${LIBRECHAT_VECTOR_DB_PASSWORD:?set LIBRECHAT_VECTOR_DB_PASSWORD in .env}
    volumes:
      - librechat-pgvector:/var/lib/postgresql/data
  rag_api:
    image: registry.librechat.ai/danny-avila/librechat-rag-api-dev-lite:latest
    restart: unless-stopped
    environment:
      DB_HOST: vectordb
      POSTGRES_DB: librechat
      POSTGRES_USER: librechat
      POSTGRES_PASSWORD: ${LIBRECHAT_VECTOR_DB_PASSWORD:?set LIBRECHAT_VECTOR_DB_PASSWORD in .env}
      RAG_PORT: 8000
    depends_on:
      - vectordb
volumes:
  librechat-images:
  librechat-uploads:
  librechat-logs:
  librechat-mongodb:
  librechat-meili:
  librechat-pgvector:
`,
			EnvContent: "LIBRECHAT_PORT=3080\nLIBRECHAT_ADMIN_PORT=3001\nLIBRECHAT_PUBLIC_URL=http://localhost:3080\nUID=1000\nGID=1000\nMEILI_MASTER_KEY=\nJWT_SECRET=\nJWT_REFRESH_SECRET=\nCREDS_KEY=\nCREDS_IV=\nADMIN_PANEL_SESSION_SECRET=\nLIBRECHAT_VECTOR_DB_PASSWORD=\nOPENAI_API_KEY=\nANTHROPIC_API_KEY=\n",
			Notes:      "Use LibreChat when customers want one shared AI chat UI across OpenAI, Anthropic, local OpenAI-compatible endpoints, agents, MCP, and files. Set all secret fields before exposing it.",
		},
		{
			ID:          "onyx",
			Name:        "Onyx",
			Description: "Open-source enterprise AI chat and knowledge platform with connectors, search, agents, and every-LLM support.",
			Category:    "ai",
			Subcategory: "workflow-rag",
			Source:      "official-github",
			Image:       "onyxdotapp/onyx-web-server:latest",
			Tags:        []string{"ai", "rag", "enterprise-search", "connectors"},
			ComposeContent: `services:
  api_server:
    image: onyxdotapp/onyx-backend:${ONYX_IMAGE_TAG:-latest}
    command: >-
      /bin/sh -c "alembic upgrade head && uvicorn onyx.main:app --host 0.0.0.0 --port 8080"
    restart: unless-stopped
    environment:
      AUTH_TYPE: ${ONYX_AUTH_TYPE:-basic}
      POSTGRES_HOST: relational_db
      POSTGRES_USER: ${ONYX_POSTGRES_USER:-postgres}
      POSTGRES_PASSWORD: ${ONYX_POSTGRES_PASSWORD:?set ONYX_POSTGRES_PASSWORD in .env}
      OPENSEARCH_HOST: opensearch
      OPENSEARCH_ADMIN_PASSWORD: ${ONYX_OPENSEARCH_PASSWORD:?set ONYX_OPENSEARCH_PASSWORD in .env}
      REDIS_HOST: cache
      MODEL_SERVER_HOST: inference_model_server
      INDEXING_MODEL_SERVER_HOST: indexing_model_server
      FILE_STORE_BACKEND: s3
      S3_ENDPOINT_URL: http://minio:9000
      S3_AWS_ACCESS_KEY_ID: ${ONYX_S3_ACCESS_KEY:?set ONYX_S3_ACCESS_KEY in .env}
      S3_AWS_SECRET_ACCESS_KEY: ${ONYX_S3_SECRET_KEY:?set ONYX_S3_SECRET_KEY in .env}
    depends_on:
      - relational_db
      - opensearch
      - cache
      - inference_model_server
      - indexing_model_server
      - minio
  background:
    image: onyxdotapp/onyx-backend:${ONYX_IMAGE_TAG:-latest}
    command: /app/scripts/supervisord_entrypoint.sh
    restart: unless-stopped
    environment:
      POSTGRES_HOST: relational_db
      POSTGRES_USER: ${ONYX_POSTGRES_USER:-postgres}
      POSTGRES_PASSWORD: ${ONYX_POSTGRES_PASSWORD:?set ONYX_POSTGRES_PASSWORD in .env}
      OPENSEARCH_HOST: opensearch
      OPENSEARCH_ADMIN_PASSWORD: ${ONYX_OPENSEARCH_PASSWORD:?set ONYX_OPENSEARCH_PASSWORD in .env}
      REDIS_HOST: cache
      MODEL_SERVER_HOST: inference_model_server
      INDEXING_MODEL_SERVER_HOST: indexing_model_server
      S3_ENDPOINT_URL: http://minio:9000
      S3_AWS_ACCESS_KEY_ID: ${ONYX_S3_ACCESS_KEY:?set ONYX_S3_ACCESS_KEY in .env}
      S3_AWS_SECRET_ACCESS_KEY: ${ONYX_S3_SECRET_KEY:?set ONYX_S3_SECRET_KEY in .env}
    depends_on:
      - api_server
  web_server:
    image: onyxdotapp/onyx-web-server:${ONYX_IMAGE_TAG:-latest}
    restart: unless-stopped
    ports:
      - "${ONYX_PORT:-3000}:3000"
    environment:
      INTERNAL_URL: http://api_server:8080
    depends_on:
      - api_server
  inference_model_server:
    image: onyxdotapp/onyx-model-server:${ONYX_IMAGE_TAG:-latest}
    restart: unless-stopped
    volumes:
      - onyx-model-cache:/app/.cache/huggingface
  indexing_model_server:
    image: onyxdotapp/onyx-model-server:${ONYX_IMAGE_TAG:-latest}
    restart: unless-stopped
    environment:
      INDEXING_ONLY: "True"
    volumes:
      - onyx-indexing-cache:/app/.cache/huggingface
  relational_db:
    image: postgres:15.2-alpine
    restart: unless-stopped
    shm_size: 1g
    command: -c 'max_connections=250'
    environment:
      POSTGRES_USER: ${ONYX_POSTGRES_USER:-postgres}
      POSTGRES_PASSWORD: ${ONYX_POSTGRES_PASSWORD:?set ONYX_POSTGRES_PASSWORD in .env}
    volumes:
      - onyx-postgres:/var/lib/postgresql/data
  opensearch:
    image: opensearchproject/opensearch:3.6.0
    restart: unless-stopped
    environment:
      discovery.type: single-node
      OPENSEARCH_INITIAL_ADMIN_PASSWORD: ${ONYX_OPENSEARCH_PASSWORD:?set ONYX_OPENSEARCH_PASSWORD in .env}
      OPENSEARCH_JAVA_OPTS: -Xms2g -Xmx2g
      bootstrap.memory_lock: "true"
    volumes:
      - onyx-opensearch:/usr/share/opensearch/data
  cache:
    image: redis:7.4-alpine
    restart: unless-stopped
    command: redis-server --save "" --appendonly no
    tmpfs:
      - /data
  minio:
    image: minio/minio:RELEASE.2025-07-23T15-54-02Z-cpuv1
    restart: unless-stopped
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: ${ONYX_S3_ACCESS_KEY:?set ONYX_S3_ACCESS_KEY in .env}
      MINIO_ROOT_PASSWORD: ${ONYX_S3_SECRET_KEY:?set ONYX_S3_SECRET_KEY in .env}
    volumes:
      - onyx-minio:/data
volumes:
  onyx-postgres:
  onyx-opensearch:
  onyx-minio:
  onyx-model-cache:
  onyx-indexing-cache:
`,
			EnvContent: "ONYX_PORT=3000\nONYX_IMAGE_TAG=latest\nONYX_AUTH_TYPE=basic\nONYX_POSTGRES_USER=postgres\nONYX_POSTGRES_PASSWORD=\nONYX_OPENSEARCH_PASSWORD=\nONYX_S3_ACCESS_KEY=\nONYX_S3_SECRET_KEY=\n",
			Notes:      "Good fit for enterprise search/RAG with connectors. Onyx is heavier than chat-only stacks; give OpenSearch and model services enough RAM before enabling broad indexing.",
		},
		{
			ID:          "khoj",
			Name:        "Khoj",
			Description: "Self-hostable personal AI second brain for docs, web answers, custom agents, automations, and local or hosted LLMs.",
			Category:    "ai",
			Subcategory: "personal-agents",
			Source:      "official-github",
			Image:       "ghcr.io/khoj-ai/khoj:latest",
			Tags:        []string{"ai", "personal-agent", "second-brain", "search"},
			ComposeContent: `services:
  database:
    image: pgvector/pgvector:pg15
    restart: unless-stopped
    environment:
      POSTGRES_USER: ${KHOJ_POSTGRES_USER:-postgres}
      POSTGRES_PASSWORD: ${KHOJ_POSTGRES_PASSWORD:?set KHOJ_POSTGRES_PASSWORD in .env}
      POSTGRES_DB: ${KHOJ_POSTGRES_DB:-postgres}
    volumes:
      - khoj-db:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${KHOJ_POSTGRES_USER:-postgres}"]
      interval: 30s
      timeout: 10s
      retries: 5
  sandbox:
    image: ghcr.io/khoj-ai/terrarium:latest
    restart: unless-stopped
  search:
    image: searxng/searxng:latest
    restart: unless-stopped
    volumes:
      - khoj-search:/etc/searxng
    environment:
      SEARXNG_BASE_URL: http://localhost:8080/
  server:
    image: ghcr.io/khoj-ai/khoj:${KHOJ_IMAGE_TAG:-latest}
    restart: unless-stopped
    command: --host="0.0.0.0" --port=42110 -vv --anonymous-mode --non-interactive
    working_dir: /app
    ports:
      - "${KHOJ_PORT:-42110}:42110"
    extra_hosts:
      - "host.docker.internal:host-gateway"
    environment:
      POSTGRES_DB: ${KHOJ_POSTGRES_DB:-postgres}
      POSTGRES_USER: ${KHOJ_POSTGRES_USER:-postgres}
      POSTGRES_PASSWORD: ${KHOJ_POSTGRES_PASSWORD:?set KHOJ_POSTGRES_PASSWORD in .env}
      POSTGRES_HOST: database
      POSTGRES_PORT: 5432
      KHOJ_DJANGO_SECRET_KEY: ${KHOJ_DJANGO_SECRET_KEY:?set KHOJ_DJANGO_SECRET_KEY in .env}
      KHOJ_DEBUG: "False"
      KHOJ_ADMIN_EMAIL: ${KHOJ_ADMIN_EMAIL:?set KHOJ_ADMIN_EMAIL in .env}
      KHOJ_ADMIN_PASSWORD: ${KHOJ_ADMIN_PASSWORD:?set KHOJ_ADMIN_PASSWORD in .env}
      KHOJ_TERRARIUM_URL: http://sandbox:8080
      KHOJ_SEARXNG_URL: http://search:8080
      OPENAI_API_KEY: ${OPENAI_API_KEY:-}
      ANTHROPIC_API_KEY: ${ANTHROPIC_API_KEY:-}
      GEMINI_API_KEY: ${GEMINI_API_KEY:-}
      OPENAI_BASE_URL: ${OPENAI_BASE_URL:-}
      KHOJ_DEFAULT_CHAT_MODEL: ${KHOJ_DEFAULT_CHAT_MODEL:-}
    volumes:
      - khoj-config:/root/.khoj
      - khoj-models:/root/.cache/huggingface
    depends_on:
      database:
        condition: service_healthy
      sandbox:
        condition: service_started
      search:
        condition: service_started
volumes:
  khoj-config:
  khoj-db:
  khoj-models:
  khoj-search:
`,
			EnvContent: "KHOJ_PORT=42110\nKHOJ_IMAGE_TAG=latest\nKHOJ_POSTGRES_DB=postgres\nKHOJ_POSTGRES_USER=postgres\nKHOJ_POSTGRES_PASSWORD=\nKHOJ_DJANGO_SECRET_KEY=\nKHOJ_ADMIN_EMAIL=admin@example.com\nKHOJ_ADMIN_PASSWORD=\nOPENAI_API_KEY=\nANTHROPIC_API_KEY=\nGEMINI_API_KEY=\nOPENAI_BASE_URL=\nKHOJ_DEFAULT_CHAT_MODEL=\n",
			Notes:      "Best for a personal/team second brain. Set a provider key or point OPENAI_BASE_URL at a local OpenAI-compatible LLM before expecting chat results.",
		},
		{
			ID:          "docsgpt",
			Name:        "DocsGPT",
			Description: "Private AI platform for document assistants, enterprise search, agent builder, document analysis, and multi-model chat.",
			Category:    "ai",
			Subcategory: "workflow-rag",
			Source:      "official-github",
			Image:       "arc53/docsgpt:develop",
			Tags:        []string{"ai", "documents", "rag", "enterprise-search"},
			ComposeContent: `services:
  frontend:
    image: arc53/docsgpt-fe:${DOCSGPT_IMAGE_TAG:-develop}
    restart: unless-stopped
    ports:
      - "${DOCSGPT_FRONTEND_PORT:-5173}:5173"
    environment:
      VITE_API_HOST: ${DOCSGPT_API_PUBLIC_URL:-http://localhost:7091}
      VITE_API_STREAMING: ${VITE_API_STREAMING:-true}
      VITE_GOOGLE_CLIENT_ID: ${VITE_GOOGLE_CLIENT_ID:-}
    depends_on:
      - backend
  backend:
    image: arc53/docsgpt:${DOCSGPT_IMAGE_TAG:-develop}
    restart: unless-stopped
    ports:
      - "${DOCSGPT_API_PORT:-7091}:7091"
    environment:
      CELERY_BROKER_URL: redis://redis:6379/0
      CELERY_RESULT_BACKEND: redis://redis:6379/1
      CACHE_REDIS_URL: redis://redis:6379/2
      POSTGRES_URI: postgresql://docsgpt:${DOCSGPT_POSTGRES_PASSWORD:?set DOCSGPT_POSTGRES_PASSWORD in .env}@postgres:5432/docsgpt
      OPENAI_API_KEY: ${OPENAI_API_KEY:-}
      ANTHROPIC_API_KEY: ${ANTHROPIC_API_KEY:-}
    volumes:
      - docsgpt-indexes:/app/indexes
      - docsgpt-inputs:/app/inputs
      - docsgpt-vectors:/app/vectors
    depends_on:
      redis:
        condition: service_started
      postgres:
        condition: service_healthy
  worker:
    image: arc53/docsgpt:${DOCSGPT_IMAGE_TAG:-develop}
    restart: unless-stopped
    command: celery -A application.app.celery worker -l INFO -B
    environment:
      CELERY_BROKER_URL: redis://redis:6379/0
      CELERY_RESULT_BACKEND: redis://redis:6379/1
      API_URL: http://backend:7091
      CACHE_REDIS_URL: redis://redis:6379/2
      POSTGRES_URI: postgresql://docsgpt:${DOCSGPT_POSTGRES_PASSWORD:?set DOCSGPT_POSTGRES_PASSWORD in .env}@postgres:5432/docsgpt
      OPENAI_API_KEY: ${OPENAI_API_KEY:-}
      ANTHROPIC_API_KEY: ${ANTHROPIC_API_KEY:-}
    volumes:
      - docsgpt-indexes:/app/indexes
      - docsgpt-inputs:/app/inputs
      - docsgpt-vectors:/app/vectors
    depends_on:
      redis:
        condition: service_started
      postgres:
        condition: service_healthy
  redis:
    image: redis:7-alpine
    restart: unless-stopped
  postgres:
    image: postgres:16-alpine
    restart: unless-stopped
    environment:
      POSTGRES_USER: docsgpt
      POSTGRES_PASSWORD: ${DOCSGPT_POSTGRES_PASSWORD:?set DOCSGPT_POSTGRES_PASSWORD in .env}
      POSTGRES_DB: docsgpt
    volumes:
      - docsgpt-postgres:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U docsgpt -d docsgpt"]
      interval: 5s
      timeout: 5s
      retries: 10
volumes:
  docsgpt-indexes:
  docsgpt-inputs:
  docsgpt-vectors:
  docsgpt-postgres:
`,
			EnvContent: "DOCSGPT_FRONTEND_PORT=5173\nDOCSGPT_API_PORT=7091\nDOCSGPT_API_PUBLIC_URL=http://localhost:7091\nDOCSGPT_IMAGE_TAG=develop\nDOCSGPT_POSTGRES_PASSWORD=\nVITE_API_STREAMING=true\nVITE_GOOGLE_CLIENT_ID=\nOPENAI_API_KEY=\nANTHROPIC_API_KEY=\n",
			Notes:      "Use DocsGPT for private document Q&A and assistants. The upstream Docker Hub tags currently use develop; pin to a tested tag before production rollout.",
		},
		{
			ID:          "openmemory-mem0",
			Name:        "OpenMemory + Mem0",
			Description: "Self-hosted Mem0/OpenMemory stack with an MCP memory API, web UI, and Qdrant vector storage.",
			Category:    "ai",
			Subcategory: "workflow-rag",
			Source:      "official-github",
			Image:       "mem0/openmemory-mcp:latest",
			Tags:        []string{"ai", "memory", "mcp", "agents"},
			ComposeContent: `services:
  mem0_store:
    image: qdrant/qdrant:latest
    restart: unless-stopped
    ports:
      - "${MEM0_QDRANT_PORT:-6333}:6333"
    volumes:
      - mem0-storage:/qdrant/storage
  openmemory-mcp:
    image: mem0/openmemory-mcp:latest
    restart: unless-stopped
    ports:
      - "${OPENMEMORY_API_PORT:-8765}:8765"
    environment:
      USER: ${OPENMEMORY_USER:?set OPENMEMORY_USER in .env}
      API_KEY: ${OPENMEMORY_API_KEY:?set OPENMEMORY_API_KEY in .env}
      QDRANT_HOST: mem0_store
      QDRANT_PORT: 6333
      OPENAI_API_KEY: ${OPENAI_API_KEY:-}
    depends_on:
      - mem0_store
  openmemory-ui:
    image: mem0/openmemory-ui:latest
    restart: unless-stopped
    ports:
      - "${OPENMEMORY_UI_PORT:-3000}:3000"
    environment:
      NEXT_PUBLIC_API_URL: ${OPENMEMORY_PUBLIC_API_URL:-http://localhost:8765}
      NEXT_PUBLIC_USER_ID: ${OPENMEMORY_USER:?set OPENMEMORY_USER in .env}
    depends_on:
      - openmemory-mcp
volumes:
  mem0-storage:
`,
			EnvContent: "OPENMEMORY_UI_PORT=3000\nOPENMEMORY_API_PORT=8765\nMEM0_QDRANT_PORT=6333\nOPENMEMORY_USER=admin\nOPENMEMORY_API_KEY=\nOPENMEMORY_PUBLIC_API_URL=http://localhost:8765\nOPENAI_API_KEY=\n",
			Notes:      "Good for adding durable memory to agents and MCP clients. Protect the API with a reverse proxy/auth layer before exposing it outside a trusted network.",
		},
		{
			ID:          "langfuse",
			Name:        "Langfuse",
			Description: "Open-source LLM observability, tracing, prompt management, metrics, datasets, playground, and evaluations platform.",
			Category:    "ai",
			Subcategory: "observability",
			Source:      "official-github",
			Image:       "langfuse/langfuse:3",
			Tags:        []string{"ai", "observability", "evals", "prompts", "tracing"},
			ComposeContent: `services:
  langfuse-worker:
    image: langfuse/langfuse-worker:3
    restart: unless-stopped
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      clickhouse:
        condition: service_healthy
      minio:
        condition: service_healthy
    environment: &langfuse-env
      NEXTAUTH_URL: ${NEXTAUTH_URL:-http://localhost:3000}
      DATABASE_URL: postgresql://postgres:${LANGFUSE_POSTGRES_PASSWORD:?set LANGFUSE_POSTGRES_PASSWORD in .env}@postgres:5432/postgres
      SALT: ${LANGFUSE_SALT:?set LANGFUSE_SALT in .env}
      ENCRYPTION_KEY: ${LANGFUSE_ENCRYPTION_KEY:?set LANGFUSE_ENCRYPTION_KEY in .env}
      CLICKHOUSE_URL: http://clickhouse:8123
      CLICKHOUSE_USER: clickhouse
      CLICKHOUSE_PASSWORD: ${LANGFUSE_CLICKHOUSE_PASSWORD:?set LANGFUSE_CLICKHOUSE_PASSWORD in .env}
      REDIS_HOST: redis
      REDIS_PORT: 6379
      REDIS_AUTH: ${LANGFUSE_REDIS_PASSWORD:?set LANGFUSE_REDIS_PASSWORD in .env}
      LANGFUSE_S3_EVENT_UPLOAD_BUCKET: langfuse
      LANGFUSE_S3_EVENT_UPLOAD_ACCESS_KEY_ID: ${LANGFUSE_MINIO_USER:?set LANGFUSE_MINIO_USER in .env}
      LANGFUSE_S3_EVENT_UPLOAD_SECRET_ACCESS_KEY: ${LANGFUSE_MINIO_PASSWORD:?set LANGFUSE_MINIO_PASSWORD in .env}
      LANGFUSE_S3_EVENT_UPLOAD_ENDPOINT: http://minio:9000
      LANGFUSE_S3_EVENT_UPLOAD_FORCE_PATH_STYLE: "true"
      LANGFUSE_S3_MEDIA_UPLOAD_BUCKET: langfuse
      LANGFUSE_S3_MEDIA_UPLOAD_ACCESS_KEY_ID: ${LANGFUSE_MINIO_USER:?set LANGFUSE_MINIO_USER in .env}
      LANGFUSE_S3_MEDIA_UPLOAD_SECRET_ACCESS_KEY: ${LANGFUSE_MINIO_PASSWORD:?set LANGFUSE_MINIO_PASSWORD in .env}
      LANGFUSE_S3_MEDIA_UPLOAD_ENDPOINT: http://minio:9000
      LANGFUSE_S3_MEDIA_UPLOAD_FORCE_PATH_STYLE: "true"
  langfuse-web:
    image: langfuse/langfuse:3
    restart: unless-stopped
    ports:
      - "${LANGFUSE_PORT:-3000}:3000"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      clickhouse:
        condition: service_healthy
      minio:
        condition: service_healthy
    environment:
      <<: *langfuse-env
      NEXTAUTH_SECRET: ${NEXTAUTH_SECRET:?set NEXTAUTH_SECRET in .env}
  clickhouse:
    image: clickhouse/clickhouse-server:latest
    restart: unless-stopped
    user: "101:101"
    environment:
      CLICKHOUSE_DB: default
      CLICKHOUSE_USER: clickhouse
      CLICKHOUSE_PASSWORD: ${LANGFUSE_CLICKHOUSE_PASSWORD:?set LANGFUSE_CLICKHOUSE_PASSWORD in .env}
    volumes:
      - langfuse-clickhouse-data:/var/lib/clickhouse
      - langfuse-clickhouse-logs:/var/log/clickhouse-server
    healthcheck:
      test: ["CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:8123/ping || exit 1"]
      interval: 5s
      timeout: 5s
      retries: 10
  minio:
    image: minio/minio:latest
    restart: unless-stopped
    command: server --address ":9000" --console-address ":9001" /data
    environment:
      MINIO_ROOT_USER: ${LANGFUSE_MINIO_USER:?set LANGFUSE_MINIO_USER in .env}
      MINIO_ROOT_PASSWORD: ${LANGFUSE_MINIO_PASSWORD:?set LANGFUSE_MINIO_PASSWORD in .env}
    volumes:
      - langfuse-minio:/data
    healthcheck:
      test: ["CMD", "mc", "ready", "local"]
      interval: 5s
      timeout: 5s
      retries: 10
  redis:
    image: redis:7
    restart: unless-stopped
    command: --requirepass ${LANGFUSE_REDIS_PASSWORD:?set LANGFUSE_REDIS_PASSWORD in .env} --maxmemory-policy noeviction
    volumes:
      - langfuse-redis:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 10
  postgres:
    image: postgres:17
    restart: unless-stopped
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${LANGFUSE_POSTGRES_PASSWORD:?set LANGFUSE_POSTGRES_PASSWORD in .env}
      POSTGRES_DB: postgres
      TZ: UTC
      PGTZ: UTC
    volumes:
      - langfuse-postgres:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 10
volumes:
  langfuse-postgres:
  langfuse-clickhouse-data:
  langfuse-clickhouse-logs:
  langfuse-minio:
  langfuse-redis:
`,
			EnvContent: "LANGFUSE_PORT=3000\nNEXTAUTH_URL=http://localhost:3000\nNEXTAUTH_SECRET=\nLANGFUSE_SALT=\nLANGFUSE_ENCRYPTION_KEY=\nLANGFUSE_POSTGRES_PASSWORD=\nLANGFUSE_REDIS_PASSWORD=\nLANGFUSE_CLICKHOUSE_PASSWORD=\nLANGFUSE_MINIO_USER=\nLANGFUSE_MINIO_PASSWORD=\n",
			Notes:      "Use Langfuse for production-grade LLM tracing and prompt/eval workflows. Generate ENCRYPTION_KEY with 32 random bytes as hex before launch.",
		},
		{
			ID:          "phoenix",
			Name:        "Arize Phoenix",
			Description: "Open-source AI observability and evaluation platform for traces, datasets, experiments, prompt management, and playgrounds.",
			Category:    "ai",
			Subcategory: "observability",
			Source:      "official-github",
			Image:       "arizephoenix/phoenix:latest",
			Tags:        []string{"ai", "observability", "tracing", "evals"},
			ComposeContent: `services:
  phoenix:
    image: arizephoenix/phoenix:latest
    restart: unless-stopped
    ports:
      - "${PHOENIX_PORT:-6006}:6006"
      - "${PHOENIX_OTLP_GRPC_PORT:-4317}:4317"
    environment:
      PHOENIX_SQL_DATABASE_URL: postgresql://phoenix:${PHOENIX_POSTGRES_PASSWORD:?set PHOENIX_POSTGRES_PASSWORD in .env}@db:5432/phoenix
    depends_on:
      - db
  db:
    image: postgres:16-alpine
    restart: unless-stopped
    environment:
      POSTGRES_USER: phoenix
      POSTGRES_PASSWORD: ${PHOENIX_POSTGRES_PASSWORD:?set PHOENIX_POSTGRES_PASSWORD in .env}
      POSTGRES_DB: phoenix
    volumes:
      - phoenix-postgres:/var/lib/postgresql/data
volumes:
  phoenix-postgres:
`,
			EnvContent: "PHOENIX_PORT=6006\nPHOENIX_OTLP_GRPC_PORT=4317\nPHOENIX_POSTGRES_PASSWORD=\n",
			Notes:      "Phoenix is lighter than Langfuse and strong for OpenTelemetry traces and evaluations. Point instrumented apps at the OTLP gRPC port.",
		},
		{
			ID:          "promptfoo",
			Name:        "promptfoo",
			Description: "Self-hosted prompt, agent, RAG, and model evaluation UI for private eval sharing and red-team workflows.",
			Category:    "ai",
			Subcategory: "evals",
			Source:      "official-docs",
			Image:       "ghcr.io/promptfoo/promptfoo:latest",
			Tags:        []string{"ai", "evals", "red-team", "security"},
			ComposeContent: `services:
  promptfoo:
    image: ghcr.io/promptfoo/promptfoo:${PROMPTFOO_IMAGE_TAG:-latest}
    restart: unless-stopped
    ports:
      - "${PROMPTFOO_PORT:-3000}:3000"
    environment:
      PROMPTFOO_SHARE_CHUNK_SIZE: ${PROMPTFOO_SHARE_CHUNK_SIZE:-10}
      PROMPTFOO_DISABLE_TELEMETRY: ${PROMPTFOO_DISABLE_TELEMETRY:-true}
      OPENAI_API_KEY: ${OPENAI_API_KEY:-}
      ANTHROPIC_API_KEY: ${ANTHROPIC_API_KEY:-}
      GOOGLE_API_KEY: ${GOOGLE_API_KEY:-}
    volumes:
      - promptfoo-data:/home/promptfoo/.promptfoo
volumes:
  promptfoo-data:
`,
			EnvContent: "PROMPTFOO_PORT=3000\nPROMPTFOO_IMAGE_TAG=latest\nPROMPTFOO_SHARE_CHUNK_SIZE=10\nPROMPTFOO_DISABLE_TELEMETRY=true\nOPENAI_API_KEY=\nANTHROPIC_API_KEY=\nGOOGLE_API_KEY=\n",
			Notes:      "Good for testing prompts, RAG outputs, agents, and red-team cases. The self-hosted open image is for individual/experimental use and has no built-in auth, so place it behind access control.",
		},
		{
			ID:          "firecrawl",
			Name:        "Firecrawl",
			Description: "Self-hosted web context API for search, scrape, crawl, extract, and agent-ready clean Markdown/structured data.",
			Category:    "ai",
			Subcategory: "search",
			Source:      "official-github",
			Image:       "ghcr.io/firecrawl/firecrawl:latest",
			Tags:        []string{"ai", "search", "scraping", "crawler", "agents"},
			ComposeContent: `services:
  api:
    image: ghcr.io/firecrawl/firecrawl:latest
    restart: unless-stopped
    command: node dist/src/harness.js --start-docker
    ports:
      - "${FIRECRAWL_PORT:-3002}:3002"
    environment:
      HOST: 0.0.0.0
      PORT: 3002
      REDIS_URL: redis://redis:6379
      REDIS_RATE_LIMIT_URL: redis://redis:6379
      PLAYWRIGHT_MICROSERVICE_URL: http://playwright-service:3000/scrape
      POSTGRES_USER: ${FIRECRAWL_POSTGRES_USER:-firecrawl}
      POSTGRES_PASSWORD: ${FIRECRAWL_POSTGRES_PASSWORD:?set FIRECRAWL_POSTGRES_PASSWORD in .env}
      POSTGRES_DB: ${FIRECRAWL_POSTGRES_DB:-firecrawl}
      POSTGRES_HOST: nuq-postgres
      POSTGRES_PORT: 5432
      USE_DB_AUTHENTICATION: ${USE_DB_AUTHENTICATION:-false}
      BULL_AUTH_KEY: ${BULL_AUTH_KEY:?set BULL_AUTH_KEY in .env}
      OPENAI_API_KEY: ${OPENAI_API_KEY:-}
      OPENAI_BASE_URL: ${OPENAI_BASE_URL:-}
      OLLAMA_BASE_URL: ${OLLAMA_BASE_URL:-}
      MODEL_NAME: ${MODEL_NAME:-}
      MODEL_EMBEDDING_NAME: ${MODEL_EMBEDDING_NAME:-}
      PROXY_SERVER: ${PROXY_SERVER:-}
      SEARXNG_ENDPOINT: ${SEARXNG_ENDPOINT:-}
    depends_on:
      - redis
      - playwright-service
      - rabbitmq
      - nuq-postgres
  playwright-service:
    image: ghcr.io/firecrawl/playwright-service:latest
    restart: unless-stopped
    environment:
      PORT: 3000
      PROXY_SERVER: ${PROXY_SERVER:-}
      BLOCK_MEDIA: ${BLOCK_MEDIA:-}
      MAX_CONCURRENT_PAGES: ${CRAWL_CONCURRENT_REQUESTS:-10}
  redis:
    image: redis:7-alpine
    restart: unless-stopped
    command: redis-server --bind 0.0.0.0
  rabbitmq:
    image: rabbitmq:3-management-alpine
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "rabbitmq-diagnostics", "-q", "check_running"]
      interval: 5s
      timeout: 5s
      retries: 5
  nuq-postgres:
    image: postgres:16-alpine
    restart: unless-stopped
    environment:
      POSTGRES_USER: ${FIRECRAWL_POSTGRES_USER:-firecrawl}
      POSTGRES_PASSWORD: ${FIRECRAWL_POSTGRES_PASSWORD:?set FIRECRAWL_POSTGRES_PASSWORD in .env}
      POSTGRES_DB: ${FIRECRAWL_POSTGRES_DB:-firecrawl}
    volumes:
      - firecrawl-postgres:/var/lib/postgresql/data
volumes:
  firecrawl-postgres:
`,
			EnvContent: "FIRECRAWL_PORT=3002\nFIRECRAWL_POSTGRES_USER=firecrawl\nFIRECRAWL_POSTGRES_PASSWORD=\nFIRECRAWL_POSTGRES_DB=firecrawl\nUSE_DB_AUTHENTICATION=false\nBULL_AUTH_KEY=\nOPENAI_API_KEY=\nOPENAI_BASE_URL=\nOLLAMA_BASE_URL=\nMODEL_NAME=\nMODEL_EMBEDDING_NAME=\nPROXY_SERVER=\nSEARXNG_ENDPOINT=\nBLOCK_MEDIA=\nCRAWL_CONCURRENT_REQUESTS=10\n",
			Notes:      "Use Firecrawl for agent web context. Self-hosting lacks Firecrawl cloud's fire-engine anti-blocking layer; use proxies and strict access control for public deployments.",
		},
		{
			ID:          "crawl4ai",
			Name:        "Crawl4AI",
			Description: "Open-source LLM-friendly crawler and scraper with secure-by-default Docker API server for Markdown/RAG extraction.",
			Category:    "ai",
			Subcategory: "search",
			Source:      "official-github",
			Image:       "unclecode/crawl4ai:latest",
			Tags:        []string{"ai", "crawler", "scraping", "rag", "secure-by-default"},
			ComposeContent: `services:
  crawl4ai:
    image: unclecode/crawl4ai:${CRAWL4AI_TAG:-latest}
    restart: unless-stopped
    user: appuser
    ports:
      - "${CRAWL4AI_PORT:-11235}:11235"
    environment:
      CRAWL4AI_API_TOKEN: ${CRAWL4AI_API_TOKEN:?set CRAWL4AI_API_TOKEN in .env}
      OPENAI_API_KEY: ${OPENAI_API_KEY:-}
      ANTHROPIC_API_KEY: ${ANTHROPIC_API_KEY:-}
      GEMINI_API_KEY: ${GEMINI_API_KEY:-}
      LLM_PROVIDER: ${LLM_PROVIDER:-}
    shm_size: 1gb
    cap_drop:
      - ALL
    security_opt:
      - no-new-privileges:true
    pids_limit: 512
    read_only: true
    tmpfs:
      - /tmp
      - /var/lib/redis
      - /var/lib/crawl4ai/outputs:mode=0700
      - /home/appuser/.cache
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:11235/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
volumes: {}
`,
			EnvContent: "CRAWL4AI_PORT=11235\nCRAWL4AI_TAG=latest\nCRAWL4AI_API_TOKEN=\nOPENAI_API_KEY=\nANTHROPIC_API_KEY=\nGEMINI_API_KEY=\nLLM_PROVIDER=\n",
			Notes:      "Crawl4AI is a strong local crawler/scraper option. Keep the API token enabled and do not expose it without a reverse proxy and firewall rules.",
		},
	}
}
