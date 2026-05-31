"use client";

import { useRouter } from "next/navigation";
import { Icon } from "@/components/icons";
import { OperatorShell } from "@/components/operator-shell";
import { QiosLogo } from "@/components/qios-logo";
import { useOperator } from "@/components/providers/operator-provider";

export default function ProfilePage() {
  const router = useRouter();
  const { session } = useOperator();

  return (
    <OperatorShell>
      <header className="sticky top-0 z-40 border-b border-border bg-surface/90 px-4 py-4 backdrop-blur-xl safe-pt">
        <h1 className="text-xl font-extrabold text-foreground">Lainnya</h1>
        <p className="mt-1 text-sm font-medium text-muted-foreground">Profil operator dan sesi kasir.</p>
      </header>

      <main className="space-y-4 px-4 pb-28 pt-5">
        <section className="rounded-[28px] border border-border bg-card p-5 text-center shadow-soft">
          <div className="mx-auto mb-4 flex h-20 w-20 items-center justify-center rounded-full border border-border-warm bg-card-high text-primary">
            <Icon name="profile" className="h-9 w-9" />
          </div>
          <h2 className="text-xl font-black text-foreground">{session.operatorName}</h2>
          <p className="mt-1 text-sm font-bold text-primary">{session.operatorCode}</p>
          <p className="mt-2 text-sm font-medium text-muted-foreground">{session.businessName}</p>
        </section>

        <section className="rounded-2xl border border-border bg-card p-4">
          <Info label="Business ID" value={session.businessId} />
          <Info label="Plan" value={session.plan} />
          <Info label="Status" value="Online" />
          <Info label="App Version" value="0.1.0" last />
        </section>

        <button className="flex h-[52px] w-full items-center justify-center gap-2 rounded-xl border border-border bg-card font-extrabold text-foreground active:scale-[0.98]" type="button">
          <Icon name="refresh" className="h-5 w-5 text-primary" />
          Refresh Data
        </button>

        <button
          onClick={() => router.push("/")}
          className="flex h-[52px] w-full items-center justify-center gap-2 rounded-xl border border-danger/30 bg-danger/10 font-extrabold text-danger active:scale-[0.98]"
          type="button"
        >
          <Icon name="logout" className="h-5 w-5" />
          Logout
        </button>

        <div className="pt-8">
          <QiosLogo />
        </div>
      </main>
    </OperatorShell>
  );
}

function Info({ label, value, last = false }: { label: string; value: string; last?: boolean }) {
  return (
    <div className={`flex items-center justify-between gap-4 py-3 ${last ? "" : "border-b border-border"}`}>
      <span className="text-sm font-semibold text-muted-foreground">{label}</span>
      <span className="text-sm font-extrabold text-foreground">{value}</span>
    </div>
  );
}
