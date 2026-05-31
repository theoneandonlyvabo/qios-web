"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";
import { Icon, type IconName } from "@/components/icons";

type NavItem = {
  href: string;
  label: string;
  icon: IconName;
  match: string[];
};

const navItems: NavItem[] = [
  { href: "/order", label: "Kasir", icon: "pos", match: ["/order", "/confirm", "/payment", "/success"] },
  { href: "/history", label: "Transaksi", icon: "receipt", match: ["/history"] },
  { href: "/products", label: "Produk", icon: "box", match: ["/products"] },
  { href: "/profile", label: "Lainnya", icon: "menu", match: ["/profile"] }
];

export function OperatorShell({
  children,
  withNav = true
}: {
  children: ReactNode;
  withNav?: boolean;
}) {
  return (
    <div className="min-h-dvh bg-background text-foreground">
      <div className="qios-phone">{children}</div>
      {withNav && <BottomNav />}
    </div>
  );
}

function BottomNav() {
  const pathname = usePathname();

  return (
    <nav className="fixed bottom-0 left-1/2 z-50 flex h-16 w-full max-w-[430px] -translate-x-1/2 items-center justify-around rounded-t-2xl border-t border-border bg-surface/90 px-3 shadow-[0_-10px_40px_rgba(0,0,0,0.45)] backdrop-blur-xl safe-pb">
      {navItems.map((item) => {
        const active = item.match.some((entry) => pathname.startsWith(entry));
        return (
          <Link
            key={item.href}
            href={item.href}
            className={`flex h-full flex-1 flex-col items-center justify-center gap-1 rounded-xl text-[11px] font-bold transition active:scale-95 ${
              active ? "text-primary" : "text-muted-foreground hover:bg-card-high hover:text-foreground"
            }`}
          >
            <Icon name={item.icon} className="h-5 w-5" />
            <span>{item.label}</span>
          </Link>
        );
      })}
    </nav>
  );
}
