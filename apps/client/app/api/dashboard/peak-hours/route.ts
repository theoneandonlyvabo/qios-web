/**
 * app/api/dashboard/peak-hours/route.ts
 * Proxy: Next.js → Go /dashboard/transactions/peak-hours
 * Query: period
 */
import { NextRequest, NextResponse } from "next/server";
import { isTestBypass, MOCK } from "@/app/api/_test-bypass";

const GO_API = process.env.GO_API_URL ?? "http://localhost:8080";
 
export async function GET(req: NextRequest) {
  const token  = req.headers.get("authorization");
  if (isTestBypass(req)) return NextResponse.json(MOCK.peakHours);
  const period = req.nextUrl.searchParams.get("period") ?? "today";
 
  const res = await fetch(
    `${GO_API}/dashboard/transactions/peak-hours?period=${period}`,
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
 