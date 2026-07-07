const projectGuides = [
  {
    id: 'librechat',
    name: 'LibreChat',
    fit: 'Multi-user AI chat, agents, MCP, files, and multi-provider routing.',
    setup: [
      'Set MEILI_MASTER_KEY, JWT secrets, credential encryption values, admin panel secret, and vector DB password.',
      'Add at least one provider key or configure a local OpenAI-compatible endpoint after first boot.',
      'Open LIBRECHAT_PORT for users and LIBRECHAT_ADMIN_PORT for administration.',
    ],
    caution: 'Put it behind HTTPS before uploads, agents, or multi-user use.',
    links: [
      ['GitHub', 'https://github.com/danny-avila/LibreChat'],
      ['Docs', 'https://docs.librechat.ai/'],
    ],
  },
  {
    id: 'onyx',
    name: 'Onyx',
    fit: 'Enterprise knowledge search and RAG with connectors.',
    setup: [
      'Set Postgres, OpenSearch, and S3/MinIO secrets before launch.',
      'Start with enough memory for OpenSearch and the model services.',
      'Configure auth and connectors in the app before indexing broad sources.',
    ],
    caution: 'This is one of the heavier templates; do not colocate it with other GPU-heavy AI stacks on a single small server.',
    links: [
      ['GitHub', 'https://github.com/onyx-dot-app/onyx'],
      ['Docs', 'https://docs.onyx.app/'],
    ],
  },
  {
    id: 'khoj',
    name: 'Khoj',
    fit: 'Personal or small-team second brain with docs, web answers, agents, and automations.',
    setup: [
      'Set Postgres password, Django secret, admin email, and admin password.',
      'Add a hosted provider key or set OPENAI_BASE_URL and KHOJ_DEFAULT_CHAT_MODEL for a local model.',
      'Open KHOJ_PORT after startup and finish onboarding.',
    ],
    caution: 'Keep reverse-proxy domain settings aligned if exposing it outside localhost.',
    links: [
      ['GitHub', 'https://github.com/khoj-ai/khoj'],
      ['Docs', 'https://docs.khoj.dev/'],
    ],
  },
  {
    id: 'docsgpt',
    name: 'DocsGPT',
    fit: 'Private document Q&A, support assistants, and enterprise search.',
    setup: [
      'Set DOCSGPT_POSTGRES_PASSWORD.',
      'Set at least one model provider key before expecting generated answers.',
      'Keep DOCSGPT_API_PUBLIC_URL reachable from user browsers.',
    ],
    caution: 'Pin a tested image tag before production because upstream compose examples commonly use develop images.',
    links: [
      ['GitHub', 'https://github.com/arc53/DocsGPT'],
    ],
  },
  {
    id: 'openmemory-mem0',
    name: 'OpenMemory + Mem0',
    fit: 'Durable memory API for agents and MCP clients.',
    setup: [
      'Set OPENMEMORY_USER and OPENMEMORY_API_KEY.',
      'Set OPENAI_API_KEY if the memory extraction path uses OpenAI models.',
      'Point MCP clients at OPENMEMORY_API_PORT and use OPENMEMORY_UI_PORT for inspection.',
    ],
    caution: 'Memory APIs can expose sensitive user context; keep them private or behind access control.',
    links: [
      ['GitHub', 'https://github.com/mem0ai/mem0'],
      ['Docs', 'https://docs.mem0.ai/'],
    ],
  },
  {
    id: 'langfuse',
    name: 'Langfuse',
    fit: 'LLM traces, prompt management, datasets, metrics, playgrounds, and eval workflows.',
    setup: [
      'Generate NEXTAUTH_SECRET, SALT, ENCRYPTION_KEY, and service passwords.',
      'Start the stack and create an org/project.',
      'Copy Langfuse keys into apps that should report traces.',
    ],
    caution: 'Back up Postgres and ClickHouse before relying on it for production telemetry.',
    links: [
      ['GitHub', 'https://github.com/langfuse/langfuse'],
      ['Docs', 'https://langfuse.com/docs'],
    ],
  },
  {
    id: 'phoenix',
    name: 'Arize Phoenix',
    fit: 'Lightweight AI tracing, experiments, datasets, prompt work, and evaluations.',
    setup: [
      'Set PHOENIX_POSTGRES_PASSWORD.',
      'Start the stack and open PHOENIX_PORT.',
      'Point OpenTelemetry instrumentation at PHOENIX_OTLP_GRPC_PORT.',
    ],
    caution: 'Place the UI behind auth if it leaves a trusted network.',
    links: [
      ['GitHub', 'https://github.com/Arize-ai/phoenix'],
      ['Docs', 'https://phoenix.arize.com/'],
    ],
  },
  {
    id: 'promptfoo',
    name: 'promptfoo',
    fit: 'Prompt, RAG, model, and agent evals, including red-team regression checks.',
    setup: [
      'Set provider keys used by evals.',
      'Start the stack and open PROMPTFOO_PORT.',
      'Keep test outputs in the included persistent volume.',
    ],
    caution: 'The open self-host image does not include built-in auth; protect it before exposure.',
    links: [
      ['GitHub', 'https://github.com/promptfoo/promptfoo'],
      ['Self-hosting', 'https://www.promptfoo.dev/docs/usage/self-hosting/'],
    ],
  },
  {
    id: 'firecrawl',
    name: 'Firecrawl',
    fit: 'Search, scrape, crawl, extract, screenshots, and agent web context API.',
    setup: [
      'Set FIRECRAWL_POSTGRES_PASSWORD and BULL_AUTH_KEY.',
      'Add model provider settings only when using AI extraction features.',
      'Call the API on FIRECRAWL_PORT after startup.',
    ],
    caution: 'Self-hosting does not include Firecrawl cloud anti-blocking; use proxies, rate limits, and private access.',
    links: [
      ['GitHub', 'https://github.com/firecrawl/firecrawl'],
      ['Self-hosting', 'https://github.com/firecrawl/firecrawl/blob/main/SELF_HOST.md'],
    ],
  },
  {
    id: 'crawl4ai',
    name: 'Crawl4AI',
    fit: 'Local crawler/scraper API that emits LLM-ready Markdown for RAG workflows.',
    setup: [
      'Set CRAWL4AI_API_TOKEN.',
      'Add provider keys only when using LLM extraction.',
      'Call the API on CRAWL4AI_PORT after the health check passes.',
    ],
    caution: 'Keep token auth enabled and avoid direct public exposure.',
    links: [
      ['GitHub', 'https://github.com/unclecode/crawl4ai'],
      ['Docs', 'https://docs.crawl4ai.com/'],
    ],
  },
];

export default function Documentation() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-gray-950">Documentation</h1>
        <p className="text-sm text-gray-600">Operational notes for Docker Compose stacks and the vetted AI catalog.</p>
      </div>

      <section className="section-panel space-y-4">
        <div>
          <h2 className="text-lg font-semibold text-gray-950">Docker Compose</h2>
          <p className="mt-1 text-sm leading-6 text-gray-600">
            Stack Manager creates normal Docker Compose projects. Review each template, set its `.env` values, then use Spin it Up or standard Compose commands such as <code className="rounded bg-gray-100 px-1 py-0.5">docker compose up -d</code>, <code className="rounded bg-gray-100 px-1 py-0.5">docker compose ps</code>, <code className="rounded bg-gray-100 px-1 py-0.5">docker compose logs</code>, and <code className="rounded bg-gray-100 px-1 py-0.5">docker compose down</code>.
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          <DocLink href="https://docs.docker.com/compose/">Docker Compose docs</DocLink>
          <DocLink href="https://docs.docker.com/compose/how-tos/gpu-support/">Compose GPU support</DocLink>
        </div>
      </section>

      <section className="section-panel space-y-3 border-amber-200 bg-amber-50">
        <h2 className="text-lg font-semibold text-amber-900">Single-GPU Rule</h2>
        <p className="text-sm leading-6 text-amber-900">
          Do not plan to run multiple heavy AI projects on one GPU at the same time. LLM inference, image generation, voice models, and larger RAG/search stacks can fight over VRAM and crash or starve each other. On a one-GPU test server, start one AI stack, verify it, shut it down, then start the next unless each service is sized and pinned to a specific available GPU.
        </p>
      </section>

      <section className="space-y-3">
        <div>
          <h2 className="text-lg font-semibold text-gray-950">Vetted AI Projects</h2>
          <p className="text-sm text-gray-600">These are the non-filler AI additions with credible upstreams and self-hosting paths.</p>
        </div>
        <div className="grid gap-3 lg:grid-cols-2">
          {projectGuides.map(project => (
            <article key={project.id} className="rounded-md border border-gray-200 bg-white p-4 shadow-sm">
              <div className="flex flex-wrap items-start justify-between gap-2">
                <div>
                  <h3 className="font-semibold text-gray-950">{project.name}</h3>
                  <div className="mt-1 font-mono text-xs text-gray-500">{project.id}</div>
                </div>
                <div className="flex flex-wrap gap-2">
                  {project.links.map(([label, href]) => <DocLink key={href} href={href} compact>{label}</DocLink>)}
                </div>
              </div>
              <p className="mt-3 text-sm leading-6 text-gray-600">{project.fit}</p>
              <ul className="mt-3 list-disc space-y-1 pl-5 text-sm leading-6 text-gray-700">
                {project.setup.map(item => <li key={item}>{item}</li>)}
              </ul>
              <p className="mt-3 rounded-md border border-blue-100 bg-blue-50 px-3 py-2 text-sm leading-6 text-blue-900">{project.caution}</p>
            </article>
          ))}
        </div>
      </section>
    </div>
  );
}

function DocLink({ href, children, compact = false }) {
  return (
    <a
      href={href}
      target="_blank"
      rel="noreferrer"
      className={`inline-flex rounded-md border border-blue-200 bg-blue-50 font-medium text-blue-800 hover:bg-blue-100 ${compact ? 'px-2 py-1 text-xs' : 'px-3 py-2 text-sm'}`}
      title={`Open ${children} in a new tab.`}
    >
      {children}
    </a>
  );
}
