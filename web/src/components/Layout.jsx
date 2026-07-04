import { Link, Outlet, useLocation } from 'react-router-dom';

export default function Layout() {
  const location = useLocation();
  const isActive = (path) => location.pathname === path ? 'bg-gray-700' : '';

  return (
    <div className="min-h-screen bg-gray-900 text-gray-100">
      <nav className="bg-gray-800 border-b border-gray-700 px-6 py-3 flex items-center justify-between">
        <div className="flex items-center gap-6">
          <Link to="/" className="text-lg font-bold text-blue-400">Compose Manager</Link>
          <Link to="/" className={`px-3 py-1 rounded text-sm hover:bg-gray-700 ${isActive('/')}`}>Dashboard</Link>
          <Link to="/settings" className={`px-3 py-1 rounded text-sm hover:bg-gray-700 ${isActive('/settings')}`}>Settings</Link>
        </div>
      </nav>
      <main className="max-w-7xl mx-auto px-6 py-6">
        <Outlet />
      </main>
    </div>
  );
}
