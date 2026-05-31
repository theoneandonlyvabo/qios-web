"use client";

import { useCallback, useMemo, useState, type ReactNode } from "react";
import { useRouter } from "next/navigation";
import { Icon } from "@/components/icons";
import { OperatorShell } from "@/components/operator-shell";
import { FakeQrCode } from "@/components/payment/fake-qr";
import { SlideToConfirm } from "@/components/payment/slide-to-confirm";
import { useOperator } from "@/components/providers/operator-provider";
import { PageTopBar } from "@/components/top-bars";
import { formatCurrency, paymentLabel } from "@/lib/format";

export default function PaymentPage() {
  const router = useRouter();
  const { session, summary, paymentMethod, createTransaction } = useOperator();
  const [cashReceived, setCashReceived] = useState("");
  const [copied, setCopied] = useState(false);
  const [confirming, setConfirming] = useState(false);

  const receivedNumber = Number(cashReceived.replace(/[^0-9]/g, ""));
  const change = useMemo(() => Math.max(0, receivedNumber - summary.total), [receivedNumber, summary.total]);

  const handleConfirm = useCallback(() => {
    if (summary.itemCount === 0) {
      router.push("/order");
      return;
    }
    setConfirming(true);
    window.setTimeout(() => {
      createTransaction();
      router.push("/success");
    }, 500);
  }, [createTransaction, router, summary.itemCount]);

  if (summary.itemCount === 0) {
    return (
      <OperatorShell>
        <PageTopBar title="Pembayaran" />
        <main className="flex min-h-[70dvh] flex-col items-center justify-center px-4 text-center">
          <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-card-high text-primary">
            <Icon name="receipt" className="h-8 w-8" />
          </div>
          <h1 className="text-xl font-extrabold text-foreground">Order kosong</h1>
          <p className="mt-2 text-sm font-medium text-muted-foreground">Buat order baru untuk melanjutkan pembayaran.</p>
          <button className="mt-6 h-12 rounded-xl bg-primary px-6 font-extrabold text-primary-foreground" onClick={() => router.push("/order")} type="button">
            Kembali ke Kasir
          </button>
        </main>
      </OperatorShell>
    );
  }

  return (
    <OperatorShell withNav={false}>
      <PageTopBar title={paymentMethod === "QRIS_STATIC" ? "Bayar via QRIS" : paymentMethod === "CASH" ? "Bayar Tunai" : "Bayar Transfer"} />

      <main className="flex min-h-[calc(100dvh-160px)] flex-col px-4 pb-32 pt-6">
        {paymentMethod === "QRIS_STATIC" && (
          <section className="flex flex-1 flex-col items-center text-center">
            <p className="mb-5 text-base font-medium text-primary">Minta pembeli scan QR di bawah.</p>
            <div className="relative mb-6 aspect-square w-full max-w-[320px]">
              <div className="absolute inset-0 rounded-[32px] bg-primary/10 blur-2xl animate-subtle-pulse" />
              <div className="relative flex h-full flex-col items-center justify-center rounded-[32px] border border-border-warm/50 bg-card/70 p-6 backdrop-blur-xl">
                <FakeQrCode />
                <div className="mt-4 flex items-center gap-2">
                  <span className="text-lg font-black tracking-widest text-white">QRIS</span>
                  <span className="h-4 w-px bg-border" />
                  <span className="text-[10px] font-extrabold uppercase tracking-tight text-muted-foreground">GPN Supported</span>
                </div>
              </div>
            </div>
          </section>
        )}

        {paymentMethod === "CASH" && (
          <section className="space-y-5">
            <div className="rounded-[28px] border border-border bg-card p-5 text-center">
              <p className="text-xs font-extrabold uppercase tracking-[0.18em] text-muted-foreground">Total Bayar</p>
              <h2 className="mt-2 text-[44px] font-black leading-none tracking-[-0.04em] text-primary">{formatCurrency(summary.total)}</h2>
            </div>
            <div className="rounded-2xl border border-border bg-card p-4">
              <label className="text-sm font-extrabold text-foreground" htmlFor="cash-received">Uang diterima</label>
              <input
                id="cash-received"
                inputMode="numeric"
                value={cashReceived}
                onChange={(event) => setCashReceived(event.target.value)}
                className="mt-3 h-14 w-full rounded-xl border border-border bg-surface px-4 text-2xl font-black text-primary placeholder:text-muted"
                placeholder="100000"
              />
              <div className="mt-4 flex items-center justify-between rounded-xl bg-card-high p-4">
                <span className="text-sm font-bold text-muted-foreground">Kembalian</span>
                <span className="text-xl font-black text-foreground">{formatCurrency(change)}</span>
              </div>
            </div>
          </section>
        )}

        {paymentMethod === "TRANSFER" && (
          <section className="space-y-5">
            <div className="rounded-[28px] border border-border bg-card p-5 text-center">
              <p className="text-xs font-extrabold uppercase tracking-[0.18em] text-muted-foreground">Total Transfer</p>
              <h2 className="mt-2 text-[44px] font-black leading-none tracking-[-0.04em] text-primary">{formatCurrency(summary.total)}</h2>
            </div>
            <div className="rounded-2xl border border-border bg-card p-4">
              <p className="text-sm font-extrabold text-foreground">Detail Rekening</p>
              <div className="mt-4 space-y-4">
                <InfoRow label="Bank" value={session.transferBankName} />
                <InfoRow
                  label="Nomor Rekening"
                  value={session.transferAccountNumber}
                  action={
                    <button
                      onClick={async () => {
                        await navigator.clipboard?.writeText(session.transferAccountNumber);
                        setCopied(true);
                        window.setTimeout(() => setCopied(false), 1300);
                      }}
                      className="flex h-9 items-center gap-1 rounded-lg bg-card-high px-3 text-xs font-extrabold text-primary"
                      type="button"
                    >
                      <Icon name="copy" className="h-4 w-4" />
                      {copied ? "Disalin" : "Copy"}
                    </button>
                  }
                />
                <InfoRow label="Atas Nama" value={session.transferAccountHolder} />
              </div>
              <p className="mt-5 rounded-xl border border-primary/20 bg-primary/10 p-3 text-xs font-bold leading-relaxed text-primary">
                Minta pembeli transfer sesuai nominal, lalu konfirmasi setelah bukti pembayaran diterima.
              </p>
            </div>
          </section>
        )}

        <section className="mt-auto w-full space-y-3 px-2 pt-6">
          <div className="text-center">
            <p className="text-[11px] font-extrabold uppercase tracking-[0.18em] text-muted-foreground">Total Bayar</p>
            <p className="mt-1 text-[38px] font-black leading-tight tracking-[-0.04em] text-primary">{formatCurrency(summary.total)}</p>
          </div>

          <div className="space-y-3 rounded-2xl border border-border/60 bg-surface/60 p-4">
            <InfoRow label="Nomor Pesanan" value="#KP-992102" />
            <InfoRow label="Merchant" value={session.businessName} />
            <InfoRow label="Metode" value={paymentLabel(paymentMethod)} />
          </div>
        </section>
      </main>

      <footer className="fixed bottom-0 left-1/2 z-50 w-full max-w-[430px] -translate-x-1/2 border-t border-border bg-card/95 p-4 backdrop-blur-xl safe-pb">
        <SlideToConfirm disabled={confirming} onConfirm={handleConfirm} />
        {confirming && <p className="mt-2 text-center text-xs font-bold text-muted-foreground">Menyimpan transaksi...</p>}
      </footer>
    </OperatorShell>
  );
}

function InfoRow({ label, value, action }: { label: string; value: string; action?: ReactNode }) {
  return (
    <div className="flex items-center justify-between gap-3 border-b border-border/40 py-2 last:border-b-0">
      <span className="text-sm font-medium text-muted-foreground">{label}</span>
      <div className="flex items-center gap-2 text-right">
        <span className="text-sm font-extrabold text-foreground">{value}</span>
        {action}
      </div>
    </div>
  );
}
