  "use client";

  import { useState } from "react";
  import { useRouter } from "next/navigation";
  import { Icon } from "@/components/icons";
  import { OperatorShell } from "@/components/operator-shell";
  import { QiosLogo } from "@/components/qios-logo";

  export default function LoginPage() {
    const router = useRouter();
    const [showCodeLogin, setShowCodeLogin] = useState(false);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState("");

    function mockLogin() {
      setError("");
      setLoading(true);
      window.setTimeout(() => router.push("/order"), 650);
    }

    function simulateInvalidQr() {
      setError("QR tidak valid. Coba scan ulang atau login dengan code.");
    }

    return (
      <OperatorShell withNav={false}>
        <main className="flex min-h-dvh flex-col bg-[radial-gradient(circle_at_50%_24%,rgba(255,181,158,0.12),transparent_20rem)] px-4 py-8 safe-pb">
          <div className="flex flex-1 flex-col items-center justify-center">
            <div className="mb-14 flex flex-col items-center gap-5 text-center">
              <div className="flex h-20 w-20 items-center justify-center rounded-[28px] bg-card-high text-primary shadow-glow">
                <Icon name="pos" className="h-9 w-9" />
              </div>
              <div>
                <h1 className="text-xl font-extrabold text-foreground">QIOS Operator</h1>
                <p className="mt-1 text-sm font-medium text-primary">Kasir cepat untuk bisnis kamu</p>
              </div>
            </div>

            <div className="w-full space-y-4">
              <button
                onClick={mockLogin}
                className="flex h-16 w-full items-center justify-center gap-3 rounded-2xl bg-primary text-base font-extrabold text-primary-foreground shadow-glow transition active:scale-[0.98] disabled:opacity-70"
                disabled={loading}
                type="button"
              >
                {loading ? (
                  "Membuka kasir..."
                ) : (
                  <>
                    <Icon name="qr" className="h-5 w-5" />
                    Scan QR Operator
                  </>
                )}
              </button>

              <p className="mx-auto max-w-[260px] text-center text-xs font-bold leading-relaxed text-muted-foreground">
                Gunakan QR dari dashboard owner untuk masuk lebih cepat.
              </p>

              <div className="flex items-center gap-3 py-1">
                <span className="h-px flex-1 bg-border" />
                <span className="text-[10px] font-extrabold uppercase tracking-[0.22em] text-muted">atau</span>
                <span className="h-px flex-1 bg-border" />
              </div>

              {!showCodeLogin ? (
                <button
                  onClick={() => setShowCodeLogin(true)}
                  className="flex h-[52px] w-full items-center justify-center gap-2 rounded-xl border border-border bg-card px-4 text-sm font-extrabold text-foreground transition hover:bg-card-high active:scale-[0.98]"
                  type="button"
                >
                  <Icon name="menu" className="h-4 w-4" />
                  Login dengan Code
                </button>
              ) : (
                <form
                  className="rounded-2xl border border-border bg-card p-4 shadow-soft"
                  onSubmit={(event) => {
                    event.preventDefault();
                    mockLogin();
                  }}
                >
                  <div className="mb-3 flex items-center justify-between">
                    <p className="text-sm font-extrabold text-foreground">Login dengan Code</p>
                    <button
                      type="button"
                      onClick={() => setShowCodeLogin(false)}
                      className="text-muted-foreground"
                      aria-label="Tutup form"
                    >
                      <Icon name="close" className="h-4 w-4" />
                    </button>
                  </div>
                  <div className="space-y-3">
                    <input className="h-12 w-full rounded-xl border border-border bg-surface px-4 text-sm font-semibold text-foreground placeholder:text-muted" placeholder="Business ID" defaultValue="QIOS-000001" />
                    <input className="h-12 w-full rounded-xl border border-border bg-surface px-4 text-sm font-semibold text-foreground placeholder:text-muted" placeholder="Operator Code" defaultValue="RZK-01" />
                    <input className="h-12 w-full rounded-xl border border-border bg-surface px-4 text-sm font-semibold text-foreground placeholder:text-muted" placeholder="Password" type="password" defaultValue="password" />
                    <button className="h-12 w-full rounded-xl bg-brand text-sm font-extrabold text-white transition active:scale-[0.98]" type="submit">
                      Masuk
                    </button>
                  </div>
                </form>
              )}

              {error && <p className="rounded-xl border border-danger/30 bg-danger/10 p-3 text-sm font-bold text-danger">{error}</p>}

              <button
                onClick={simulateInvalidQr}
                className="mx-auto block text-xs font-bold text-muted-foreground underline decoration-border underline-offset-4"
                type="button"
              >
                Simulasikan QR error
              </button>
            </div>
          </div>

          <div className="mt-6 flex justify-center">
            <QiosLogo />
          </div>
        </main>
      </OperatorShell>
    );
  }
