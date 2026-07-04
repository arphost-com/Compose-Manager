import { useState } from 'react';

export default function Settings() {
  const [apiKey, setApiKey] = useState(localStorage.getItem('cm_api_key') || '');

  const save = () => {
    localStorage.setItem('cm_api_key', apiKey);
    window.location.reload();
  };

  return (
    <div className="max-w-md mx-auto mt-12">
      <div className="bg-gray-800 rounded-lg p-6 border border-gray-700">
        <h1 className="text-xl font-bold mb-4">Compose Manager</h1>
        <p className="text-gray-400 text-sm mb-4">Enter your API key to connect.</p>
        <label className="block text-sm text-gray-400 mb-1">API Key</label>
        <input
          type="password"
          value={apiKey}
          onChange={e => setApiKey(e.target.value)}
          placeholder="Enter API key"
          className="w-full bg-gray-700 border border-gray-600 rounded px-3 py-2 text-white mb-4 focus:outline-none focus:border-blue-500"
          onKeyDown={e => e.key === 'Enter' && save()}
        />
        <button onClick={save}
          className="w-full bg-blue-600 hover:bg-blue-500 text-white py-2 rounded">
          Connect
        </button>
        {localStorage.getItem('cm_api_key') && (
          <button onClick={() => { localStorage.removeItem('cm_api_key'); window.location.reload(); }}
            className="w-full mt-2 bg-gray-700 hover:bg-gray-600 text-gray-300 py-2 rounded">
            Disconnect
          </button>
        )}
      </div>
    </div>
  );
}
