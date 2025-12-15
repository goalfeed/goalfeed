import React, { useEffect, useState } from 'react';
import { apiClient } from '../utils/api';

type HAStatus = {
  connected: boolean;
  source: 'env' | 'config' | 'unset' | string;
  message: string;
  url: string;
  tokenSet: boolean;
};

const HASettings: React.FC = () => {
  const [status, setStatus] = useState<HAStatus | null>(null);
  // Removed unused configuredUrl state to satisfy ESLint
  const [configuredTokenSet, setConfiguredTokenSet] = useState(false);
  const [urlInput, setUrlInput] = useState('');
  const [tokenInput, setTokenInput] = useState('');
  const [clearToken, setClearToken] = useState(false);
  const [saving, setSaving] = useState(false);

  const load = async () => {
    const [cfg, st] = await Promise.all([
      apiClient.get('/api/homeassistant/config'),
      apiClient.get('/api/homeassistant/status'),
    ]);
    if (cfg.data.success) {
      setConfiguredTokenSet(!!cfg.data.data.configured.tokenSet);
      setUrlInput(cfg.data.data.configured.url || '');
    }
    if (st.data.success) {
      setStatus(st.data.data);
    }
  };

  useEffect(() => { load(); }, []);

  const save = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    try {
      await apiClient.post('/api/homeassistant/config', {
        url: urlInput,
        accessToken: clearToken ? '' : tokenInput,
        clearToken,
      });
      await load();
      setTokenInput('');
      setClearToken(false);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="p-6">
      <div className="border-b border-white/10 pb-4 mb-6">
        <h2 className="text-xl font-semibold text-white">Home Assistant Integration</h2>
        <p className="text-slate-300 text-sm">Configure URL and token, and view connection status.</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-black/30 border border-white/10 rounded-xl p-5">
          <h3 className="text-white font-medium mb-4">Connection Status</h3>
          {status ? (
            <div className="space-y-2 text-sm">
              <div className="flex items-center space-x-2">
                <span className={`w-3 h-3 rounded-full ${status.connected ? 'bg-green-400' : 'bg-red-400'}`}></span>
                <span className="text-slate-200">{status.connected ? 'Connected' : 'Not Connected'}</span>
              </div>
              <div className="text-slate-300">URL: <span className="text-slate-200">{status.url || '—'}</span></div>
              <div className="text-slate-300">Token set: <span className="text-slate-200">{status.tokenSet ? 'Yes' : 'No'}</span></div>
              <div className="text-slate-300">Source: <span className="text-slate-200">{status.source}</span></div>
              {!status.connected && status.message && (
                <div className="text-slate-400">Reason: {status.message}</div>
              )}
              <button onClick={load} className="mt-3 px-3 py-1.5 bg-slate-700 hover:bg-slate-600 text-white rounded-md">Re-check</button>
            </div>
          ) : (
            <div className="text-slate-400">Loading…</div>
          )}
        </div>

        <form onSubmit={save} className="bg-black/30 border border-white/10 rounded-xl p-5">
          <h3 className="text-white font-medium mb-4">Configuration</h3>
          <div className="space-y-4">
            <div>
              <label className="block text-slate-300 text-sm mb-1">Home Assistant URL</label>
              <input className="w-full bg-black/40 border border-white/10 rounded-md px-3 py-2 text-white"
                     value={urlInput}
                     onChange={(e) => setUrlInput(e.target.value)}
                     placeholder="http://homeassistant.local:8123" />
            </div>
            <div>
              <label className="block text-slate-300 text-sm mb-1">Access Token</label>
              <input className="w-full bg-black/40 border border-white/10 rounded-md px-3 py-2 text-white"
                     value={tokenInput}
                     onChange={(e) => setTokenInput(e.target.value)}
                     placeholder={configuredTokenSet ? '•••••••• (configured)' : 'Long-Lived Access Token'}
                     type="password" />
              <div className="mt-2 flex items-center space-x-2">
                <input id="clearToken" type="checkbox" checked={clearToken} onChange={e => setClearToken(e.target.checked)} />
                <label htmlFor="clearToken" className="text-slate-300 text-sm">Clear stored token</label>
              </div>
            </div>
            <div className="pt-2">
              <button disabled={saving} className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white rounded-md">
                {saving ? 'Saving…' : 'Save Configuration'}
              </button>
            </div>
          </div>

          <p className="text-xs text-slate-400 mt-4">If running as a Home Assistant add-on, Supervisor-provided URL/token will override stored config.</p>
        </form>
      </div>
    </div>
  );
};

export default HASettings;



