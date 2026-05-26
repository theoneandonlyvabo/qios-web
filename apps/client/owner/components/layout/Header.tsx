'use client';

import { useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

// ==========================================
// HEADER — mobile top bar + hamburger drawer
// ==========================================

const menuItems = [
  { label: 'Dashboard', path: '/dashboard' },
  { label: 'Analytics', path: '/analytics' },
  { label: 'Transactions', path: '/history' },
  { label: 'Operators', path: '/operators' },
  { label: 'Products', path: '/products' },
];

export default function Header() {
  const [open, setOpen] = useState(false);
  const pathname = usePathname();

  return (
    <>
      {/* Mobile top bar */}
      <header className="md:hidden flex items-center justify-between p-4 sticky top-0 z-40 w-full border-b bg-[#161616] border-[#242424]">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-lg bg-[#CA400A] flex items-center justify-center font-black text-lg text-white">
            Q
          </div>
          <div>
            <span className="font-black text-lg tracking-tight block text-white">QIOS</span>
            <span className="text-[9px] font-bold uppercase tracking-wider block text-gray-400">
              Skalar Solutions
            </span>
          </div>
        </div>
        <button
          onClick={() => setOpen(true)}
          className="p-2.5 rounded-lg border border-[#242424] text-gray-400"
          aria-label="Open menu"
        >
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
      </header>

      {/* Mobile drawer */}
      {open && (
        <div className="md:hidden fixed inset-0 z-50 flex">
          <div className="fixed inset-0 bg-black/70" onClick={() => setOpen(false)} />
          <div className="relative w-4/5 max-w-sm h-full flex flex-col justify-between p-6 bg-[#161616] z-10">
            <div className="space-y-6">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-lg bg-[#CA400A] flex items-center justify-center font-bold text-xl text-white">
                  Q
                </div>
                <div>
                  <span className="font-black text-xl tracking-tight block text-white">QIOS</span>
                  <span className="text-[10px] font-bold block uppercase tracking-wider text-gray-400">
                    Skalar Solutions
                  </span>
                </div>
              </div>
              <nav className="space-y-1">
                {menuItems.map((item) => (
                  <Link
                    key={item.path}
                    href={item.path}
                    onClick={() => setOpen(false)}
                    className={`w-full flex items-center px-3 py-3 rounded-lg text-sm font-semibold transition-all ${
                      pathname.startsWith(item.path)
                        ? 'bg-[#212121] text-[#CA400A] border-l-4 border-[#CA400A]'
                        : 'text-gray-400 hover:text-white hover:bg-[#1A1A1A]'
                    }`}
                  >
                    {item.label}
                  </Link>
                ))}
              </nav>
            </div>
            <div className="pt-4 border-t border-[#242424]">
              <button
                onClick={() => {
                  localStorage.removeItem('qios_access_token');
                  window.location.href = process.env.NEXT_PUBLIC_AUTH_URL ?? '/';
                }}
                className="w-full bg-[#CA400A] text-white text-xs font-bold py-3 px-4 rounded-lg"
              >
                LOGOUT
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
