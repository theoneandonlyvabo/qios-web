'use client';

import { useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';

// Splash screen — tampil setiap kali buka app, redirect ke dashboard setelah 2 detik
export default function SplashPage() {
  const router = useRouter();
  const redirected = useRef(false);

  useEffect(() => {
    if (redirected.current) return;
    redirected.current = true;

    const timer = setTimeout(() => {
      router.replace('/dashboard');
    }, 2000);

    return () => clearTimeout(timer);
  }, [router]);

  return (
    <div className="fixed inset-0 flex flex-col items-center justify-center bg-[#0F0F0F] text-white">
      <div className="flex flex-col items-center gap-6 text-center max-w-xs px-4">

        {/* Logo */}
        <div className="relative flex items-center justify-center">
          <div className="w-20 h-20 rounded-[22px] bg-[#CA400A] flex items-center justify-center shadow-2xl shadow-[#CA400A]/40">
            <svg width="36" height="36" viewBox="0 0 36 36" fill="none">
              <rect x="4" y="4" width="12" height="12" rx="2.5" fill="white" fillOpacity="0.9"/>
              <rect x="20" y="4" width="12" height="12" rx="2.5" fill="white" fillOpacity="0.9"/>
              <rect x="4" y="20" width="12" height="12" rx="2.5" fill="white" fillOpacity="0.9"/>
              <rect x="20" y="20" width="12" height="12" rx="2.5" fill="white" fillOpacity="0.4"/>
            </svg>
          </div>
          <div className="absolute w-28 h-28 rounded-[30px] border border-[#CA400A]/20 pointer-events-none" />
        </div>

        {/* Teks */}
        <div className="space-y-2">
          <h2 className="text-xl font-black tracking-tight text-white">Data Sedang Diambil</h2>
          <p className="text-gray-400 text-xs leading-relaxed">
            Mohon tunggu, Skalar Solutions sedang mengambil data
            transaksi, operator, dan resep bumbu Warung Kopi Senja...
          </p>
        </div>

        {/* Spinner */}
        <div className="w-7 h-7 border-2 border-[#CA400A]/25 border-t-[#CA400A] rounded-full animate-spin" />

      </div>
    </div>
  );
}
