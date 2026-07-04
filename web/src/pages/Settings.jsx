import { useState } from 'react';

export default function Settings() {
  const [apiKey, setApiKey] = useState(localStorage.getItem('cm_api_key') || '');

  const save = () => {
    localStorage.setItem('cm_api_key', apiKey);
    window.location.reload();
  };

  return (
    <div className="max-w-md mx-auto mt-12">
      <div className="section-panel">
        <h1 className="mb-4 text-xl font-semibold text-gray-950">API Access</h1>
        <p className="mb-4 text-sm text-gray-600">Enter the API key configured on the Compose Manager server.</p>
        <label className="mb-1 block text-sm font-medium text-gray-700" title="Stored in this browser's local storage and sent as X-API-Key.">API Key</label>
        <input
          type="password"
          value={apiKey}
          onChange={e => setApiKey(e.target.value)}
          placeholder="Enter API key"
          className="input mb-4"
          onKeyDown={e => e.key === 'Enter' && save()}
        />
        <button onClick={save}
          title="Save this API key and reload the app."
          className="btn-primary w-full">
          Connect
        </button>
        {localStorage.getItem('cm_api_key') && (
          <button onClick={() => { localStorage.removeItem('cm_api_key'); window.location.reload(); }}
            title="Remove the stored API key from this browser."
            className="btn-secondary mt-2 w-full">
            Disconnect
          </button>
        )}
      </div>
    </div>
  );
}
