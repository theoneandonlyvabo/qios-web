"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode
} from "react";
import { operatorSession, products, seedTransactions } from "@/data/mock";
import { generateOrderId } from "@/lib/format";
import { readStorage, writeStorage } from "@/lib/storage";
import type {
  CartItem,
  OperatorSession,
  OrderSummary,
  PaymentMethod,
  Product,
  ProductCategory,
  Transaction
} from "@/types";

type OperatorContextValue = {
  session: OperatorSession;
  products: Product[];
  categories: ProductCategory[];
  cart: CartItem[];
  selectedCategory: ProductCategory;
  searchQuery: string;
  paymentMethod: PaymentMethod;
  note: string;
  transactions: Transaction[];
  summary: OrderSummary;
  setSelectedCategory: (category: ProductCategory) => void;
  setSearchQuery: (query: string) => void;
  setPaymentMethod: (method: PaymentMethod) => void;
  setNote: (note: string) => void;
  addProduct: (productId: string) => void;
  removeProduct: (productId: string) => void;
  clearCart: () => void;
  createTransaction: () => Transaction;
  voidTransaction: (transactionId: string, reason: string) => void;
  getLastTransaction: () => Transaction | null;
};

const OperatorContext = createContext<OperatorContextValue | null>(null);

const CART_KEY = "qios.operator.cart";
const TRANSACTION_KEY = "qios.operator.transactions";
const LAST_TRANSACTION_KEY = "qios.operator.lastTransaction";

function normalizeCart(cart: CartItem[]): CartItem[] {
  return cart.filter((item) => item.quantity > 0);
}

function buildSummary(cart: CartItem[]): OrderSummary {
  const items = cart
    .map((item) => {
      const product = products.find((entry) => entry.id === item.productId);
      if (!product) return null;

      return {
        productId: product.id,
        productName: product.name,
        category: product.category,
        quantity: item.quantity,
        unitPrice: product.price,
        subtotal: product.price * item.quantity,
        emoji: product.emoji,
        gradient: product.gradient
      };
    })
    .filter((item): item is NonNullable<typeof item> => Boolean(item));

  const subtotal = items.reduce((total, item) => total + item.subtotal, 0);
  const tax = 0;
  const total = subtotal + tax;
  const itemCount = items.reduce((total, item) => total + item.quantity, 0);

  return { itemCount, subtotal, tax, total, items };
}

export function OperatorProvider({ children }: { children: ReactNode }) {
  const [cart, setCart] = useState<CartItem[]>([]);
  const [transactions, setTransactions] = useState<Transaction[]>(seedTransactions);
  const [selectedCategory, setSelectedCategory] = useState<ProductCategory>("Semua");
  const [searchQuery, setSearchQuery] = useState("");
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>("QRIS_STATIC");
  const [note, setNote] = useState("");

  useEffect(() => {
    setCart(readStorage<CartItem[]>(CART_KEY, []));
    setTransactions(readStorage<Transaction[]>(TRANSACTION_KEY, seedTransactions));
  }, []);

  useEffect(() => {
    writeStorage(CART_KEY, cart);
  }, [cart]);

  useEffect(() => {
    writeStorage(TRANSACTION_KEY, transactions);
  }, [transactions]);

  const summary = useMemo(() => buildSummary(cart), [cart]);

  const addProduct = useCallback((productId: string) => {
    setCart((current) => {
      const exists = current.find((item) => item.productId === productId);
      if (!exists) return [...current, { productId, quantity: 1 }];
      return current.map((item) =>
        item.productId === productId ? { ...item, quantity: item.quantity + 1 } : item
      );
    });
  }, []);

  const removeProduct = useCallback((productId: string) => {
    setCart((current) =>
      normalizeCart(
        current.map((item) =>
          item.productId === productId ? { ...item, quantity: item.quantity - 1 } : item
        )
      )
    );
  }, []);

  const clearCart = useCallback(() => {
    setCart([]);
    setNote("");
    setPaymentMethod("QRIS_STATIC");
  }, []);

  const createTransaction = useCallback(() => {
    const snapshot = buildSummary(cart);
    const now = new Date();
    const transaction: Transaction = {
      id: globalThis.crypto?.randomUUID?.() ?? `tx-${Date.now()}`,
      orderId: generateOrderId(now),
      status: "CONFIRMED",
      paymentMethod,
      total: snapshot.total,
      subtotal: snapshot.subtotal,
      tax: snapshot.tax,
      items: snapshot.items,
      createdAt: now.toISOString(),
      confirmedAt: now.toISOString(),
      createdBy: operatorSession.operatorName,
      confirmedBy: operatorSession.operatorName,
      businessName: operatorSession.businessName,
      note: note.trim() || undefined
    };

    setTransactions((current) => [transaction, ...current]);
    writeStorage(LAST_TRANSACTION_KEY, transaction);
    clearCart();
    return transaction;
  }, [cart, clearCart, note, paymentMethod]);

  const voidTransaction = useCallback((transactionId: string, reason: string) => {
    const now = new Date().toISOString();
    setTransactions((current) =>
      current.map((transaction) =>
        transaction.id === transactionId && transaction.status === "PENDING"
          ? {
              ...transaction,
              status: "VOIDED",
              voidedAt: now,
              voidedBy: operatorSession.operatorName,
              voidReason: reason
            }
          : transaction
      )
    );
  }, []);

  const getLastTransaction = useCallback(() => {
    return readStorage<Transaction | null>(LAST_TRANSACTION_KEY, null);
  }, []);

  const value = useMemo<OperatorContextValue>(
    () => ({
      session: operatorSession,
      products,
      categories: ["Semua", "Coffee", "Non Coffee", "Food", "Dessert"],
      cart,
      selectedCategory,
      searchQuery,
      paymentMethod,
      note,
      transactions,
      summary,
      setSelectedCategory,
      setSearchQuery,
      setPaymentMethod,
      setNote,
      addProduct,
      removeProduct,
      clearCart,
      createTransaction,
      voidTransaction,
      getLastTransaction
    }),
    [
      addProduct,
      cart,
      clearCart,
      createTransaction,
      getLastTransaction,
      note,
      paymentMethod,
      removeProduct,
      searchQuery,
      selectedCategory,
      summary,
      transactions,
      voidTransaction
    ]
  );

  return <OperatorContext.Provider value={value}>{children}</OperatorContext.Provider>;
}

export function useOperator() {
  const context = useContext(OperatorContext);
  if (!context) {
    throw new Error("useOperator must be used inside OperatorProvider");
  }
  return context;
}
