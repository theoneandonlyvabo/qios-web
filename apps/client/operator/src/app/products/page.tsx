"use client";

import { useState } from "react";
import { Icon } from "@/components/icons";
import { OperatorShell } from "@/components/operator-shell";
import { useOperator } from "@/components/providers/operator-provider";
import { formatCurrency } from "@/lib/format";

export default function ProductsPage() {
  const { products } = useOperator();
  const [query, setQuery] = useState("");

  const filtered = products.filter((product) => product.name.toLowerCase().includes(query.toLowerCase()));

  return (
    <OperatorShell>
      <header className="sticky top-0 z-40 border-b border-border bg-surface/90 px-4 py-4 backdrop-blur-xl safe-pt">
        <h1 className="text-xl font-extrabold text-foreground">Produk</h1>
        <p className="mt-1 text-sm font-medium text-muted-foreground">Katalog read-only untuk operator.</p>
      </header>

      <main className="space-y-4 px-4 pb-28 pt-4">
        <div className="relative">
          <Icon name="search" className="pointer-events-none absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-primary" />
          <input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Cari produk..."
            className="h-12 w-full rounded-xl border border-border bg-card pl-11 pr-4 text-sm font-semibold text-foreground placeholder:text-muted"
          />
        </div>

        <div className="rounded-2xl border border-primary/20 bg-primary/10 p-4 text-sm font-bold leading-relaxed text-primary">
          Untuk edit produk, harga, atau resep, hubungi tim Skalar. Mereka bakal bantu update biar data tetap rapi.
        </div>

        <section className="space-y-3">
          {filtered.map((product) => (
            <article key={product.id} className="flex items-center justify-between rounded-2xl border border-border bg-card p-3">
              <div className="flex items-center gap-3">
                <div className={`flex h-14 w-14 items-center justify-center rounded-xl bg-gradient-to-br ${product.gradient}`}>
                  <span className="text-2xl">{product.emoji}</span>
                </div>
                <div>
                  <p className="font-extrabold text-foreground">{product.name}</p>
                  <p className="text-xs font-bold text-muted-foreground">{product.category}</p>
                </div>
              </div>
              <div className="text-right">
                <p className="font-extrabold text-primary">{formatCurrency(product.price)}</p>
                <p className="text-[10px] font-extrabold text-success">AVAILABLE</p>
              </div>
            </article>
          ))}
        </section>
      </main>
    </OperatorShell>
  );
}
