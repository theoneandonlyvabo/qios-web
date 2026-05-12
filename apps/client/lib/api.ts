/**
 * lib/api.ts
 * Base fetch helper — token disimpan di memory, bukan localStorage (aturan CLAUDE.md)
 * Semua request ke Go HARUS lewat /api/... bukan langsung ke :8080
 */

let accessToken: string | null = null;

export function setAccessToken(token: string) { accessToken = token; }
export function getAccessToken() { return accessToken; }
export function clearAccessToken() { accessToken = null; }

type FetchOptions = {
  method?: "GET" | "POST" | "PATCH" | "PUT" | "DELETE";
  body?: unknown;
  params?: Record<string, string | number | undefined>;
};

export async function apiFetch<T>(path: string, options: FetchOptions = {}): Promise<T> {
  const { method = "GET", body, params } = options;

  let url = path;
  if (params) {
    const qs = new URLSearchParams();
    Object.entries(params).forEach(([k, v]) => { if (v !== undefined) qs.set(k, String(v)); });
    const str = qs.toString();
    if (str) url += "?" + str;
  }

  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (accessToken) headers["Authorization"] = `Bearer ${accessToken}`;

  const res = await fetch(url, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  if (res.status === 401) {
    const refreshed = await tryRefresh();
    if (refreshed) {
      headers["Authorization"] = `Bearer ${accessToken}`;
      const retry = await fetch(url, { method, headers, body: body ? JSON.stringify(body) : undefined });
      if (!retry.ok) throw new ApiError(retry.status, "Unauthorized setelah refresh");
      return retry.json();
    }
    clearAccessToken();
    if (typeof window !== "undefined") window.location.href = "/login";
    throw new ApiError(401, "Sesi habis, silakan login kembali");
  }

  if (!res.ok) {
    const errBody = await res.json().catch(() => ({}));
    throw new ApiError(res.status, errBody?.error ?? "Terjadi kesalahan");
  }

  return res.json();
}

async function tryRefresh(): Promise<boolean> {
  try {
    const res = await fetch("/api/auth/refresh", { method: "POST" });
    if (!res.ok) return false;
    const json = await res.json();
    if (json?.data?.access_token) { setAccessToken(json.data.access_token); return true; }
    return false;
  } catch { return false; }
}

export class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = "ApiError";
  }
}