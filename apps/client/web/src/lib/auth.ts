export type LoginRole = "cashier" | "owner";

type ApiResponse<TData> = {
  success: boolean;
  data: TData | null;
  error: string | null;
};

type OwnerLoginData = {
  access_token: string;
  user: {
    id: string;
    name: string;
    email: string;
    role: "owner";
    business?: unknown;
  };
};

type CashierLoginData = {
  access_token: string;
  operator: {
    id: string;
    business_id: string;
    name: string;
    operator_code: string;
    is_active: boolean;
  };
  business: {
    id: string;
    name: string;
    category?: string;
    qris_static_payload?: string | null;
  };
};

export type AuthSession =
  | {
      accessToken: string;
      role: "owner";
      profile: OwnerLoginData["user"];
    }
  | {
      accessToken: string;
      role: "cashier";
      profile: CashierLoginData["operator"];
      business: CashierLoginData["business"];
    };

const API_BASE_URL =
  process.env.NEXT_PUBLIC_QIOS_API_URL?.replace(/\/$/, "") ??
  "http://localhost:8080";

export async function loginOwner(credentials: {
  email: string;
  password: string;
}): Promise<AuthSession> {
  const payload = await postJson<OwnerLoginData>("/auth/login", credentials);

  return {
    accessToken: payload.access_token,
    role: "owner",
    profile: payload.user,
  };
}

export async function loginCashier(credentials: {
  businessId: string;
  operatorCode: string;
  password: string;
}): Promise<AuthSession> {
  const payload = await postJson<CashierLoginData>("/operator/auth/login", {
    business_id: credentials.businessId,
    operator_code: credentials.operatorCode,
    password: credentials.password,
  });

  return {
    accessToken: payload.access_token,
    role: "cashier",
    profile: payload.operator,
    business: payload.business,
  };
}

export async function loginCashierQr(credentials: {
  qrToken: string;
}): Promise<AuthSession> {
  const payload = await postJson<CashierLoginData>("/kasir/auth/login/qr", {
    qr_token: credentials.qrToken,
  });

  return {
    accessToken: payload.access_token,
    role: "cashier",
    profile: payload.operator,
    business: payload.business,
  };
}

export function persistSession(role: LoginRole, session: AuthSession) {
  window.localStorage.setItem("qios.auth.access_token", session.accessToken);
  window.localStorage.setItem("qios.auth.role", role);
  window.localStorage.setItem("qios.auth.profile", JSON.stringify(session));
}

async function postJson<TData>(
  path: string,
  body: Record<string, string>,
): Promise<TData> {
  let response: Response;

  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      body: JSON.stringify(body),
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      method: "POST",
    });
  } catch {
    throw new Error("Tidak bisa terhubung ke server QIOS.");
  }

  const payload = await parseApiResponse<TData>(response);

  if (!response.ok || !payload.success || !payload.data) {
    throw new Error(payload.error ?? "Credential tidak valid.");
  }

  return payload.data;
}

async function parseApiResponse<TData>(
  response: Response,
): Promise<ApiResponse<TData>> {
  try {
    const payload = (await response.json()) as ApiResponse<TData>;

    return {
      data: payload.data ?? null,
      error: payload.error ?? null,
      success: Boolean(payload.success),
    };
  } catch {
    return {
      data: null,
      error: response.ok
        ? "Response server tidak valid."
        : `Request gagal dengan status ${response.status}.`,
      success: false,
    };
  }
}
