"use client";

import {
  FormEvent,
  useEffect,
  useMemo,
  useState,
} from "react";
import { useRouter } from "next/navigation";
import {
  LoginRole,
  loginCashier,
  loginCashierQr,
  loginOwner,
  loginOwnerGoogle,
  persistSession,
} from "@/lib/auth";

const roleOptions: Array<{ label: string; value: LoginRole }> = [
  { label: "Kasir", value: "cashier" },
  { label: "Owner", value: "owner" },
];

type CashierLoginMethod = "qr" | "credential";

const cashierMethodOptions: Array<{
  label: string;
  value: CashierLoginMethod;
}> = [
  { label: "Generate QR", value: "qr" },
  { label: "Kode Operator", value: "credential" },
];

const features = [
  {
    icon: "grid",
    title: "Dashboard Terintegrasi",
    description: "Pantau data finansial bisnis dalam satu tampilan intuitif.",
  },
  {
    icon: "bolt",
    title: "Otomasi Transaksi",
    description: "Pencatatan transaksi lebih cepat dan rekonsiliasi lebih rapi.",
  },
  {
    icon: "shield",
    title: "Keamanan Enterprise",
    description: "Akses owner dan kasir dipisahkan dengan scope autentikasi.",
  },
];

const animatedHeadlines = [
  "Data Bisnis Anda.",
  "Arus Kas Anda.",
  "Operasional Anda.",
  "Insight Bisnis Anda.",
];

const GOOGLE_IDENTITY_SCRIPT_ID = "google-identity-services";
const GOOGLE_CLIENT_ID = process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID ?? "";

const roleCopy: Record<
  LoginRole,
  {
    title: string;
    description: string;
    identifierLabel: string;
    identifierPlaceholder: string;
    button: string;
  }
> = {
  cashier: {
    title: "Masuk Kasir",
    description: "Akses workspace kasir untuk mencatat transaksi outlet.",
    identifierLabel: "Operator Code",
    identifierPlaceholder: "kasir-1",
    button: "Masuk Ke Kasir",
  },
  owner: {
    title: "Masuk Owner",
    description: "Akses portal manajemen QIOS untuk memantau bisnis Anda.",
    identifierLabel: "Email Bisnis",
    identifierPlaceholder: "Masukkan email bisnis",
    button: "Masuk Ke Dashboard",
  },
};

type GoogleTokenResponse = {
  access_token?: string;
  error?: string;
  error_description?: string;
};

type GoogleProfile = {
  email?: string;
  name?: string;
};

declare global {
  interface Window {
    google?: {
      accounts: {
        oauth2: {
          initTokenClient: (options: {
            callback: (response: GoogleTokenResponse) => void;
            client_id: string;
            scope: string;
          }) => {
            requestAccessToken: (options?: { prompt?: string }) => void;
          };
        };
      };
    };
  }
}

function createQrLoginCode() {
  const randomPart =
    typeof crypto !== "undefined" && "randomUUID" in crypto
      ? crypto.randomUUID().slice(0, 8)
      : Math.random().toString(36).slice(2, 10);

  return `QIOS-KASIR-${randomPart.toUpperCase()}`;
}

function loadGoogleIdentityScript() {
  if (window.google?.accounts?.oauth2) {
    return Promise.resolve();
  }

  return new Promise<void>((resolve, reject) => {
    const existingScript = document.getElementById(GOOGLE_IDENTITY_SCRIPT_ID);

    if (existingScript) {
      existingScript.addEventListener("load", () => resolve(), { once: true });
      existingScript.addEventListener(
        "error",
        () => reject(new Error("Google login gagal dimuat.")),
        { once: true },
      );
      return;
    }

    const script = document.createElement("script");
    script.async = true;
    script.defer = true;
    script.id = GOOGLE_IDENTITY_SCRIPT_ID;
    script.src = "https://accounts.google.com/gsi/client";
    script.onload = () => resolve();
    script.onerror = () => reject(new Error("Google login gagal dimuat."));
    document.head.appendChild(script);
  });
}

async function fetchGoogleProfile(accessToken: string): Promise<GoogleProfile> {
  const response = await fetch("https://www.googleapis.com/oauth2/v3/userinfo", {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  });

  if (!response.ok) {
    throw new Error("Profil Google tidak bisa dibaca.");
  }

  return (await response.json()) as GoogleProfile;
}

export function LoginPage() {
  const router = useRouter();
  const [role, setRole] = useState<LoginRole>("owner");
  const [identifier, setIdentifier] = useState("");
  const [password, setPassword] = useState("");
  const [businessId, setBusinessId] = useState(
    process.env.NEXT_PUBLIC_QIOS_BUSINESS_ID ?? "",
  );
  const [cashierLoginMethod, setCashierLoginMethod] =
    useState<CashierLoginMethod>("qr");
  const [qrLoginCode, setQrLoginCode] = useState(createQrLoginCode);
  const [qrLoginStatus, setQrLoginStatus] = useState(
    "Menunggu scan dari HP kasir...",
  );
  const [error, setError] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [isGoogleLoading, setIsGoogleLoading] = useState(false);
  const [isDark, setIsDark] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  const activeCopy = roleCopy[role];
  const isCashier = role === "cashier";
  const isCashierQr = isCashier && cashierLoginMethod === "qr";

  useEffect(() => {
    if (!isCashierQr || !qrLoginCode) return;

    let isCancelled = false;
    let isChecking = false;

    const checkQrLogin = async () => {
      if (isChecking || isCancelled) return;
      isChecking = true;

      try {
        const session = await loginCashierQr({ qrToken: qrLoginCode });

        if (isCancelled) return;

        persistSession("cashier", session);
        router.replace("/kasir/dashboard");
      } catch {
        if (!isCancelled) {
          setQrLoginStatus("QR aktif. Scan dari HP kasir untuk login.");
        }
      } finally {
        isChecking = false;
      }
    };

    void checkQrLogin();
    const intervalId = window.setInterval(checkQrLogin, 2500);

    return () => {
      isCancelled = true;
      window.clearInterval(intervalId);
    };
  }, [isCashierQr, qrLoginCode, router]);

  const canSubmit = useMemo(() => {
    if (isLoading || isGoogleLoading) return false;
    if (!identifier.trim() || !password.trim()) return false;
    if (isCashier && !businessId.trim()) return false;
    return true;
  }, [businessId, identifier, isCashier, isGoogleLoading, isLoading, password]);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");

    if (isCashierQr) {
      return;
    }

    if (!isCashierQr && !identifier.trim()) {
      setError(`${activeCopy.identifierLabel} wajib diisi.`);
      return;
    }

    if (!isCashierQr && !password.trim()) {
      setError("Kata sandi wajib diisi.");
      return;
    }

    if (isCashier && !isCashierQr && !businessId.trim()) {
      setError("Business ID wajib diisi untuk login Kasir sesuai API QIOS.");
      return;
    }

    setIsLoading(true);

    try {
      const session = isCashier
        ? await loginCashier({
              businessId: businessId.trim(),
              operatorCode: identifier.trim(),
              password,
            })
        : await loginOwner({
            email: identifier.trim(),
            password,
          });

      persistSession(role, session);
      router.replace(isCashier ? "/kasir/dashboard" : "/dashboard");
    } catch (caughtError) {
      setError(
        caughtError instanceof Error
          ? caughtError.message
          : "Login gagal. Silakan coba lagi.",
      );
    } finally {
      setIsLoading(false);
    }
  }

  async function handleOwnerGoogleLogin() {
    setError("");

    if (!GOOGLE_CLIENT_ID) {
      setError("Google Client ID belum dikonfigurasi untuk login email bisnis.");
      return;
    }

    setIsGoogleLoading(true);

    try {
      await loadGoogleIdentityScript();

      const tokenClient = window.google?.accounts.oauth2.initTokenClient({
        callback: (response) => {
          void completeOwnerGoogleLogin(response);
        },
        client_id: GOOGLE_CLIENT_ID,
        scope: "openid email profile",
      });

      if (!tokenClient) {
        throw new Error("Google login belum siap.");
      }

      tokenClient.requestAccessToken({ prompt: "select_account" });
      window.setTimeout(() => setIsGoogleLoading(false), 60000);
    } catch (caughtError) {
      setIsGoogleLoading(false);
      setError(
        caughtError instanceof Error
          ? caughtError.message
          : "Login Google gagal dimuat.",
      );
    }
  }

  async function completeOwnerGoogleLogin(response: GoogleTokenResponse) {
    try {
      if (response.error) {
        throw new Error(
          response.error_description ?? "Akun Google belum dipilih.",
        );
      }

      if (!response.access_token) {
        throw new Error("Akun Google belum dipilih.");
      }

      const profile = await fetchGoogleProfile(response.access_token);

      if (!profile.email || !profile.name) {
        throw new Error("Email bisnis dari Google tidak valid.");
      }

      const session = await loginOwnerGoogle({
        email: profile.email,
        name: profile.name,
      });

      persistSession("owner", session);
      router.replace("/dashboard");
    } catch (caughtError) {
      setError(
        caughtError instanceof Error
          ? caughtError.message
          : "Email bisnis belum terdaftar sebagai owner.",
      );
      setIsGoogleLoading(false);
    }
  }

  return (
    <main
      className={`relative min-h-dvh overflow-x-hidden transition-colors duration-300 lg:h-screen lg:overflow-hidden ${
        isDark
          ? "bg-[#0d0d0f] text-white"
          : "bg-[#fffaf7] text-[#141414]"
      }`}
    >
      <ThemeToggle isDark={isDark} onToggle={() => setIsDark((value) => !value)} />

      <div
        className={`qios-page-glow pointer-events-none absolute inset-0 ${
          isDark
            ? "bg-[radial-gradient(circle_at_78%_30%,rgba(220,38,38,0.18),transparent_32%),linear-gradient(90deg,rgba(255,255,255,0.05)_1px,transparent_1px)]"
            : "bg-[radial-gradient(circle_at_78%_30%,rgba(218,45,12,0.10),transparent_32%),linear-gradient(90deg,rgba(20,20,20,0.05)_1px,transparent_1px)]"
        } bg-[length:auto,50%_100%]`}
      />

      <div className="relative mx-auto grid min-h-dvh w-full max-w-7xl grid-cols-1 gap-6 px-4 py-5 sm:px-6 md:px-8 lg:h-screen lg:min-h-0 lg:grid-cols-[1fr_0.82fr] lg:items-center lg:gap-10 lg:px-10 lg:py-5 xl:gap-12 xl:px-12 xl:py-7">
        <HeroPanel isDark={isDark} />

        <section className="qios-login-enter order-1 flex items-center justify-center pt-14 lg:order-2 lg:h-full lg:pt-0">
          <div className="w-full max-w-[25.75rem]">
            <div className="mb-5 flex lg:hidden">
              <BrandMark />
            </div>

            <div
              className={`w-full overflow-hidden rounded-[1.75rem] border shadow-[0_24px_70px_rgba(211,47,18,0.14)] backdrop-blur ${
                isDark
                  ? "border-white/10 bg-[#171719]/92"
                  : "border-black/10 bg-white/82"
              }`}
            >
              <div className="p-5 sm:p-6 lg:p-5 xl:p-6">
                <div
                  className="qios-auth-copy mb-4 text-center lg:mb-5"
                  key={role}
                >
                  <h1 className="text-3xl font-black tracking-normal sm:text-[1.95rem]">
                    {activeCopy.title}
                  </h1>
                  <p
                    className={`mx-auto mt-2 max-w-sm text-sm leading-6 ${
                      isDark ? "text-white/58" : "text-black/52"
                    }`}
                  >
                    {activeCopy.description}
                  </p>
                </div>

                <div
                  className={`mb-4 grid grid-cols-2 rounded-2xl p-1 ${
                    isDark ? "bg-white/8" : "bg-[#f3ebe7]"
                  }`}
                  role="tablist"
                  aria-label="Pilih role login"
                >
                  {roleOptions.map((option) => {
                    const isActive = option.value === role;

                    return (
                      <button
                        aria-selected={isActive}
                        className={`h-10 rounded-xl text-sm font-extrabold transition-all duration-300 ease-out active:scale-[0.98] ${
                          isActive
                            ? "scale-[1.01] bg-[#df2600] text-white shadow-[0_12px_24px_rgba(223,38,0,0.24)]"
                            : isDark
                              ? "text-white/50 hover:text-white"
                              : "text-black/48 hover:text-black"
                        }`}
                        key={option.value}
                        onClick={() => {
                          setRole(option.value);
                          setError("");
                          setIdentifier("");
                          setPassword("");
                        }}
                        role="tab"
                        type="button"
                      >
                        {option.label}
                      </button>
                    );
                  })}
                </div>

                <form
                  className="qios-auth-panel space-y-3"
                  key={`${role}-${cashierLoginMethod}`}
                  onSubmit={handleSubmit}
                >
                  {isCashier ? (
                    <div
                      className={`grid grid-cols-2 rounded-2xl p-1 ${
                        isDark ? "bg-white/8" : "bg-[#f3ebe7]"
                      }`}
                      role="tablist"
                      aria-label="Pilih metode login kasir"
                    >
                      {cashierMethodOptions.map((option) => {
                        const isActive = option.value === cashierLoginMethod;

                        return (
                          <button
                            aria-selected={isActive}
                            className={`h-9 rounded-xl text-xs font-extrabold transition-all duration-300 ease-out active:scale-[0.98] sm:text-sm ${
                              isActive
                                ? "scale-[1.01] bg-[#df2600] text-white shadow-[0_10px_20px_rgba(223,38,0,0.22)]"
                                : isDark
                                  ? "text-white/50 hover:text-white"
                                  : "text-black/48 hover:text-black"
                            }`}
                            key={option.value}
                            onClick={() => {
                              setCashierLoginMethod(option.value);
                              setError("");
                              setQrLoginStatus("");
                            }}
                            role="tab"
                            type="button"
                          >
                            {option.label}
                          </button>
                        );
                      })}
                    </div>
                  ) : null}

                  {isCashierQr ? (
                    <GeneratedQrLoginPanel
                      code={qrLoginCode}
                      isDark={isDark}
                      onGenerate={() => {
                        setQrLoginCode(createQrLoginCode());
                        setQrLoginStatus("QR baru dibuat. Scan dari HP kasir.");
                      }}
                      status={qrLoginStatus}
                    />
                  ) : (
                    <>
                      {isCashier ? (
                        <Field
                          autoComplete="organization"
                          id="business-id"
                          isDark={isDark}
                          label="Business ID"
                          onChange={setBusinessId}
                          placeholder="Masukkan ID bisnis"
                          value={businessId}
                        />
                      ) : null}

                      <Field
                        autoComplete={isCashier ? "username" : "email"}
                        id="identifier"
                        inputMode={isCashier ? "text" : "email"}
                        isDark={isDark}
                        label={activeCopy.identifierLabel}
                        onChange={setIdentifier}
                        placeholder={activeCopy.identifierPlaceholder}
                        type={isCashier ? "text" : "email"}
                        value={identifier}
                      />

                      <div>
                        <div className="mb-1.5 flex items-center justify-between gap-4">
                          <label
                            className={`text-sm font-bold ${
                              isDark ? "text-white/74" : "text-black/70"
                            }`}
                            htmlFor="password"
                          >
                            Kata Sandi
                          </label>
                          <button
                            className="text-sm font-bold text-[#df2600] hover:text-[#ff3b12]"
                            type="button"
                          >
                            Lupa sandi?
                          </button>
                        </div>
                        <div className="relative">
                          <input
                            autoComplete="current-password"
                            className={`h-11 w-full rounded-2xl border px-4 pr-16 text-sm font-semibold outline-none transition placeholder:font-medium focus:border-[#df2600] focus:ring-4 focus:ring-[#df2600]/10 sm:h-12 sm:px-5 sm:text-base ${
                              isDark
                                ? "border-white/10 bg-white/[0.04] text-white placeholder:text-white/25"
                                : "border-black/8 bg-[#fff8f5] text-black placeholder:text-black/30"
                            }`}
                            id="password"
                            onChange={(event) =>
                              setPassword(event.target.value)
                            }
                            placeholder="Masukkan kata sandi"
                            type={showPassword ? "text" : "password"}
                            value={password}
                          />
                          <button
                            aria-label={
                              showPassword
                                ? "Sembunyikan kata sandi"
                                : "Tampilkan kata sandi"
                            }
                            className={`absolute right-3 top-1/2 flex h-8 min-w-12 -translate-y-1/2 items-center justify-center rounded-full px-2 text-xs font-bold ${
                              isDark
                                ? "text-white/50 hover:bg-white/10 hover:text-white"
                                : "text-black/45 hover:bg-black/5 hover:text-black"
                            }`}
                            onClick={() => setShowPassword((value) => !value)}
                            type="button"
                          >
                            {showPassword ? "Hide" : "Show"}
                          </button>
                        </div>
                      </div>
                    </>
                  )}

                  {error ? (
                    <div
                      className={`rounded-2xl border px-4 py-3 text-sm font-semibold ${
                        isDark
                          ? "border-red-400/25 bg-red-500/10 text-red-200"
                          : "border-red-100 bg-red-50 text-red-700"
                      }`}
                    >
                      {error}
                    </div>
                  ) : null}

                  {!isCashierQr ? (
                    <button
                      className="mt-1 flex h-11 w-full items-center justify-center gap-3 rounded-2xl bg-[#df2600] px-5 text-sm font-black text-white shadow-[0_16px_28px_rgba(223,38,0,0.30)] transition hover:bg-[#c82100] disabled:cursor-not-allowed disabled:bg-black/20 disabled:shadow-none sm:h-12 sm:text-base"
                      disabled={!canSubmit}
                      type="submit"
                    >
                      {isLoading ? "Memproses..." : activeCopy.button}
                      <span aria-hidden="true">&gt;</span>
                    </button>
                  ) : null}

                  {!isCashier ? (
                    <OwnerGoogleLoginButton
                      disabled={isLoading || isGoogleLoading}
                      isDark={isDark}
                      isLoading={isGoogleLoading}
                      onClick={handleOwnerGoogleLogin}
                    />
                  ) : null}
                </form>
              </div>

              <div
                className={`border-t px-5 py-3 text-center text-xs sm:px-8 sm:text-sm ${
                  isDark
                    ? "border-white/10 bg-white/[0.03] text-white/42"
                    : "border-black/8 bg-[#fff7f3] text-black/45"
                }`}
              >
                Belum punya akun?{" "}
                <span className="font-extrabold text-[#df2600]">
                  Hubungi Skalar
                </span>
              </div>
            </div>
          </div>
        </section>
      </div>
    </main>
  );
}

function OwnerGoogleLoginButton({
  disabled,
  isDark,
  isLoading,
  onClick,
}: {
  disabled: boolean;
  isDark: boolean;
  isLoading: boolean;
  onClick: () => void;
}) {
  return (
    <div className="pt-1">
      <div
        className={`mb-3 flex items-center gap-3 text-[0.65rem] font-black uppercase tracking-wide ${
          isDark ? "text-white/28" : "text-black/28"
        }`}
      >
        <span
          className={`h-px flex-1 ${isDark ? "bg-white/10" : "bg-black/8"}`}
        />
        <span>atau</span>
        <span
          className={`h-px flex-1 ${isDark ? "bg-white/10" : "bg-black/8"}`}
        />
      </div>
      <button
        className={`flex h-11 w-full items-center justify-center gap-3 rounded-2xl border px-5 text-sm font-black transition disabled:cursor-not-allowed disabled:opacity-60 sm:h-12 ${
          isDark
            ? "border-white/10 bg-white/[0.04] text-white hover:bg-white/[0.08]"
            : "border-black/8 bg-white text-[#141414] shadow-[0_12px_26px_rgba(17,17,17,0.06)] hover:border-[#df2600]/20 hover:bg-[#fff8f5]"
        }`}
        disabled={disabled}
        onClick={onClick}
        type="button"
      >
        <GoogleLogo />
        {isLoading ? "Membuka pilihan email..." : "Masuk dengan Email Bisnis"}
      </button>
    </div>
  );
}

function GoogleLogo() {
  return (
    <span className="flex h-7 w-7 items-center justify-center rounded-full bg-white">
      <svg
        aria-hidden="true"
        className="h-4 w-4"
        viewBox="0 0 24 24"
      >
        <path
          d="M21.805 10.023h-9.58v3.955h5.514c-.238 1.277-.96 2.36-2.04 3.087v2.565h3.302c1.934-1.781 3.049-4.405 3.049-7.514 0-.713-.064-1.4-.245-2.093Z"
          fill="#4285F4"
        />
        <path
          d="M12.225 22c2.76 0 5.077-.913 6.769-2.47l-3.302-2.565c-.913.612-2.08.974-3.467.974-2.666 0-4.923-1.8-5.728-4.22H3.084v2.648C4.765 19.704 8.22 22 12.225 22Z"
          fill="#34A853"
        />
        <path
          d="M6.497 13.719a6.01 6.01 0 0 1 0-3.837V7.234H3.084a10.006 10.006 0 0 0 0 9.133l3.413-2.648Z"
          fill="#FBBC05"
        />
        <path
          d="M12.225 5.962c1.502 0 2.85.516 3.91 1.529l2.934-2.934C17.3 2.913 14.983 2 12.225 2 8.22 2 4.765 4.296 3.084 7.234l3.413 2.648c.805-2.42 3.062-3.92 5.728-3.92Z"
          fill="#EA4335"
        />
      </svg>
    </span>
  );
}

function GeneratedQrLoginPanel({
  code,
  isDark,
  onGenerate,
  status,
}: {
  code: string;
  isDark: boolean;
  onGenerate: () => void;
  status: string;
}) {
  const displayCode = code || "QIOS-KASIR";

  return (
    <div
      className={`rounded-2xl border p-3.5 ${
        isDark
          ? "border-white/10 bg-white/[0.04]"
          : "border-black/8 bg-[#fff8f5]"
      }`}
    >
      <div className="grid gap-3 sm:grid-cols-[7.25rem_1fr]">
        <div
          className={`relative flex aspect-square items-center justify-center overflow-hidden rounded-2xl border p-3 ${
            isDark
              ? "border-white/10 bg-black/30"
              : "border-[#f0d8d0] bg-white"
          }`}
        >
          <QrCodeSvg value={displayCode} />
          <span className="absolute left-2.5 top-2.5 h-4 w-4 border-l-2 border-t-2 border-[#df2600]" />
          <span className="absolute right-2.5 top-2.5 h-4 w-4 border-r-2 border-t-2 border-[#df2600]" />
          <span className="absolute bottom-2.5 left-2.5 h-4 w-4 border-b-2 border-l-2 border-[#df2600]" />
          <span className="absolute bottom-2.5 right-2.5 h-4 w-4 border-b-2 border-r-2 border-[#df2600]" />
        </div>

        <div className="flex min-w-0 flex-col justify-center">
          <p className="text-sm font-black text-[#df2600]">
            Generate QR login
          </p>
          <p
            className={`mt-1 text-xs leading-5 ${
              isDark ? "text-white/54" : "text-black/50"
            }`}
          >
            Scan QR ini dari HP kasir. Web akan login otomatis setelah QR disetujui.
          </p>

          <div className="mt-3 flex gap-2">
            <button
              className="h-9 flex-1 rounded-xl bg-[#df2600] px-3 text-xs font-black text-white shadow-[0_10px_20px_rgba(223,38,0,0.20)] transition hover:bg-[#c82100]"
              onClick={onGenerate}
              type="button"
            >
              {code ? "Generate Ulang" : "Generate QR"}
            </button>
          </div>
        </div>
      </div>

      <div
        className={`mt-3 rounded-2xl border px-4 py-3 ${
          isDark
            ? "border-white/10 bg-white/[0.04]"
            : "border-black/8 bg-white"
        }`}
      >
        <p
          className={`text-[0.65rem] font-black uppercase tracking-wide ${
            isDark
              ? "text-white/42"
              : "text-black/32"
          }`}
        >
          Kode QR
        </p>
        <p
          className={`mt-1 break-all text-sm font-black ${
            code
              ? isDark
                ? "text-white"
                : "text-[#141414]"
              : isDark
                ? "text-white/35"
                : "text-black/32"
          }`}
        >
          {code || "Belum digenerate"}
        </p>
        {status ? (
          <p
            className={`mt-2 text-xs font-semibold ${
              isDark ? "text-white/48" : "text-black/45"
            }`}
          >
            {status}
          </p>
        ) : null}
      </div>
    </div>
  );
}

const QR_ALPHANUMERIC_CHARS = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ $%*+-./:";
const QR_SIZE = 21;

function QrCodeSvg({ value }: { value: string }) {
  const matrix = createQrMatrix(value);

  return (
    <svg
      aria-label="QR login kasir"
      className="h-full w-full"
      role="img"
      viewBox={`0 0 ${QR_SIZE} ${QR_SIZE}`}
    >
      <rect fill="white" height={QR_SIZE} width={QR_SIZE} />
      {matrix.map((row, rowIndex) =>
        row.map((isDark, columnIndex) =>
          isDark ? (
            <rect
              fill="#df2600"
              height="1"
              key={`${rowIndex}-${columnIndex}`}
              width="1"
              x={columnIndex}
              y={rowIndex}
            />
          ) : null,
        ),
      )}
    </svg>
  );
}

function createQrMatrix(rawValue: string) {
  const value =
    rawValue
      .toUpperCase()
      .split("")
      .filter((character) => QR_ALPHANUMERIC_CHARS.includes(character))
      .join("")
      .slice(0, 25) || "QIOS-KASIR";
  const dataCodewords = createQrDataCodewords(value);
  const errorCodewords = createQrErrorCodewords(dataCodewords, 7);
  const codewordBits = [...dataCodewords, ...errorCodewords].flatMap((codeword) =>
    numberToBits(codeword, 8),
  );
  const matrix: Array<Array<boolean | null>> = Array.from(
    { length: QR_SIZE },
    () => Array.from({ length: QR_SIZE }, () => null),
  );
  const reserved = Array.from({ length: QR_SIZE }, () =>
    Array.from({ length: QR_SIZE }, () => false),
  );

  const setFunctionModule = (row: number, column: number, isDark: boolean) => {
    if (row < 0 || column < 0 || row >= QR_SIZE || column >= QR_SIZE) return;
    matrix[row][column] = isDark;
    reserved[row][column] = true;
  };

  drawFinderPattern(0, 0, setFunctionModule);
  drawFinderPattern(0, QR_SIZE - 7, setFunctionModule);
  drawFinderPattern(QR_SIZE - 7, 0, setFunctionModule);

  for (let index = 8; index < QR_SIZE - 8; index += 1) {
    const isDark = index % 2 === 0;
    setFunctionModule(6, index, isDark);
    setFunctionModule(index, 6, isDark);
  }

  setFunctionModule(13, 8, true);
  drawFormatBits(matrix, reserved, 0);

  let bitIndex = 0;
  let upward = true;

  for (let rightColumn = QR_SIZE - 1; rightColumn >= 1; rightColumn -= 2) {
    if (rightColumn === 6) rightColumn -= 1;

    for (let verticalIndex = 0; verticalIndex < QR_SIZE; verticalIndex += 1) {
      const row = upward ? QR_SIZE - 1 - verticalIndex : verticalIndex;

      for (let columnOffset = 0; columnOffset < 2; columnOffset += 1) {
        const column = rightColumn - columnOffset;

        if (reserved[row][column]) continue;

        const shouldMask = (row + column) % 2 === 0;
        const bit = codewordBits[bitIndex] === 1;
        matrix[row][column] = bit !== shouldMask;
        bitIndex += 1;
      }
    }

    upward = !upward;
  }

  drawFormatBits(matrix, reserved, 0);

  return matrix.map((row) => row.map(Boolean));
}

function createQrDataCodewords(value: string) {
  const bits: number[] = [];
  const appendBits = (number: number, length: number) => {
    bits.push(...numberToBits(number, length));
  };

  appendBits(0b0010, 4);
  appendBits(value.length, 9);

  for (let index = 0; index < value.length; index += 2) {
    const first = QR_ALPHANUMERIC_CHARS.indexOf(value[index]);
    const second =
      index + 1 < value.length
        ? QR_ALPHANUMERIC_CHARS.indexOf(value[index + 1])
        : -1;

    if (second >= 0) {
      appendBits(first * 45 + second, 11);
    } else {
      appendBits(first, 6);
    }
  }

  const capacityBits = 19 * 8;
  const terminatorLength = Math.min(4, capacityBits - bits.length);
  appendBits(0, terminatorLength);

  while (bits.length % 8 !== 0) bits.push(0);

  const codewords: number[] = [];
  for (let index = 0; index < bits.length; index += 8) {
    codewords.push(bitsToNumber(bits.slice(index, index + 8)));
  }

  const padCodewords = [0xec, 0x11];
  let padIndex = 0;
  while (codewords.length < 19) {
    codewords.push(padCodewords[padIndex % 2]);
    padIndex += 1;
  }

  return codewords;
}

function createQrErrorCodewords(data: number[], degree: number) {
  const generator = createReedSolomonGenerator(degree);
  const result = Array.from({ length: degree }, () => 0);

  for (const codeword of data) {
    const factor = codeword ^ result.shift()!;
    result.push(0);

    for (let index = 0; index < degree; index += 1) {
      result[index] ^= gfMultiply(generator[index + 1], factor);
    }
  }

  return result;
}

function createReedSolomonGenerator(degree: number) {
  let result = [1];

  for (let index = 0; index < degree; index += 1) {
    result = multiplyPolynomials(result, [1, gfPow(2, index)]);
  }

  return result;
}

function multiplyPolynomials(left: number[], right: number[]) {
  const result = Array.from({ length: left.length + right.length - 1 }, () => 0);

  left.forEach((leftValue, leftIndex) => {
    right.forEach((rightValue, rightIndex) => {
      result[leftIndex + rightIndex] ^= gfMultiply(leftValue, rightValue);
    });
  });

  return result;
}

function gfPow(value: number, power: number) {
  let result = 1;

  for (let index = 0; index < power; index += 1) {
    result = gfMultiply(result, value);
  }

  return result;
}

function gfMultiply(left: number, right: number) {
  let result = 0;
  let value = left;
  let multiplier = right;

  while (multiplier > 0) {
    if ((multiplier & 1) !== 0) result ^= value;
    value <<= 1;
    if ((value & 0x100) !== 0) value ^= 0x11d;
    multiplier >>= 1;
  }

  return result;
}

function drawFinderPattern(
  row: number,
  column: number,
  setModule: (row: number, column: number, isDark: boolean) => void,
) {
  for (let deltaRow = -1; deltaRow <= 7; deltaRow += 1) {
    for (let deltaColumn = -1; deltaColumn <= 7; deltaColumn += 1) {
      const targetRow = row + deltaRow;
      const targetColumn = column + deltaColumn;
      const isInFinder =
        deltaRow >= 0 && deltaRow <= 6 && deltaColumn >= 0 && deltaColumn <= 6;
      const isDark =
        isInFinder &&
        (deltaRow === 0 ||
          deltaRow === 6 ||
          deltaColumn === 0 ||
          deltaColumn === 6 ||
          (deltaRow >= 2 &&
            deltaRow <= 4 &&
            deltaColumn >= 2 &&
            deltaColumn <= 4));

      setModule(targetRow, targetColumn, isDark);
    }
  }
}

function drawFormatBits(
  matrix: Array<Array<boolean | null>>,
  reserved: boolean[][],
  mask: number,
) {
  const formatBits = createFormatBits(mask);
  const setFormatModule = (row: number, column: number, bitIndex: number) => {
    matrix[row][column] = ((formatBits >> bitIndex) & 1) !== 0;
    reserved[row][column] = true;
  };

  for (let index = 0; index <= 5; index += 1) setFormatModule(8, index, index);
  setFormatModule(8, 7, 6);
  setFormatModule(8, 8, 7);
  setFormatModule(7, 8, 8);
  for (let index = 9; index < 15; index += 1) {
    setFormatModule(14 - index, 8, index);
  }

  for (let index = 0; index < 8; index += 1) {
    setFormatModule(QR_SIZE - 1 - index, 8, index);
  }
  for (let index = 8; index < 15; index += 1) {
    setFormatModule(8, QR_SIZE - 15 + index, index);
  }
}

function createFormatBits(mask: number) {
  const errorCorrectionLevel = 0b01;
  const data = (errorCorrectionLevel << 3) | mask;
  let remainder = data << 10;

  for (let bitIndex = 14; bitIndex >= 10; bitIndex -= 1) {
    if (((remainder >> bitIndex) & 1) !== 0) {
      remainder ^= 0x537 << (bitIndex - 10);
    }
  }

  return ((data << 10) | remainder) ^ 0x5412;
}

function numberToBits(value: number, length: number) {
  return Array.from({ length }, (_, index) => (value >> (length - 1 - index)) & 1);
}

function bitsToNumber(bits: number[]) {
  return bits.reduce((result, bit) => (result << 1) | bit, 0);
}

function ThemeToggle({
  isDark,
  onToggle,
}: {
  isDark: boolean;
  onToggle: () => void;
}) {
  return (
    <div className="fixed right-4 top-4 z-20 flex items-center gap-2 sm:right-5 sm:top-5 sm:gap-3">
      <span
        className={`hidden text-sm font-bold sm:block ${
          isDark ? "text-white/70" : "text-black/55"
        }`}
      >
        {isDark ? "Dark" : "Light"}
      </span>
      <button
        aria-label="Toggle dark mode"
        aria-pressed={isDark}
        className={`flex h-9 w-[4.1rem] items-center rounded-full border p-1 transition ${
          isDark
            ? "justify-end border-white/15 bg-white/12"
            : "justify-start border-black/10 bg-white"
        } shadow-sm`}
        onClick={onToggle}
        type="button"
      >
        <span className="flex h-7 w-7 items-center justify-center rounded-full bg-[#141414] text-xs font-black text-white shadow-md">
          {isDark ? "D" : "L"}
        </span>
      </button>
    </div>
  );
}

function BrandMark() {
  return (
    <div className="flex items-center gap-3">
      <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-[#df2600] text-lg font-black text-white shadow-[0_14px_24px_rgba(223,38,0,0.22)] sm:h-12 sm:w-12 sm:text-xl">
        Q
      </div>
      <span className="text-2xl font-black tracking-normal">QIOS</span>
    </div>
  );
}

function HeroPanel({ isDark }: { isDark: boolean }) {
  return (
    <section className="order-2 flex flex-col justify-center pb-8 pt-2 lg:order-1 lg:h-full lg:min-h-0 lg:py-0">
      <div className="qios-hero-brand mb-6 hidden pt-4 lg:mb-7 lg:flex xl:mb-8">
        <BrandMark />
      </div>

      <h2 className="qios-hero-title max-w-3xl text-4xl font-black leading-[1.08] tracking-normal sm:text-5xl lg:text-[clamp(2.55rem,4vw,3.9rem)]">
        <span className="block lg:whitespace-nowrap">Kendali Penuh Atas</span>
        <span
          className="qios-rotator text-[clamp(2rem,9vw,2.5rem)] text-[#df2600] sm:text-5xl lg:text-[clamp(2.55rem,4vw,3.9rem)]"
          aria-live="polite"
        >
          {animatedHeadlines.map((headline, index) => (
            <span
              className="qios-rotator__item"
              key={headline}
              style={{ animationDelay: `${index * 2.4}s` }}
            >
              {headline}
            </span>
          ))}
        </span>
      </h2>

      <p
        className={`qios-hero-copy mt-5 max-w-2xl text-base leading-7 sm:text-lg lg:mt-6 lg:leading-8 ${
          isDark ? "text-white/58" : "text-black/55"
        }`}
      >
        Satu sistem terintegrasi untuk manajemen finansial dan operasional.
        Dirancang khusus untuk UMKM yang siap naik kelas.
      </p>

      <div className="qios-hero-showcase mt-8 grid gap-4 sm:mt-9 lg:mt-10 lg:grid-cols-[1fr_16rem]">
        <div className="grid gap-3">
          {features.map((feature) => (
            <div
              className={`group flex items-center gap-4 rounded-2xl border p-3.5 shadow-[0_14px_34px_rgba(17,17,17,0.045)] transition ${
                isDark
                  ? "border-white/10 bg-white/[0.04]"
                  : "border-black/8 bg-white/72"
              }`}
              key={feature.title}
            >
              <div
                className={`flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl ${
                  isDark ? "bg-[#df2600]/14" : "bg-[#fff0ea]"
                }`}
              >
                <FeatureIcon name={feature.icon} />
              </div>
              <div className="min-w-0">
                <h3 className="text-[0.98rem] font-extrabold leading-5 tracking-[-0.01em]">
                  {feature.title}
                </h3>
                <p
                  className={`mt-1.5 line-clamp-2 text-[0.9rem] leading-6 ${
                    isDark ? "text-white/50" : "text-black/48"
                  }`}
                >
                  {feature.description}
                </p>
              </div>
            </div>
          ))}
        </div>

        <div
          className={`relative overflow-hidden rounded-3xl border p-5 shadow-[0_18px_44px_rgba(223,38,0,0.08)] ${
            isDark
              ? "border-white/10 bg-[#df2600]/10"
              : "border-[#f4d7cd] bg-[#fff0ea]"
          }`}
        >
          <div className="absolute -right-10 -top-10 h-28 w-28 rounded-full bg-[#df2600]/10" />
          <div className="absolute -bottom-16 left-10 h-28 w-28 rounded-full bg-white/70" />

          <div className="relative flex h-full flex-col justify-between">
            <div>
              <div className="mb-5 flex h-12 w-12 items-center justify-center rounded-2xl bg-[#df2600] text-white shadow-[0_12px_24px_rgba(223,38,0,0.22)]">
                <svg
                  aria-hidden="true"
                  className="h-6 w-6"
                  fill="none"
                  viewBox="0 0 24 24"
                >
                  <path
                    d="M5 18V9M12 18V5M19 18v-7"
                    stroke="currentColor"
                    strokeLinecap="round"
                    strokeWidth="2.4"
                  />
                </svg>
              </div>

              <p className="text-base font-black leading-5 text-[#df2600]">
                Satu Sistem untuk Bisnis Bertumbuh
              </p>
              <p
                className={`mt-3 text-sm leading-6 ${
                  isDark ? "text-white/56" : "text-black/52"
                }`}
              >
                QIOS menyatukan pencatatan, pemantauan, dan insight operasional
                dalam pengalaman yang sederhana.
              </p>
            </div>

            <div className="mt-5 flex flex-wrap gap-2">
              {["Real-time", "Terstruktur", "Insightful"].map((item) => (
                <span
                  className="rounded-full bg-white/80 px-3 py-1.5 text-[0.7rem] font-black uppercase tracking-wide text-[#df2600] shadow-sm"
                  key={item}
                >
                  {item}
                </span>
              ))}
            </div>
          </div>
        </div>
      </div>

      <div
        className={`mt-6 flex flex-wrap gap-x-5 gap-y-2 text-xs font-semibold uppercase tracking-normal sm:text-sm lg:mt-7 ${
          isDark ? "text-white/30" : "text-black/30"
        }`}
      >
        <span>(c) 2026 QIOS System</span>
        <span className={isDark ? "text-white/14" : "text-black/14"}>|</span>
        <span>Powered by Skalar</span>
      </div>
    </section>
  );
}

function FeatureIcon({ name }: { name: string }) {
  if (name === "bolt") {
    return (
      <svg
        aria-hidden="true"
        className="h-5 w-5 text-[#df2600]"
        fill="none"
        viewBox="0 0 24 24"
      >
        <path
          d="M13 2 5 13h6l-1 9 9-13h-6l1-7Z"
          fill="currentColor"
        />
      </svg>
    );
  }

  if (name === "shield") {
    return (
      <svg
        aria-hidden="true"
        className="h-5 w-5 text-[#df2600]"
        fill="none"
        viewBox="0 0 24 24"
      >
        <path
          d="M12 3 19 6.2v5.7c0 4.1-2.7 7.4-7 9.1-4.3-1.7-7-5-7-9.1V6.2L12 3Z"
          stroke="currentColor"
          strokeLinejoin="round"
          strokeWidth="2"
        />
        <path
          d="m9.4 12 1.8 1.8 3.6-4"
          stroke="currentColor"
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="2"
        />
      </svg>
    );
  }

  return (
    <svg
      aria-hidden="true"
      className="h-5 w-5 text-[#df2600]"
      fill="none"
      viewBox="0 0 24 24"
    >
      <rect height="6" rx="2" stroke="currentColor" strokeWidth="2" width="6" x="4" y="4" />
      <rect height="6" rx="2" stroke="currentColor" strokeWidth="2" width="6" x="14" y="4" />
      <rect height="6" rx="2" stroke="currentColor" strokeWidth="2" width="6" x="4" y="14" />
      <rect height="6" rx="2" stroke="currentColor" strokeWidth="2" width="6" x="14" y="14" />
    </svg>
  );
}

type FieldProps = {
  autoComplete?: string;
  id: string;
  inputMode?: "email" | "text";
  isDark: boolean;
  label: string;
  onChange: (value: string) => void;
  placeholder: string;
  type?: "email" | "password" | "text";
  value: string;
};

function Field({
  autoComplete,
  id,
  inputMode,
  isDark,
  label,
  onChange,
  placeholder,
  type = "text",
  value,
}: FieldProps) {
  return (
    <label className="block" htmlFor={id}>
      <span
        className={`mb-1.5 block text-sm font-bold ${
          isDark ? "text-white/74" : "text-black/70"
        }`}
      >
        {label}
      </span>
      <input
        autoComplete={autoComplete}
        className={`h-11 w-full rounded-2xl border px-4 text-sm font-semibold outline-none transition placeholder:font-medium focus:border-[#df2600] focus:ring-4 focus:ring-[#df2600]/10 sm:h-12 sm:px-5 sm:text-base ${
          isDark
            ? "border-white/10 bg-white/[0.04] text-white placeholder:text-white/25"
            : "border-black/8 bg-[#fff8f5] text-black placeholder:text-black/30"
        }`}
        id={id}
        inputMode={inputMode}
        onChange={(event) => onChange(event.target.value)}
        placeholder={placeholder}
        type={type}
        value={value}
      />
    </label>
  );
}
