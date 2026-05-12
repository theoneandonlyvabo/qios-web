"use client";

import { usePathname } from "next/navigation";
import {
  LayoutDashboard, BarChart2, BrainCircuit,
  History, Users, LogOut, Sun, Moon,
} from "lucide-react";

const navItems = [
  { label: "Dashboard",    icon: LayoutDashboard, href: "/dashboard" },
  { label: "Statistics",   icon: BarChart2,        href: "/statistics" },
  { label: "AI Analytics", icon: BrainCircuit,     href: "/analytics" },
  { label: "History",      icon: History,          href: "/history" },
  { label: "Operators",    icon: Users,            href: "/operators" },
];

interface SidebarProps {
  dark: boolean;
  onToggleDark: () => void;
}

export default function Sidebar({ dark, onToggleDark }: SidebarProps) {
  const pathname = usePathname();

  const bg          = dark ? "#141414" : "#ffffff";
  const border      = dark ? "#222"    : "#f0ece8";
  const muted       = dark ? "#555"    : "#aaa";
  const text        = dark ? "#f5f5f5" : "#111";

  return (
    <aside style={{
      width: 220, flexShrink: 0,
      background: bg,
      borderRight: `1px solid ${border}`,
      display: "flex", flexDirection: "column",
      height: "100vh", position: "sticky", top: 0,
    }}>
      {/* Logo */}
      <div style={{ padding: "24px 20px 20px", borderBottom: `1px solid ${border}` }}>
        <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
          <div style={{
            width: 34, height: 34, borderRadius: 9,
            background: "#E84C1F",
            display: "flex", alignItems: "center", justifyContent: "center",
          }}>
            <span style={{ color: "#fff", fontWeight: 900, fontSize: 16, letterSpacing: -1 }}>Q</span>
          </div>
          <div>
            <p style={{ margin: 0, fontWeight: 800, fontSize: 16, color: text, letterSpacing: -0.5 }}>QIOS</p>
            <p style={{ margin: 0, fontSize: 11, color: muted }}>Warung Makan Bu Siti</p>
          </div>
        </div>
      </div>

      {/* Nav */}
      <nav style={{ flex: 1, padding: "12px", display: "flex", flexDirection: "column", gap: 2 }}>
        <p style={{
          fontSize: 10, fontWeight: 700, color: muted,
          letterSpacing: 1, padding: "8px 8px 4px", textTransform: "uppercase",
        }}>
          Menu
        </p>
        {navItems.map((item) => {
          const active = pathname === item.href;
          const Icon   = item.icon;
          return (
            <a key={item.label} href={item.href} style={{
              display: "flex", alignItems: "center", gap: 10,
              padding: "9px 10px", borderRadius: 10,
              background: active ? "rgba(232,76,31,0.1)" : "transparent",
              color:      active ? "#E84C1F" : dark ? "#999" : "#666",
              fontWeight: active ? 700 : 500,
              fontSize: 13, textDecoration: "none",
              transition: "background 0.15s, color 0.15s",
            }}>
              <Icon size={17} strokeWidth={active ? 2.5 : 2} />
              {item.label}
              {active && (
                <div style={{
                  marginLeft: "auto", width: 6, height: 6,
                  borderRadius: "50%", background: "#E84C1F",
                }} />
              )}
            </a>
          );
        })}
      </nav>

      {/* Bottom */}
      <div style={{ padding: "12px", borderTop: `1px solid ${border}`, display: "flex", flexDirection: "column", gap: 4 }}>
        <button onClick={onToggleDark} style={{
          display: "flex", alignItems: "center", gap: 10,
          padding: "9px 10px", borderRadius: 10, border: "none",
          background: "transparent", cursor: "pointer",
          color: dark ? "#999" : "#666", fontSize: 13, fontWeight: 500,
          width: "100%", textAlign: "left",
        }}>
          {dark ? <Sun size={17} /> : <Moon size={17} />}
          {dark ? "Mode Terang" : "Mode Gelap"}
        </button>
        <button style={{
          display: "flex", alignItems: "center", gap: 10,
          padding: "9px 10px", borderRadius: 10, border: "none",
          background: "transparent", cursor: "pointer",
          color: "#E84C1F", fontSize: 13, fontWeight: 600,
          width: "100%", textAlign: "left",
        }}>
          <LogOut size={17} />
          Keluar
        </button>
      </div>
    </aside>
  );
}