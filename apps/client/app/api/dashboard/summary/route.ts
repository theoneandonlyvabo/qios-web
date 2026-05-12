/**
 * app/api/dashboard/summary/route.ts
 * Proxy: Next.js → Go /dashboard/summary
 * Query: period = today | this_week | this_month | last_month
 */
import { NextRequest, NextResponse } from "next/server";
 
const GO_API = process.env.GO_API_URL ?? "http://localhost:8080";
 
export async function GET(req: NextRequest) {
  const token  = req.headers.get("authorization");
  const period = req.nextUrl.searchParams.get("period") ?? "today";
 
  const res = await fetch(`${GO_API}/dashboard/summary?period=${period}`, {
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: token } : {}),
    },
    cache: "no-store",
  });
 
  const data = await res.json();
  return NextResponse.json(data, { status: res.status });
}