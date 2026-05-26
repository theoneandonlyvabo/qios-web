'use client';

import React, { useState } from 'react';
import { INITIAL_OPERATORS } from '../../lib/mockData';
import { Operator } from '../../types';

export default function OperatorsPage() {
  const [operators, setOperators] = useState<Operator[]>(INITIAL_OPERATORS);
  const [showAddModal, setShowAddModal] = useState(false);
  const [showQrModal, setShowQrModal] = useState<Operator | null>(null);
  const [alertInfo, setAlertInfo] = useState<{ title: string; message: string } | null>(null);

  // Form State untuk Tambah Operator Baru
  const [newOpName, setNewOpName] = useState('');
  const [newOpCode, setNewOpCode] = useState('');
  const [newOpPassword, setNewOpPassword] = useState('');

  // Konfigurasi Lisensi Maksimal Operator (Sesuai paket Enterprise = 5)
  const maxOperators = 5;

  // Handler Aktivasi/Deaktivasi Operator Kasir
  const handleToggleStatus = (id: string) => {
    setOperators(prev =>
      prev.map(op => {
        if (op.id === id) {
          return { ...op, is_active: !op.is_active };
        }
        return op;
      })
    );
  };

  // Handler Hapus Operator
  const handleDeleteOperator = (id: string, name: string) => {
    setOperators(prev => prev.filter(op => op.id !== id));
    setAlertInfo({
      title: 'Operator Dihapus',
      message: `Akun kasir "${name}" telah berhasil dihapus dari sistem QIOS.`
    });
  };

  // Handler Regenerasi Token Login QR Code
  const handleRegenerateQr = (id: string) => {
    setOperators(prev =>
      prev.map(op => {
        if (op.id === id) {
          const newToken = `QIOS-OP-QR-REGEN-${Math.random().toString(36).substring(2, 10).toUpperCase()}`;
          const updatedOp = { ...op, qr_token: newToken };
          if (showQrModal?.id === id) {
            setShowQrModal(updatedOp);
          }
          return updatedOp;
        }
        return op;
      })
    );
  };

  // Handler Kirim Form Tambah Operator Baru
  const handleAddSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!newOpName.trim() || !newOpCode.trim() || !newOpPassword.trim()) {
      setAlertInfo({
        title: 'Formulir Tidak Lengkap',
        message: 'Semua kolom isian wajib diisi.'
      });
      return;
    }

    // Validasi Limit Kuota Lisensi
    if (operators.length >= maxOperators) {
      setAlertInfo({
        title: 'Batas Lisensi Tercapai',
        message: `Lisensi paket bisnis Anda membatasi maksimal ${maxOperators} operator kasir aktif. Silakan hubungi tim Skalar Solutions untuk melakukan upgrade kuota.`
      });
      return;
    }

    // Validasi Duplikasi Kode Kasir
    const isCodeExists = operators.some(op => op.operator_code.toLowerCase() === newOpCode.toLowerCase());
    if (isCodeExists) {
      setAlertInfo({
        title: 'Kode Operator Duplikat',
        message: `Kode "${newOpCode}" sudah digunakan oleh operator lain. Gunakan kode yang unik.`
      });
      return;
    }

    const newOperator: Operator = {
      id: `op-${Date.now()}`,
      name: newOpName,
      operator_code: newOpCode.toLowerCase(),
      is_active: true,
      qr_token: `QIOS-OP-QR-${newOpCode.toUpperCase()}-${Math.random().toString(36).substring(2, 7).toUpperCase()}`,
      created_at: new Date().toISOString()
    };

    setOperators(prev => [...prev, newOperator]);
    setShowAddModal(false);
    setNewOpName('');
    setNewOpCode('');
    setNewOpPassword('');
    
    setAlertInfo({
      title: 'Kasir Berhasil Dibuat',
      message: `Operator kasir "${newOpName}" telah terdaftar dan siap melayani transaksi.`
    });
  };

  return (
    <div className="p-4 md:p-8 space-y-6 animate-fadeIn text-gray-900 dark:text-white">
      
      {/* ATAS: KUOTA LISENSI DAN TRIGGER TAMBAH */}
      <div className="p-5 rounded-xl border border-gray-200 dark:border-[#242424] bg-white dark:bg-[#161616] flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 shadow-sm">
        <div>
          <h3 className="font-extrabold text-sm md:text-base uppercase tracking-wider text-gray-800 dark:text-white">
            Kuota Lisensi Operator Aktif
          </h3>
          <p className="text-xs md:text-sm text-gray-400 dark:text-gray-500 mt-1">
            Terpakai: <span className="font-extrabold text-gray-800 dark:text-white">{operators.length}</span> dari maksimal <span className="font-extrabold text-gray-800 dark:text-white">{maxOperators} kasir</span>.
          </p>
        </div>

        <button
          onClick={() => {
            if (operators.length >= maxOperators) {
              setAlertInfo({
                title: 'Batas Kuota Lisensi',
                message: `Anda telah mencapai batas kuota maksimal (${maxOperators} operator). Hubungi Account Manager Skalar untuk membeli slot lisensi tambahan.`
              });
              return;
            }
            setShowAddModal(true);
          }}
          className="bg-[#CA400A] hover:bg-[#E04E15] text-xs text-white px-5 py-3 rounded-xl font-bold uppercase transition-all shadow-md shadow-[#CA400A]/10 w-full sm:w-auto text-center h-12 flex items-center justify-center"
        >
          + Tambah Kasir Baru
        </button>
      </div>

      {/* GRID KARTU DAFTAR OPERATOR */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
        {operators.map((op) => (
          <div key={op.id} className="rounded-xl border border-gray-200 dark:border-[#242424] bg-white dark:bg-[#161616] p-5 md:p-6 space-y-4 shadow-sm flex flex-col justify-between">
            
            <div className="space-y-4">
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-3">
                  <div className="w-12 h-12 rounded-xl text-[#CA400A] bg-[#CA400A]/10 font-black text-lg flex items-center justify-center border border-[#CA400A]/20">
                    {op.name.charAt(0).toUpperCase()}
                  </div>
                  <div>
                    <span className="font-bold text-sm block text-gray-800 dark:text-white">{op.name}</span>
                    <span className="text-[10px] font-bold uppercase tracking-wider text-gray-400 dark:text-gray-500 block mt-0.5">
                      Kode: {op.operator_code}
                    </span>
                  </div>
                </div>

                {/* Sakelar Toggle Status Keaktifan Kasir */}
                <button
                  onClick={() => handleToggleStatus(op.id)}
                  className={`w-11 h-6 rounded-full transition-colors relative focus:outline-none ${
                    op.is_active ? 'bg-[#CA400A]' : 'bg-gray-300 dark:bg-[#333333]'
                  }`}
                >
                  <span className={`w-4 h-4 bg-white rounded-full absolute top-1 transition-all shadow ${
                    op.is_active ? 'right-1' : 'left-1'
                  }`}></span>
                </button>
              </div>

              <div className="text-xs pt-3 border-t border-gray-100 dark:border-[#242424] flex justify-between text-gray-400 dark:text-gray-500 font-semibold">
                <span>Terdaftar sejak:</span>
                <span className="text-gray-700 dark:text-gray-300">
                  {new Date(op.created_at).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })}
                </span>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-3 pt-4 border-t border-gray-100 dark:border-[#242424] mt-2">
              <button
                onClick={() => setShowQrModal(op)}
                className="text-xs font-bold py-2.5 rounded-lg text-center transition-colors uppercase h-10 flex items-center justify-center bg-gray-100 hover:bg-gray-200 dark:bg-[#242424] dark:hover:bg-[#2F2F2F] text-gray-800 dark:text-white border border-gray-300 dark:border-[#2D2D2D]"
              >
                Scan QR
              </button>
              <button
                onClick={() => handleDeleteOperator(op.id, op.name)}
                className="bg-red-50 hover:bg-red-100 dark:bg-red-950/20 dark:hover:bg-red-900/40 text-red-700 dark:text-red-400 text-xs font-bold py-2.5 rounded-lg text-center transition-colors uppercase h-10 flex items-center justify-center border border-red-200 dark:border-red-900/30"
              >
                Hapus
              </button>
            </div>

          </div>
        ))}
      </div>

      {/* ==========================================
          MODALS & LAYERS
          ========================================== */}

      {/* MODAL: TAMBAH OPERATOR BARU */}
      {showAddModal && (
        <div className="fixed inset-0 bg-black/80 flex items-end sm:items-center justify-center p-0 sm:p-4 z-50 animate-fadeIn">
          <div className="bg-white dark:bg-[#161616] w-full sm:max-w-md rounded-t-2xl sm:rounded-2xl p-6 space-y-4 shadow-2xl border border-gray-200 dark:border-[#2C2C2C]">
            <div className="flex justify-between items-center pb-3 border-b border-gray-100 dark:border-[#242424]">
              <h3 className="font-extrabold text-sm uppercase tracking-wider text-gray-800 dark:text-white">Tambah Operator Kasir</h3>
              <button onClick={() => setShowAddModal(false)} className="text-lg p-2 text-gray-400 dark:text-gray-500 hover:text-gray-800 dark:hover:text-white">✕</button>
            </div>

            <form onSubmit={handleAddSubmit} className="space-y-4">
              <div className="space-y-1.5">
                <label className="block text-[10px] font-bold uppercase tracking-wider text-gray-400 dark:text-gray-500">Nama Kasir</label>
                <input
                  type="text"
                  value={newOpName}
                  onChange={(e) => setNewOpName(e.target.value)}
                  placeholder="Contoh: Siti Aisyah"
                  className="w-full border border-gray-300 dark:border-[#383838] bg-white dark:bg-[#242424] text-xs px-3 py-2.5 rounded-lg focus:outline-none h-11 text-gray-800 dark:text-white focus:border-[#CA400A]"
                  required
                />
              </div>

              <div className="space-y-1.5">
                <label className="block text-[10px] font-bold uppercase tracking-wider text-gray-400 dark:text-gray-500">Kode Operator (Unik)</label>
                <input
                  type="text"
                  value={newOpCode}
                  onChange={(e) => setNewOpCode(e.target.value)}
                  placeholder="Contoh: kasir-siti"
                  className="w-full border border-gray-300 dark:border-[#383838] bg-white dark:bg-[#242424] text-xs px-3 py-2.5 rounded-lg focus:outline-none h-11 text-gray-800 dark:text-white focus:border-[#CA400A]"
                  required
                />
              </div>

              <div className="space-y-1.5">
                <label className="block text-[10px] font-bold uppercase tracking-wider text-gray-400 dark:text-gray-500">Kata Sandi Akses</label>
                <input
                  type="password"
                  value={newOpPassword}
                  onChange={(e) => setNewOpPassword(e.target.value)}
                  placeholder="Password rahasia untuk otentikasi kasir"
                  className="w-full border border-gray-300 dark:border-[#383838] bg-white dark:bg-[#242424] text-xs px-3 py-2.5 rounded-lg focus:outline-none h-11 text-gray-800 dark:text-white focus:border-[#CA400A]"
                  required
                />
              </div>

              <button
                type="submit"
                className="w-full bg-[#CA400A] hover:bg-[#E04E15] text-white text-xs font-bold py-3.5 rounded-lg uppercase tracking-wider h-12 flex items-center justify-center transition-colors"
              >
                Konfirmasi Tambah
              </button>
            </form>
          </div>
        </div>
      )}

      {/* MODAL: QR CODE LOGIN */}
      {showQrModal && (
        <div className="fixed inset-0 bg-black/85 flex items-end sm:items-center justify-center p-0 sm:p-4 z-50 animate-fadeIn">
          <div className="bg-white dark:bg-[#161616] w-full sm:max-w-sm rounded-t-2xl sm:rounded-2xl p-6 space-y-6 shadow-2xl text-center border border-gray-200 dark:border-[#2C2C2C]">
            <div>
              <h3 className="font-extrabold text-sm uppercase tracking-wider text-gray-800 dark:text-white">QR Code Login Operator</h3>
              <p className="text-[11px] text-gray-400 dark:text-gray-500 mt-1">Arahkan kamera tablet kasir ke layar ini untuk masuk instan</p>
            </div>

            {/* Representasi Visual Mock QR Code */}
            <div className="w-48 h-48 bg-white p-3 rounded-xl mx-auto flex flex-col justify-between items-center border border-gray-200 shadow-sm relative">
              <div className="w-full h-full border-4 border-black/5 p-2 flex flex-col justify-between">
                <div className="flex justify-between w-full">
                  <div className="w-10 h-10 bg-black rounded"></div>
                  <div className="w-10 h-10 bg-black rounded"></div>
                </div>
                <span className="text-[9px] font-mono font-bold text-black uppercase tracking-widest break-all">
                  {showQrModal.qr_token.substring(0, 15)}
                </span>
                <div className="flex justify-between w-full items-end">
                  <div className="w-10 h-10 bg-black rounded"></div>
                  <div className="w-6 h-6 bg-[#CA400A] rounded animate-pulse"></div>
                </div>
              </div>
            </div>

            <div className="space-y-2">
              <span className="text-xs font-bold block text-gray-800 dark:text-white">{showQrModal.name}</span>
              <span className="text-[10px] text-gray-500 dark:text-gray-400 font-mono break-all bg-gray-50 dark:bg-[#212121] px-2 py-2 rounded border border-gray-200 dark:border-[#2D2D2D] block">
                Token: {showQrModal.qr_token}
              </span>
            </div>

            <div className="flex flex-col gap-3">
              <button
                onClick={() => handleRegenerateQr(showQrModal.id)}
                className="w-full bg-amber-50 hover:bg-amber-100 dark:bg-amber-950/30 dark:hover:bg-amber-900/40 text-amber-700 dark:text-amber-500 text-xs font-bold py-3 rounded-lg border border-amber-200 dark:border-amber-900/20 uppercase tracking-wider h-11 flex items-center justify-center"
              >
                Regenerasi QR
              </button>
              <button
                onClick={() => setShowQrModal(null)}
                className="w-full bg-gray-100 hover:bg-gray-200 dark:bg-[#242424] dark:hover:bg-[#2F2F2F] text-gray-800 dark:text-white border border-gray-300 dark:border-[#2D2D2D] text-xs font-bold py-3 rounded-lg uppercase tracking-wider h-11 flex items-center justify-center"
              >
                Selesai
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ALERT DIALOG UTAMA */}
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
              className="w-full bg-[#CA400A] text-white text-xs font-bold py-2.5 rounded-lg uppercase h-11 flex items-center justify-center"
            >
              Mengerti
            </button>
          </div>
        </div>
      )}

    </div>
  );
}