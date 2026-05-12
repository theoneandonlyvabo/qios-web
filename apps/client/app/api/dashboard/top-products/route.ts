/**
 * app/api/dashboard/top-products/route.ts
 * Proxy: Next.js → Go /dashboard/products/top
 * Query: period, limit
 */
import { NextRequest, NextResponse } from "next/server";
 
const GO_API = process.env.GO_API_URL ?? "http://localhost:8080";
 
export async function GET(req: NextRequest) {
  const token  = req.headers.get("authorization");
  const period = req.nextUrl.searchParams.get("period") ?? "today";
  const limit  = req.nextUrl.searchParams.get("limit")  ?? "5";
 
  const res = await fetch(
    `${GO_API}/dashboard/products/top?period=${period}&limit=${limit}`,
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
 