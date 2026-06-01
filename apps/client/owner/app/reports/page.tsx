'use client';

import React, { useState, useMemo } from 'react';
import { INITIAL_TRANSACTIONS, INITIAL_PRODUCTS } from '../../lib/mockData';

export default function ReportsPage() {
  const [reportTab, setReportTab] = useState<'daily' | 'monthly' | 'consumption'>('daily');
  const [reportDate, setReportDate] = useState('2026-05-25');
  const [reportMonth, setReportMonth] = useState('2026-05');
  const [selectedRange, setSelectedRange] = useState({ start: '2026-05-20', end: '2026-05-25' });
  const [exportProgress, setExportProgress] = useState<{ active: boolean; progress: number; downloadUrl?: string } | null>(null);
  const [alertInfo, setAlertInfo] = useState<{ title: string; message: string } | null>(null);

  // ==========================================
  // KALKULASI LOG KONSUMSI BAHAN BAKU DARI RESEP
  // ==========================================
  const calculations = useMemo(() => {
    const confirmedTx = INITIAL_TRANSACTIONS.filter(t => t.status === 'CONFIRMED');
    const totalRevenue = confirmedTx.reduce((sum, t) => sum + t.total_amount, 0);
    const totalTransactions = confirmedTx.length;

    // Kalkulasi volume per-item yang terjual
    const productSalesMap: Record<string, { name: string; total_sold: number; total_revenue: number }> = {};
    confirmedTx.forEach(tx => {
      tx.items.forEach(item => {
        if (!productSalesMap[item.product_id]) {
          productSalesMap[item.product_id] = { name: item.product_name, total_sold: 0, total_revenue: 0 };
        }
        productSalesMap[item.product_id].total_sold += item.quantity;
        productSalesMap[item.product_id].total_revenue += item.subtotal;
      });
    });

    // Kalkulasi konsumsi bahan baku bersarang (nested recipes)
    const ingredientConsumptionMap: Record<string, { total: number; unit: string; usedIn: Record<string, number> }> = {};
    confirmedTx.forEach(tx => {
      tx.items.forEach(item => {
        const fullProduct = INITIAL_PRODUCTS.find(p => p.id === item.product_id);
        if (fullProduct && fullProduct.recipe) {
          fullProduct.recipe.forEach(ing => {
            if (!ingredientConsumptionMap[ing.name]) {
              ingredientConsumptionMap[ing.name] = { total: 0, unit: ing.unit, usedIn: {} };
            }
            const currentItemUsage = ing.qty * item.quantity;
            ingredientConsumptionMap[ing.name].total += currentItemUsage;
            if (!ingredientConsumptionMap[ing.name].usedIn[item.product_name]) {
              ingredientConsumptionMap[ing.name].usedIn[item.product_name] = 0;
            }
            ingredientConsumptionMap[ing.name].usedIn[item.product_name] += currentItemUsage;
          });
        }
      });
    });

    return {
      totalRevenue,
      totalTransactions,
      topProducts: Object.values(productSalesMap),
      ingredientConsumption: Object.entries(ingredientConsumptionMap).map(([name, data]) => ({
        name,
        total: data.total,
        unit: data.unit,
        usedIn: Object.entries(data.usedIn).map(([pName, qty]) => ({ pName, qty }))
      }))
    };
  }, []);

  const formatRupiah = (num: number) => {
    return 'Rp ' + num.toLocaleString('id-ID');
  };

  const handleExportReport = (type: string, format: 'pdf' | 'csv') => {
    setExportProgress({ active: true, progress: 5 });
    
    let tick = 5;
    const interval = setInterval(() => {
      tick += 25;
      if (tick >= 100) {
        clearInterval(interval);
        setExportProgress({
          active: true,
          progress: 100,
          downloadUrl: `#`,
        });
      } else {
        setExportProgress(prev => prev ? { ...prev, progress: tick } : null);
      }
    }, 300);
  };

  return (
    <div className="p-4 md:p-8 space-y-6 animate-fadeIn text-gray-900 dark:text-white">
      
      {/* SELEKTOR TAB LAPORAN */}
      <div className="border-b border-gray-200 dark:border-[#242424] flex gap-6 overflow-x-auto">
        {[
          { id: 'daily', label: 'Harian' },
          { id: 'monthly', label: 'Bulanan' },
          { id: 'consumption', label: 'Bahan Baku' },
        ].map(tab => (
          <button
            key={tab.id}
            onClick={() => setReportTab(tab.id as any)}
            className={`pb-3 text-xs md:text-sm font-bold border-b-2 uppercase tracking-wide transition-all whitespace-nowrap ${
              reportTab === tab.id
                ? 'border-[#CA400A] text-[#CA400A]'
                : 'border-transparent text-gray-500 hover:text-gray-900 dark:hover:text-white'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* KONTEN TAB LAPORAN: HARIAN */}
      {reportTab === 'daily' && (
        <div className="space-y-6 animate-fadeIn">
          <div className="p-4 rounded-xl border border-gray-200 dark:border-[#242424] bg-white dark:bg-[#161616] flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 shadow-sm">
            <div className="flex items-center gap-3">
              <span className="text-xs md:text-sm font-bold text-gray-400 dark:text-gray-500 uppercase">Tanggal:</span>
              <input
                type="date"
                value={reportDate}
                onChange={(e) => setReportDate(e.target.value)}
                className="text-xs px-3 py-1.5 rounded border border-gray-300 dark:border-[#383838] bg-white dark:bg-[#242424] font-bold focus:outline-none h-10 text-gray-700 dark:text-white"
              />
            </div>
            <button 
              onClick={() => handleExportReport('Harian', 'pdf')}
              className="bg-[#CA400A] hover:bg-[#E04E15] text-xs text-white px-4 py-2.5 rounded-lg font-bold uppercase transition-colors h-10 shadow-md shadow-[#CA400A]/10"
            >
              Export PDF
            </button>
          </div>

          <div className="rounded-xl border border-gray-200 dark:border-[#242424] bg-white dark:bg-[#161616] p-5 md:p-6 shadow-sm">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8 text-center border-b border-gray-100 dark:border-[#242424] pb-6">
              <div>
                <span className="text-xs uppercase font-bold text-gray-400 dark:text-gray-500">Pendapatan Bersih</span>
                <span className="text-2xl md:text-3xl font-black block mt-1 text-gray-800 dark:text-white">{formatRupiah(calculations.totalRevenue)}</span>
              </div>
              <div>
                <span className="text-xs uppercase font-bold text-gray-400 dark:text-gray-500">Transaksi Selesai</span>
                <span className="text-2xl md:text-3xl font-black block mt-1 text-gray-800 dark:text-white">{calculations.totalTransactions}</span>
              </div>
              <div>
                <span className="text-xs uppercase font-bold text-gray-400 dark:text-gray-500">Breakdown Tunai / QRIS</span>
                <span className="text-2xl md:text-3xl font-black block mt-1 text-gray-800 dark:text-white">50% / 50%</span>
              </div>
            </div>

            <h4 className="font-bold text-xs uppercase tracking-wider mb-4 text-gray-400 dark:text-gray-500">Breakdown Penjualan Per Menu</h4>
            <div className="divide-y divide-gray-100 dark:divide-[#242424]">
              {calculations.topProducts.map((p, idx) => (
                <div key={idx} className="py-3 flex items-center justify-between text-xs md:text-sm">
                  <span className="font-semibold text-gray-800 dark:text-white">{p.name}</span>
                  <span className="font-bold text-gray-600 dark:text-gray-400">{p.total_sold} Porsi ({formatRupiah(p.total_revenue)})</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* KONTEN TAB LAPORAN: BULANAN */}
      {reportTab === 'monthly' && (
        <div className="space-y-6 animate-fadeIn">
          <div className="p-4 rounded-xl border border-gray-200 dark:border-[#242424] bg-white dark:bg-[#161616] flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 shadow-sm">
            <div className="flex items-center gap-3">
              <span className="text-xs md:text-sm font-bold text-gray-400 dark:text-gray-500 uppercase">Bulan:</span>
              <input
                type="month"
                value={reportMonth}
                onChange={(e) => setReportMonth(e.target.value)}
                className="text-xs px-3 py-1.5 rounded border border-gray-300 dark:border-[#383838] bg-white dark:bg-[#242424] font-bold focus:outline-none h-10 text-gray-700 dark:text-white"
              />
            </div>
            <button 
              onClick={() => handleExportReport('Bulanan', 'csv')}
              className="bg-[#CA400A] hover:bg-[#E04E15] text-xs text-white px-4 py-2.5 rounded-lg font-bold uppercase transition-colors h-10 shadow-md shadow-[#CA400A]/10"
            >
              Export CSV
            </button>
          </div>

          <div className="rounded-xl border border-gray-200 dark:border-[#242424] bg-white dark:bg-[#161616] p-6 text-center shadow-sm">
            <svg className="w-12 h-12 text-[#CA400A] mx-auto mb-3" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
            </svg>
            <h3 className="font-extrabold text-base md:text-lg text-gray-800 dark:text-white">Laporan Konsolidasi Bulanan</h3>
            <p className="text-xs md:text-sm text-gray-400 dark:text-gray-500 max-w-sm mx-auto mt-2 leading-relaxed">
              Total perputaran kas terekam bersih sebesar <span className="font-bold text-gray-800 dark:text-white">{formatRupiah(calculations.totalRevenue)}</span> dari akumulasi POS kasir.
            </p>
          </div>
        </div>
      )}

      {/* KONTEN TAB LAPORAN: BAHAN BAKU */}
      {reportTab === 'consumption' && (
        <div className="space-y-6 animate-fadeIn">
          <div className="p-4 rounded-xl border border-gray-200 dark:border-[#242424] bg-white dark:bg-[#161616] flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between shadow-sm">
            <div className="flex flex-wrap items-center gap-2">
              <input
                type="date"
                value={selectedRange.start}
                onChange={(e) => setSelectedRange({ ...selectedRange, start: e.target.value })}
                className="text-xs px-3 py-1.5 rounded border border-gray-300 dark:border-[#383838] bg-white dark:bg-[#242424] font-bold focus:outline-none h-10 text-gray-700 dark:text-white"
              />
              <span className="text-gray-400 font-bold text-xs">s/d</span>
              <input
                type="date"
                value={selectedRange.end}
                onChange={(e) => setSelectedRange({ ...selectedRange, end: e.target.value })}
                className="text-xs px-3 py-1.5 rounded border border-gray-300 dark:border-[#383838] bg-white dark:bg-[#242424] font-bold focus:outline-none h-10 text-gray-700 dark:text-white"
              />
            </div>
            <button 
              onClick={() => handleExportReport('Bahan_Baku', 'pdf')}
              className="bg-[#CA400A] hover:bg-[#E04E15] text-xs text-white px-4 py-2.5 rounded-lg font-bold uppercase transition-colors h-10 w-full sm:w-auto shadow-md shadow-[#CA400A]/10"
            >
              Export PDF
            </button>
          </div>

          <div className="rounded-xl border border-gray-200 dark:border-[#242424] bg-white dark:bg-[#161616] p-5 md:p-6 shadow-sm">
            <div className="flex justify-between items-center pb-4 border-b border-gray-100 dark:border-[#242424] mb-6">
              <h3 className="font-extrabold text-xs uppercase tracking-wider text-gray-400 dark:text-gray-500">
                Log Konsumsi Bahan Baku Bersarang
              </h3>
              <span className="text-[10px] text-gray-400 dark:text-gray-500 font-bold uppercase tracking-widest">Auto-POS Sync</span>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {calculations.ingredientConsumption.map((ing, idx) => (
                <div key={idx} className="p-4 rounded-xl border border-gray-200 dark:border-[#2D2D2D] bg-gray-50 dark:bg-[#212121] space-y-3 shadow-sm">
                  <div className="flex justify-between items-center">
                    <span className="font-black text-sm uppercase tracking-wider text-gray-800 dark:text-white">{ing.name}</span>
                    <span className="text-sm font-black text-[#CA400A]">
                      {ing.total.toLocaleString('id-ID')} {ing.unit}
                    </span>
                  </div>
                  <div className="pt-2 border-t border-gray-200 dark:border-gray-500/10 text-xs text-gray-400 space-y-1">
                    <span className="font-bold block text-[10px] uppercase text-gray-400 dark:text-gray-500 mb-1">Dipakai untuk:</span>
                    {ing.usedIn.map((used, uIdx) => (
                      <div key={uIdx} className="flex justify-between text-gray-600 dark:text-gray-300">
                        <span>{used.pName}</span>
                        <span className="text-gray-800 dark:text-white font-bold">{used.qty} {ing.unit}</span>
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* EXPORT PROGRESS DIALOG */}
      {exportProgress && exportProgress.active && (
        <div className="fixed inset-0 bg-black/85 flex items-end sm:items-center justify-center p-0 sm:p-4 z-50 animate-fadeIn">
          <div className="bg-white dark:bg-[#161616] w-full sm:max-w-md rounded-t-2xl sm:rounded-2xl p-6 text-center space-y-4 border border-gray-200 dark:border-[#2C2C2C] shadow-2xl">
            <h3 className="font-extrabold text-sm uppercase tracking-wider text-gray-800 dark:text-white">Memproses Kompresi File...</h3>
            <div className="w-full bg-gray-200 dark:bg-[#242424] h-3 rounded-full overflow-hidden relative border border-gray-300 dark:border-[#383838]">
              <div className="h-full bg-[#CA400A] rounded-full transition-all duration-300" style={{ width: `${exportProgress.progress}%` }}></div>
            </div>
            <span className="text-xs font-semibold block text-gray-500 dark:text-gray-400">Progres Ekspor: {exportProgress.progress}%</span>

            {exportProgress.progress >= 100 && (
              <div className="pt-4 border-t border-gray-200 dark:border-[#242424] space-y-3">
                <span className="text-emerald-500 dark:text-emerald-400 text-xs font-bold block">✓ Laporan PDF/CSV Berhasil Dibuat</span>
                <div className="flex flex-col sm:flex-row gap-3">
                  <button
                    onClick={() => {
                      setAlertInfo({ title: 'Download Selesai', message: 'Laporan digital telah disimpan ke memori penyimpanan lokal Anda.' });
                      setExportProgress(null);
                    }}
                    className="w-full bg-[#CA400A] text-white text-xs font-bold py-3 rounded-lg block uppercase tracking-wider text-center h-12 flex items-center justify-center"
                  >
                    Unduh Sekarang
                  </button>
                  <button onClick={() => setExportProgress(null)} className="w-full bg-gray-100 dark:bg-[#242424] text-gray-800 dark:text-white border border-gray-300 dark:border-[#2D2D2D] text-xs font-bold py-3 rounded-lg h-12">
                    Batal
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {alertInfo && (
        <div className="fixed inset-0 bg-black/90 flex items-center justify-center p-4 z-[100] animate-fadeIn">
          <div className="bg-white dark:bg-[#161616] w-full max-w-sm rounded-2xl p-6 text-center space-y-4 shadow-2xl border border-gray-200 dark:border-[#2D2D2D]">
            <div className="w-10 h-10 rounded-full bg-[#CA400A]/10 text-[#CA400A] border border-[#CA400A]/30 flex items-center justify-center mx-auto">
              <svg className="w-6 h-6 text-[#CA400A]" fill="none" stroke="currentColor" strokeWidth="2.5" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div>
              <h4 className="font-black text-sm uppercase tracking-wider text-gray-900 dark:text-white">{alertInfo.title}</h4>
              <p className="text-xs text-gray-500 dark:text-gray-400 mt-2 leading-relaxed">{alertInfo.message}</p>
            </div>
            <button onClick={() => setAlertInfo(null)} className="w-full bg-[#CA400A] text-white text-xs font-bold py-2.5 rounded-lg uppercase h-11 flex items-center justify-center">
              Mengerti
            </button>
          </div>
        </div>
      )}

    </div>
  );
}