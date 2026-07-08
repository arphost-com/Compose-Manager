import { useState } from 'react';
import { auth, totp } from '../api/client';

export default function Login() {
  const [form, setForm] = useState({ username: '', password: '' });
  const [totpStep, setTotpStep] = useState(false);
  const [totpToken, setTotpToken] = useState('');
  const [totpCode, setTotpCode] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const submit = async (event) => {
    event.preventDefault();
    setLoading(true);
    setError('');
    try {
      const res = await auth.login(form);
      if (res.data.totp_required) {
        setTotpToken(res.data.totp_token);
        setTotpStep(true);
        setLoading(false);
        return;
      }
      localStorage.setItem('cm_token', res.data.token);
      localStorage.setItem('cm_user', JSON.stringify(res.data.user));
      localStorage.removeItem('cm_api_key');
      window.location.href = '/';
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const submitTotp = async (event) => {
    event.preventDefault();
    setLoading(true);
    setError('');
    try {
      const res = await totp.login(totpToken, totpCode);
      localStorage.setItem('cm_token', res.data.token);
      localStorage.setItem('cm_user', JSON.stringify(res.data.user));
      localStorage.removeItem('cm_api_key');
      window.location.href = '/';
    } catch (err) {
      setError(err.message);
      if (err.message?.includes('expired')) {
        setTotpStep(false);
        setTotpToken('');
        setTotpCode('');
      }
    } finally {
      setLoading(false);
    }
  };

  const resetTotp = () => {
    setTotpStep(false);
    setTotpToken('');
    setTotpCode('');
    setError('');
  };

  return (
    <div className="min-h-screen bg-gray-50 px-6 py-12 text-gray-950">
      {!totpStep ? (
        <form onSubmit={submit} className="mx-auto mt-12 max-w-md section-panel space-y-4">
          <div>
            <h1 className="text-xl font-semibold">Stack Manager</h1>
            <p className="mt-1 text-sm text-gray-600">Sign in with your username and password.</p>
          </div>
          {error && <div className="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800">{error}</div>}
          <label className="block text-sm">
            <span className="mb-1 block font-medium text-gray-700">Username</span>
            <input value={form.username} onChange={e => setForm({ ...form, username: e.target.value })} className="input" autoComplete="username" autoFocus />
          </label>
          <label className="block text-sm">
            <span className="mb-1 block font-medium text-gray-700">Password</span>
            <input type="password" value={form.password} onChange={e => setForm({ ...form, password: e.target.value })} className="input" autoComplete="current-password" />
          </label>
          <button disabled={loading} className="btn-primary w-full">{loading ? 'Signing in...' : 'Sign In'}</button>
        </form>
      ) : (
        <form onSubmit={submitTotp} className="mx-auto mt-12 max-w-md section-panel space-y-4">
          <div>
            <h1 className="text-xl font-semibold">Two-Factor Authentication</h1>
            <p className="mt-1 text-sm text-gray-600">Enter the 6-digit code from your authenticator app, or a backup code.</p>
          </div>
          {error && <div className="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800">{error}</div>}
          <label className="block text-sm">
            <span className="mb-1 block font-medium text-gray-700">Code</span>
            <input
              value={totpCode}
              onChange={e => setTotpCode(e.target.value.replace(/[^0-9a-fA-F]/g, ''))}
              className="input text-center text-2xl tracking-widest"
              maxLength={8}
              autoComplete="one-time-code"
              inputMode="numeric"
              autoFocus
              placeholder="000000"
            />
          </label>
          <button disabled={loading} className="btn-primary w-full">{loading ? 'Verifying...' : 'Verify'}</button>
          <button type="button" onClick={resetTotp} className="btn-secondary w-full">Back to login</button>
        </form>
      )}
    </div>
  );
}
