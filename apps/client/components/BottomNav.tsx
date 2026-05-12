"use client";

import { usePathname } from "next/navigation";
import { LayoutDashboard, BarChart2, BrainCircuit, History, Users } from "lucide-react";

const navItems = [
  { label: "Dashboard",  icon: LayoutDashboard, href: "/dashboard" },
  { label: "Statistics", icon: BarChart2,        href: "/statistics" },
  { label: "Analytics",  icon: BrainCircuit,     href: "/analytics" },
  { label: "History",    icon: History,          href: "/history" },
  { label: "Operators",  icon: Users,            href: "/operators" },
];

interface BottomNavProps {
  dark: boolean;
}

export default function BottomNav({ dark }: BottomNavProps) {
  const pathname = usePathname();

  const bg     = dark ? "#141414" : "#ffffff";
  const border = dark ? "#222"    : "#f0ece8";

  return (
    <nav style={{
      position: "fixed", bottom: 0, left: 0, right: 0,
      background: bg,
      borderTop: `1px solid ${border}`,
      display: "flex",
      zIndex: 100,
    }}>
      {navItems.map((item) => {
        const active = pathname === item.href;
        const Icon   = item.icon;
        return (
          <a key={item.label} href={item.href} style={{
            flex: 1,
            display: "flex", flexDirection: "column", alignItems: "center",
            gap: 3, padding: "10px 4px 14px",
            color: active ? "#E84C1F" : dark ? "#555" : "#bbb",
            textDecoration: "none",
            fontSize: 9, fontWeight: active ? 700 : 500,
          }}>
            <Icon size={20} strokeWidth={active ? 2.5 : 1.8} />
            {item.label.split(" ")[0]}
          </a>
        );
      })}
    </nav>
  );
}