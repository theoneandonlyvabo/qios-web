"use client";

import { useState, useEffect, useCallback } from "react";
import {
  LineChart, Line, XAxis, YAxis, CartesianGrid,
  Tooltip, ResponsiveContainer,
} from "recharts";
import {
  ShoppingCart, DollarSign, TrendingUp,
  Users, ArrowUpRight, ArrowDownRight, Clock, Package,
  RefreshCw,
} from "lucide-react";
import { apiFetch } from "@/lib/api";

// ─── Types (sesuai qios-api.yml) ─────────────────────────────────────────────

type DashboardSummary = {
  period: string;
  total_revenue: number;
  total_transactions: number;
  avg_transaction_value: number;
  revenue_change_pct: number;
  transaction_change_pct: number;
};

type DailyTrend = {
  date: string;
  total_transactions: number;
  total_revenue: number;
};

type TopProduct = {
  product_id: string;
  product_name: string;
  total_sold: number;
  total_revenue: number;
  category: string | null;
};

type HourlyDistribution = {
  hour: number;
  total_transactions: number;
  avg_transactions: number;
};

type ApiResponse<T> = { success: boolean; data: T; error: string | null };

type Period = "today" | "this_week" | "this_month" | "last_month";

const PERIOD_LABEL: Record<Period, string> = {
  today: "Hari ini",
  this_week: "Minggu ini",
  this_month: "Bulan ini",
  last_month: "Bulan lalu",
};

// ─── Helpers ──────────────────────────────────────────────────────────────────

function formatRupiah(n: number) {
  return "Rp " + n.toLocaleString("id-ID");
}

function formatPct(n: number) {
  return (n >= 0 ? "+" : "") + n.toFixed(1) + "%";
}

function last7Days() {
  const end = new Date();
  const start = new Date();
  start.setDate(end.getDate() - 6);
  const fmt = (d: Date) => d.toISOString().split("T")[0];
  return { start_date: fmt(start), end_date: fmt(end) };
}

function shortDay(dateStr: string) {
  const days = ["Min", "Sen", "Sel", "Rab", "Kam", "Jum", "Sab"];
  return days[new Date(dateStr).getDay()];
}

// ─── Tooltip ──────────────────────────────────────────────────────────────────

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
      <p style={{ fontSize: 13, fontWeight: 700, margin: 0 }}>
        {formatRupiah(payload[0].value)}
      </p>
    </div>
  );
}

// ─── Peak Hours Bar ───────────────────────────────────────────────────────────

function PeakHoursBar({ data }: { data: HourlyDistribution[] }) {
  if (!data.length) return <div style={{ height: 64, color: "var(--muted)", fontSize: 12, display: "flex", alignItems: "center" }}>Belum ada data</div>;
  const max = Math.max(...data.map((h) => h.total_transactions));
  return (
    <div style={{ display: "flex", alignItems: "flex-end", gap: 4, height: 64 }}>
      {data.map((h) => {
        const pct    = max > 0 ? (h.total_transactions / max) * 100 : 0;
        const isPeak = h.total_transactions === max;
        return (
          <div key={h.hour} style={{ flex: 1, display: "flex", flexDirection: "column", alignItems: "center", gap: 3, height: "100%" }}>
            <div style={{ flex: 1, width: "100%", display: "flex", alignItems: "flex-end" }}>
              <div style={{
                width: "100%", borderRadius: 3, height: `${pct}%`,
                background: isPeak ? "#E84C1F" : "var(--bar-bg)",
              }} />
            </div>
            <span style={{ fontSize: 9, color: "var(--muted)" }}>
              {h.hour === 8 || h.hour === 12 || h.hour === 19 ? String(h.hour).padStart(2, "0") : ""}
            </span>
          </div>
        );
      })}
    </div>
  );
}

// ─── Skeleton ─────────────────────────────────────────────────────────────────

function Skeleton({ h = 20 }: { h?: number }) {
  return (
    <div style={{
      width: "100%", height: h, borderRadius: 8,
      background: "var(--bar-bg)",
      animation: "skpulse 1.5s ease-in-out infinite",
    }} />
  );
}

// ─── Page ─────────────────────────────────────────────────────────────────────

export default function DashboardPage() {
  const [chartMode, setChartMode] = useState<"revenue" | "transactions">("revenue");
  const [period, setPeriod]       = useState<Period>("today");

  const [summary,     setSummary]     = useState<DashboardSummary | null>(null);
  const [trend,       setTrend]       = useState<DailyTrend[]>([]);
  const [topProducts, setTopProducts] = useState<TopProduct[]>([]);
  const [peakHours,   setPeakHours]   = useState<HourlyDistribution[]>([]);
  const [loading,     setLoading]     = useState(true);
  const [error,       setError]       = useState<string | null>(null);

  const fetchAll = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const { start_date, end_date } = last7Days();
      const [s, t, tp, ph] = await Promise.all([
        apiFetch<ApiResponse<DashboardSummary>>(`/api/dashboard/summary?period=${period}`),
        apiFetch<ApiResponse<DailyTrend[]>>(`/api/dashboard/trend?start_date=${start_date}&end_date=${end_date}`),
        apiFetch<ApiResponse<TopProduct[]>>(`/api/dashboard/top-products?period=${period}&limit=5`),
        apiFetch<ApiResponse<HourlyDistribution[]>>(`/api/dashboard/peak-hours?period=${period}`),
      ]);
      if (s.success)  setSummary(s.data);
      if (t.success)  setTrend(t.data);
      if (tp.success) setTopProducts(tp.data);
      if (ph.success) setPeakHours(ph.data);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Gagal memuat data");
    } finally {
      setLoading(false);
    }
  }, [period]);

  useEffect(() => { fetchAll(); }, [fetchAll]);

  const today = new Date().toLocaleDateString("id-ID", {
    weekday: "long", year: "numeric", month: "long", day: "numeric",
  });

  const chartData = trend.map((d) => ({
    day: shortDay(d.date),
    revenue: d.total_revenue,
    transactions: d.total_transactions,
  }));

  const peakHour = peakHours.length
    ? peakHours.reduce((a, b) => a.total_transactions > b.total_transactions ? a : b)
    : null;

  const card: React.CSSProperties = {
    background: "var(--card-bg)",
    border: "1px solid var(--card-border)",
    borderRadius: 16,
  };

  return (
    <>
      <style>{`
        :root {
          --card-bg: #ffffff; --card-border: #f0ece8;
          --muted: #999; --bar-bg: #e8e3dc;
          --tab-on-bg: #111; --tab-on-text: #fff;
          --tab-off-bg: #f0ece8; --tab-off-text: #666;
          --grid-line: #eee;
        }
        html.dark {
          --card-bg: #1a1a1a; --card-border: #232323;
          --muted: #555; --bar-bg: #2a2a2a;
          --tab-on-bg: #fff; --tab-on-text: #111;
          --tab-off-bg: #222; --tab-off-text: #666;
          --grid-line: #222;
        }
        @keyframes skpulse { 0%,100%{opacity:1} 50%{opacity:0.4} }
      `}</style>

      {/* Header */}
      <div style={{ marginBottom: 24 }}>
        <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: 10, flexWrap: "wrap", gap: 8 }}>
          <div style={{ display: "flex", alignItems: "center", gap: 6, background: "rgba(22,163,74,0.1)", color: "#16a34a", padding: "5px 10px", borderRadius: 8, fontSize: 11, fontWeight: 700 }}>
            <span style={{ width: 6, height: 6, borderRadius: "50%", background: "#16a34a", display: "inline-block" }} />
            Data Langsung
          </div>

          {/* Period filter */}
          <div style={{ display: "flex", gap: 5, flexWrap: "wrap" }}>
            {(Object.keys(PERIOD_LABEL) as Period[]).map((p) => (
              <button key={p} onClick={() => setPeriod(p)} style={{
                padding: "5px 10px", borderRadius: 7, border: "none", cursor: "pointer",
                fontSize: 11, fontWeight: 600,
                background: period === p ? "#E84C1F" : "var(--tab-off-bg)",
                color:      period === p ? "#fff"    : "var(--tab-off-text)",
              }}>
                {PERIOD_LABEL[p]}
              </button>
            ))}
            <button onClick={fetchAll} style={{
              padding: "5px 8px", borderRadius: 7, border: "none", cursor: "pointer",
              background: "var(--tab-off-bg)", color: "var(--muted)", display: "flex",
            }}>
              <RefreshCw size={13} />
            </button>
          </div>
        </div>
        <h1 style={{ margin: 0, fontSize: 22, fontWeight: 800, letterSpacing: -0.5 }}>Dashboard</h1>
        <p style={{ margin: "4px 0 0", fontSize: 12, color: "var(--muted)" }}>{today}</p>
      </div>

      {/* Error */}
      {error && (
        <div style={{ background: "rgba(220,38,38,0.1)", border: "1px solid rgba(220,38,38,0.2)", borderRadius: 10, padding: "10px 14px", marginBottom: 16, fontSize: 13, color: "#dc2626" }}>
          ⚠ {error} —{" "}
          <button onClick={fetchAll} style={{ background: "none", border: "none", color: "#dc2626", fontWeight: 700, cursor: "pointer" }}>
            Coba lagi
          </button>
        </div>
      )}

      {/* Metric Cards */}
      <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(160px, 1fr))", gap: 12, marginBottom: 14 }}>
        {[
          { label: "Revenue",       value: formatRupiah(summary?.total_revenue ?? 0),        change: summary ? formatPct(summary.revenue_change_pct) : null,     positive: (summary?.revenue_change_pct ?? 0) >= 0,     icon: DollarSign,  sub: "vs periode lalu" },
          { label: "Total Transaksi", value: String(summary?.total_transactions ?? 0),        change: summary ? formatPct(summary.transaction_change_pct) : null, positive: (summary?.transaction_change_pct ?? 0) >= 0, icon: ShoppingCart, sub: "vs periode lalu" },
          { label: "Rata-rata Order", value: formatRupiah(summary?.avg_transaction_value ?? 0), change: null, positive: true, icon: TrendingUp, sub: "per transaksi" },
          { label: "Operator",      value: "-",  change: null, positive: true, icon: Users, sub: "aktif" },
        ].map(({ label, value, change, positive, icon: Icon, sub }) => (
          <div key={label} style={{ ...card, padding: 16 }}>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginBottom: 12 }}>
              <span style={{ fontSize: 12, color: "var(--muted)" }}>{label}</span>
              <div style={{ width: 30, height: 30, borderRadius: 8, background: "rgba(232,76,31,0.1)", color: "#E84C1F", display: "flex", alignItems: "center", justifyContent: "center" }}>
                <Icon size={15} />
              </div>
            </div>
            {loading ? <Skeleton h={26} /> : <p style={{ margin: 0, fontSize: 18, fontWeight: 800 }}>{value}</p>}
            <div style={{ display: "flex", alignItems: "center", gap: 4, marginTop: 8 }}>
              {change && (positive ? <ArrowUpRight size={12} color="#16a34a" /> : <ArrowDownRight size={12} color="#dc2626" />)}
              {change && <span style={{ fontSize: 11, fontWeight: 700, color: positive ? "#16a34a" : "#dc2626" }}>{change}</span>}
              <span style={{ fontSize: 11, color: "var(--muted)" }}>{sub}</span>
            </div>
          </div>
        ))}
      </div>

      {/* Chart */}
      <div style={{ ...card, padding: 20, marginBottom: 12 }}>
        <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: 16, flexWrap: "wrap", gap: 8 }}>
          <div>
            <p style={{ margin: 0, fontSize: 15, fontWeight: 700 }}>Tren Transaksi</p>
            <p style={{ margin: "2px 0 0", fontSize: 12, color: "var(--muted)" }}>7 hari terakhir</p>
          </div>
          <div style={{ display: "flex", gap: 6 }}>
            {(["revenue", "transactions"] as const).map((mode) => (
              <button key={mode} onClick={() => setChartMode(mode)} style={{
                padding: "6px 14px", borderRadius: 8, border: "none", cursor: "pointer",
                fontSize: 12, fontWeight: 600,
                background: chartMode === mode ? "var(--tab-on-bg)" : "var(--tab-off-bg)",
                color:      chartMode === mode ? "var(--tab-on-text)" : "var(--tab-off-text)",
              }}>
                {mode === "revenue" ? "Revenue" : "Transaksi"}
              </button>
            ))}
          </div>
        </div>
        {loading ? <Skeleton h={180} /> : (
          <ResponsiveContainer width="100%" height={180}>
            <LineChart data={chartData} margin={{ top: 4, right: 4, left: -20, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--grid-line)" vertical={false} />
              <XAxis dataKey="day" tick={{ fontSize: 11, fill: "var(--muted)" as string }} axisLine={false} tickLine={false} />
              <YAxis tick={{ fontSize: 10, fill: "var(--muted)" as string }} axisLine={false} tickLine={false}
                tickFormatter={(v) => chartMode === "revenue" ? `${(v/1000000).toFixed(1)}jt` : `${v}`}
              />
              <Tooltip content={<CustomTooltip />} />
              <Line type="monotone" dataKey={chartMode === "revenue" ? "revenue" : "transactions"}
                stroke="#E84C1F" strokeWidth={2.5}
                dot={{ fill: "#E84C1F", strokeWidth: 0, r: 3.5 }}
                activeDot={{ r: 5, fill: "#E84C1F", strokeWidth: 2, stroke: "var(--card-bg)" }}
              />
            </LineChart>
          </ResponsiveContainer>
        )}
      </div>

      {/* Peak hours + ringkasan */}
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 12, marginBottom: 12 }}>
        <div style={{ ...card, padding: 16 }}>
          <div style={{ display: "flex", alignItems: "center", gap: 7, marginBottom: 12 }}>
            <Clock size={14} color="var(--muted)" />
            <p style={{ margin: 0, fontSize: 13, fontWeight: 700 }}>Jam Sibuk</p>
          </div>
          {loading ? <Skeleton h={64} /> : <PeakHoursBar data={peakHours} />}
          <p style={{ margin: "8px 0 0", fontSize: 11, color: "var(--muted)" }}>
            {peakHour ? <>Puncak: <strong>{String(peakHour.hour).padStart(2,"0")}:00</strong></> : "Belum ada data"}
          </p>
        </div>
        <div style={{ ...card, padding: 16 }}>
          <p style={{ margin: "0 0 10px", fontSize: 13, fontWeight: 700 }}>Ringkasan</p>
          {[
            { label: "Transaksi", value: String(summary?.total_transactions ?? "-"), icon: ShoppingCart },
            { label: "Jam puncak", value: peakHour ? `${String(peakHour.hour).padStart(2,"0")}:00` : "-", icon: Clock },
            { label: "Avg order",  value: formatRupiah(summary?.avg_transaction_value ?? 0), icon: Package },
          ].map(({ label, value, icon: Icon }) => (
            <div key={label} style={{ display: "flex", alignItems: "center", justifyContent: "space-between", padding: "7px 0", borderBottom: "1px solid var(--card-border)" }}>
              <div style={{ display: "flex", alignItems: "center", gap: 7 }}>
                <Icon size={13} color="var(--muted)" />
                <span style={{ fontSize: 12, color: "var(--muted)" }}>{label}</span>
              </div>
              <span style={{ fontSize: 12, fontWeight: 700 }}>{loading ? "..." : value}</span>
            </div>
          ))}
        </div>
      </div>

      {/* Top Products */}
      <div style={{ ...card, padding: 20 }}>
        <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: 16 }}>
          <p style={{ margin: 0, fontSize: 15, fontWeight: 700 }}>Produk Terlaris</p>
          <span style={{ background: "#E84C1F", color: "#fff", fontSize: 11, fontWeight: 700, borderRadius: 6, padding: "3px 9px" }}>
            {PERIOD_LABEL[period]}
          </span>
        </div>
        {loading ? (
          <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
            {[1,2,3,4,5].map((i) => <Skeleton key={i} h={44} />)}
          </div>
        ) : topProducts.length === 0 ? (
          <p style={{ color: "var(--muted)", fontSize: 13, textAlign: "center", padding: "20px 0" }}>
            Belum ada data produk untuk periode ini
          </p>
        ) : topProducts.map((p, i) => (
          <div key={p.product_id} style={{ display: "flex", alignItems: "center", gap: 12, padding: "10px 8px", borderRadius: 10, cursor: "default" }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = "var(--tab-off-bg)"; }}
            onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = "transparent"; }}
          >
            <div style={{ width: 26, height: 26, borderRadius: 7, flexShrink: 0, display: "flex", alignItems: "center", justifyContent: "center", fontSize: 11, fontWeight: 700, background: i === 0 ? "rgba(232,76,31,0.1)" : "var(--tab-off-bg)", color: i === 0 ? "#E84C1F" : "var(--muted)" }}>
              {i + 1}
            </div>
            <div style={{ flex: 1, minWidth: 0 }}>
              <p style={{ margin: 0, fontSize: 13, fontWeight: 600, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>{p.product_name}</p>
              <p style={{ margin: 0, fontSize: 11, color: "var(--muted)" }}>{p.total_sold} terjual{p.category ? ` · ${p.category}` : ""}</p>
            </div>
            <div style={{ textAlign: "right", flexShrink: 0 }}>
              <p style={{ margin: 0, fontSize: 12, fontWeight: 700 }}>{formatRupiah(p.total_revenue)}</p>
            </div>
          </div>
        ))}
      </div>
    </>
  );
}