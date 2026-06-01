import type { PaymentMethod } from "@/types";

const API_URL = process.env.NEXT_PUBLIC_QIOS_API_URL ?? "http://localhost:8080";

type ApiResponse<T> = {
  success: boolean;
  data: T;
  error: string | null;
};

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_URL}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {})
    },
    credentials: "include"
  });

  const json = (await response.json()) as ApiResponse<T>;

  if (!response.ok || !json.success) {
    throw new Error(json.error ?? `Request failed: ${response.status}`);
  }

  return json.data;
}

export const operatorApi = {
  loginQr(qrToken: string) {
    return request<{ access_token: string }>("/kasir/auth/login/qr", {
      method: "POST",
      body: JSON.stringify({ qr_token: qrToken })
    });
  },

  loginCredential(body: { business_id: string; operator_code: string; password: string }) {
    return request<{ access_token: string }>("/kasir/auth/login", {
      method: "POST",
      body: JSON.stringify(body)
    });
  },

  getProducts() {
    return request<unknown>("/products?is_available=true");
  },

  createOrder(items: Array<{ product_id: string; quantity: number }>) {
    return request<unknown>("/orders", {
      method: "POST",
      body: JSON.stringify({ items })
    });
  },

  confirmCheckout(orderId: string, body: { payment_method: PaymentMethod; note?: string }) {
    return request<unknown>(`/orders/${orderId}/checkout/confirm`, {
      method: "POST",
      body: JSON.stringify(body)
    });
  },

  todayTransactions() {
    return request<unknown>("/transactions");
  },

  voidOrder(orderId: string, voidReason: string) {
    return request<unknown>(`/orders/${orderId}/void`, {
      method: "POST",
      body: JSON.stringify({ void_reason: voidReason })
    });
  }
};

export function loginWithQR(qr_token: string) {
  return request<{ access_token: string }>("/kasir/auth/login/qr", {
    method: "POST",
    body: JSON.stringify({ qr_token })
  });
}

export function loginWithCredential(body: {
  business_id: string;
  operator_code: string;
  password: string;
}) {
  return request<{ access_token: string }>("/kasir/auth/login", {
    method: "POST",
    body: JSON.stringify(body)
  });
}

