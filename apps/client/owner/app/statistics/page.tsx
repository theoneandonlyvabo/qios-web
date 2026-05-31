'use client';

import React, { useState, useMemo } from 'react';
import { INITIAL_PRODUCTS, INITIAL_TRANSACTIONS } from '../../lib/mockData';

export default function StatisticsPage() {
  const [timeframe, setTimeframe] = useState('7 Hari');

  // Urutkan produk berdasarkan performa kuantitas penjualan tertinggi
  const productPerformance = useMemo(() => {
    return [...INITIAL_PRODUCTS].sort((a, b) => b.total_sold - a.total_sold);
  }, []);

  const formatRupiah = (num: number) => {
    return 'Rp ' + num.toLocaleString('id-ID');
  };

  return (
    <div className="p-4 md:p-8 space-y-6 animate-fadeIn text-gray-900 dark:text-white">
      
      {/* BARIS KONTROL ATAS */}
      <div className="p-4 rounded-xl border border-gray-200 dark:border-[#242424] bg-white dark:bg-[#161616] flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 shadow-sm">
        <div className="flex items-center gap-3">
          <span className="text-xs md:text-sm font-bold text-gray-400 dark:text-gray-500 uppercase tracking-wider">Timeframe Analisis:</span>
          <div className="inline-flex rounded-lg p-1 bg-gray-100 dark:bg-[#242424] border border-gray-200 dark:border-[#3E3E3E]">
            {['7 Hari', '30 Hari', '3 Bulan'].map((p) => (
              <button
                key={p}
                onClick={() => setTimeframe(p)}
                className={`px-3 py-1 text-xs font-bold rounded h-8 flex items-center transition-all ${
                  timeframe === p
                    ? 'bg-[#CA400A] text-white shadow'
                    : 'text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'
                }`}
              >
                {p}
              </button>
            ))}
          </div>
        </div>

        <div className="text-xs md:text-sm text-gray-500 dark:text-gray-400 font-semibold">
          Menganalisis total <span className="font-extrabold text-gray-800 dark:text-white">{INITIAL_TRANSACTIONS.length} Transaksi</span> dari buku besar.
        </div>
      </div>

      {/* METRIK PERSENTASE KONTRIBUSI */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {[
          { label: 'Kategori Kopi', val: '65.2%', sub: 'Kontribusi menu berbasis espresso', color: 'text-[#CA400A]' },
          { label: 'Kategori Non-Kopi', val: '22.8%', sub: 'Kontribusi Matcha & teh premium', color: 'text-emerald-600 dark:text-emerald-400' },
          { label: 'Kategori Makanan', val: '12.0%', sub: 'Sinergi penjualan pastry/croissant', color: 'text-amber-600 dark:text-amber-500' }
        ].map((item, idx) => (
          <div key={idx} className="p-5 rounded-xl border border-gray-200 dark:border-[#242424] bg-white dark:bg-[#161616] shadow-sm">
            <span className="text-xs font-bold text-gray-400 dark:text-gray-500 uppercase block tracking-wider">{item.label}</span>
            <span className={`text-3xl font-black block mt-2 tracking-tight ${item.color}`}>{item.val}</span>
            <span className="text-[11px] text-gray-500 dark:text-gray-400 block mt-1 font-semibold">{item.sub}</span>
          </div>
        ))}
      </div>

      {/* TABEL PERFORMA PRODUK DETAIL */}
      <div className="rounded-xl border border-gray-200 dark:border-[#242424] bg-white dark:bg-[#161616] p-4 md:p-6 overflow-hidden shadow-sm">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between pb-4 border-b border-gray-100 dark:border-[#2D2D2D] mb-6 gap-2">
          <h3 className="font-bold text-sm md:text-base uppercase tracking-wider text-gray-800 dark:text-white">
            Peringkat Kontribusi Penjualan Menu
          </h3>
          <span className="text-[10px] text-gray-400 dark:text-gray-500 font-bold uppercase tracking-widest">
            Audit Real-time
          </span>
        </div>
        
        <div className="overflow-x-auto -mx-4 px-4 md:mx-0 md:px-0">
          <table className="w-full text-left text-sm min-w-[600px]">
            <thead>
              <tr className="border-b border-gray-200 dark:border-[#2D2D2D] text-[11px] uppercase font-bold text-gray-400 dark:text-gray-500 tracking-wider bg-gray-50 dark:bg-[#212121]">
                <th className="p-3">Posisi</th>
                <th className="p-3">Nama Produk</th>
                <th className="p-3">Kategori</th>
                <th className="p-3 text-right">Porsi Terjual</th>
                <th className="p-3 text-right">Harga Satuan</th>
                <th className="p-3 text-right">Total Kontribusi Revenue</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 dark:divide-[#242424]">
              {productPerformance.map((p, idx) => (
                <tr key={p.id} className="transition-all hover:bg-gray-50/50 dark:hover:bg-[#1C1C1C]">
                  <td className="p-3 font-black text-[#CA400A]">#{idx + 1}</td>
                  <td className="p-3 font-bold text-gray-800 dark:text-white">{p.name}</td>
                  <td className="p-3">
                    <span className="text-xs px-2.5 py-0.5 rounded font-bold uppercase border border-gray-200 dark:border-[#313131] bg-gray-100 dark:bg-[#242424] text-gray-600 dark:text-gray-400">
                      {p.category}
                    </span>
                  </td>
                  <td className="p-3 text-right font-bold text-gray-800 dark:text-white">{p.total_sold} Porsi</td>
                  <td className="p-3 text-right text-gray-400 dark:text-gray-500 font-semibold">{formatRupiah(p.price)}</td>
                  <td className="p-3 text-right font-black text-[#CA400A]">
                    {formatRupiah(p.price * p.total_sold)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

    </div>
  );
}