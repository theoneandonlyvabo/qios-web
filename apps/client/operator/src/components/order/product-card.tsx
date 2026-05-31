"use client";

import { Icon } from "@/components/icons";
import { formatCurrency } from "@/lib/format";
import type { Product } from "@/types";

export function ProductCard({
  product,
  quantity,
  onAdd,
  onRemove
}: {
  product: Product;
  quantity: number;
  onAdd: () => void;
  onRemove: () => void;
}) {
  const selected = quantity > 0;

  return (
    <article
      className={`relative flex min-h-[222px] flex-col overflow-hidden rounded-2xl border bg-card transition-all ${
        selected
          ? "border-primary shadow-[0_0_20px_rgba(255,181,158,0.16)]"
          : "border-border hover:border-border-warm"
      }`}
    >
      <div className={`relative flex h-28 items-center justify-center bg-gradient-to-br ${product.gradient}`}>
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_20%,rgba(255,181,158,0.2),transparent_45%)]" />
        <span className="relative text-5xl drop-shadow-2xl">{product.emoji}</span>
        {product.badge && (
          <span className="absolute right-2 top-2 rounded-full bg-primary px-2 py-1 text-[10px] font-extrabold text-primary-foreground">
            {product.badge}
          </span>
        )}
      </div>

      <div className="flex flex-1 flex-col justify-between p-3">
        <div>
          <h3 className="truncate text-sm font-extrabold text-foreground">{product.name}</h3>
          <p className={`mt-1 text-sm font-extrabold ${selected ? "text-primary" : "text-muted-foreground"}`}>
            {formatCurrency(product.price)}
          </p>
        </div>

        {selected ? (
          <div className="mt-3 flex h-10 items-center justify-between rounded-xl bg-card-high p-1">
            <button
              aria-label={`Kurangi ${product.name}`}
              onClick={onRemove}
              className="flex h-8 w-8 items-center justify-center rounded-lg bg-surface text-foreground transition active:scale-95"
              type="button"
            >
              <Icon name="minus" className="h-4 w-4" />
            </button>
            <span className="text-sm font-extrabold text-foreground">{quantity}</span>
            <button
              aria-label={`Tambah ${product.name}`}
              onClick={onAdd}
              className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground transition active:scale-95"
              type="button"
            >
              <Icon name="plus" className="h-4 w-4" />
            </button>
          </div>
        ) : (
          <button
            onClick={onAdd}
            className="mt-3 h-10 rounded-xl bg-card-high text-sm font-extrabold text-foreground transition hover:bg-card-highest active:scale-[0.98]"
            type="button"
          >
            Tambah
          </button>
        )}
      </div>
    </article>
  );
}
