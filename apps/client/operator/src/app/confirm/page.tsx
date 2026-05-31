"use client";

import { useRouter } from "next/navigation";
import { Icon } from "@/components/icons";
import { OperatorShell } from "@/components/operator-shell";
import { PaymentMethodCard } from "@/components/payment/payment-method-card";
import { useOperator } from "@/components/providers/operator-provider";
import { PageTopBar } from "@/components/top-bars";
import { formatCurrency } from "@/lib/format";
import type { PaymentMethod } from "@/types";

const paymentMethods: PaymentMethod[] = ["CASH", "QRIS_STATIC", "TRANSFER"];

export default function ConfirmPage() {
  const router = useRouter();
  const { summary, paymentMethod, setPaymentMethod, note, setNote } = useOperator();

  if (summary.itemCount === 0) {
    return (
      <OperatorShell>
        <PageTopBar title="Ringkasan Order" />
        <main className="flex min-h-[70dvh] flex-col items-center justify-center px-4 text-center">
          <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-card-high text-primary">
            <Icon name="receipt" className="h-8 w-8" />
          </div>
          <h1 className="text-xl font-extrabold text-foreground">Belum ada item</h1>
          <p className="mt-2 max-w-[260px] text-sm font-medium text-muted-foreground">Tambahkan produk dulu sebelum masuk ke pembayaran.</p>
          <button
            onClick={() => router.push("/order")}
            className="mt-6 h-12 rounded-xl bg-primary px-6 font-extrabold text-primary-foreground"
            type="button"
          >
            Kembali ke Kasir
          </button>
        </main>
      </OperatorShell>
    );
  }

  return (
    <OperatorShell withNav={false}>
      <PageTopBar title="Ringkasan Order" />
      <main className="pb-[260px] pt-6">
        <section className="space-y-3 px-4">
          <div className="flex items-end justify-between">
            <h2 className="text-lg font-extrabold text-foreground">Detail Pesanan</h2>
            <p className="text-xs font-extrabold text-primary">{summary.itemCount} Items</p>
          </div>

          {summary.items.map((item) => (
            <article key={item.productId} className="flex items-center justify-between rounded-2xl border border-border-warm/60 bg-card p-3">
              <div className="flex min-w-0 items-center gap-3">
                <div className={`flex h-14 w-14 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br ${item.gradient}`}>
                  <span className="text-2xl">{item.emoji}</span>
                </div>
                <div className="min-w-0">
                  <p className="truncate text-sm font-extrabold text-foreground">{item.productName}</p>
                  <p className="text-xs font-bold text-muted-foreground">Qty: {item.quantity} × {formatCurrency(item.unitPrice)}</p>
                </div>
              </div>
              <p className="ml-3 shrink-0 text-sm font-extrabold text-primary">{formatCurrency(item.subtotal)}</p>
            </article>
          ))}
        </section>

        <section className="mx-4 mt-7 rounded-3xl border border-border-warm/70 bg-card-high p-5">
          <div className="flex items-center justify-between text-sm font-semibold text-muted-foreground">
            <span>Subtotal</span>
            <span className="text-foreground">{formatCurrency(summary.subtotal)}</span>
          </div>
          <div className="mt-4 flex items-center justify-between text-sm font-semibold text-muted-foreground">
            <span>Pajak</span>
            <span className="text-foreground">{formatCurrency(summary.tax)}</span>
          </div>
          <div className="my-5 h-px bg-border" />
          <div className="flex items-end justify-between gap-6 text-primary">
            <p className="text-xl font-extrabold leading-tight">Total<br />Akhir</p>
            <p className="text-[40px] font-black leading-none tracking-[-0.04em]">{formatCurrency(summary.total)}</p>
          </div>
        </section>
      </main>

      <section className="fixed bottom-0 left-1/2 z-50 w-full max-w-[430px] -translate-x-1/2 border-t border-border bg-card/95 px-4 pt-5 backdrop-blur-xl safe-pb">
        <p className="mb-3 text-lg font-extrabold text-foreground">Metode Pembayaran</p>
        <div className="grid grid-cols-3 gap-2">
          {paymentMethods.map((method) => (
            <PaymentMethodCard
              key={method}
              method={method}
              compact
              selected={paymentMethod === method}
              onSelect={() => setPaymentMethod(method)}
            />
          ))}
        </div>

        <textarea
          value={note}
          onChange={(event) => setNote(event.target.value)}
          className="mt-4 min-h-16 w-full resize-none rounded-xl border border-border bg-surface px-4 py-3 text-sm font-semibold text-foreground placeholder:text-muted"
          placeholder="Catatan transaksi opsional"
        />

        <button
          onClick={() => router.push("/payment")}
          className="mt-4 flex h-14 w-full items-center justify-center gap-3 rounded-xl bg-brand text-lg font-extrabold text-white shadow-[0_16px_36px_rgba(202,64,10,0.22)] transition active:scale-[0.98]"
          type="button"
        >
          Pilih Pembayaran
          <Icon name="chevron-right" className="h-5 w-5" />
        </button>
        <p className="mt-2 text-center text-xs font-extrabold text-muted-foreground">Pastikan nominal pesanan sudah sesuai</p>
      </section>
    </OperatorShell>
  );
}
