'use client';

import React, { useState, useMemo } from 'react';
import { INITIAL_TRANSACTIONS, INITIAL_OPERATORS } from '../../lib/mockData';
import { Transaction } from '../../types';

export default function HistoryPage() {
  const [transactions, setTransactions] = useState<Transaction[]>(INITIAL_TRANSACTIONS);
  const [selectedTx, setSelectedTx] = useState<Transaction | null>(null);
  const [showVoidDialog, setShowVoidDialog] = useState(false);
  const [voidReason, setVoidReason] = useState('');
  
  // State Filter Pencarian & Kategori
  const [searchQuery, setSearchQuery] = useState('');
  const [filterStatus, setFilterStatus] = useState('ALL');
  const [filterPayment, setFilterPayment] = useState('ALL');
  const [filterOperator, setFilterOperator] = useState('ALL');

  // Custom alert state
  const [alertInfo, setAlertInfo] = useState<{ title: string; message: string } | null>(null);

  // ==========================================
  // PENYARINGAN DATA TRANSAKSI (REAL-TIME)
  // ==========================================
  const filteredTransactions = useMemo(() => {
    return transactions.filter(t => {
      const matchSearch = searchQuery === '' || 
        t.order_id.toLowerCase().includes(searchQuery.toLowerCase()) ||
        (t.note && t.note.toLowerCase().includes(searchQuery.toLowerCase()));
      
      const matchStatus = filterStatus === 'ALL' || t.status === filterStatus;
      const matchPayment = filterPayment === 'ALL' || t.payment_method === filterPayment;
      const matchOperator = filterOperator === 'ALL' || t.created_by_operator_id === filterOperator;

      return matchSearch && matchStatus && matchPayment && matchOperator;
    });
  }, [transactions, searchQuery, filterStatus, filterPayment, filterOperator]);

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

  const handleConfirmVoid = () => {
    if (!voidReason.trim()) {
      setAlertInfo({
        title: 'Validasi Gagal',
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
      message: `Transaksi ${targetTx.order_id} berhasil di-void.`
    });
  };

  return (
    <div className="p-4 md:p-8 space-y-6 animate-fadeIn text-gray-900 dark:text-white">
      
      {/* PANEL FILTER PENCARIAN DETAIL */}
      <div className="p-5 rounded-xl border border-gray-200 dark:border-[#242424] bg-white dark:bg-[#161616] space-y-4 shadow-sm">
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4">
          <div>
            <label className="block text-[11px] font-bold uppercase tracking-wider mb-1.5 text-gray-400 dark:text-gray-500">Cari Order ID / Catatan</label>
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Contoh: QM-102526"
              className="w-full border border-gray-300 dark:border-[#383838] bg-white dark:bg-[#242424] text-xs px-3 py-2 rounded-lg focus:outline-none transition-all h-10"
            />
          </div>

          <div>
            <label className="block text-[11px] font-bold uppercase tracking-wider mb-1.5 text-gray-400 dark:text-gray-500">Status</label>
            <select
              value={filterStatus}
              onChange={(e) => setFilterStatus(e.target.value)}
              className="w-full border border-gray-300 dark:border-[#383838] bg-white dark:bg-[#242424] text-xs px-3 py-2 rounded-lg focus:outline-none h-10"
            >
              <option value="ALL">SEMUA STATUS</option>
              <option value="CONFIRMED">CONFIRMED</option>
              <option value="PENDING">PENDING</option>
              <option value="VOIDED">VOIDED</option>
            </select>
          </div>

          <div>
            <label className="block text-[11px] font-bold uppercase tracking-wider mb-1.5 text-gray-400 dark:text-gray-500">Metode</label>
            <select
              value={filterPayment}
              onChange={(e) => setFilterPayment(e.target.value)}
              className="w-full border border-gray-300 dark:border-[#383838] bg-white dark:bg-[#242424] text-xs px-3 py-2 rounded-lg focus:outline-none h-10"
            >
              <option value="ALL">SEMUA METODE</option>
              <option value="CASH">CASH</option>
              <option value="QRIS_STATIC">QRIS STATIS</option>
              <option value="TRANSFER">TRANSFER</option>
            </select>
          </div>

          <div>
            <label className="block text-[11px] font-bold uppercase tracking-wider mb-1.5 text-gray-400 dark:text-gray-500">Operator</label>
            <select
              value={filterOperator}
              onChange={(e) => setFilterOperator(e.target.value)}
              className="w-full border border-gray-300 dark:border-[#383838] bg-white dark:bg-[#242424] text-xs px-3 py-2 rounded-lg focus:outline-none h-10"
            >
              <option value="ALL">SEMUA KASIR</option>
              {INITIAL_OPERATORS.map(op => (
                <option key={op.id} value={op.id}>{op.name}</option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {/* TABEL LOG RIWAYAT TRANSAKSI */}
      <div className="rounded-xl border border-gray-200 dark:border-[#242424] bg-white dark:bg-[#161616] overflow-hidden shadow-sm">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm min-w-[700px]">
            <thead className="border-b border-gray-200 dark:border-[#2D2D2D] text-[10px] uppercase font-bold text-gray-400 dark:text-gray-500 tracking-wider bg-gray-50 dark:bg-[#212121]">
              <tr>
                <th className="p-4">Order ID</th>
                <th className="p-4">Waktu Transaksi</th>
                <th className="p-4">Operator Kasir</th>
                <th className="p-4 text-right">Total Amount</th>
                <th className="p-4">Metode Bayar</th>
                <th className="p-4">Status</th>
                <th className="p-4 text-center">Aksi</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 dark:divide-[#242424]">
              {filteredTransactions.map((tItem) => (
                <tr key={tItem.id} className="transition-colors hover:bg-gray-50/50 dark:hover:bg-[#1C1C1C]">
                  <td className="p-4 font-bold text-[#CA400A]">{tItem.order_id}</td>
                  <td className="p-4 text-xs text-gray-500 dark:text-gray-400">{formatDate(tItem.created_at)}</td>
                  <td className="p-4 font-semibold text-gray-800 dark:text-white">{tItem.created_by_operator_name}</td>
                  <td className="p-4 text-right font-bold text-gray-800 dark:text-white">{formatRupiah(tItem.total_amount)}</td>
                  <td className="p-4">
                    <span className="text-[10px] px-2.5 py-1 rounded font-bold uppercase border border-gray-200 dark:border-[#313131] bg-gray-100 dark:bg-[#242424] text-gray-600 dark:text-gray-400">
                      {tItem.payment_method || 'PENDING'}
                    </span>
                  </td>
                  <td className="p-4">
                    <span className={`text-[10px] uppercase px-2.5 py-1 rounded font-extrabold ${
                      tItem.status === 'CONFIRMED' ? 'bg-emerald-50 dark:bg-emerald-950/40 text-emerald-700 dark:text-emerald-400 border border-emerald-200 dark:border-emerald-900/30' :
                      tItem.status === 'PENDING' ? 'bg-amber-50 dark:bg-amber-950/40 text-amber-700 dark:text-amber-400 border border-amber-200 dark:border-amber-900/30' :
                      'bg-red-50 dark:bg-red-950/40 text-red-700 dark:text-red-400 border border-red-200 dark:border-red-900/30'
                    }`}>
                      {tItem.status}
                    </span>
                  </td>
                  <td className="p-4 text-center">
                    <button
                      onClick={() => setSelectedTx(tItem)}
                      className="text-xs font-bold px-3.5 py-1.5 rounded-lg bg-gray-100 hover:bg-gray-200 dark:bg-[#242424] dark:hover:bg-[#2F2F2F] text-gray-800 dark:text-white border border-gray-300 dark:border-[#2D2D2D] transition-all h-9"
                    >
                      DETAIL
                    </button>
                  </td>
                </tr>
              ))}
              {filteredTransactions.length === 0 && (
                <tr>
                  <td colSpan={7} className="text-center py-12 text-gray-400 dark:text-gray-500 font-medium">
                    Tidak ada data transaksi yang cocok dengan filter pencarian.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* ==========================================
          MODAL DETAIL & VOID DIALOGS
          ========================================== */}
      {selectedTx && (
        <div className="fixed inset-0 bg-black/80 flex items-end sm:items-center justify-center p-0 sm:p-4 z-50 animate-fadeIn">
          <div className="bg-white dark:bg-[#161616] w-full sm:max-w-xl rounded-t-2xl sm:rounded-2xl p-6 space-y-6 shadow-2xl overflow-y-auto max-h-[90vh] border-t sm:border border-gray-200 dark:border-[#2C2C2C]">
            <div className="flex items-center justify-between pb-4 border-b border-gray-100 dark:border-[#242424]">
              <div>
                <h3 className="font-extrabold text-md text-[#CA400A] uppercase tracking-wider">Detail Log Transaksi</h3>
                <span className="text-xs text-gray-400 dark:text-gray-500 font-bold">ID: {selectedTx.order_id}</span>
              </div>
              <button onClick={() => setSelectedTx(null)} className="font-extrabold text-xl p-2 -mr-2 text-gray-400 dark:text-gray-500 hover:text-white">✕</button>
            </div>

            <div className="grid grid-cols-2 gap-4 text-xs text-gray-500 dark:text-gray-400">
              <div>Waktu Dibuat: <span className="font-semibold block text-gray-800 dark:text-white">{formatDate(selectedTx.created_at)}</span></div>
              <div>Metode Bayar: <span className="font-semibold block uppercase text-gray-800 dark:text-white">{selectedTx.payment_method || 'PENDING'}</span></div>
              <div>Kasir Operator: <span className="font-semibold block text-gray-800 dark:text-white">{selectedTx.created_by_operator_name}</span></div>
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
                  <span>TOTAL AKUMULASI</span>
                  <span className="text-[#CA400A]">{formatRupiah(selectedTx.total_amount)}</span>
                </div>
              </div>
            </div>

            {selectedTx.note && (
              <div className="p-3 rounded-lg border border-gray-100 dark:border-[#2D2D2D] bg-gray-50 dark:bg-[#212121] text-xs">
                <span className="font-bold block text-gray-400 dark:text-gray-500 uppercase tracking-wider text-[10px]">Catatan:</span>
                <p className="text-gray-600 dark:text-gray-300 mt-0.5 italic">"{selectedTx.note}"</p>
              </div>
            )}

            {selectedTx.status === 'VOIDED' && (
              <div className="p-3 bg-red-50 dark:bg-red-950/20 rounded-lg border border-red-200 dark:border-red-900/30 text-xs space-y-1">
                <span className="font-bold block text-red-600 dark:text-red-400 uppercase tracking-wider text-[10px]">Alasan Void:</span>
                <p className="text-red-700 dark:text-red-300 font-semibold">"{selectedTx.void_reason}"</p>
              </div>
            )}

            <div className="flex flex-col sm:flex-row gap-3 pt-4 border-t border-gray-100 dark:border-[#242424]">
              {selectedTx.status !== 'VOIDED' && (
                <button
                  onClick={() => setShowVoidDialog(true)}
                  className="w-full bg-red-50 hover:bg-red-100 dark:bg-red-950/40 dark:hover:bg-red-900/50 text-red-700 dark:text-red-400 text-xs font-bold py-3.5 rounded-lg uppercase tracking-wider border border-red-200 dark:border-red-900/30 h-12 flex items-center justify-center"
                >
                  Void Transaksi
                </button>
              )}
              <button
                onClick={() => setSelectedTx(null)}
                className="w-full bg-gray-100 hover:bg-gray-200 dark:bg-[#242424] dark:hover:bg-[#2F2F2F] text-gray-800 dark:text-white border border-gray-300 dark:border-[#2D2D2D] text-xs font-bold py-3.5 rounded-lg uppercase tracking-wider h-12 flex items-center justify-center"
              >
                Tutup Detail
              </button>
            </div>
          </div>
        </div>
      )}

      {showVoidDialog && (
        <div className="fixed inset-0 bg-black/95 flex items-end sm:items-center justify-center p-0 sm:p-4 z-[60] animate-fadeIn">
          <div className="bg-white dark:bg-[#161616] w-full sm:max-w-md rounded-t-2xl sm:rounded-2xl border border-red-200 dark:border-red-900/40 p-6 space-y-4 shadow-2xl">
            <h4 className="font-black text-red-600 dark:text-red-400 text-md uppercase tracking-wider">Konfirmasi Void Transaksi</h4>
            <p className="text-xs text-gray-600 dark:text-gray-300 leading-relaxed">
              Anda akan membatalkan transaksi ini secara permanen dari buku besar finansial outlet.
            </p>
            <div>
              <label className="block text-[10px] font-bold text-gray-500 uppercase tracking-wider mb-1.5">Alasan Pembatalan (Wajib)</label>
              <textarea
                value={voidReason}
                onChange={(e) => setVoidReason(e.target.value)}
                placeholder="Contoh: Salah input pesanan"
                className="w-full h-24 border focus:border-red-500 focus:ring-1 focus:ring-red-500 text-xs p-3 rounded-lg focus:outline-none transition-all resize-none bg-white dark:bg-[#242424] border-gray-300 dark:border-[#383838]"
                required
              />
            </div>
            <div className="flex flex-col sm:flex-row gap-3 pt-2">
              <button onClick={handleConfirmVoid} className="w-full bg-[#CA400A] hover:bg-[#E04E15] text-white text-xs font-bold py-3.5 rounded-lg uppercase tracking-wider h-12 flex items-center justify-center">
                Ya, Void Transaksi
              </button>
              <button onClick={() => { setShowVoidDialog(false); setVoidReason(''); }} className="w-full bg-gray-100 hover:bg-gray-200 dark:bg-[#242424] text-gray-800 dark:text-white border border-gray-300 text-xs font-bold py-3.5 rounded-lg h-12 flex items-center justify-center">
                Batal
              </button>
            </div>
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