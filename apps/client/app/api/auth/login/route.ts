import { NextRequest, NextResponse } from "next/server";

const GO_API = process.env.GO_API_URL ?? "http://localhost:8080";

export async function POST(req: NextRequest) {
  const body = await req.json();

  const res = await fetch(`${GO_API}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });

  const data = await res.json();
  const response = NextResponse.json(data, { status: res.status });

  if (res.ok && data?.data?.access_token) {
    response.cookies.set("qios_access_token", data.data.access_token, {
      httpOnly: true,
      secure: process.env.NODE_ENV === "production",
      sameSite: "strict",
      maxAge: 900,
      path: "/",
    });
  }

  const setCookie = res.headers.get("set-cookie");
  if (setCookie) response.headers.append("set-cookie", setCookie);

  return response;
}
