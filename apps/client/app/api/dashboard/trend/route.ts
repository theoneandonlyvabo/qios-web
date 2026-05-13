/**
 * app/api/dashboard/trend/route.ts
 * Proxy: Next.js → Go /dashboard/transactions/trend
 * Query: start_date, end_date (format YYYY-MM-DD)
 */
import { NextRequest, NextResponse } from "next/server";

const GO_API = process.env.GO_API_URL ?? "http://localhost:8080";

export async function GET(req: NextRequest) {
  const token     = req.headers.get("authorization");
  const startDate = req.nextUrl.searchParams.get("start_date") ?? "";
  const endDate   = req.nextUrl.searchParams.get("end_date") ?? "";

  const qs = new URLSearchParams();
  if (startDate) qs.set("start_date", startDate);
  if (endDate)   qs.set("end_date", endDate);

  const res = await fetch(
    `${GO_API}/dashboard/transactions/trend${qs.toString() ? "?" + qs : ""}`,
    {
      headers: {
        "Content-Type": "application/json",
        ...(token ? { Authorization: token } : {}),
      },
      cache: "no-store",
    }
  );

  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}
