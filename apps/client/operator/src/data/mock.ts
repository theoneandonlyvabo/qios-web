import type { OperatorSession, Product, ProductCategory, Transaction } from "@/types";

export const operatorSession: OperatorSession = {
  businessId: "QIOS-000001",
  businessName: "Warung Kopi Senja",
  operatorId: "op-rizky",
  operatorCode: "RZK-01",
  operatorName: "Rizky",
  plan: "Skalar Max",
  qrisStaticPayload: "00020101021126680014ID.CO.QRIS.WWW01189360091100000000000215ID102001234567890303UMI51440014ID.CO.QRIS.WWW0215QIOS0000015204549953033605802ID5917Warung Kopi Senja6007Jakarta61051234062070703A016304B2E1",
  transferBankName: "BCA",
  transferAccountNumber: "1234567890",
  transferAccountHolder: "PT Skalar Solusi Digital"
};

export const categories: ProductCategory[] = ["Semua", "Coffee", "Non Coffee", "Food", "Dessert"];

export const products: Product[] = [
  {
    id: "kopi-susu-senja",
    name: "Kopi Susu Senja",
    category: "Coffee",
    price: 22000,
    isAvailable: true,
    badge: "Terlaris",
    emoji: "☕",
    gradient: "from-amber-950 via-stone-900 to-black"
  },
  {
    id: "americano",
    name: "Americano",
    category: "Coffee",
    price: 18000,
    isAvailable: true,
    emoji: "🧊",
    gradient: "from-slate-950 via-zinc-900 to-black"
  },
  {
    id: "croissant",
    name: "Croissant",
    category: "Food",
    price: 25000,
    isAvailable: true,
    emoji: "🥐",
    gradient: "from-orange-950 via-stone-900 to-black"
  },
  {
    id: "matcha-latte",
    name: "Matcha Latte",
    category: "Non Coffee",
    price: 28000,
    isAvailable: true,
    emoji: "🍵",
    gradient: "from-emerald-950 via-stone-900 to-black"
  },
  {
    id: "double-espresso",
    name: "Double Espresso",
    category: "Coffee",
    price: 22500,
    isAvailable: true,
    emoji: "☕",
    gradient: "from-neutral-950 via-stone-900 to-black"
  },
  {
    id: "avocado-toast-xl",
    name: "Avocado Toast XL",
    category: "Food",
    price: 45000,
    isAvailable: true,
    emoji: "🥑",
    gradient: "from-lime-950 via-stone-900 to-black"
  },
  {
    id: "glazed-donut",
    name: "Glazed Artisan Donut",
    category: "Dessert",
    price: 18000,
    isAvailable: true,
    emoji: "🍩",
    gradient: "from-rose-950 via-stone-900 to-black"
  },
  {
    id: "iced-tea",
    name: "Iced Lemon Tea",
    category: "Non Coffee",
    price: 16000,
    isAvailable: true,
    emoji: "🍋",
    gradient: "from-yellow-950 via-stone-900 to-black"
  }
];

const now = new Date("2026-05-05T14:32:00+07:00");

export const seedTransactions: Transaction[] = [
  {
    id: "tx-0001",
    orderId: "QM-0001",
    status: "CONFIRMED",
    paymentMethod: "QRIS_STATIC",
    total: 85000,
    subtotal: 85000,
    tax: 0,
    items: [
      {
        productId: "kopi-susu-senja",
        productName: "Kopi Susu Senja",
        category: "Coffee",
        quantity: 2,
        unitPrice: 22000,
        subtotal: 44000,
        emoji: "☕",
        gradient: "from-amber-950 via-stone-900 to-black"
      },
      {
        productId: "americano",
        productName: "Americano",
        category: "Coffee",
        quantity: 1,
        unitPrice: 18000,
        subtotal: 18000,
        emoji: "🧊",
        gradient: "from-slate-950 via-zinc-900 to-black"
      },
      {
        productId: "croissant",
        productName: "Croissant",
        category: "Food",
        quantity: 1,
        unitPrice: 23000,
        subtotal: 23000,
        emoji: "🥐",
        gradient: "from-orange-950 via-stone-900 to-black"
      }
    ],
    createdAt: now.toISOString(),
    confirmedAt: now.toISOString(),
    createdBy: "Rizky",
    confirmedBy: "Rizky",
    businessName: operatorSession.businessName
  },
  {
    id: "tx-0002",
    orderId: "QM-0002",
    status: "CONFIRMED",
    paymentMethod: "CASH",
    total: 120000,
    subtotal: 120000,
    tax: 0,
    items: [],
    createdAt: "2026-05-05T14:15:00+07:00",
    confirmedAt: "2026-05-05T14:15:00+07:00",
    createdBy: "Rizky",
    confirmedBy: "Rizky",
    businessName: operatorSession.businessName
  },
  {
    id: "tx-0003",
    orderId: "QM-0003",
    status: "CONFIRMED",
    paymentMethod: "QRIS_STATIC",
    total: 45000,
    subtotal: 45000,
    tax: 0,
    items: [],
    createdAt: "2026-05-05T13:45:00+07:00",
    confirmedAt: "2026-05-05T13:45:00+07:00",
    createdBy: "Rizky",
    confirmedBy: "Rizky",
    businessName: operatorSession.businessName
  },
  {
    id: "tx-0004",
    orderId: "QM-0004",
    status: "PENDING",
    paymentMethod: "TRANSFER",
    total: 350000,
    subtotal: 350000,
    tax: 0,
    items: [],
    createdAt: "2026-05-05T12:30:00+07:00",
    createdBy: "Rizky",
    businessName: operatorSession.businessName
  }
];
