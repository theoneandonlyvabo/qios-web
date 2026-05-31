import type { ReactNode } from "react";
import type { Metadata, Viewport } from "next";
import "./globals.css";
import { AppProviders } from "@/components/providers/app-providers";

export const metadata: Metadata = {
  title: "QIOS Operator",
  description: "Mobile-first cashier web app for QIOS operators.",
  applicationName: "QIOS Operator",
  appleWebApp: {
    capable: true,
    title: "QIOS Operator",
    statusBarStyle: "black-translucent"
  }
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  maximumScale: 1,
  viewportFit: "cover",
  themeColor: "#0d0e0f"
};

export default function RootLayout({ children }: Readonly<{ children: ReactNode }>) {
  return (
    <html lang="id" className="dark">
      <body>
        <AppProviders>{children}</AppProviders>
      </body>
    </html>
  );
}
