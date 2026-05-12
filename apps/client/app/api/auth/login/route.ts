import { NextRequest, NextResponse } from "next/server";

const GO_API = process.env.GO_API_URL ?? "http://localhost:8080";

// ─── TEST BYPASS ──────────────────────────────────────────────────────────────
// Remove this block before production deploy.
const TEST_USER = {
  email: "admin-qios",
  password: "admin123",
  response: {
    success: true,
    data: {
      access_token: "test-token-bypass",
      user: {
        id: "test-user-id",
        name: "Admin QIOS",
        email: "admin@qios.id",
        role: "owner",
        business: {
          id: "test-business-id",
          name: "QIOS Demo Business",
          xendit_status: "ACTIVE",
        },
      },
    },
    error: null,
  },
};
// ─────────────────────────────────────────────────────────────────────────────

export async function POST(req: NextRequest) {
  const body = await req.json();

  if (body.email === TEST_USER.email && body.password === TEST_USER.password) {
    return NextResponse.json(TEST_USER.response);
  }

  const res = await fetch(`${GO_API}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });

  const data = await res.json();
  const response = NextResponse.json(data, { status: res.status });

  const setCookie = res.headers.get("set-cookie");
  if (setCookie) response.headers.set("set-cookie", setCookie);

  return response;
}
