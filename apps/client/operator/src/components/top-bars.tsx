"use client";

import { useRouter } from "next/navigation";
import { Icon } from "@/components/icons";
import { useOperator } from "@/components/providers/operator-provider";

export function BusinessHeader() {
  const { session } = useOperator();

  return (
    <header className="sticky top-0 z-40 flex h-16 items-center justify-between border-b border-border bg-surface/90 px-4 backdrop-blur-xl safe-pt">
      <div className="flex items-center gap-3">
        <div className="h-9 w-9 overflow-hidden rounded-full border border-border-warm bg-gradient-to-br from-brand via-orange-900 to-black shadow-glow">
          <div className="flex h-full w-full items-center justify-center text-sm font-black text-white">Q</div>
        </div>
        <div>
          <p className="text-lg font-extrabold leading-tight text-foreground">{session.businessName}</p>
          <p className="text-sm font-medium text-muted-foreground">{session.operatorName}</p>
        </div>
      </div>
      <div className="flex items-center gap-2">
        <span className="h-2.5 w-2.5 rounded-full bg-success shadow-[0_0_14px_rgba(34,197,94,0.8)]" />
      </div>
    </header>
  );
}

export function PageTopBar({ title, action = "history" }: { title: string; action?: "history" | "none" }) {
  const router = useRouter();

  return (
    <header className="sticky top-0 z-40 flex h-14 items-center justify-between border-b border-border bg-surface/90 px-4 backdrop-blur-xl safe-pt">
      <div className="flex items-center gap-3">
        <button
          aria-label="Kembali"
          onClick={() => router.back()}
          className="flex h-9 w-9 items-center justify-center rounded-full text-primary transition hover:bg-card-high active:scale-95"
          type="button"
        >
          <Icon name="arrow-left" className="h-5 w-5" />
        </button>
        <h1 className="text-lg font-extrabold text-primary">{title}</h1>
      </div>
      {action === "history" && <Icon name="history" className="h-5 w-5 text-primary" />}
    </header>
  );
}
