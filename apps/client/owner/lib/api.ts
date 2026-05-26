// ==========================================
// HTTP CLIENT TERPUSAT — semua API call lewat sini
// Tidak ada fetch() langsung di komponen
// ==========================================

const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080';

type RequestOptions = Omit<RequestInit, 'body'> & {
  body?: unknown;
};

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { body, headers, ...rest } = options;

  const response = await fetch(`${BASE_URL}${path}`, {
    ...rest,
    headers: {
      'Content-Type': 'application/json',
      ...headers,
    },
    body: body !== undefined ? JSON.stringify(body) : undefined,
    credentials: 'include', // kirim HttpOnly cookie refresh token
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: response.statusText }));
    throw new Error(error?.error ?? `Request failed: ${response.status}`);
  }

  return response.json() as Promise<T>;
}

// ---- Auth ----
export const authApi = {
  logout: () => request<void>('/auth/logout', { method: 'POST' }),
  refresh: () => request<{ data: { access_token: string } }>('/auth/refresh', { method: 'POST' }),
};

// ---- Dashboard ----
export const dashboardApi = {
  getSummary: (period?: string) =>
    request(`/dashboard/summary${period ? `?period=${period}` : ''}`),
  getTrend: (startDate: string, endDate: string) =>
    request(`/dashboard/transactions/trend?start_date=${startDate}&end_date=${endDate}`),
  getTopProducts: (period?: string, limit = 5) =>
    request(`/dashboard/products/top?limit=${limit}${period ? `&period=${period}` : ''}`),
  getPeakHours: (period?: string) =>
    request(`/dashboard/transactions/peak-hours${period ? `?period=${period}` : ''}`),
};

// ---- Transactions ----
export const transactionsApi = {
  list: (params?: Record<string, string>) => {
    const query = params ? '?' + new URLSearchParams(params).toString() : '';
    return request(`/transactions${query}`);
  },
  getById: (id: string) => request(`/transactions/${id}`),
  void: (id: string, voidReason: string) =>
    request(`/transactions/${id}/void`, { method: 'POST', body: { void_reason: voidReason } }),
};

// ---- Operators ----
export const operatorsApi = {
  list: () => request('/business/operators'),
  create: (data: { name: string; operator_code: string; password: string }) =>
    request('/business/operators', { method: 'POST', body: data }),
  update: (id: string, data: { name?: string; is_active?: boolean }) =>
    request(`/business/operators/${id}`, { method: 'PUT', body: data }),
  delete: (id: string) => request(`/business/operators/${id}`, { method: 'DELETE' }),
  regenerateQr: (id: string) =>
    request(`/business/operators/${id}/regenerate-qr`, { method: 'POST' }),
};

// ---- Products ----
export const productsApi = {
  list: (params?: Record<string, string>) => {
    const query = params ? '?' + new URLSearchParams(params).toString() : '';
    return request(`/products${query}`);
  },
  getById: (id: string) => request(`/products/${id}`),
};

// ---- Analytics ----
export const analyticsApi = {
  getOverview: (startDate: string, endDate: string, compareWith = 'previous_period') =>
    request(
      `/analytics/overview?start_date=${startDate}&end_date=${endDate}&compare_with=${compareWith}`
    ),
};

// ---- Insight ----
export const insightApi = {
  list: () => request('/insight'),
};
