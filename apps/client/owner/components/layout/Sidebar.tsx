'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';

// ==========================================
// SIDEBAR — navigasi utama owner dashboard
// ==========================================

const menuItems = [
  {
    key: 'dashboard',
    label: 'Dashboard',
    path: '/dashboard',
    icon: (
      <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
          d="M4 6a2 2 0 012-2h2a2 2 0 012 2v4a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v4a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v4a2 2 0 01-2 2H6a2 2 0 01-2-2v-4zM14 16a2 2 0 012-2h2a2 2 0 012 2v4a2 2 0 01-2 2h-2a2 2 0 01-2-2v-4z" />
      </svg>
    ),
  },
  {
    key: 'analytics',
    label: 'Analytics',
    path: '/analytics',
    icon: (
      <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
          d="M13 10V3L4 14h7v7l9-11h-7z" />
      </svg>
    ),
  },
  {
    key: 'transactions',
    label: 'Transactions',
    path: '/history',
    icon: (
      <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
          d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
    ),
  },
  {
    key: 'operators',
    label: 'Operators',
    path: '/operators',
    icon: (
      <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
          d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
      </svg>
    ),
  },
  {
    key: 'products',
    label: 'Products',
    path: '/products',
    icon: (
      <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
          d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
      </svg>
    ),
  },
];

export default function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="hidden md:flex w-64 flex-col justify-between border-r shrink-0 bg-[#161616] border-[#242424]">
      <div>
        {/* Logo */}
        <div className="p-6 flex items-center gap-3">
          <div className="w-10 h-10 rounded-lg bg-[#CA400A] flex items-center justify-center font-bold text-xl text-white shadow-lg">
            Q
          </div>
          <div>
            <span className="font-extrabold text-xl tracking-tight block text-white">QIOS</span>
            <span className="text-[10px] font-bold uppercase tracking-wider text-gray-400">
              Skalar Solutions
            </span>
          </div>
        </div>

        {/* Nav */}
        <nav className="px-4 py-2 space-y-1">
          {menuItems.map((item) => {
            const isActive = pathname.startsWith(item.path);
            return (
              <Link
                key={item.key}
                href={item.path}
                className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-semibold transition-all duration-200 ${
                  isActive
                    ? 'bg-[#212121] text-[#CA400A] border-l-4 border-[#CA400A]'
                    : 'text-gray-400 hover:text-white hover:bg-[#1A1A1A]'
                }`}
              >
                {item.icon}
                <span>{item.label}</span>
              </Link>
            );
          })}
        </nav>
      </div>

      {/* User info + logout */}
      <div className="p-4 border-t border-[#242424] flex items-center justify-between">
        <div className="flex items-center gap-3 truncate">
          <div className="w-9 h-9 rounded-full bg-[#CA400A]/20 text-[#CA400A] font-bold flex items-center justify-center border border-[#CA400A]/30 text-sm shrink-0">
            OW
          </div>
          <div className="truncate">
            <span className="font-bold text-sm block truncate text-white">Owner</span>
            <span className="text-xs block truncate text-gray-400">QIOS Dashboard</span>
          </div>
        </div>
        <button
          onClick={() => {
            // Logout: clear token, redirect ke login (Dev 3)
            localStorage.removeItem('qios_access_token');
            window.location.href = process.env.NEXT_PUBLIC_AUTH_URL ?? '/';
          }}
          className="p-2 rounded-lg text-gray-400 hover:text-white hover:bg-[#242424] transition-all"
          title="Logout"
        >
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
              d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
          </svg>
        </button>
      </div>
    </aside>
  );
}
