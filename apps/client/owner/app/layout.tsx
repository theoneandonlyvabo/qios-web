'use client';

import React, { useState, useMemo, useEffect } from 'react';
import { usePathname, useRouter } from 'next/navigation';
import './globals.css';

// ==========================================
// DATA MOCK TERINTEGRASI UNTUK PREVIEW MANDIRI
// ==========================================

const INITIAL_PRODUCTS = [
  {
    id: 'prod-1',
    name: 'Kopi Susu Gula Aren',
    price: 22000,
    category: 'Kopi',
    description: 'Espresso blend khas Senja dengan susu segar UHT dan manisnya gula aren alami asli Sukabumi.',
    recipe: [
      { name: 'Susu UHT', qty: 150, unit: 'ml' },
      { name: 'Espresso', qty: 30, unit: 'ml' },
      { name: 'Gula Aren', qty: 20, unit: 'ml' }
    ],
    is_available: true,
    total_sold: 139
  },
  {
    id: 'prod-2',
    name: 'Americano',
    price: 18000,
    category: 'Kopi',
    description: 'Double shot espresso dari biji kopi arabika pilihan dengan air mineral berkualitas tinggi.',
    recipe: [
      { name: 'Espresso', qty: 60, unit: 'ml' }
    ],
    is_available: true,
    total_sold: 95
  },
  {
    id: 'prod-3',
    name: 'Matcha Latte',
    price: 25000,
    category: 'Non-Kopi',
    description: 'Bubuk matcha murni impor Jepang dengan susu UHT creamy dan sedikit sentuhan vanilla syrup.',
    recipe: [
      { name: 'Bubuk Matcha', qty: 15, unit: 'g' },
      { name: 'Susu UHT', qty: 180, unit: 'ml' }
    ],
    is_available: true,
    total_sold: 74
  },
  {
    id: 'prod-4',
    name: 'Croissant Butter',
    price: 28000,
    category: 'Makanan',
    description: 'Pastry mentega berlapis klasik, renyah di luar dan lembut berongga di bagian dalam.',
    recipe: [
      { name: 'Croissant Mentah', qty: 1, unit: 'pcs' }
    ],
    is_available: true,
    total_sold: 48
  },
  {
    id: 'prod-5',
    name: 'Cafe Latte',
    price: 24000,
    category: 'Kopi',
    description: 'Espresso shot seimbang dipadukan dengan foam susu halus tebal berstandar latte art.',
    recipe: [
      { name: 'Espresso', qty: 30, unit: 'ml' },
      { name: 'Susu UHT', qty: 160, unit: 'ml' }
    ],
    is_available: true,
    total_sold: 40
  }
];

const formatRupiah = (num: number) => {
  if (num === undefined || num === null) return 'Rp 0';
  return 'Rp ' + num.toLocaleString('id-ID');
};

export default function RootLayout({
  children,
}: {
  children?: React.ReactNode;
}) {
  const [isDarkMode, setIsDarkMode] = useState(true);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

  const pathname = usePathname();
  const router = useRouter();

  const userEmail = 'alayavaro@skalar.id';

  // Load theme preference
  useEffect(() => {
    const saved = localStorage.getItem('qios-owner-theme');
    if (saved !== null) setIsDarkMode(saved === 'dark');
  }, []);

  // Sync data-theme attribute + persist
  useEffect(() => {
    document.documentElement.dataset.theme = isDarkMode ? 'dark' : 'light';
    localStorage.setItem('qios-owner-theme', isDarkMode ? 'dark' : 'light');
  }, [isDarkMode]);

  // Carousel Teks Berjalan
  const [activeIndex, setActiveIndex] = useState(0);
  useEffect(() => {
    const interval = setInterval(() => {
      setActiveIndex((prev) => (prev === 0 ? 1 : 0));
    }, 4500);
    return () => clearInterval(interval);
  }, []);

  const menuItems = [
    {
      key: 'dashboard',
      label: 'Dashboard',
      path: '/dashboard',
      icon: (
        <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6a2 2 0 012-2h2a2 2 0 012 2v4a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v4a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v4a2 2 0 01-2 2H6a2 2 0 01-2-2v-4zM14 16a2 2 0 012-2h2a2 2 0 012 2v4a2 2 0 01-2 2h-2a2 2 0 01-2-2v-4z" />
        </svg>
      ),
    },
    {
      key: 'statistics',
      label: 'Statistik Penjualan',
      path: '/statistics',
      icon: (
        <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
        </svg>
      ),
    },
    {
      key: 'analytics',
      label: 'AI Analytics',
      path: '/analytics',
      icon: (
        <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
      ),
      badge: true,
    },
    {
      key: 'reports',
      label: 'Laporan Finansial',
      path: '/reports',
      icon: (
        <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
      ),
    },
    {
      key: 'history',
      label: 'Riwayat Transaksi',
      path: '/history',
      icon: (
        <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      ),
    },
    {
      key: 'operators',
      label: 'Kelola Operator',
      path: '/operators',
      icon: (
        <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
        </svg>
      ),
    },
    {
      key: 'products',
      label: 'Menu & Produk',
      path: '/products',
      icon: (
        <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
        </svg>
      ),
    },
  ];

  const handleNavigate = (path: string) => {
    router.push(path);
    setIsMobileMenuOpen(false);
  };

  const activeTab = useMemo(() => {
    if (!pathname) return 'dashboard';
    const currentRoute = pathname.replace(/^\//, '');
    const match = menuItems.find(item => item.key === currentRoute);
    return match ? currentRoute : 'dashboard';
  }, [pathname]); // eslint-disable-line react-hooks/exhaustive-deps

  const [activeProducts] = useState(INITIAL_PRODUCTS);
  const [activeCategory, setActiveCategory] = useState('SEMUA');
  const categories = useMemo(() => {
    return ['SEMUA', ...Array.from(new Set(activeProducts.map(p => p.category)))];
  }, [activeProducts]);
  const filteredProducts = useMemo(() => {
    if (activeCategory === 'SEMUA') return activeProducts;
    return activeProducts.filter(p => p.category === activeCategory);
  }, [activeCategory, activeProducts]);

  const renderFallbackContent = () => {
    if (activeTab === 'products') {
      return (
        <div className="space-y-6 animate-fadeIn">
          <div className={`p-4 md:p-6 rounded-xl border flex items-start gap-4 ${isDarkMode ? 'bg-[#1C110C] border-[#CA400A]/30' : 'bg-orange-50/60 border-orange-200'}`}>
            <div className="w-12 h-12 rounded-lg bg-[#CA400A]/10 text-[#CA400A] flex items-center justify-center shrink-0 border border-[#CA400A]/20">
              <svg className="w-6 h-6 text-[#CA400A]" fill="none" stroke="currentColor" strokeWidth="2.5" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" /></svg>
            </div>
            <div>
              <h3 className="font-bold text-sm md:text-base uppercase tracking-wider qios-text-title">Katalog Produk & Resep Read-Only</h3>
              <p className="text-xs md:text-sm mt-1 qios-text-muted">Perubahan harga atau revisi komposisi resep wajib diserahkan kepada tim Skalar Solutions luring.</p>
            </div>
          </div>

          <div className="flex flex-wrap items-center gap-2 pb-2">
            {categories.map((cat) => (
              <button
                key={cat}
                onClick={() => setActiveCategory(cat)}
                className={`px-4 py-2.5 text-xs font-bold rounded-lg uppercase tracking-wider border transition-all ${
                  activeCategory === cat
                    ? 'bg-[#CA400A] text-white border-transparent shadow-md'
                    : 'qios-card qios-text-muted hover:qios-text-title'
                }`}
              >
                {cat === 'SEMUA' ? 'Semua Menu' : cat}
              </button>
            ))}
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {filteredProducts.map((p, idx) => (
              <div key={idx} className="rounded-xl border p-4 md:p-6 flex flex-col justify-between hover:border-[#CA400A]/20 transition-all qios-card">
                <div>
                  <div className="flex justify-between items-start mb-2">
                    <span className="font-black text-sm md:text-md qios-text-title">{p.name}</span>
                    <span className="text-[10px] px-2 py-0.5 rounded font-extrabold uppercase border qios-badge">{p.category}</span>
                  </div>
                  <span className="text-sm md:text-md font-bold text-[#CA400A] block mb-3">{formatRupiah(p.price)}</span>
                  <p className="text-xs leading-relaxed mb-4 qios-text-muted">{p.description}</p>
                </div>
                <div className="pt-4 border-t space-y-2 qios-divider">
                  <span className="text-[10px] font-bold text-gray-500 uppercase block">Resep Komposisi:</span>
                  <div className="flex flex-wrap gap-1.5">
                    {p.recipe.map((ing, ingIdx) => (
                      <span key={ingIdx} className="text-xs px-2 py-1 rounded font-semibold border qios-badge">{ing.name}: {ing.qty} {ing.unit}</span>
                    ))}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      );
    }

    return (
      <div className="flex flex-col items-center justify-center min-h-[50vh] text-center p-6 space-y-4">
        <div className="w-12 h-12 rounded-2xl bg-[#CA400A] flex items-center justify-center font-black text-2xl text-white">Q</div>
        <div className="space-y-1">
          <h3 className="font-bold text-lg qios-text-title">Modul QIOS Owner Aktif</h3>
          <p className="text-sm max-w-sm mx-auto qios-text-muted">
            Gunakan sidebar di sebelah kiri untuk berpindah rute operasional. Tampilan menu dan interaksi sudah terpasang rapi di Next.js lokal Anda.
          </p>
        </div>
      </div>
    );
  };

  const isSplash = pathname === '/';

  const layoutContent = (
    <div className="flex flex-col md:flex-row min-h-screen w-full transition-colors duration-200 overflow-x-hidden">

      {/* MOBILE HEADER */}
      {!isSplash && (
      <header className="md:hidden flex items-center justify-between p-4 sticky top-0 z-40 w-full border-b transition-colors duration-200 qios-sidebar">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-lg bg-[#CA400A] flex items-center justify-center font-black text-lg text-white shadow-md">Q</div>
          <div>
            <span className="font-black text-lg tracking-tight block qios-text-title">QIOS</span>
            <span className="text-[9px] font-bold uppercase tracking-wider block qios-text-muted">Skalar Solutions</span>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setIsDarkMode(!isDarkMode)}
            className={`p-2.5 rounded-lg border transition-all ${isDarkMode ? 'text-yellow-500' : 'text-indigo-900'}`}
          >
            {isDarkMode ? (
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364-6.364l-.707.707M6.343 17.657l-.707.707m0-12.728l.707.707m12.728 12.728l.707-.707M12 8a4 4 0 100 8 4 4 0 000-8z" /></svg>
            ) : (
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" /></svg>
            )}
          </button>
          <button
            onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
            className="p-2.5 rounded-lg border transition-all"
          >
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" /></svg>
          </button>
        </div>
      </header>
      )}

      {/* MOBILE DRAWER */}
      {!isSplash && isMobileMenuOpen && (
        <div className="md:hidden fixed inset-0 z-50 flex animate-fadeIn">
          <div className="fixed inset-0 bg-black/70" onClick={() => setIsMobileMenuOpen(false)}></div>
          <div className="relative w-4/5 max-w-sm h-full flex flex-col justify-between p-6 transition-colors duration-200 z-10 qios-sidebar">
            <div className="space-y-6">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-lg bg-[#CA400A] flex items-center justify-center font-bold text-xl text-white">Q</div>
                <div>
                  <span className="font-black text-xl tracking-tight block qios-text-title">QIOS</span>
                  <span className="text-[10px] font-bold block uppercase tracking-wider qios-text-muted">Skalar Solutions</span>
                </div>
              </div>
              <nav className="space-y-1">
                {menuItems.map(item => (
                  <button
                    key={item.key}
                    onClick={() => handleNavigate(item.path)}
                    className={`w-full flex items-center justify-between px-3 py-3 rounded-lg text-sm font-semibold transition-all ${pathname.startsWith(item.path) ? 'qios-nav-active' : 'qios-nav-item'}`}
                  >
                    <div className="flex items-center gap-3">
                      {item.icon}
                      <span>{item.label}</span>
                    </div>
                  </button>
                ))}
              </nav>
            </div>
            <div className="pt-4 border-t qios-divider">
              <div className="flex items-center gap-3 truncate mb-4">
                <div className="w-10 h-10 rounded-full bg-[#CA400A]/20 text-[#CA400A] font-bold flex items-center justify-center border border-[#CA400A]/30">OW</div>
                <div className="truncate">
                  <span className="font-bold text-sm block truncate qios-text-title">Alayavaro Rachmadia</span>
                  <span className="text-xs block truncate qios-text-muted">{userEmail}</span>
                </div>
              </div>
              <button onClick={() => router.push('/login')} className="w-full bg-[#CA400A] text-white text-xs font-bold py-3 px-4 rounded-lg">LOGOUT</button>
            </div>
          </div>
        </div>
      )}

      {/* DESKTOP SIDEBAR */}
      {!isSplash && (
      <aside className="hidden md:flex w-64 flex-col justify-between border-r shrink-0 transition-colors duration-200 z-10 qios-sidebar">
        <div>
          <div className="p-6 flex items-center gap-3">
            <div className="w-10 h-10 rounded-lg bg-[#CA400A] flex items-center justify-center font-bold text-xl text-white shadow-lg shadow-[#CA400A]/20">Q</div>
            <div>
              <span className="font-extrabold text-xl tracking-tight block qios-text-title">QIOS</span>
              {/* VERTICAL SMOOTH CAROUSEL */}
              <div className="relative h-[16px] overflow-hidden select-none">
                <div
                  className="transition-transform duration-800"
                  style={{ transform: `translateY(-${activeIndex * 16}px)` }}
                >
                  <div className="text-[10px] font-bold uppercase tracking-wider h-[16px] flex items-center qios-text-muted">Skalar Solutions</div>
                  <div className="text-[10px] font-bold uppercase tracking-wider h-[16px] flex items-center qios-text-muted">Kendali Bisnis Anda</div>
                </div>
              </div>
            </div>
          </div>

          <div className="p-4 mx-4 my-1 rounded-xl relative overflow-hidden transition-colors duration-200 qios-sidebar-profile">
            <span className="text-[11px] font-bold text-gray-500 uppercase block">Merchant</span>
            <span className="font-bold text-md block truncate qios-text-title">Warung Kopi Senja</span>
            <span className="inline-block mt-2 px-2 py-0.5 rounded text-[10px] bg-[#CA400A] text-white font-bold uppercase tracking-wider">SKALAR MAX</span>
          </div>

          <nav className="px-4 py-4 space-y-1">
            {menuItems.map(item => (
              <button
                key={item.key}
                onClick={() => handleNavigate(item.path)}
                className={`w-full flex items-center justify-between px-3 py-2.5 rounded-lg text-sm font-semibold transition-all duration-200 ${pathname.startsWith(item.path) ? 'qios-nav-active' : 'qios-nav-item'}`}
              >
                <div className="flex items-center gap-3">
                  {item.icon}
                  <span>{item.label}</span>
                </div>
              </button>
            ))}
          </nav>
        </div>

        <div className="p-4 border-t flex items-center justify-between qios-divider">
          <div className="flex items-center gap-3 truncate">
            <div className="w-10 h-10 rounded-full bg-[#CA400A]/20 text-[#CA400A] font-bold flex items-center justify-center border border-[#CA400A]/30">OW</div>
            <div className="truncate">
              <span className="font-bold text-sm block truncate qios-text-title">Alayavaro Rachmadia</span>
              <span className="text-xs block truncate qios-text-muted">{userEmail}</span>
            </div>
          </div>
          <button
            onClick={() => setIsDarkMode(!isDarkMode)}
            className={`p-2 rounded-lg border transition-all ${isDarkMode ? 'text-yellow-500 border-transparent hover:bg-[var(--surface)]' : 'text-[#CA400A] border-transparent hover:bg-[var(--surface)]'}`}
          >
            {isDarkMode ? (
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364-6.364l-.707.707M6.343 17.657l-.707.707m0-12.728l.707.707m12.728 12.728l.707-.707M12 8a4 4 0 100 8 4 4 0 000-8z" /></svg>
            ) : (
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" /></svg>
            )}
          </button>
        </div>
      </aside>
      )}

      {/* AREA UTAMA */}
      <main className="flex-1 overflow-y-auto min-h-screen">
        {children ? children : renderFallbackContent()}
      </main>
    </div>
  );

  return (
    <html lang="id" data-theme={isDarkMode ? 'dark' : 'light'} suppressHydrationWarning>
      <body suppressHydrationWarning>
        {layoutContent}
      </body>
    </html>
  );
}
