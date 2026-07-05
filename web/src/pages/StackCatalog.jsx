import { useEffect, useMemo, useState } from 'react';
import { projects, stackTemplates } from '../api/client';

const CATEGORY_LABELS = {
  ai: 'AI',
  automation: 'Automation',
  cms: 'CMS',
  database: 'Database',
  devtools: 'Dev Tools',
  docs: 'Docs',
  files: 'Files',
  management: 'Management',
  media: 'Media',
  monitoring: 'Monitoring',
  proxy: 'Proxy',
  queue: 'Queue',
  security: 'Security',
  web: 'Web',
};

const CATEGORY_DESCRIPTIONS = {
  all: 'Every built-in stack. Use the search box or the category chips to narrow the list.',
  ai: 'AI/ML stacks: LLM inference, image and voice generation, vector databases, RAG workflows, search, code assistants, and personal agent gateways.',
  automation: 'Task automation, workflow engines, cron schedulers, and notification systems.',
  cms: 'Content management systems, blog platforms, and e-commerce storefronts.',
  database: 'SQL, NoSQL, and graph databases with persistent volumes ready to be shared with other stacks.',
  devtools: 'Developer tools: CI/CD servers, Git forges, in-browser IDEs, code intelligence, and diagram editors.',
  docs: 'Wikis, technical documentation, and knowledge bases.',
  files: 'File sync, share, and object storage servers.',
  management: 'Docker and infrastructure management dashboards.',
  media: 'Media servers, libraries, and download automation.',
  monitoring: 'Metrics, uptime, logging, and observability.',
  proxy: 'Reverse proxies, load balancers, and forward proxies — most also handle TLS termination.',
  queue: 'Message queues, brokers, and workflow orchestrators.',
  security: 'Authentication, SSO, VPNs, and security scanners.',
  web: 'Web servers and static hosting.',
};

const SUBCATEGORY_LABELS = {
  'llm-inference': 'LLM inference',
  'code-assistants': 'Code assistants',
  'personal-agents': 'Personal AI agents',
  'image-generation': 'Image generation',
  'voice-speech': 'Voice / speech',
  'vector-db': 'Vector DB',
  'workflow-rag': 'Workflow / RAG',
  'search': 'Search',
};

const SUBCATEGORY_DESCRIPTIONS = {
  'all': 'Every AI template. Pick a sub-category to focus.',
  'llm-inference': 'Run large language models locally or expose an OpenAI-compatible API for other services to hit.',
  'code-assistants': 'In-editor and CLI AI coding assistants — pair one with an LLM inference stack for a fully local setup.',
  'personal-agents': 'Self-hosted personal AI agents that bridge Discord, Slack, WhatsApp, Signal, Telegram, iMessage and other chat apps to an LLM. OpenClaw-style.',
  'image-generation': 'Stable Diffusion and other text-to-image / image-to-image tools. Most benefit from an NVIDIA GPU.',
  'voice-speech': 'Text-to-speech, automatic speech recognition, and voice cloning servers.',
  'vector-db': 'Vector similarity databases for embeddings and RAG. Pair with an LLM inference stack.',
  'workflow-rag': 'RAG pipelines and LLM workflow builders — orchestration for chains, tools, and document ingestion.',
  'search': 'Full-text and hybrid search engines. Some support vector search alongside classic BM25.',
};

const SUBCATEGORY_ORDER = [
  'llm-inference',
  'code-assistants',
  'personal-agents',
  'image-generation',
  'voice-speech',
  'vector-db',
  'workflow-rag',
  'search',
];

function labelForCategory(cat) {
  return CATEGORY_LABELS[cat] || (cat ? cat.charAt(0).toUpperCase() + cat.slice(1) : cat);
}

function labelForSubcategory(sub) {
  return SUBCATEGORY_LABELS[sub] || sub;
}

const EMPTY_FORM = { name: '', compose_content: '', env_content: '', inactive: false, overwrite: false };

export default function StackCatalog() {
  const [templates, setTemplates] = useState([]);
  const [query, setQuery] = useState('');
  const [category, setCategory] = useState('all');
  const [subcategory, setSubcategory] = useState('all');
  const [message, setMessage] = useState(null);
  const [selected, setSelected] = useState(null);
  const [form, setForm] = useState(EMPTY_FORM);
  const [submitting, setSubmitting] = useState(false);

  const load = async () => {
    try {
      const res = await stackTemplates.list();
      setTemplates(res.data || []);
    } catch (err) {
      setMessage({ type: 'error', text: err.message });
    }
  };

  useEffect(() => { load(); }, []);
  useEffect(() => { setSubcategory('all'); }, [category]);
  useEffect(() => {
    if (!selected) return;
    const onKey = (event) => { if (event.key === 'Escape') closeModal(); };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [selected]);

  const categoryCounts = useMemo(() => {
    const counts = {};
    templates.forEach(t => { counts[t.category] = (counts[t.category] || 0) + 1; });
    return counts;
  }, [templates]);

  const categories = useMemo(() => {
    const set = new Set(templates.map(t => t.category));
    const order = ['ai', 'web', 'proxy', 'cms', 'database', 'devtools', 'docs', 'files', 'management', 'media', 'monitoring', 'queue', 'security', 'automation'];
    const known = order.filter(cat => set.has(cat));
    const extra = Array.from(set).filter(cat => !order.includes(cat)).sort();
    return ['all', ...known, ...extra];
  }, [templates]);

  const subcategoryCounts = useMemo(() => {
    const counts = {};
    templates.filter(t => t.category === 'ai').forEach(t => {
      const sub = t.subcategory || 'other';
      counts[sub] = (counts[sub] || 0) + 1;
    });
    return counts;
  }, [templates]);

  const aiSubcategories = useMemo(() => {
    const set = new Set(templates.filter(t => t.category === 'ai').map(t => t.subcategory || 'other'));
    const known = SUBCATEGORY_ORDER.filter(sub => set.has(sub));
    const extra = Array.from(set).filter(sub => !SUBCATEGORY_ORDER.includes(sub)).sort();
    return ['all', ...known, ...extra];
  }, [templates]);

  const filtered = templates.filter(template => {
    const q = query.trim().toLowerCase();
    if (category !== 'all' && template.category !== category) return false;
    if (category === 'ai' && subcategory !== 'all' && (template.subcategory || 'other') !== subcategory) return false;
    if (!q) return true;
    return [template.name, template.description, template.category, template.subcategory, ...(template.tags || [])].filter(Boolean).join(' ').toLowerCase().includes(q);
  });

  const openTemplate = (template) => {
    setSelected(template);
    setForm({
      name: template.id,
      compose_content: template.compose_content,
      env_content: template.env_content || '',
      inactive: false,
      overwrite: false,
    });
    setMessage(null);
  };

  const closeModal = () => {
    if (submitting) return;
    setSelected(null);
    setForm(EMPTY_FORM);
  };

  const spinItUp = async (event) => {
    event.preventDefault();
    setSubmitting(true);
    setMessage({ type: 'running', text: `Creating ${form.name}...` });
    try {
      await projects.create(form);
      setMessage({ type: 'ok', text: `Spun up ${form.name}. Head to the Dashboard to check status.` });
      setSelected(null);
      setForm(EMPTY_FORM);
    } catch (err) {
      setMessage({ type: 'error', text: err.message });
    } finally {
      setSubmitting(false);
    }
  };

  const activeCategoryDescription = CATEGORY_DESCRIPTIONS[category] || '';
  const activeSubDescription = category === 'ai' ? (SUBCATEGORY_DESCRIPTIONS[subcategory] || '') : '';

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-gray-950">Stack Catalog</h1>
          <p className="text-sm text-gray-600">Click a stack to open the editor, review the compose.yml + .env, then Spin it Up.</p>
        </div>
        <button onClick={load} className="btn-secondary" title="Reload the built-in stack catalog.">Refresh</button>
      </div>

      {message && <div className={`rounded border px-4 py-3 text-sm ${message.type === 'error' ? 'border-red-200 bg-red-50 text-red-900' : message.type === 'running' ? 'border-blue-200 bg-blue-50 text-blue-900' : 'border-green-200 bg-green-50 text-green-900'}`}>{message.text}</div>}

      <div className="section-panel space-y-3">
        <input className="input w-full" value={query} onChange={e => setQuery(e.target.value)} placeholder="search templates, tags, categories" title="Filter catalog templates by name, description, tags, or (sub)category." />
        <div className="flex flex-wrap gap-2" title="Choose a category. AI has its own sub-category selector below.">
          {categories.map(cat => {
            const total = cat === 'all' ? templates.length : (categoryCounts[cat] || 0);
            const active = category === cat;
            return (
              <button
                key={cat}
                type="button"
                onClick={() => setCategory(cat)}
                title={cat === 'all' ? 'Show every category.' : `Show only ${labelForCategory(cat)} templates.`}
                className={`rounded-full border px-3 py-1 text-xs font-medium ${active ? 'border-blue-500 bg-blue-500 text-white' : 'border-gray-200 bg-white text-gray-700 hover:border-blue-300'}`}
              >
                {cat === 'all' ? 'All' : labelForCategory(cat)} <span className={`ml-1 ${active ? 'text-blue-100' : 'text-gray-400'}`}>{total}</span>
              </button>
            );
          })}
        </div>
        {category === 'ai' && (
          <div className="flex flex-wrap gap-2 border-t border-gray-100 pt-3" title="AI sub-categories.">
            {aiSubcategories.map(sub => {
              const total = sub === 'all' ? (categoryCounts.ai || 0) : (subcategoryCounts[sub] || 0);
              const active = subcategory === sub;
              return (
                <button
                  key={sub}
                  type="button"
                  onClick={() => setSubcategory(sub)}
                  title={sub === 'all' ? 'Show every AI template.' : `Show only AI ${labelForSubcategory(sub)} templates.`}
                  className={`rounded-full border px-3 py-1 text-xs font-medium ${active ? 'border-purple-500 bg-purple-500 text-white' : 'border-gray-200 bg-white text-gray-700 hover:border-purple-300'}`}
                >
                  {sub === 'all' ? 'All AI' : labelForSubcategory(sub)} <span className={`ml-1 ${active ? 'text-purple-100' : 'text-gray-400'}`}>{total}</span>
                </button>
              );
            })}
          </div>
        )}
        {(activeCategoryDescription || activeSubDescription) && (
          <div className="rounded-md border border-blue-100 bg-blue-50 px-3 py-2 text-sm text-blue-900">
            <div>{activeCategoryDescription}</div>
            {activeSubDescription && <div className="mt-1 text-blue-800">{activeSubDescription}</div>}
          </div>
        )}
      </div>

      <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
        {filtered.map(template => (
          <button key={template.id} type="button" onClick={() => openTemplate(template)} className="min-w-0 rounded-md border border-gray-200 bg-white p-4 text-left text-sm shadow-sm hover:border-blue-300 focus:outline-none focus:ring-2 focus:ring-blue-200" title={`${template.name} — click to open the editor and spin it up.`}>
            <div className="flex items-start justify-between gap-2">
              <div className="min-w-0 flex-1">
                <div className="truncate font-medium text-gray-950" title={template.name}>{template.name}</div>
                <div className="mt-1 text-xs text-gray-500 line-clamp-2">{template.description}</div>
              </div>
              <div className="flex shrink-0 flex-col items-end gap-1">
                <Badge>{labelForCategory(template.category)}</Badge>
                {template.subcategory && <Badge tone="purple">{labelForSubcategory(template.subcategory)}</Badge>}
              </div>
            </div>
            <div className="mt-3 flex flex-wrap gap-1">
              {(template.tags || []).map(tag => <Badge key={tag} tone="cyan">{tag}</Badge>)}
            </div>
            {template.image && <div className="mt-3 min-w-0 break-all font-mono text-xs text-gray-500">{template.image}</div>}
          </button>
        ))}
        {filtered.length === 0 && <div className="py-12 text-center text-sm text-gray-500 md:col-span-2 xl:col-span-3 2xl:col-span-4">No templates match the current filters.</div>}
      </div>

      {selected && (
        <div className="fixed inset-0 z-40 flex items-center justify-center bg-gray-950/60 p-4" onClick={closeModal} role="dialog" aria-modal="true">
          <div className="flex max-h-[90vh] w-full max-w-4xl flex-col overflow-hidden rounded-lg bg-white shadow-2xl" onClick={e => e.stopPropagation()}>
            <div className="flex items-start justify-between gap-3 border-b border-gray-200 px-5 py-4">
              <div>
                <div className="flex items-center gap-2">
                  <h2 className="text-lg font-semibold text-gray-950">{selected.name}</h2>
                  <Badge>{labelForCategory(selected.category)}</Badge>
                  {selected.subcategory && <Badge tone="purple">{labelForSubcategory(selected.subcategory)}</Badge>}
                </div>
                <p className="mt-1 text-sm text-gray-600">{selected.notes || selected.description}</p>
              </div>
              <button type="button" onClick={closeModal} className="btn-secondary" title="Close without creating (Esc).">Close</button>
            </div>

            <form onSubmit={spinItUp} className="flex flex-1 flex-col gap-4 overflow-y-auto p-5">
              <div className="grid gap-3 md:grid-cols-[1fr_auto_auto]">
                <label className="block text-sm" title="Folder name to create under the Docker root.">
                  <span className="mb-1 block font-medium text-gray-700">Project name</span>
                  <input required disabled={submitting} value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} className="input" placeholder="example-stack" />
                </label>
                <label className="flex items-end gap-2 pb-2 text-sm text-gray-700" title="Create the folder but skip starting the stack until it's marked active.">
                  <input type="checkbox" disabled={submitting} checked={form.inactive} onChange={e => setForm({ ...form, inactive: e.target.checked })} />
                  Start inactive
                </label>
                <label className="flex items-end gap-2 pb-2 text-sm text-gray-700" title="Allow replacing compose.yml and .env if the folder already exists.">
                  <input type="checkbox" disabled={submitting} checked={form.overwrite} onChange={e => setForm({ ...form, overwrite: e.target.checked })} />
                  Overwrite existing
                </label>
              </div>
              <label className="block text-sm" title="Editable compose.yml before creation.">
                <span className="mb-1 block font-medium text-gray-700">compose.yml</span>
                <textarea disabled={submitting} required className="textarea h-72 font-mono" value={form.compose_content} onChange={e => setForm({ ...form, compose_content: e.target.value })} />
              </label>
              <label className="block text-sm" title="Editable .env with default settings for this stack.">
                <span className="mb-1 block font-medium text-gray-700">.env</span>
                <textarea disabled={submitting} className="textarea h-40 font-mono" value={form.env_content} onChange={e => setForm({ ...form, env_content: e.target.value })} />
              </label>
              <div className="flex flex-wrap items-center justify-end gap-2 border-t border-gray-200 pt-3">
                <button type="button" onClick={closeModal} disabled={submitting} className="btn-secondary" title="Cancel without creating (Esc).">Cancel</button>
                <button type="submit" disabled={submitting} className="btn-primary" title="Create the project folder, write compose.yml + .env, and (unless Start inactive is checked) start it.">
                  {submitting ? 'Spinning it up...' : 'Spin it Up'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

function Badge({ tone = 'gray', children }) {
  const tones = {
    gray: 'bg-gray-100 text-gray-700',
    cyan: 'bg-cyan-100 text-cyan-800',
    purple: 'bg-purple-100 text-purple-800',
  };
  return <span className={`inline-flex rounded px-2 py-0.5 text-xs font-medium ${tones[tone] || tones.gray}`}>{children}</span>;
}
