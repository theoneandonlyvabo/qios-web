// Remove this file before production deploy.

export const TEST_TOKEN = "test-token-bypass";

export const isTestBypass = (req: { headers: { get(k: string): string | null }; cookies: { get(k: string): { value: string } | undefined } }) => {
  const auth = req.headers.get("authorization");
  if (auth === `Bearer ${TEST_TOKEN}`) return true;
  const cookie = req.cookies.get("qios_access_token")?.value;
  return cookie === TEST_TOKEN;
};

const today = new Date().toISOString().split("T")[0];
const days = ["Min", "Sen", "Sel", "Rab", "Kam", "Jum", "Sab"];

function last7(): { date: string }[] {
  return Array.from({ length: 7 }, (_, i) => {
    const d = new Date();
    d.setDate(d.getDate() - (6 - i));
    return { date: d.toISOString().split("T")[0] };
  });
}

export const MOCK = {
  summary: {
    success: true,
    data: {
      period: "today",
      total_revenue: 4750000,
      total_transactions: 23,
      avg_transaction_value: 206521,
      revenue_change_pct: 12.5,
      transaction_change_pct: 8.3,
    },
    error: null,
  },

  trend: {
    success: true,
    data: last7().map((d, i) => ({
      date: d.date,
      total_transactions: 15 + Math.floor(Math.random() * 20),
      total_revenue: 2000000 + i * 300000 + Math.floor(Math.random() * 500000),
    })),
    error: null,
  },

  topProducts: {
    success: true,
    data: [
      { product_id: "p1", product_name: "Nasi Goreng Spesial", total_sold: 42, total_revenue: 1470000, category: "Makanan" },
      { product_id: "p2", product_name: "Es Teh Manis", total_sold: 38, total_revenue: 380000, category: "Minuman" },
      { product_id: "p3", product_name: "Ayam Bakar", total_sold: 29, total_revenue: 1305000, category: "Makanan" },
      { product_id: "p4", product_name: "Jus Alpukat", total_sold: 21, total_revenue: 630000, category: "Minuman" },
      { product_id: "p5", product_name: "Soto Ayam", total_sold: 18, total_revenue: 630000, category: "Makanan" },
    ],
    error: null,
  },

  peakHours: {
    success: true,
    data: Array.from({ length: 14 }, (_, i) => {
      const hour = i + 8;
      const tx = hour === 12 ? 18 : hour === 13 ? 15 : hour === 18 ? 12 : 2 + Math.floor(Math.random() * 6);
      return { hour, total_transactions: tx, avg_transactions: tx };
    }),
    error: null,
  },
};
