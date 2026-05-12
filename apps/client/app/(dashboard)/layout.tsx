"use client";

import { useState, useEffect } from "react";
import { Menu, Sun, Moon } from "lucide-react";
import Sidebar   from "@/components/Sidebar";
import BottomNav from "@/components/BottomNav";

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const [dark, setDark]               = useState(true);
  const [sidebarOpen, setSidebarOpen] = useState(false);

  // Tambahkan/hapus class "dark" di <html> setiap kali dark berubah
  useEffect(() => {
    if (dark) {
      document.documentElement.classList.add("dark");
    } else {
      document.documentElement.classList.remove("dark");
    }
  }, [dark]);

  const bg         = dark ? "#111111" : "#faf8f6";
  const cardBg     = dark ? "#1a1a1a" : "#ffffff";
  const cardBorder = dark ? "#232323" : "#f0ece8";
  const text       = dark ? "#f5f5f5" : "#111111";
  const muted      = dark ? "#555"    : "#999";

  return (
    <div style={{
      display: "flex",
      background: bg,
      minHeight: "100vh",
      color: text,
      fontFamily: "'DM Sans', system-ui, sans-serif",
    }}>

      {/* Sidebar desktop */}
      <div className="sidebar-desktop">
        <Sidebar dark={dark} onToggleDark={() => setDark(!dark)} />
      </div>

      {/* Sidebar mobile overlay */}
      {sidebarOpen && (
        <div
          onClick={() => setSidebarOpen(false)}
          style={{
            position: "fixed", inset: 0,
            background: "rgba(0,0,0,0.5)",
            zIndex: 200, display: "flex",
          }}
        >
          <div onClick={(e) => e.stopPropagation()} style={{ width: 240 }}>
            <Sidebar dark={dark} onToggleDark={() => setDark(!dark)} />
          </div>
        </div>
      )}

      {/* Main area */}
      <div style={{ flex: 1, display: "flex", flexDirection: "column", minWidth: 0 }}>

        {/* Mobile topbar */}
        <header
          className="mobile-topbar"
          style={{
            display: "flex", alignItems: "center",
            justifyContent: "space-between",
            padding: "14px 16px",
            background: cardBg,
            borderBottom: `1px solid ${cardBorder}`,
          }}
        >
          <button
            onClick={() => setSidebarOpen(true)}
            style={{
              background: "transparent", border: `1px solid ${cardBorder}`,
              borderRadius: 10, padding: "7px 10px", cursor: "pointer",
              display: "flex", alignItems: "center", gap: 8,
              color: text, fontSize: 13,
            }}
          >
            <Menu size={18} />
            <span style={{ fontWeight: 800, color: "#E84C1F" }}>QIOS</span>
          </button>

          <button
            onClick={() => setDark(!dark)}
            style={{
              background: "transparent", border: `1px solid ${cardBorder}`,
              borderRadius: 10, padding: "7px 10px", cursor: "pointer",
              color: muted, display: "flex",
            }}
          >
            {dark ? <Sun size={16} /> : <Moon size={16} />}
          </button>
        </header>

        {/* Page content */}
        <main style={{ flex: 1, padding: "28px 28px 100px", minWidth: 0 }}>
          {children}
        </main>
      </div>

      {/* Bottom nav mobile */}
      <div className="bottom-nav-mobile">
        <BottomNav dark={dark} />
      </div>

      <style>{`
        .sidebar-desktop   { display: flex; }
        .mobile-topbar     { display: none; }
        .bottom-nav-mobile { display: none; }

        @media (max-width: 767px) {
          .sidebar-desktop   { display: none !important; }
          .mobile-topbar     { display: flex !important; }
          .bottom-nav-mobile { display: block !important; }
        }
      `}</style>
    </div>
  );
}