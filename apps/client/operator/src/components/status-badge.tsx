import type { OrderStatus, PaymentMethod } from "@/types";
import { paymentLabel, statusLabel } from "@/lib/format";

export function StatusBadge({ status }: { status: OrderStatus }) {
  const style = {
    CONFIRMED: "border-success/25 bg-success/10 text-success",
    PENDING: "border-warning/25 bg-warning/10 text-warning",
    VOIDED: "border-danger/25 bg-danger/10 text-danger"
  }[status];

  return (
    <span className={`inline-flex rounded-full border px-2 py-0.5 text-[10px] font-extrabold uppercase tracking-wide ${style}`}>
      {statusLabel(status)}
    </span>
  );
}

export function PaymentBadge({ method }: { method: PaymentMethod }) {
  return (
    <span className="inline-flex rounded-full border border-border-warm/70 bg-card-high px-2 py-0.5 text-[10px] font-bold text-primary">
      {paymentLabel(method)}
    </span>
  );
}
