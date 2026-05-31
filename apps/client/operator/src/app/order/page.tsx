"use client";

import { BusinessHeader } from "@/components/top-bars";
import { Icon } from "@/components/icons";
import { OfflineBanner } from "@/components/offline-banner";
import { OperatorShell } from "@/components/operator-shell";
import { ProductCard } from "@/components/order/product-card";
import { StickyCart } from "@/components/order/sticky-cart";
import { useOperator } from "@/components/providers/operator-provider";
import { formatCurrency } from "@/lib/format";

export default function OrderPage() {
  const {
    products,
    categories,
    cart,
    selectedCategory,
    searchQuery,
    summary,
    setSelectedCategory,
    setSearchQuery,
    addProduct,
    removeProduct
  } = useOperator();

  const filteredProducts = products.filter((product) => {
    const matchCategory = selectedCategory === "Semua" || product.category === selectedCategory;
    const matchSearch = product.name.toLowerCase().includes(searchQuery.toLowerCase().trim());
    return product.isAvailable && matchCategory && matchSearch;
  });

  function quantityOf(productId: string) {
    return cart.find((item) => item.productId === productId)?.quantity ?? 0;
  }

  return (
    <OperatorShell>
      <BusinessHeader />
      <OfflineBanner />

      <main className="pb-36">
        <section className="px-4 py-6 text-center">
          <p className="mb-1 text-sm font-bold text-muted-foreground">Total Pesanan</p>
          <h1 className="text-[42px] font-black leading-tight tracking-[-0.04em] text-primary">
            {formatCurrency(summary.total)}
          </h1>
        </section>

        <section className="px-4">
          <div className="relative">
            <Icon name="search" className="pointer-events-none absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-primary" />
            <input
              value={searchQuery}
              onChange={(event) => setSearchQuery(event.target.value)}
              className="h-12 w-full rounded-xl border border-border bg-card pl-11 pr-4 text-sm font-semibold text-foreground placeholder:text-muted-foreground focus:border-primary"
              placeholder="Cari produk..."
              type="search"
            />
          </div>
        </section>

        <section className="no-scrollbar mt-4 overflow-x-auto px-4">
          <div className="flex min-w-max gap-2 pb-1">
            {categories.map((category) => {
              const active = selectedCategory === category;
              return (
                <button
                  key={category}
                  onClick={() => setSelectedCategory(category)}
                  className={`h-9 rounded-full px-4 text-xs font-extrabold transition active:scale-95 ${
                    active ? "bg-foreground text-surface" : "border border-border bg-card text-foreground"
                  }`}
                  type="button"
                >
                  {category}
                </button>
              );
            })}
          </div>
        </section>

        <section className="mt-5 grid grid-cols-2 gap-4 px-4">
          {filteredProducts.map((product) => (
            <ProductCard
              key={product.id}
              product={product}
              quantity={quantityOf(product.id)}
              onAdd={() => addProduct(product.id)}
              onRemove={() => removeProduct(product.id)}
            />
          ))}
        </section>

        {filteredProducts.length === 0 && (
          <section className="px-4 pt-16 text-center">
            <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-card-high text-primary">
              <Icon name="search" className="h-7 w-7" />
            </div>
            <h2 className="text-lg font-extrabold text-foreground">Produk tidak ditemukan</h2>
            <p className="mt-2 text-sm font-medium text-muted-foreground">Coba kata kunci lain atau pilih kategori berbeda.</p>
          </section>
        )}
      </main>

      <StickyCart itemCount={summary.itemCount} total={summary.total} />
    </OperatorShell>
  );
}
