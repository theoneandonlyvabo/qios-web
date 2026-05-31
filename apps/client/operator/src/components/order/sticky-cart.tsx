"use client";

import { useRouter } from "next/navigation";
import { Icon } from "@/components/icons";
import { formatCurrency } from "@/lib/format";

export function StickyCart({ itemCount, total }: { itemCount: number; total: number }) {
  const router = useRouter();
  if (itemCount === 0) return null;

  return (
    <div className="fixed bottom-16 left-1/2 z-40 w-full max-w-[430px] -translate-x-1/2 bg-gradient-to-t from-surface via-surface/90 to-transparent px-4 pb-4 pt-6">
      <div className="flex items-center justify-between rounded-2xl border border-border bg-card-high p-3 shadow-soft">
        <div>
          <p className="text-xs font-extrabold text-muted-foreground">{itemCount} item terpilih</p>
          <p className="text-lg font-extrabold text-primary">{formatCurrency(total)}</p>
        </div>
        <button
          onClick={() => router.push("/confirm")}
          className="flex h-12 items-center gap-2 rounded-xl bg-primary px-5 text-base font-extrabold text-primary-foreground transition active:scale-95"
          type="button"
        >
          Konfirmasi
          <Icon name="chevron-right" className="h-5 w-5" />
        </button>
      </div>
    </div>
  );
}
