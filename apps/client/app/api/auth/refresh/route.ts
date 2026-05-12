/**
 * app/api/auth/refresh/route.ts
 * Proxy: Next.js → Go /auth/refresh
 * Cookie httpOnly refresh_token diteruskan otomatis dari browser ke Go
 */
import { NextRequest, NextResponse } from "next/server";
 
const GO_API = process.env.GO_API_URL ?? "http://localhost:8080";
 
export async function POST(req: NextRequest) {
  const cookieHeader = req.headers.get("cookie") ?? "";
 
  const res = await fetch(`${GO_API}/auth/refresh`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...(cookieHeader ? { Cookie: cookieHeader } : {}),
    },
  });
 
  const data = await res.json();
  const response = NextResponse.json(data, { status: res.status });
 
  const setCookie = res.headers.get("set-cookie");
  if (setCookie) response.headers.set("set-cookie", setCookie);
 
  return response;
}
 