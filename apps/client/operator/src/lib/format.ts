import type { PaymentMethod, OrderStatus } from "@/types";

export function formatCurrency(value: number): string {
  return `Rp ${new Intl.NumberFormat("id-ID", {
    maximumFractionDigits: 0
  }).format(value)}`;
}

export function formatShortDate(value: Date | string): string {
  const date = typeof value === "string" ? new Date(value) : value;
  return new Intl.DateTimeFormat("id-ID", {
    weekday: "long",
    day: "numeric",
    month: "long",
    year: "numeric"
  }).format(date);
}

export function formatTime(value: Date | string): string {
  const date = typeof value === "string" ? new Date(value) : value;
  return new Intl.DateTimeFormat("id-ID", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false
  }).format(date);
}

export function paymentLabel(method: PaymentMethod): string {
  switch (method) {
    case "CASH":
      return "Tunai";
    case "QRIS_STATIC":
      return "QRIS";
    case "TRANSFER":
      return "Transfer";
  }
}

export function paymentDescription(method: PaymentMethod): string {
  switch (method) {
    case "CASH":
      return "Bayar di kasir";
    case "QRIS_STATIC":
      return "Scan QR merchant";
    case "TRANSFER":
      return "Transfer bank";
  }
}

export function statusLabel(status: OrderStatus): string {
  switch (status) {
    case "CONFIRMED":
      return "CONFIRMED";
    case "PENDING":
      return "PENDING";
    case "VOIDED":
      return "VOIDED";
  }
}

export function generateOrderId(seed = new Date()): string {
  const date = seed;
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, "0");
  const d = String(date.getDate()).padStart(2, "0");
  const random = Math.floor(Math.random() * 0xffff)
    .toString(16)
    .toUpperCase()
    .padStart(4, "0");
  return `QM-000001-${y}${m}${d}-${random}`;
}
