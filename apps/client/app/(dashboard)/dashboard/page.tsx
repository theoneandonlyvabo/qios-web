"use client";

import { useState } from "react";
import {
  LineChart, Line, XAxis, YAxis, CartesianGrid,
  Tooltip, ResponsiveContainer,
} from "recharts";
import {
  ShoppingCart, DollarSign, TrendingUp, TrendingDown,
  Users, ArrowUpRight, ArrowDownRight, Clock, Package,
} from "lucide-react";

// ─── Mock Data (swap ke API real di Week 3) ──────────────────────────────────

const trendData = [
  { day: "Sen", revenue: 1200000, transactions: 24 },
  { day: "Sel", revenue: 1850000, transactions: 37 },
  { day: "Rab", revenue: 1400000, transactions: 28 },
  { day: "Kam", revenue: 2100000, transactions: 42 },
  { day: "Jum", revenue: 2600000, transactions: 52 },
  { day: "Sab", revenue: 3200000, transactions: 64 },
  { day: "Min", revenue: 2800000, transactions: 56 },
];

const topProducts = [
  { id: 1, name: "Kopi Susu Gula Aren",  sold: 142, revenue: 4970000, trend: "up"   },
  { id: 2, name: "Nasi Goreng Spesial",  sold: 118, revenue: 5310000, trend: "up"   },
  { id: 3, name: "Es Teh Manis",         sold: 97,  revenue: 1455000, trend: "down" },
  { id: 4, name: "Mie Ayam Bakso",       sold: 85,  revenue: 3825000, trend: "up"   },
  { id: 5, name: "Jus Alpukat",          sold: 63,  revenue: 2205000, trend: "down" },
];

const summaryCards = [
  { label: "Revenue Hari Ini", value: "Rp 2.847.000", change: "+12.4%", positive: true,  icon: DollarSign, sub: "vs kemarin" },
  { label: "Total Transaksi",  value: "56",            change: "+8.7%",  positive: true,  icon: ShoppingCart, sub: "vs kemarin" },
  { label: "Rata-rata Order",  value: "Rp 50.800",     change: "-3.2%",  positive: false, icon: TrendingUp, sub: "vs kemarin" },
  { label: "Operator Aktif",   value: "3 / 4",         change: "0%",     positive: true,  icon: Users, sub: "terdaftar" },
];

const peakHours = [
  { hour: "08", count: 4  }, { hour: "09", count: 8  }, { hour: "10", count: 12 },
  { hour: "11", count: 18 }, { hour: "12", count: 24 }, { hour: "13", count: 20 },
  { hour: "14", count: 14 }, { hour: "15", count: 10 }, { hour: "16", count: 8  },
  { hour: "17", count: 6  }, { hour: "18", count: 9  }, { hour: "19", count: 7  },
];

// ─── Helpers ─────────────────────────────────────────────────────────────────

function formatRupiah(n: number) {
  return "Rp " + n.toLocaleString("id-ID");
}

// ─── Tooltip ─────────────────────────────────────────────────────────────────

function CustomTooltip({ active, payload, label }: {
  active?: boolean; payload?: { value: number }[]; label?: string;
}) {
  if (!active || !payload?.length) return null;
  return (
    <div style={{
      background: "var(--card-bg)", border: "1px solid var(--card-border)",
      borderRadius: 10, padding: "10px 14px",
    }}>
      <p style={{ fontSize: 11, color: "var(--muted)", marginBottom: 2 }}>{label}</p>
      <p style={{ fontSize: 13, fontWeight: 700, margin: 0 }}>{formatRupiah(payload[0].value)}</p>
    </div>
  );
}

// ─── Peak Hours Bar ───────────────────────────────────────────────────────────

function PeakHoursBar() {
  const max = Math.max(...peakHours.map((h) => h.count));
  return (
    <div style={{ display: "flex", alignItems: "flex-end", gap: 4, height: 64 }}>
      {peakHours.map((h) => {
        const pct    = (h.count / max) * 100;
        const isPeak = h.count === max;
        return (
          <div key={h.hour} style={{ flex: 1, display: "flex", flexDirection: "column", alignItems: "center", gap: 3, height: "100%" }}>
            <div style={{ flex: 1, width: "100%", display: "flex", alignItems: "flex-end" }}>
              <div style={{
                width: "100%", borderRadius: 3,
                height: `${pct}%`,
                background: isPeak ? "#E84C1F" : "var(--bar-bg)",
                transition: "height 0.3s",
              }} />
            </div>
            <span style={{ fontSize: 9, color: "var(--muted)" }}>
              {h.hour === "08" || h.hour === "12" || h.hour === "19" ? h.hour : ""}
            </span>
          </div>
        );
      })}
    </div>
  );
}

// ─── Page ─────────────────────────────────────────────────────────────────────

export default function DashboardPage() {
  const [chartMode, setChartMode] = useState<"revenue" | "transactions">("revenue");

  const today = new Date().toLocaleDateString("id-ID", {
    weekday: "long", year: "numeric", month: "long", day: "numeric",
  });

  return (
    <>
      {/* CSS variables — supaya sinkron dengan dark/light dari layout */}
      <style>{`
        :root {
          --card-bg:     #ffffff;
          --card-border: #f0ece8;
          --muted:       #999;
          --bar-bg:      #e8e3dc;
          --tab-on-bg:   #111;
          --tab-on-text: #fff;
          --tab-off-bg:  #f0ece8;
          --tab-off-text:#666;
          --grid-line:   #eee;
        }
        html.dark {
          --card-bg:     #1a1a1a;
          --card-border: #232323;
          --muted:       #555;
          --bar-bg:      #2a2a2a;
          --tab-on-bg:   #fff;
          --tab-on-text: #111;
          --tab-off-bg:  #222;
          --tab-off-text:#666;
          --grid-line:   #222;
        }
      `}</style>

      {/* Page header */}
      <div style={{ marginBottom: 24 }}>
        <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 6 }}>
          <div style={{
            display: "flex", alignItems: "center", gap: 6,
            background: "rgba(22,163,74,0.1)", color: "#16a34a",
            padding: "5px 10px", borderRadius: 8, fontSize: 11, fontWeight: 700,
          }}>
            <span style={{ width: 6, height: 6, borderRadius: "50%", background: "#16a34a", display: "inline-block" }} />
            Data Langsung
          </div>
        </div>
        <h1 style={{ margin: 0, fontSize: 22, fontWeight: 800, letterSpacing: -0.5, lineHeight: 1.2 }}>
          Dashboard
        </h1>
        <p style={{ margin: "4px 0 0", fontSize: 12, color: "var(--muted)" }}>{today}</p>
      </div>

      {/* Metric Cards */}
      <div style={{
        display: "grid",
        gridTemplateColumns: "repeat(auto-fit, minmax(160px, 1fr))",
        gap: 12, marginBottom: 14,
      }}>
        {summaryCards.map(({ label, value, change, positive, icon: Icon, sub }) => (
          <div key={label} style={{
            background: "var(--card-bg)", border: "1px solid var(--card-border)",
            borderRadius: 16, padding: 16,
          }}>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginBottom: 12 }}>
              <span style={{ fontSize: 12, color: "var(--muted)", lineHeight: 1.3 }}>{label}</span>
              <div style={{
                width: 30, height: 30, borderRadius: 8,
                background: "rgba(232,76,31,0.1)", color: "#E84C1F",
                display: "flex", alignItems: "center", justifyContent: "center",
              }}>
                <Icon size={15} />
              </div>
            </div>
            <p style={{ margin: 0, fontSize: 20, fontWeight: 800, letterSpacing: -0.5 }}>{value}</p>
            <div style={{ display: "flex", alignItems: "center", gap: 4, marginTop: 8 }}>
              {positive
                ? <ArrowUpRight size={12} color="#16a34a" />
                : <ArrowDownRight size={12} color="#dc2626" />}
              <span style={{ fontSize: 11, fontWeight: 700, color: positive ? "#16a34a" : "#dc2626" }}>{change}</span>
              <span style={{ fontSize: 11, color: "var(--muted)" }}>{sub}</span>
            </div>
          </div>
        ))}
      </div>

      {/* Chart */}
      <div style={{ display: "grid", gridTemplateColumns: "1fr", gap: 12, marginBottom: 12 }}>
        <div style={{
          background: "var(--card-bg)", border: "1px solid var(--card-border)",
          borderRadius: 16, padding: 20,
        }}>
          <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: 16, flexWrap: "wrap", gap: 8 }}>
            <div>
              <p style={{ margin: 0, fontSize: 15, fontWeight: 700, letterSpacing: -0.3 }}>Tren Minggu Ini</p>
              <p style={{ margin: "2px 0 0", fontSize: 12, color: "var(--muted)" }}>7 hari terakhir</p>
            </div>
            <div style={{ display: "flex", gap: 6 }}>
              {(["revenue", "transactions"] as const).map((mode) => (
                <button key={mode} onClick={() => setChartMode(mode)} style={{
                  padding: "6px 14px", borderRadius: 8, border: "none", cursor: "pointer",
                  fontSize: 12, fontWeight: 600, transition: "all 0.15s",
                  background: chartMode === mode ? "var(--tab-on-bg)"   : "var(--tab-off-bg)",
                  color:      chartMode === mode ? "var(--tab-on-text)" : "var(--tab-off-text)",
                }}>
                  {mode === "revenue" ? "Revenue" : "Transaksi"}
                </button>
              ))}
            </div>
          </div>
          <ResponsiveContainer width="100%" height={180}>
            <LineChart data={trendData} margin={{ top: 4, right: 4, left: -20, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--grid-line)" vertical={false} />
              <XAxis dataKey="day" tick={{ fontSize: 11, fill: "var(--muted)" }} axisLine={false} tickLine={false} />
              <YAxis tick={{ fontSize: 10, fill: "var(--muted)" }} axisLine={false} tickLine={false}
                tickFormatter={(v) => chartMode === "revenue" ? `${(v / 1000000).toFixed(1)}jt` : `${v}`}
              />
              <Tooltip content={<CustomTooltip />} />
              <Line
                type="monotone"
                dataKey={chartMode === "revenue" ? "revenue" : "transactions"}
                stroke="#E84C1F" strokeWidth={2.5}
                dot={{ fill: "#E84C1F", strokeWidth: 0, r: 3.5 }}
                activeDot={{ r: 5, fill: "#E84C1F", strokeWidth: 2, stroke: "var(--card-bg)" }}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* Peak hours + quick stats */}
        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 12 }}>
          <div style={{
            background: "var(--card-bg)", border: "1px solid var(--card-border)",
            borderRadius: 16, padding: 16,
          }}>
            <div style={{ display: "flex", alignItems: "center", gap: 7, marginBottom: 12 }}>
              <Clock size={14} color="var(--muted)" />
              <p style={{ margin: 0, fontSize: 13, fontWeight: 700 }}>Jam Sibuk</p>
            </div>
            <PeakHoursBar />
            <p style={{ margin: "8px 0 0", fontSize: 11, color: "var(--muted)" }}>
              Puncak: <strong>12:00 – 13:00</strong>
            </p>
          </div>

          <div style={{
            background: "var(--card-bg)", border: "1px solid var(--card-border)",
            borderRadius: 16, padding: 16,
          }}>
            <p style={{ margin: "0 0 12px", fontSize: 13, fontWeight: 700 }}>Ringkasan</p>
            {[
              { label: "Produk aktif",   value: "24", icon: Package },
              { label: "Order pending",  value: "0",  icon: Clock   },
              { label: "Operator aktif", value: "3",  icon: Users   },
            ].map(({ label, value, icon: Icon }) => (
              <div key={label} style={{
                display: "flex", alignItems: "center", justifyContent: "space-between",
                padding: "7px 0", borderBottom: "1px solid var(--card-border)",
              }}>
                <div style={{ display: "flex", alignItems: "center", gap: 7 }}>
                  <Icon size={13} color="var(--muted)" />
                  <span style={{ fontSize: 12, color: "var(--muted)" }}>{label}</span>
                </div>
                <span style={{ fontSize: 13, fontWeight: 700 }}>{value}</span>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Top Products */}
      <div style={{
        background: "var(--card-bg)", border: "1px solid var(--card-border)",
        borderRadius: 16, padding: 20,
      }}>
        <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: 16 }}>
          <p style={{ margin: 0, fontSize: 15, fontWeight: 700, letterSpacing: -0.3 }}>Produk Terlaris</p>
          <span style={{
            background: "#E84C1F", color: "#fff",
            fontSize: 11, fontWeight: 700, borderRadius: 6, padding: "3px 9px",
          }}>Minggu ini</span>
        </div>

        {topProducts.map((p, i) => (
          <div key={p.id} style={{
            display: "flex", alignItems: "center", gap: 12,
            padding: "10px 8px", borderRadius: 10, cursor: "default",
          }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = "var(--tab-off-bg)"; }}
            onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = "transparent"; }}
          >
            <div style={{
              width: 26, height: 26, borderRadius: 7, flexShrink: 0,
              display: "flex", alignItems: "center", justifyContent: "center",
              fontSize: 11, fontWeight: 700,
              background: i === 0 ? "rgba(232,76,31,0.1)" : "var(--tab-off-bg)",
              color:      i === 0 ? "#E84C1F"             : "var(--muted)",
            }}>
              {i + 1}
            </div>
            <div style={{ flex: 1, minWidth: 0 }}>
              <p style={{ margin: 0, fontSize: 13, fontWeight: 600, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>
                {p.name}
              </p>
              <p style={{ margin: 0, fontSize: 11, color: "var(--muted)" }}>{p.sold} terjual</p>
            </div>
            <div style={{ textAlign: "right", flexShrink: 0 }}>
              <p style={{ margin: 0, fontSize: 12, fontWeight: 700 }}>{formatRupiah(p.revenue)}</p>
              <div style={{ display: "flex", alignItems: "center", gap: 3, justifyContent: "flex-end", marginTop: 2 }}>
                {p.trend === "up"
                  ? <TrendingUp size={10} color="#16a34a" />
                  : <TrendingDown size={10} color="#dc2626" />}
                <span style={{ fontSize: 10, color: p.trend === "up" ? "#16a34a" : "#dc2626" }}>
                  {p.trend === "up" ? "naik" : "turun"}
                </span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </>
  );
}