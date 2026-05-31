"use client";

import { Icon, type IconName } from "@/components/icons";
import type { PaymentMethod } from "@/types";

const paymentMeta: Record<PaymentMethod, { icon: IconName; title: string; description: string }> = {
  CASH: { icon: "cash", title: "Tunai", description: "Bayar di kasir" },
  QRIS_STATIC: { icon: "qr", title: "QRIS Static", description: "Direkomendasikan" },
  TRANSFER: { icon: "bank", title: "Transfer", description: "BCA, Mandiri, BNI" }
};

export function PaymentMethodCard({
  method,
  selected,
  onSelect,
  compact = false
}: {
  method: PaymentMethod;
  selected: boolean;
  onSelect: () => void;
  compact?: boolean;
}) {
  const meta = paymentMeta[method];

  if (compact) {
    return (
      <button
        onClick={onSelect}
        className={`flex min-h-[86px] flex-col items-center justify-center gap-2 rounded-xl border p-3 transition active:scale-95 ${
          selected ? "border-primary bg-brand/10 text-primary" : "border-border bg-card text-muted-foreground"
        }`}
        type="button"
      >
        <Icon name={meta.icon} className="h-5 w-5" />
        <span className="text-xs font-extrabold">{meta.title}</span>
      </button>
    );
  }

  return (
    <button
      onClick={onSelect}
      className={`flex w-full items-center justify-between rounded-2xl border p-4 text-left transition active:scale-[0.98] ${
        selected ? "border-primary bg-brand/10 shadow-glow" : "border-border bg-card hover:bg-card-high"
      }`}
      type="button"
    >
      <div className="flex items-center gap-3">
        <div className={`flex h-11 w-11 items-center justify-center rounded-full ${selected ? "bg-primary/20 text-primary" : "bg-card-high text-muted-foreground"}`}>
          <Icon name={meta.icon} className="h-5 w-5" />
        </div>
        <div>
          <p className="text-sm font-extrabold text-foreground">{meta.title}</p>
          <p className="text-xs font-bold text-primary">{meta.description}</p>
        </div>
      </div>
      {selected && (
        <div className="flex h-6 w-6 items-center justify-center rounded-full bg-primary text-primary-foreground">
          <Icon name="check" className="h-4 w-4" />
        </div>
      )}
    </button>
  );
}
