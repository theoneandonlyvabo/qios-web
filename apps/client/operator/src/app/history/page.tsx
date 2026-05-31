"use client";

import { useMemo, useState, type ReactNode } from "react";
import { Icon } from "@/components/icons";
import { OperatorShell } from "@/components/operator-shell";
import { StatusBadge } from "@/components/status-badge";
import { useOperator } from "@/components/providers/operator-provider";
import { formatCurrency, formatShortDate, formatTime, paymentLabel } from "@/lib/format";
import type { Transaction } from "@/types";

export default function HistoryPage() {
  const { transactions } = useOperator();
  const [selected, setSelected] = useState<Transaction | null>(null);

  const summary = useMemo(() => {
    const confirmed = transactions.filter((transaction) => transaction.status === "CONFIRMED");
    return {
      count: transactions.length,
      revenue: confirmed.reduce((total, transaction) => total + transaction.total, 0)
    };
  }, [transactions]);

  return (
    <OperatorShell>
      <header className="sticky top-0 z-40 flex h-16 items-center justify-between border-b border-border bg-surface/90 px-4 backdrop-blur-xl safe-pt">
        <div className="flex items-center gap-2">
          <Icon name="menu" className="h-5 w-5 text-primary" />
          <p className="text-lg font-extrabold text-foreground">QIOS Business</p>
        </div>
        <div className="flex items-center gap-3">
          <div className="flex h-8 w-8 items-center justify-center rounded-full border border-border bg-card-high text-muted-foreground">
            <Icon name="profile" className="h-4 w-4" />
          </div>
          <span className="h-3 w-3 rounded-full bg-primary" />
        </div>
      </header>

      <main className="space-y-6 px-4 pb-28 pt-5">
        <section>
          <h1 className="text-xl font-extrabold text-foreground">Riwayat Hari Ini</h1>
          <p className="mt-1 text-sm font-medium text-primary">{formatShortDate(new Date("2026-05-05T10:00:00+07:00"))}</p>
        </section>

        <section className="grid grid-cols-2 gap-3">
          <SummaryCard icon="receipt" label="Total Transaksi" value={String(summary.count)} />
          <SummaryCard icon="cash" label="Total Omzet" value={formatCurrency(summary.revenue)} accent />
        </section>

        <section>
          <div className="mb-3 flex items-end justify-between">
            <h2 className="text-lg font-extrabold text-foreground">Daftar Transaksi</h2>
            <button className="text-xs font-extrabold text-primary" type="button">Lihat Semua</button>
          </div>

          <div className="overflow-hidden rounded-2xl border border-border bg-card">
            {transactions.map((transaction) => (
              <button
                key={transaction.id}
                onClick={() => setSelected(transaction)}
                className="flex w-full items-start justify-between gap-3 border-b border-border p-4 text-left transition last:border-b-0 hover:bg-card-high active:scale-[0.99]"
                type="button"
              >
                <div className="flex min-w-0 items-center gap-3">
                  <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-full bg-card-highest text-primary">
                    <Icon name={transaction.paymentMethod === "QRIS_STATIC" ? "qr" : transaction.paymentMethod === "CASH" ? "cash" : "bank"} className="h-5 w-5" />
                  </div>
                  <div className="min-w-0">
                    <p className="truncate text-lg font-extrabold text-foreground">{transaction.orderId}</p>
                    <p className="text-xs font-extrabold text-primary">{formatTime(transaction.createdAt)} • {paymentLabel(transaction.paymentMethod)}</p>
                  </div>
                </div>
                <div className="shrink-0 text-right">
                  <p className="text-lg font-extrabold text-foreground">{formatCurrency(transaction.total)}</p>
                  <div className="mt-1 flex justify-end">
                    <StatusBadge status={transaction.status} />
                  </div>
                </div>
              </button>
            ))}
          </div>
        </section>
      </main>

      {selected && <TransactionSheet transaction={selected} onClose={() => setSelected(null)} />}
    </OperatorShell>
  );
}

function SummaryCard({ icon, label, value, accent = false }: { icon: "receipt" | "cash"; label: string; value: string; accent?: boolean }) {
  return (
    <article className="relative overflow-hidden rounded-2xl border border-border bg-card-high p-4 shadow-soft">
      <div className="absolute -right-5 -top-5 h-20 w-20 rounded-full bg-primary/10 blur-2xl" />
      <div className="relative flex items-center gap-1.5 text-primary">
        <Icon name={icon} className="h-4 w-4" />
        <p className="text-xs font-extrabold">{label}</p>
      </div>
      <p className={`relative mt-3 font-black ${accent ? "text-lg text-primary" : "text-2xl text-foreground"}`}>{value}</p>
    </article>
  );
}

function TransactionSheet({ transaction, onClose }: { transaction: Transaction; onClose: () => void }) {
  const { voidTransaction } = useOperator();
  const [showVoid, setShowVoid] = useState(false);
  const [reason, setReason] = useState("");

  return (
    <div className="fixed inset-0 z-[80] flex items-end justify-center bg-black/60 px-0" onClick={onClose}>
      <div
        className="w-full max-w-[430px] rounded-t-[28px] border border-border bg-surface p-4 shadow-[0_-30px_100px_rgba(0,0,0,0.65)] safe-pb"
        onClick={(event) => event.stopPropagation()}
      >
        <div className="mx-auto mb-4 h-1.5 w-12 rounded-full bg-card-highest" />
        <div className="mb-4 flex items-start justify-between gap-4">
          <div>
            <p className="text-xs font-extrabold uppercase tracking-[0.18em] text-muted-foreground">Detail Transaksi</p>
            <h3 className="mt-1 text-xl font-black text-foreground">{transaction.orderId}</h3>
          </div>
          <button onClick={onClose} className="flex h-9 w-9 items-center justify-center rounded-full bg-card-high text-muted-foreground" type="button">
            <Icon name="close" className="h-4 w-4" />
          </button>
        </div>

        <div className="rounded-2xl border border-border bg-card p-4">
          <Info label="Status" value={<StatusBadge status={transaction.status} />} />
          <Info label="Metode" value={paymentLabel(transaction.paymentMethod)} />
          <Info label="Waktu" value={`${formatTime(transaction.createdAt)} WIB`} />
          <Info label="Kasir" value={transaction.createdBy} />
          <Info label="Total" value={formatCurrency(transaction.total)} last />
        </div>

        <div className="mt-4 rounded-2xl border border-border bg-card p-4">
          <p className="mb-3 text-sm font-extrabold text-foreground">Item Pesanan</p>
          {transaction.items.length === 0 ? (
            <p className="text-sm font-medium text-muted-foreground">Item snapshot tidak tersedia di data contoh.</p>
          ) : (
            <div className="space-y-3">
              {transaction.items.map((item) => (
                <div key={item.productId} className="flex justify-between gap-3 text-sm">
                  <div>
                    <p className="font-extrabold text-foreground">{item.productName}</p>
                    <p className="text-xs font-bold text-muted-foreground">{item.quantity} × {formatCurrency(item.unitPrice)}</p>
                  </div>
                  <p className="font-extrabold text-primary">{formatCurrency(item.subtotal)}</p>
                </div>
              ))}
            </div>
          )}
        </div>

        {transaction.status === "PENDING" && (
          <div className="mt-4">
            {!showVoid ? (
              <button onClick={() => setShowVoid(true)} className="h-12 w-full rounded-xl border border-danger/30 bg-danger/10 font-extrabold text-danger" type="button">
                Void Transaksi
              </button>
            ) : (
              <div className="rounded-2xl border border-danger/25 bg-danger/10 p-4">
                <p className="text-sm font-extrabold text-danger">Alasan void wajib diisi</p>
                <textarea
                  value={reason}
                  onChange={(event) => setReason(event.target.value)}
                  placeholder="Contoh: pelanggan batal bayar"
                  className="mt-3 min-h-20 w-full rounded-xl border border-danger/20 bg-surface p-3 text-sm font-semibold text-foreground placeholder:text-muted"
                />
                <button
                  disabled={reason.trim().length < 3}
                  onClick={() => {
                    voidTransaction(transaction.id, reason.trim());
                    onClose();
                  }}
                  className="mt-3 h-12 w-full rounded-xl bg-danger font-extrabold text-white disabled:opacity-40"
                  type="button"
                >
                  Konfirmasi Void
                </button>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

function Info({ label, value, last = false }: { label: string; value: ReactNode; last?: boolean }) {
  return (
    <div className={`flex items-center justify-between gap-4 py-3 ${last ? "" : "border-b border-border"}`}>
      <span className="text-sm font-semibold text-muted-foreground">{label}</span>
      <span className="text-right text-sm font-extrabold text-foreground">{value}</span>
    </div>
  );
}
