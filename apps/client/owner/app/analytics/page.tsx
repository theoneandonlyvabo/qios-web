'use client';

import React, { useState } from 'react';
import { INITIAL_INSIGHTS } from '../../lib/mockData';

export default function AnalyticsPage() {
  const [expandedInsight, setExpandedInsight] = useState<string | null>(null);

  return (
    <div className="p-4 md:p-8 space-y-6 animate-fadeIn text-gray-900 dark:text-white">
      
      {/* HEADER BANNER AI */}
      <div className="p-5 md:p-6 rounded-xl border border-gray-200 dark:border-[#CA400A]/20 bg-gradient-to-r from-orange-50/60 to-white dark:from-[#1E120D] dark:to-[#161616] shadow-sm">
        <div className="flex flex-col sm:flex-row items-start gap-4">
          <div className="w-12 h-12 rounded-lg bg-[#CA400A]/10 text-[#CA400A] flex items-center justify-center shrink-0 border border-[#CA400A]/20">
            <svg className="w-6 h-6 animate-pulse" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
          </div>
          <div className="space-y-1">
            <h3 className="font-extrabold text-base uppercase tracking-wider text-gray-800 dark:text-white">
              Sistem Analisis Rule-Based AI QIOS (v1.0)
            </h3>
            <p className="text-xs md:text-sm text-gray-500 dark:text-gray-400 leading-relaxed">
              Platform membaca tren transaksi outlet Warung Kopi Senja secara otonom. Tidak ada chatbot berbelit-belit; kami menyajikan pola operasional dan anomali secara terstruktur sehingga Anda siap mengambil keputusan bisnis taktis dengan akurasi tinggi.
            </p>
          </div>
        </div>
      </div>

      {/* INSIGHT CARDS GRID */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {INITIAL_INSIGHTS.map((card) => (
          <div 
            key={card.id} 
            className="rounded-xl border border-gray-200 dark:border-[#242424] bg-white dark:bg-[#161616] p-5 md:p-6 flex flex-col justify-between hover:border-[#CA400A]/30 transition-all shadow-sm"
          >
            <div>
              <div className="flex items-center justify-between pb-4 border-b border-gray-100 dark:border-[#242424] mb-4">
                <div className="flex items-center gap-2.5">
                  {/* VEKTOR IKON ADAPTIF */}
                  {card.type === 'trend' && (
                    <svg className="w-5 h-5 text-[#CA400A]" fill="none" stroke="currentColor" strokeWidth="2.5" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
                    </svg>
                  )}
                  {card.type === 'consumption' && (
                    <svg className="w-5 h-5 text-[#CA400A]" fill="none" stroke="currentColor" strokeWidth="2.5" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                    </svg>
                  )}
                  {card.type === 'opportunity' && (
                    <svg className="w-5 h-5 text-emerald-500" fill="none" stroke="currentColor" strokeWidth="2.5" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
                    </svg>
                  )}
                  {card.type === 'warning' && (
                    <svg className="w-5 h-5 text-amber-500" fill="none" stroke="currentColor" strokeWidth="2.5" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                    </svg>
                  )}
                  <span className="font-extrabold text-sm uppercase tracking-wide text-gray-800 dark:text-white">
                    {card.title}
                  </span>
                </div>
                <span className="text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-widest">
                  Active
                </span>
              </div>

              <p className="text-xs md:text-sm leading-relaxed mb-6 text-gray-600 dark:text-gray-300">
                {card.narrative}
              </p>
            </div>

            <div className="space-y-3">
              {expandedInsight === card.id ? (
                <div className="p-4 rounded-lg border border-gray-100 dark:border-[#2D2D2D] bg-gray-50 dark:bg-[#212121] animate-fadeIn text-xs space-y-2">
                  <span className="font-bold block uppercase tracking-wider text-[10px] text-gray-800 dark:text-white">
                    Metadata Validasi Agregasi:
                  </span>
                  <div className="grid grid-cols-2 gap-2 text-gray-500 dark:text-gray-400 font-medium">
                    <div>Mulai: <span className="text-gray-700 dark:text-gray-300 font-semibold">{card.source_data_window.start_date}</span></div>
                    <div>Akhir: <span className="text-gray-700 dark:text-gray-300 font-semibold">{card.source_data_window.end_date}</span></div>
                  </div>
                  <p className="text-[#CA400A] font-bold text-[11px] pt-1">
                    Rule Confidence Score: 96.5%
                  </p>
                  <button 
                    onClick={() => setExpandedInsight(null)}
                    className="w-full text-center mt-2 pt-2 border-t border-gray-200 dark:border-[#3A3A3A] font-bold text-[11px] uppercase tracking-wider text-gray-400 dark:text-gray-500 hover:text-[#CA400A]"
                  >
                    Tutup Detail
                  </button>
                </div>
              ) : (
                <button
                  onClick={() => setExpandedInsight(card.id)}
                  className="w-full font-bold text-xs py-3 px-4 rounded-lg transition-colors h-11 text-center uppercase tracking-wider bg-gray-100 hover:bg-gray-200 dark:bg-[#242424] dark:hover:bg-[#2F2F2F] text-gray-800 dark:text-white border border-gray-300 dark:border-[#2D2D2D]"
                >
                  Buka Data Pendukung
                </button>
              )}

              <div className="flex justify-between items-center pt-3 border-t border-gray-100 dark:border-[#1E1E1E] text-[10px] text-gray-400 dark:text-gray-500 font-semibold">
                <span>Metrik Analisis: 30 Hari</span>
                <span>Diperbarui Hari Ini</span>
              </div>
            </div>

          </div>
        ))}
      </div>

    </div>
  );
}