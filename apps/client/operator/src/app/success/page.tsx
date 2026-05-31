"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Icon } from "@/components/icons";
import { OperatorShell } from "@/components/operator-shell";
import { useOperator } from "@/components/providers/operator-provider";
import { seedTransactions } from "@/data/mock";
import { formatCurrency, formatTime, paymentLabel } from "@/lib/format";
import type { Transaction } from "@/types";

export default function SuccessPage() {
  const router = useRouter();
  const { getLastTransaction } = useOperator();
  const [transaction, setTransaction] = useState<Transaction | null>(null);

  useEffect(() => {
    setTransaction(getLastTransaction() ?? seedTransactions[0]);
  }, [getLastTransaction]);

  const tx = transaction ?? seedTransactions[0];

  return (
    <OperatorShell withNav={false}>
      <main className="relative flex min-h-dvh flex-col overflow-hidden bg-[radial-gradient(circle_at_50%_34%,rgba(255,181,158,0.12),transparent_19rem)] px-4 py-8 safe-pb">
        <section className="flex flex-1 flex-col items-center justify-center text-center">
          <div className="relative mb-7 flex items-center justify-center">
            <div className="absolute h-28 w-28 rounded-full bg-primary/20 animate-ping" style={{ animationDuration: "3s" }} />
            <div className="relative flex h-24 w-24 items-center justify-center rounded-full border border-primary/25 bg-primary/10 text-primary backdrop-blur-sm">
              <Icon name="check" className="h-11 w-11" />
            </div>
          </div>

          <h1 className="text-xl font-extrabold text-foreground">Transaksi Berhasil</h1>
          <p className="mt-1 text-sm font-bold text-muted-foreground">Pembayaran telah diterima</p>
          <p className="mt-8 text-[44px] font-black leading-none tracking-[-0.04em] text-primary">{formatCurrency(tx.total)}</p>

          <div className="mt-8 w-full max-w-[340px] rounded-2xl border border-border bg-card/80 p-4 backdrop-blur-xl shadow-soft">
            <Info label="Order ID" value={tx.orderId} />
            <Info label="Metode" value={paymentLabel(tx.paymentMethod)} />
            <Info label="Waktu" value={`Hari ini, ${formatTime(tx.confirmedAt ?? tx.createdAt)} WIB`} last />
          </div>
        </section>

        <footer className="mx-auto w-full max-w-[340px] space-y-3">
          <button
            onClick={() => router.push("/order")}
            className="h-[52px] w-full rounded-full bg-primary text-base font-extrabold text-primary-foreground transition active:scale-[0.98]"
            type="button"
          >
            Order Baru
          </button>
          <button
            onClick={() => router.push("/history")}
            className="h-[52px] w-full rounded-full border border-border-warm bg-transparent text-base font-extrabold text-primary transition hover:bg-card active:scale-[0.98]"
            type="button"
          >
            Lihat Riwayat
          </button>
        </footer>
      </main>
    </OperatorShell>
  );
}

function Info({ label, value, last = false }: { label: string; value: string; last?: boolean }) {
  return (
    <div className={`flex items-center justify-between py-3 ${last ? "" : "border-b border-border"}`}>
      <span className="text-sm font-semibold text-muted-foreground">{label}</span>
      <span className="max-w-[190px] truncate text-right text-sm font-extrabold text-foreground">{value}</span>
    </div>
  );
}
