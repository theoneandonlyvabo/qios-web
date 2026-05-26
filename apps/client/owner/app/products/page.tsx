'use client';

import React, { useState, useMemo } from 'react';
import { INITIAL_PRODUCTS } from '../../lib/mockData';

export default function ProductsPage() {
  const [activeCategory, setActiveCategory] = useState<string>('SEMUA');

  // Mengambil daftar kategori unik secara dinamis dari mock data produk
  const categories = useMemo(() => {
    const list = new Set(INITIAL_PRODUCTS.map(p => p.category));
    return ['SEMUA', ...Array.from(list)];
  }, []);

  // Menyaring produk berdasarkan kategori yang dipilih
  const filteredProducts = useMemo(() => {
    if (activeCategory === 'SEMUA') return INITIAL_PRODUCTS;
    return INITIAL_PRODUCTS.filter(p => p.category === activeCategory);
  }, [activeCategory]);

  const formatRupiah = (num: number) => {
    return 'Rp ' + num.toLocaleString('id-ID');
  };

  return (
    <div className="p-4 md:p-8 space-y-6 animate-fadeIn text-gray-900 dark:text-white">
      
      {/* BANNER REKAYASA GUARD (Owner Read-Only Notice) */}
      <div className="p-5 md:p-6 rounded-xl border border-orange-200 dark:border-[#CA400A]/30 bg-orange-50/60 dark:bg-[#1C110C] shadow-sm">
        <div className="flex items-start gap-4">
          <div className="w-12 h-12 rounded-lg bg-[#CA400A]/10 text-[#CA400A] flex items-center justify-center shrink-0 border border-[#CA400A]/20">
            <svg className="w-6 h-6 text-[#CA400A]" fill="none" stroke="currentColor" strokeWidth="2.5" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
          </div>
          <div className="space-y-1">
            <h3 className="font-extrabold text-sm md:text-base uppercase tracking-wider text-gray-800 dark:text-white">
              Katalog Produk & Resep Read-Only
            </h3>
            <p className="text-xs md:text-sm text-gray-500 dark:text-gray-400 leading-relaxed">
              Untuk menjaga kestabilan data historis pembukuan keuangan dan perhitungan log konsumsi persediaan bahan baku, penyesuaian harga, penambahan menu baru, atau revisi komposisi resep wajib diserahkan kepada tim Skalar Solutions. Silakan hubungi Account Manager Anda untuk perubahan data secara luring.
            </p>
          </div>
        </div>
      </div>

      {/* FILTER KATEGORI MENU */}
      <div className="flex flex-wrap items-center gap-2 pb-2">
        {categories.map((cat) => (
          <button
            key={cat}
            onClick={() => setActiveCategory(cat)}
            className={`px-4 py-2.5 text-xs font-bold rounded-lg uppercase tracking-wider border transition-all ${
              activeCategory === cat
                ? 'bg-[#CA400A] text-white border-transparent shadow-md'
                : 'bg-white dark:bg-[#161616] border-gray-200 dark:border-[#242424] text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white shadow-sm'
            }`}
          >
            {cat === 'SEMUA' ? 'Semua Menu' : cat}
          </button>
        ))}
      </div>

      {/* GRID DAFTAR PRODUK DAN DETAIL RESEP */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredProducts.map((p) => (
          <div 
            key={p.id} 
            className="rounded-xl border border-gray-200 dark:border-[#242424] bg-white dark:bg-[#161616] p-5 md:p-6 flex flex-col justify-between hover:border-[#CA400A]/20 transition-all shadow-sm group"
          >
            <div className="space-y-3">
              <div className="flex justify-between items-start gap-2">
                <span className="font-black text-sm md:text-base text-gray-800 dark:text-white group-hover:text-[#CA400A] transition-colors">
                  {p.name}
                </span>
                <span className="text-[10px] px-2 py-0.5 rounded-md font-extrabold uppercase border border-gray-200 dark:border-[#313131] bg-gray-100 dark:bg-[#242424] text-gray-600 dark:text-gray-400 shrink-0">
                  {p.category}
                </span>
              </div>

              <span className="text-sm md:text-base font-extrabold text-[#CA400A] block">
                {formatRupiah(p.price)}
              </span>

              <p className="text-xs md:text-sm text-gray-500 dark:text-gray-400 leading-relaxed">
                {p.description}
              </p>
            </div>

            {/* KOMPOSISI DETAIL BAHAN BAKU */}
            <div className="pt-4 border-t border-gray-100 dark:border-[#242424] mt-4 space-y-2">
              <span className="text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-wider block">
                Komposisi Takaran Resep:
              </span>
              <div className="flex flex-wrap gap-1.5">
                {p.recipe && p.recipe.map((ing, ingIdx) => (
                  <span 
                    key={ingIdx} 
                    className="text-[11px] px-2.5 py-1 rounded-md font-semibold border border-gray-200 dark:border-[#2D2D2D] bg-gray-50 dark:bg-[#212121] text-gray-600 dark:text-gray-400"
                  >
                    {ing.name}: {ing.qty} {ing.unit}
                  </span>
                ))}
              </div>
            </div>

          </div>
        ))}

        {filteredProducts.length === 0 && (
          <div className="col-span-full text-center py-12 text-gray-400 dark:text-gray-500 font-medium">
            Tidak ada menu yang terdaftar di dalam kategori ini.
          </div>
        )}
      </div>

    </div>
  );
}