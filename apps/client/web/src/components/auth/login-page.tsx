"use client";

import { FormEvent, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import {
  LoginRole,
  loginCashier,
  loginOwner,
  persistSession,
} from "@/lib/auth";

const roleOptions: Array<{ label: string; value: LoginRole }> = [
  { label: "Kasir", value: "cashier" },
  { label: "Owner", value: "owner" },
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
    identifierPlaceholder: "admin@kopi-senja.com",
    button: "Masuk Ke Dashboard",
  },
};

export function LoginPage() {
  const router = useRouter();
  const [role, setRole] = useState<LoginRole>("owner");
  const [identifier, setIdentifier] = useState("");
  const [password, setPassword] = useState("");
  const [businessId, setBusinessId] = useState(
    process.env.NEXT_PUBLIC_QIOS_BUSINESS_ID ?? "",
  );
  const [error, setError] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [isDark, setIsDark] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  const activeCopy = roleCopy[role];
  const isCashier = role === "cashier";

  const canSubmit = useMemo(() => {
    if (isLoading) return false;
    if (!identifier.trim() || !password.trim()) return false;
    if (isCashier && !businessId.trim()) return false;
    return true;
  }, [businessId, identifier, isCashier, isLoading, password]);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");

    if (!identifier.trim()) {
      setError(`${activeCopy.identifierLabel} wajib diisi.`);
      return;
    }

    if (!password.trim()) {
      setError("Kata sandi wajib diisi.");
      return;
    }

    if (isCashier && !businessId.trim()) {
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
      router.replace(isCashier ? "/kasir/dashboard" : "/owner/dashboard");
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
        className={`pointer-events-none absolute inset-0 ${
          isDark
            ? "bg-[radial-gradient(circle_at_78%_30%,rgba(220,38,38,0.18),transparent_32%),linear-gradient(90deg,rgba(255,255,255,0.05)_1px,transparent_1px)]"
            : "bg-[radial-gradient(circle_at_78%_30%,rgba(218,45,12,0.10),transparent_32%),linear-gradient(90deg,rgba(20,20,20,0.05)_1px,transparent_1px)]"
        } bg-[length:auto,50%_100%]`}
      />

      <div className="relative mx-auto grid min-h-dvh w-full max-w-7xl grid-cols-1 gap-6 px-4 py-5 sm:px-6 md:px-8 lg:h-screen lg:min-h-0 lg:grid-cols-[1fr_0.82fr] lg:items-center lg:gap-10 lg:px-10 lg:py-5 xl:gap-12 xl:px-12 xl:py-7">
        <HeroPanel isDark={isDark} />

        <section className="order-1 flex items-center justify-center pt-14 lg:order-2 lg:h-full lg:pt-0">
          <div className="w-full max-w-[28.5rem]">
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
              <div className="p-5 sm:p-7 lg:p-7 xl:p-8">
              <div className="mb-5 text-center lg:mb-6">
                <h1 className="text-3xl font-black tracking-normal sm:text-[2.15rem]">
                  {activeCopy.title}
                </h1>
                <p
                  className={`mx-auto mt-2 max-w-sm text-sm leading-6 sm:text-base ${
                    isDark ? "text-white/58" : "text-black/52"
                  }`}
                >
                  {activeCopy.description}
                </p>
              </div>

              <div
                className={`mb-5 grid grid-cols-2 rounded-2xl p-1 ${
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
                      className={`h-11 rounded-xl text-sm font-extrabold transition ${
                        isActive
                          ? "bg-[#df2600] text-white shadow-[0_12px_24px_rgba(223,38,0,0.24)]"
                          : isDark
                            ? "text-white/50 hover:text-white"
                            : "text-black/48 hover:text-black"
                      }`}
                      key={option.value}
                      onClick={() => {
                        setRole(option.value);
                        setError("");
                        setIdentifier("");
                      }}
                      role="tab"
                      type="button"
                    >
                      {option.label}
                    </button>
                  );
                })}
              </div>

              <form className="space-y-4" onSubmit={handleSubmit}>
                {isCashier ? (
                  <Field
                    autoComplete="organization"
                    id="business-id"
                    isDark={isDark}
                    label="Business ID"
                    onChange={setBusinessId}
                    placeholder="00000000-0000-0000-0000-000000000000"
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
                  <div className="mb-2 flex items-center justify-between gap-4">
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
                      className={`h-12 w-full rounded-2xl border px-4 pr-16 text-sm font-semibold outline-none transition placeholder:font-medium focus:border-[#df2600] focus:ring-4 focus:ring-[#df2600]/10 sm:h-[3.25rem] sm:px-5 sm:text-base ${
                        isDark
                          ? "border-white/10 bg-white/[0.04] text-white placeholder:text-white/25"
                          : "border-black/8 bg-[#fff8f5] text-black placeholder:text-black/30"
                      }`}
                      id="password"
                      onChange={(event) => setPassword(event.target.value)}
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

                <button
                  className="mt-1 flex h-12 w-full items-center justify-center gap-3 rounded-2xl bg-[#df2600] px-5 text-sm font-black text-white shadow-[0_16px_28px_rgba(223,38,0,0.30)] transition hover:bg-[#c82100] disabled:cursor-not-allowed disabled:bg-black/20 disabled:shadow-none sm:h-[3.25rem] sm:text-base"
                  disabled={!canSubmit}
                  type="submit"
                >
                  {isLoading ? "Memproses..." : activeCopy.button}
                  <span aria-hidden="true">&gt;</span>
                </button>
              </form>
            </div>

              <div
                className={`border-t px-5 py-4 text-center text-xs sm:px-8 sm:text-sm ${
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
      <div className="mb-8 hidden lg:mb-10 lg:flex xl:mb-12">
        <BrandMark />
      </div>

      <h2 className="max-w-3xl text-4xl font-black leading-[1.08] tracking-normal sm:text-5xl lg:text-[clamp(2.55rem,4vw,3.9rem)]">
        <span className="block whitespace-nowrap">Kendali Penuh Atas</span>
        <span
          className="qios-rotator text-[#df2600]"
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
        className={`mt-5 max-w-2xl text-base leading-7 sm:text-lg lg:mt-6 lg:leading-8 ${
          isDark ? "text-white/58" : "text-black/55"
        }`}
      >
        Satu sistem terintegrasi untuk manajemen finansial dan operasional.
        Dirancang khusus untuk UMKM yang siap naik kelas.
      </p>

      <div className="mt-8 space-y-4 sm:mt-9 sm:space-y-5 lg:mt-10">
        {features.map((feature) => (
          <div className="flex items-start gap-4" key={feature.title}>
            <div
              className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl sm:h-11 sm:w-11 ${
                isDark ? "bg-[#df2600]/14" : "bg-[#fff0ea]"
              }`}
            >
              <FeatureIcon name={feature.icon} />
            </div>
            <div>
              <h3 className="text-base font-black sm:text-lg">
                {feature.title}
              </h3>
              <p
                className={`mt-1 text-sm leading-6 sm:text-base ${
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
        className={`mt-8 flex flex-wrap gap-x-5 gap-y-2 text-xs font-semibold uppercase tracking-normal sm:text-sm lg:mt-9 ${
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
    return <span className="text-lg font-black text-[#df2600]">Z</span>;
  }

  if (name === "shield") {
    return <span className="text-base font-black text-[#df2600]">O</span>;
  }

  return (
    <span className="grid grid-cols-2 gap-1">
      <span className="h-2 w-2 rounded-[0.2rem] border-2 border-[#df2600]" />
      <span className="h-2 w-2 rounded-[0.2rem] border-2 border-[#df2600]" />
      <span className="h-2 w-2 rounded-[0.2rem] border-2 border-[#df2600]" />
      <span className="h-2 w-2 rounded-[0.2rem] border-2 border-[#df2600]" />
    </span>
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
        className={`mb-2 block text-sm font-bold ${
          isDark ? "text-white/74" : "text-black/70"
        }`}
      >
        {label}
      </span>
      <input
        autoComplete={autoComplete}
        className={`h-12 w-full rounded-2xl border px-4 text-sm font-semibold outline-none transition placeholder:font-medium focus:border-[#df2600] focus:ring-4 focus:ring-[#df2600]/10 sm:h-[3.25rem] sm:px-5 sm:text-base ${
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
