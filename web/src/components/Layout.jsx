import { Link, Outlet, useLocation } from 'react-router-dom';

export default function Layout() {
  const location = useLocation();
  const isActive = (path) => location.pathname === path ? 'bg-gray-100 text-gray-950' : 'text-gray-600';

  return (
    <div className="min-h-screen bg-gray-50 text-gray-950">
      <nav className="border-b border-gray-200 bg-white px-6 py-3">
        <div className="flex items-center gap-6">
          <Link to="/" className="text-lg font-semibold text-blue-700">Compose Manager</Link>
          <Link to="/" className={`rounded-md px-3 py-1.5 text-sm hover:bg-gray-100 ${isActive('/')}`}>Dashboard</Link>
          <Link to="/settings" className={`rounded-md px-3 py-1.5 text-sm hover:bg-gray-100 ${isActive('/settings')}`}>Settings</Link>
        </div>
      </nav>
      <main className="max-w-7xl mx-auto px-6 py-6">
        <Outlet />
      </main>
    </div>
  );
}
