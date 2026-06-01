'use client';

import React, { useState, useMemo } from 'react';
import { INITIAL_TRANSACTIONS, INITIAL_PRODUCTS } from '../../lib/mockData';
import { Transaction } from '../../types';

export default function DashboardPage() {
  // Menggunakan state lokal untuk transaksi agar interaksi live (seperti Void) dapat memperbarui UI secara real-time
  const [transactions, setTransactions] = useState<Transaction[]>(INITIAL_TRANSACTIONS);
  const [selectedTx, setSelectedTx] = useState<Transaction | null>(null);
  const [showVoidDialog, setShowVoidDialog] = useState(false);
  const [voidReason, setVoidReason] = useState('');
  
  // Custom alert modal state
  const [alertInfo, setAlertInfo] = useState<{ title: string; message: string } | null>(null);

  // ==========================================
  // KALKULASI METRIK SECARA REAL-TIME
  // ==========================================
  const metrics = useMemo(() => {
    const confirmedTx = transactions.filter(t => t.status === 'CONFIRMED');
    const totalRevenue = confirmedTx.reduce((sum, t) => sum + t.total_amount, 0);
    const totalTransactions = confirmedTx.length;
    const avgOrderValue = totalTransactions > 0 ? totalRevenue / totalTransactions : 0;
    const pendingCount = transactions.filter(t => t.status === 'PENDING').length;
    const voidedCount = transactions.filter(t => t.status === 'VOIDED').length;

    // Hitung 5 Menu Terlaris berdasarkan transaksi yang dikonfirmasi
    const productSalesMap: Record<string, { name: string; category: string; total_sold: number; total_revenue: number }> = {};
    confirmedTx.forEach(tx => {
      tx.items.forEach(item => {
        const fullProduct = INITIAL_PRODUCTS.find(p => p.id === item.product_id);
        const category = fullProduct ? fullProduct.category : 'Umum';
        
        if (!productSalesMap[item.product_id]) {
          productSalesMap[item.product_id] = { 
            name: item.product_name, 
            category: category,
            total_sold: 0, 
            total_revenue: 0 
          };
        }
        productSalesMap[item.product_id].total_sold += item.quantity;
        productSalesMap[item.product_id].total_revenue += item.subtotal;
      });
    });

    const topProducts = Object.values(productSalesMap)
      .sort((a, b) => b.total_sold - a.total_sold)
      .slice(0, 5);

    return {
      totalRevenue,
      totalTransactions,
      avgOrderValue,
      pendingCount,
      voidedCount,
      topProducts
    };
  }, [transactions]);

  // Formatting Rupiah Helper
  const formatRupiah = (num: number) => {
    return 'Rp ' + num.toLocaleString('id-ID');
  };

  const formatDate = (isoStr: string) => {
    const d = new Date(isoStr);
    return d.toLocaleDateString('id-ID', {
      day: 'numeric',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  // Handler Konfirmasi Void Transaksi
  const handleConfirmVoid = () => {
    if (!voidReason.trim()) {
      setAlertInfo({
        title: 'Gagal',
        message: 'Alasan pembatalan (Void Reason) wajib diisi untuk kebutuhan audit log POS.'
      });
      return;
    }

    if (!selectedTx) return;

    const targetTx = selectedTx;
    setTransactions(prev =>
      prev.map(t => {
        if (t.id === targetTx.id) {
          return {
            ...t,
            status: 'VOIDED',
            void_reason: voidReason,
            voided_at: new Date().toISOString()
          };
        }
        return t;
      })
    );

    setShowVoidDialog(false);
    setSelectedTx(null);
    setVoidReason('');
    setAlertInfo({
      title: 'Berhasil Void',
      message: `Transaksi ${targetTx.order_id} berhasil di-void dari pembukuan harian.`
    });
  };

  return (
    <div className="p-4 md:p-8 space-y-6 animate-fadeIn text-gray-900 dark:text-white">
      
      {/* 1. METRICS GRID */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
        {[
          { label: 'Total Revenue', value: formatRupiah(metrics.totalRevenue), sub: '+12.4% vs pekan lalu', border: 'border-[#CA400A]', text: 'text-gray-900 dark:text-white' },
          { label: 'Transaksi CONFIRMED', value: metrics.totalTransactions, sub: '+8.1% vs pekan lalu', border: 'border-gray-200 dark:border-[#2D2D2D]', text: 'text-[#CA400A]' },
          { label: 'Avg Order Value (AOV)', value: formatRupiah(Math.round(metrics.avgOrderValue)), sub: '+3.9% vs pekan lalu', border: 'border-gray-200 dark:border-[#2D2D2D]', text: 'text-gray-900 dark:text-white' },
          { label: 'Transaksi PENDING', value: metrics.pendingCount, sub: 'Butuh konfirmasi kasir', border: 'border-amber-200 dark:border-amber-900/30', text: 'text-amber-600 dark:text-amber-500' },
          { label: 'Transaksi VOIDED', value: metrics.voidedCount, sub: 'Log pembatalan tercatat', border: 'border-red-200 dark:border-red-950/30', text: 'text-red-600 dark:text-red-500' },
        ].map((card, idx) => (
          <div key={idx} className={`p-5 rounded-xl bg-white dark:bg-[#161616] border ${card.border} hover:border-[#CA400A]/50 transition-colors shadow-sm`}>
            <span className="text-xs font-bold text-gray-400 dark:text-gray-500 uppercase block tracking-wider">{card.label}</span>
            <span className={`text-2xl font-black block mt-2 tracking-tight ${card.text}`}>{card.value}</span>
            <span className="text-[11px] text-gray-500 dark:text-gray-400 block mt-1 font-semibold">{card.sub}</span>
          </div>
        ))}
      </div>

      {/* 2. CHARTS AREA */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        
        {/* Revenue Trend Line Chart */}
        <div className="lg:col-span-2 p-5 md:p-6 rounded-xl bg-white dark:bg-[#161616] border border-gray-200 dark:border-[#242424] shadow-sm">
          <div className="flex items-center justify-between mb-6">
            <h3 className="font-bold text-sm md:text-base uppercase tracking-wider text-gray-800 dark:text-white">Tren Revenue Harian</h3>
            <span className="bg-[#CA400A]/10 text-[#CA400A] text-[10px] md:text-xs px-2.5 py-1 rounded font-bold uppercase">
              7 Hari Terakhir
            </span>
          </div>

          <div className="h-48 md:h-64 relative">
            <svg className="w-full h-full" viewBox="0 0 500 200" preserveAspectRatio="none">
              <defs>
                <linearGradient id="chartGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="#CA400A" stopOpacity="0.4"/>
                  <stop offset="100%" stopColor="#CA400A" stopOpacity="0"/>
                </linearGradient>
              </defs>
              <line x1="0" y1="40" x2="500" y2="40" className="stroke-gray-200 dark:stroke-[#252525]" strokeDasharray="3,3" />
              <line x1="0" y1="85" x2="500" y2="85" className="stroke-gray-200 dark:stroke-[#252525]" strokeDasharray="3,3" />
              <line x1="0" y1="130" x2="500" y2="130" className="stroke-gray-200 dark:stroke-[#252525]" strokeDasharray="3,3" />
              <line x1="0" y1="175" x2="500" y2="175" className="stroke-gray-200 dark:stroke-[#252525]" strokeDasharray="3,3" />

              <path
                d="M0,170 Q40,150 80,165 T160,110 T240,120 T320,80 T400,95 T480,45"
                fill="none"
                stroke="#CA400A"
                strokeWidth="3.5"
                strokeLinecap="round"
              />
              <path
                d="M0,170 Q40,150 80,165 T160,110 T240,120 T320,80 T400,95 T480,45 L500,200 L0,200 Z"
                fill="url(#chartGrad)"
              />
              <circle cx="80" cy="165" r="5" className="fill-white dark:fill-[#161616]" stroke="#CA400A" strokeWidth="2" />
              <circle cx="160" cy="110" r="5" className="fill-white dark:fill-[#161616]" stroke="#CA400A" strokeWidth="2" />
              <circle cx="240" cy="120" r="5" className="fill-white dark:fill-[#161616]" stroke="#CA400A" strokeWidth="2" />
              <circle cx="320" cy="80" r="5" className="fill-white dark:fill-[#161616]" stroke="#CA400A" strokeWidth="2" />
              <circle cx="400" cy="95" r="5" className="fill-white dark:fill-[#161616]" stroke="#CA400A" strokeWidth="2" />
              <circle cx="480" cy="45" r="6" fill="#CA400A" className="stroke-white dark:stroke-[#161616]" strokeWidth="2.5" />
            </svg>

            <div className="flex justify-between text-[9px] md:text-[11px] font-bold mt-2 text-gray-400 dark:text-gray-500">
              <span>Sel (19)</span>
              <span>Rab (20)</span>
              <span>Kam (21)</span>
              <span>Jum (22)</span>
              <span>Sab (23)</span>
              <span>Min (24)</span>
              <span className="text-[#CA400A]">Sen (25)</span>
            </div>
          </div>
        </div>

        {/* Peak Hours Custom SVG Bar Chart */}
        <div className="p-4 md:p-6 rounded-xl border bg-white dark:bg-[#161616] border-gray-200 dark:border-[#242424] shadow-sm">
          <h3 className="font-bold text-sm md:text-base uppercase tracking-wider mb-6 text-gray-800 dark:text-white">Distribusi Jam Tersibuk</h3>
          <div className="space-y-4">
            {[
              { range: '08:00 - 10:00 (Pagi)', vol: 15, pct: 30, level: 'Senggang', highlight: false },
              { range: '12:00 - 14:00 (Siang)', vol: 41, pct: 82, level: 'Puncak', highlight: true },
              { range: '16:00 - 18:00 (Sore)', vol: 27, pct: 54, level: 'Ramai', highlight: false },
              { range: '19:00 - 21:00 (Malam)', vol: 17, pct: 34, level: 'Senggang', highlight: false }
            ].map((hour, idx) => (
              <div key={idx} className="space-y-1">
                <div className="flex justify-between text-xs">
                  <span className={`font-semibold ${hour.highlight ? 'text-[#CA400A]' : 'text-gray-700 dark:text-gray-300'}`}>
                    {hour.range}
                  </span>
                  <span className="text-gray-400 dark:text-gray-500 font-bold">{hour.vol} Tx</span>
                </div>
                
                <div className="w-full h-3 rounded-full overflow-hidden relative border border-gray-200 dark:border-[#303030] bg-gray-100 dark:bg-[#242424]">
                  <div 
                    className={`h-full rounded-full transition-all duration-1000 ${
                      hour.highlight ? 'bg-gradient-to-r from-[#CA400A] to-[#E54D16]' : 'bg-gray-400 dark:bg-gray-500'
                    }`} 
                    style={{ width: `${hour.pct}%` }}
                  ></div>
                </div>
              </div>
            ))}
          </div>
        </div>

      </div>

      {/* 3. LOWER SECTION: TOP PRODUCTS & RECENT TRANSACTIONS */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        
        {/* Top Products */}
        <div className="p-4 md:p-6 rounded-xl border bg-white dark:bg-[#161616] border-gray-200 dark:border-[#242424] shadow-sm">
          <h3 className="font-bold text-sm md:text-base uppercase tracking-wider mb-4 text-gray-800 dark:text-white">5 Menu Terlaris Pekan Ini</h3>
          <div className="divide-y divide-gray-200 dark:divide-[#242424]">
            {metrics.topProducts.map((p, idx) => (
              <div key={idx} className="py-3 flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <span className="text-sm font-black text-[#CA400A]">#{idx + 1}</span>
                  <div>
                    <span className="font-bold text-sm block text-gray-800 dark:text-white">{p.name}</span>
                    <span className="text-xs text-gray-400 dark:text-gray-500">{p.category}</span>
                  </div>
                </div>
                <div className="text-right">
                  <span className="font-bold text-sm block text-gray-800 dark:text-white">{p.total_sold} Porsi</span>
                  <span className="text-[11px] text-gray-400 dark:text-gray-500 font-semibold">{formatRupiah(p.total_revenue)}</span>
                </div>
              </div>
            ))}
            {metrics.topProducts.length === 0 && (
              <p className="text-center py-6 text-sm text-gray-400">Belum ada data penjualan.</p>
            )}
          </div>
        </div>

        {/* Recent Transactions List */}
        <div className="lg:col-span-2 p-4 md:p-6 rounded-xl border bg-white dark:bg-[#161616] border-gray-200 dark:border-[#242424] shadow-sm">
          <h3 className="font-bold text-sm md:text-base uppercase tracking-wider mb-4 text-gray-800 dark:text-white">Aktivitas Transaksi Terakhir</h3>
          
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm min-w-[500px]">
              <thead>
                <tr className="border-b border-gray-200 dark:border-[#2D2D2D] text-[10px] uppercase font-bold text-gray-400 dark:text-gray-500 tracking-wider">
                  <th className="pb-3">Order ID</th>
                  <th className="pb-3">Waktu</th>
                  <th className="pb-3">Kasir</th>
                  <th className="pb-3 text-right">Total Amount</th>
                  <th className="pb-3">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100 dark:divide-[#1F1F1F]">
                {transactions.slice(0, 5).map((t, idx) => (
                  <tr 
                    key={idx} 
                    onClick={() => setSelectedTx(t)}
                    className="hover:bg-gray-50 dark:hover:bg-[#1C1C1C] transition-colors cursor-pointer"
                  >
                    <td className="py-3 font-bold text-[#CA400A]">{t.order_id}</td>
                    <td className="py-3 text-xs text-gray-500 dark:text-gray-400">{formatDate(t.created_at)}</td>
                    <td className="py-3 font-semibold text-gray-700 dark:text-gray-300">{t.created_by_operator_name}</td>
                    <td className="py-3 text-right font-bold text-gray-800 dark:text-white">{formatRupiah(t.total_amount)}</td>
                    <td className="py-3">
                      <span className={`text-[10px] font-extrabold uppercase px-2 py-0.5 rounded ${
                        t.status === 'CONFIRMED' ? 'bg-emerald-50 dark:bg-emerald-950/40 text-emerald-700 dark:text-emerald-400 border border-emerald-200 dark:border-emerald-900/30' :
                        t.status === 'PENDING' ? 'bg-amber-50 dark:bg-amber-950/40 text-amber-700 dark:text-amber-400 border border-amber-200 dark:border-amber-900/30' :
                        'bg-red-50 dark:bg-red-950/40 text-red-700 dark:text-red-400 border border-red-200 dark:border-red-900/30'
                      }`}>
                        {t.status}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

      </div>

      {/* ==========================================
          MODALS & DIALOG LAYERS
          ========================================== */}

      {/* TRANSACTION DETAIL MODAL */}
      {selectedTx && (
        <div className="fixed inset-0 bg-black/80 flex items-end sm:items-center justify-center p-0 sm:p-4 z-50 animate-fadeIn">
          <div className="bg-white dark:bg-[#161616] w-full sm:max-w-xl rounded-t-2xl sm:rounded-2xl p-6 space-y-6 shadow-2xl overflow-y-auto max-h-[90vh] border-t sm:border border-gray-200 dark:border-[#2C2C2C]">
            
            <div className="flex items-center justify-between pb-4 border-b border-gray-100 dark:border-[#242424]">
              <div>
                <h3 className="font-extrabold text-md text-[#CA400A] uppercase tracking-wider">Detail Transaksi POS</h3>
                <span className="text-xs text-gray-400 dark:text-gray-500 font-bold">ID: {selectedTx.order_id}</span>
              </div>
              <button 
                onClick={() => setSelectedTx(null)}
                className="font-extrabold text-xl p-2 -mr-2 text-gray-400 dark:text-gray-500 hover:text-gray-700 dark:hover:text-white"
              >
                ✕
              </button>
            </div>

            <div className="grid grid-cols-2 gap-4 text-xs text-gray-500 dark:text-gray-400">
              <div>Waktu: <span className="font-semibold block text-gray-800 dark:text-white">{formatDate(selectedTx.created_at)}</span></div>
              <div>Metode: <span className="font-semibold block uppercase text-gray-800 dark:text-white">{selectedTx.payment_method || 'PENDING'}</span></div>
              <div>Kasir: <span className="font-semibold block text-gray-800 dark:text-white">{selectedTx.created_by_operator_name}</span></div>
              <div>
                Status: 
                <span className={`block font-extrabold uppercase mt-0.5 ${
                  selectedTx.status === 'CONFIRMED' ? 'text-emerald-600 dark:text-emerald-400' :
                  selectedTx.status === 'PENDING' ? 'text-amber-600 dark:text-amber-500' : 'text-red-600 dark:text-red-500'
                }`}>
                  {selectedTx.status}
                </span>
              </div>
            </div>

            <div className="space-y-3">
              <span className="text-xs font-bold text-gray-400 dark:text-gray-500 uppercase tracking-wider block">Items Terjual:</span>
              <div className="rounded-xl p-3 border border-gray-100 dark:border-[#2D2D2D] bg-gray-50 dark:bg-[#212121] divide-y divide-gray-200 dark:divide-[#2D2D2D]">
                {selectedTx.items.map((item, idx) => (
                  <div key={idx} className="py-2.5 flex justify-between text-xs text-gray-700 dark:text-gray-300">
                    <div>
                      <span className="font-bold block text-gray-800 dark:text-white">{item.product_name}</span>
                      <span className="text-gray-400 dark:text-gray-500">{item.quantity} x {formatRupiah(item.unit_price)}</span>
                    </div>
                    <span className="font-bold text-gray-800 dark:text-white flex items-center">{formatRupiah(item.subtotal)}</span>
                  </div>
                ))}

                <div className="pt-3 mt-1 flex justify-between font-black text-sm text-gray-900 dark:text-white border-t border-gray-200 dark:border-[#3F3F3F]">
                  <span>TOTAL PEMBAYARAN</span>
                  <span className="text-[#CA400A]">{formatRupiah(selectedTx.total_amount)}</span>
                </div>
              </div>
            </div>

            {selectedTx.note && (
              <div className="p-3 rounded-lg border border-gray-100 dark:border-[#2D2D2D] bg-gray-50 dark:bg-[#212121] text-xs">
                <span className="font-bold block text-gray-400 dark:text-gray-500 uppercase tracking-wider text-[10px]">Catatan Kasir:</span>
                <p className="text-gray-600 dark:text-gray-300 mt-0.5 italic">"{selectedTx.note}"</p>
              </div>
            )}

            {selectedTx.status === 'VOIDED' && (
              <div className="p-3 bg-red-50 dark:bg-red-950/20 rounded-lg border border-red-200 dark:border-red-900/30 text-xs space-y-1">
                <span className="font-bold block text-red-600 dark:text-red-400 uppercase tracking-wider text-[10px]">Alasan Pembatalan (Audit Log):</span>
                <p className="text-red-700 dark:text-red-300 font-semibold">"{selectedTx.void_reason}"</p>
                {selectedTx.voided_at && (
                  <span className="text-[10px] text-gray-400 dark:text-gray-500 block">Dibatalkan pada: {formatDate(selectedTx.voided_at)}</span>
                )}
              </div>
            )}

            <div className="flex flex-col sm:flex-row gap-3 pt-4 border-t border-gray-100 dark:border-[#242424]">
              {selectedTx.status !== 'VOIDED' && (
                <button
                  onClick={() => setShowVoidDialog(true)}
                  className="w-full bg-red-50 hover:bg-red-100 dark:bg-red-950/40 dark:hover:bg-red-900/50 text-red-700 dark:text-red-400 text-xs font-bold py-3.5 rounded-lg uppercase tracking-wider border border-red-200 dark:border-red-900/30 transition-colors h-12 flex items-center justify-center"
                >
                  Void Transaksi
                </button>
              )}
              <button
                onClick={() => setSelectedTx(null)}
                className="w-full bg-gray-100 hover:bg-gray-200 dark:bg-[#242424] dark:hover:bg-[#2F2F2F] text-gray-800 dark:text-white border border-gray-300 dark:border-[#2D2D2D] text-xs font-bold py-3.5 rounded-lg uppercase tracking-wider transition-colors h-12 flex items-center justify-center"
              >
                Tutup Detail
              </button>
            </div>

          </div>
        </div>
      )}

      {/* VOID CONFIRMATION DIALOG */}
      {showVoidDialog && (
        <div className="fixed inset-0 bg-black/95 flex items-end sm:items-center justify-center p-0 sm:p-4 z-[60] animate-fadeIn">
          <div className="bg-white dark:bg-[#161616] w-full sm:max-w-md rounded-t-2xl sm:rounded-2xl border border-red-200 dark:border-red-900/40 p-6 space-y-4 shadow-2xl">
            <h4 className="font-black text-red-600 dark:text-red-400 text-md uppercase tracking-wider">Konfirmasi Void Transaksi</h4>
            <p className="text-xs text-gray-600 dark:text-gray-300 leading-relaxed">
              Anda akan membatalkan transaksi secara permanen. Tindakan ini akan mengurangkan data pendapatan harian serta mengembalikan log konsumsi bahan baku.
            </p>

            <div>
              <label className="block text-[10px] font-bold text-gray-500 uppercase tracking-wider mb-1.5">
                Alasan Pembatalan (Wajib Diisi)
              </label>
              <textarea
                value={voidReason}
                onChange={(e) => setVoidReason(e.target.value)}
                placeholder="Contoh: Salah pilih metode pembayaran oleh kasir"
                className="w-full h-24 border focus:border-red-500 focus:ring-1 focus:ring-red-500 text-xs p-3 rounded-lg focus:outline-none transition-all resize-none bg-white dark:bg-[#242424] border-gray-300 dark:border-[#383838]"
                required
              />
            </div>

            <div className="flex flex-col sm:flex-row gap-3 pt-2">
              <button
                onClick={handleConfirmVoid}
                className="w-full bg-[#CA400A] hover:bg-[#E04E15] text-white text-xs font-bold py-3.5 rounded-lg uppercase tracking-wider h-12 flex items-center justify-center"
              >
                Ya, Void Transaksi
              </button>
              <button
                onClick={() => {
                  setShowVoidDialog(false);
                  setVoidReason('');
                }}
                className="w-full bg-gray-100 hover:bg-gray-200 dark:bg-[#242424] dark:hover:bg-[#2F2F2F] text-gray-800 dark:text-white border border-gray-300 dark:border-[#2D2D2D] text-xs font-bold py-3.5 rounded-lg uppercase tracking-wider h-12 flex items-center justify-center"
              >
                Batal
              </button>
            </div>
          </div>
        </div>
      )}

      {/* GLOBAL CUSTOM ALERT DIALOG */}
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
            <button
              onClick={() => setAlertInfo(null)}
              className="w-full bg-[#CA400A] text-white text-xs font-bold py-2.5 rounded-lg uppercase h-11 flex items-center justify-center tracking-wider focus:outline-none"
            >
              Mengerti
            </button>
          </div>
        </div>
      )}

    </div>
  );
}