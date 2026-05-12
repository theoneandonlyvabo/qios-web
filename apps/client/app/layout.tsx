import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "QIOS — Kendali Penuh Atas Data Bisnis Anda",
  description: "Sistem manajemen keuangan dan operasional untuk UMKM Indonesia.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="id">
      <body>{children}</body>
    </html>
  );
}