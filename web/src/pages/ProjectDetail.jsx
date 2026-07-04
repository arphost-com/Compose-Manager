import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { projects, debug as debugApi, security, backup, dbadmin } from '../api/client';

export default function ProjectDetail() {
  const { name } = useParams();
  const [project, setProject] = useState(null);
  const [activeTab, setActiveTab] = useState('overview');
  const [tabData, setTabData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [tabLoading, setTabLoading] = useState(false);
  const [actionResult, setActionResult] = useState(null);

  const fetchProject = async () => {
    try {
      const res = await projects.get(name);
      setProject(res.data);
    } catch (err) {
      setActionResult({ status: 'error', error: err.message });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchProject(); }, [name]);

  const loadTab = async (tab) => {
    setActiveTab(tab);
    setTabData(null);
    setTabLoading(true);
    try {
      let res;
      switch (tab) {
        case 'logs': res = await debugApi.logs(name); break;
        case 'stats': res = await debugApi.stats(name); break;
        case 'security': res = await security.scan(name); break;
        case 'backups': res = await backup.listProject(name); break;
        case 'databases': res = await dbadmin.health(name); break;
        case 'inspect': res = await debugApi.inspect(name); break;
        default: return;
      }
      setTabData(res.data);
    } catch (err) {
      setTabData({ error: err.message });
    } finally {
      setTabLoading(false);
    }
  };

  const runAction = async (action) => {
    try {
      setActionResult({ status: 'running', action });
      let res;
      switch (action) {
        case 'pull': res = await projects.pull(name); break;
        case 'up': res = await projects.up(name); break;
        case 'down': res = await projects.down(name); break;
        case 'restart': res = await projects.restart(name); break;
        case 'update': res = await projects.update(name); break;
        case 'backup': res = await backup.create(name); break;
        case 'db-dump': res = await dbadmin.dump(name); break;
      }
      setActionResult({ status: 'done', action, result: res.data });
      fetchProject();
    } catch (err) {
      setActionResult({ status: 'error', action, error: err.message });
    }
  };

  if (loading) return <div className="text-center py-12 text-gray-400">Loading...</div>;
  if (!project) return <div className="text-center py-12 text-red-400">Project not found</div>;

  const tabs = ['overview', 'logs', 'stats', 'security', 'backups', 'databases', 'inspect'];

  return (
    <div>
      {/* Header */}
      <div className="flex items-center gap-4 mb-6">
        <Link to="/" className="text-gray-400 hover:text-gray-200">&larr; Back</Link>
        <span className={`w-3 h-3 rounded-full ${project.running ? 'bg-green-500' : 'bg-gray-500'}`} />
        <h1 className="text-2xl font-bold">{project.name}</h1>
        {project.inactive && <span className="text-xs bg-yellow-800 text-yellow-200 px-2 py-0.5 rounded">inactive</span>}
      </div>

      {/* Actions */}
      <div className="flex gap-2 mb-6">
        {['pull', 'up', 'down', 'restart', 'update', 'backup', 'db-dump'].map(action => (
          <button key={action} onClick={() => runAction(action)}
            className="px-3 py-1.5 text-sm bg-gray-700 hover:bg-gray-600 rounded capitalize">
            {action}
          </button>
        ))}
      </div>

      {/* Action result */}
      {actionResult && (
        <div className={`mb-4 p-3 rounded text-sm ${
          actionResult.status === 'running' ? 'bg-blue-900 text-blue-200' :
          actionResult.status === 'error' ? 'bg-red-900 text-red-200' :
          'bg-green-900 text-green-200'
        }`}>
          {actionResult.status === 'running' ? `Running ${actionResult.action}...` :
           actionResult.status === 'error' ? `Error: ${actionResult.error}` :
           `${actionResult.action} completed`}
          <button onClick={() => setActionResult(null)} className="ml-4 underline">dismiss</button>
        </div>
      )}

      {/* Tabs */}
      <div className="flex gap-1 mb-4 border-b border-gray-700">
        {tabs.map(tab => (
          <button key={tab} onClick={() => tab === 'overview' ? setActiveTab(tab) : loadTab(tab)}
            className={`px-4 py-2 text-sm capitalize rounded-t ${
              activeTab === tab ? 'bg-gray-700 text-white' : 'text-gray-400 hover:text-gray-200'
            }`}>{tab}</button>
        ))}
      </div>

      {/* Tab content */}
      <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
        {activeTab === 'overview' && (
          <div>
            <p className="text-gray-400 text-sm mb-2">Directory: <code className="text-gray-300">{project.dir}</code></p>
            <p className="text-gray-400 text-sm mb-4">Compose file: <code className="text-gray-300">{project.compose_file}</code></p>
            <h3 className="font-semibold mb-2">Containers ({project.containers?.length || 0})</h3>
            {project.containers?.map(c => (
              <div key={c.name} className="flex items-center gap-3 py-2 border-b border-gray-700 last:border-0">
                <span className={`w-2 h-2 rounded-full ${c.state === 'running' ? 'bg-green-500' : 'bg-gray-500'}`} />
                <span className="font-mono text-sm">{c.name}</span>
                <span className="text-gray-400 text-sm">{c.image}</span>
                <span className="text-gray-500 text-xs">{c.state}</span>
                {c.ports && <span className="text-gray-500 text-xs">{c.ports}</span>}
              </div>
            ))}
          </div>
        )}

        {tabLoading && <div className="text-center py-8 text-gray-400">Loading...</div>}

        {!tabLoading && tabData && activeTab === 'logs' && (
          <pre className="text-xs text-gray-300 whitespace-pre-wrap max-h-96 overflow-y-auto font-mono">
            {Array.isArray(tabData) ? tabData.map(l => l.output).join('\n') : JSON.stringify(tabData, null, 2)}
          </pre>
        )}

        {!tabLoading && tabData && activeTab === 'stats' && (
          <div>
            {tabData.stats?.map((s, i) => (
              <div key={i} className="flex gap-6 py-2 border-b border-gray-700 last:border-0 text-sm">
                <span className="font-mono w-48">{s.container}</span>
                <span>CPU: {s.cpu}</span>
                <span>Memory: {s.memory}</span>
                <span>Net: {s.net_io}</span>
                <span>Block: {s.block_io}</span>
                <span>PIDs: {s.pids}</span>
              </div>
            ))}
          </div>
        )}

        {!tabLoading && tabData && activeTab === 'security' && (
          <div>
            {tabData.summary && (
              <div className="flex gap-4 mb-4">
                {Object.entries(tabData.summary).map(([sev, count]) => (
                  <span key={sev} className={`px-3 py-1 rounded text-sm ${
                    sev === 'high' || sev === 'critical' ? 'bg-red-900 text-red-200' :
                    sev === 'medium' ? 'bg-yellow-900 text-yellow-200' :
                    'bg-gray-700 text-gray-300'
                  }`}>{sev}: {count}</span>
                ))}
              </div>
            )}
            {tabData.findings?.map((f, i) => (
              <div key={i} className="py-2 border-b border-gray-700 last:border-0 text-sm">
                <span className={`px-2 py-0.5 rounded text-xs mr-2 ${
                  f.severity === 'high' || f.severity === 'critical' ? 'bg-red-900 text-red-200' :
                  f.severity === 'medium' ? 'bg-yellow-900 text-yellow-200' :
                  'bg-gray-700 text-gray-300'
                }`}>{f.severity}</span>
                <span className="text-gray-400 text-xs mr-2">[{f.category}]</span>
                {f.description}
              </div>
            ))}
            {(!tabData.findings || tabData.findings.length === 0) && (
              <p className="text-green-400">No security findings</p>
            )}
          </div>
        )}

        {!tabLoading && tabData && activeTab === 'backups' && (
          <div>
            {Array.isArray(tabData) && tabData.map((b, i) => (
              <div key={i} className="flex items-center gap-4 py-2 border-b border-gray-700 last:border-0 text-sm">
                <span className="font-mono">{b.id}</span>
                <span className="text-gray-400">{(b.size_bytes / 1024 / 1024).toFixed(1)} MB</span>
                <span className="text-gray-500">{new Date(b.created_at).toLocaleString()}</span>
              </div>
            ))}
            {(!tabData || tabData.length === 0) && <p className="text-gray-400">No backups found</p>}
          </div>
        )}

        {!tabLoading && tabData && activeTab === 'databases' && (
          <div>
            {tabData.checks?.map((c, i) => (
              <div key={i} className="flex items-center gap-4 py-2 border-b border-gray-700 last:border-0 text-sm">
                <span className={`w-2 h-2 rounded-full ${c.healthy ? 'bg-green-500' : 'bg-red-500'}`} />
                <span className="font-mono">{c.container}</span>
                <span className="text-gray-400">{c.engine}</span>
                <span className="text-gray-500">{c.healthy ? 'healthy' : 'unhealthy'}</span>
              </div>
            ))}
            {(!tabData.checks || tabData.checks.length === 0) && <p className="text-gray-400">No database containers found</p>}
          </div>
        )}

        {!tabLoading && tabData && activeTab === 'inspect' && (
          <pre className="text-xs text-gray-300 whitespace-pre-wrap max-h-96 overflow-y-auto font-mono">
            {JSON.stringify(tabData.inspections, null, 2)}
          </pre>
        )}

        {!tabLoading && tabData?.error && (
          <p className="text-red-400">Error: {tabData.error}</p>
        )}
      </div>
    </div>
  );
}
