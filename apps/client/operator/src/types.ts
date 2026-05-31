export type PaymentMethod = "CASH" | "QRIS_STATIC" | "TRANSFER";

export type OrderStatus = "PENDING" | "CONFIRMED" | "VOIDED";

export type ProductCategory = "Semua" | "Coffee" | "Non Coffee" | "Food" | "Dessert";

export type Product = {
  id: string;
  name: string;
  category: Exclude<ProductCategory, "Semua">;
  price: number;
  isAvailable: boolean;
  badge?: string;
  emoji: string;
  gradient: string;
};

export type CartItem = {
  productId: string;
  quantity: number;
};

export type OrderItemSnapshot = {
  productId: string;
  productName: string;
  category: string;
  quantity: number;
  unitPrice: number;
  subtotal: number;
  emoji: string;
  gradient: string;
};

export type OrderSummary = {
  itemCount: number;
  subtotal: number;
  tax: number;
  total: number;
  items: OrderItemSnapshot[];
};

export type Transaction = {
  id: string;
  orderId: string;
  status: OrderStatus;
  paymentMethod: PaymentMethod;
  total: number;
  subtotal: number;
  tax: number;
  items: OrderItemSnapshot[];
  createdAt: string;
  confirmedAt?: string;
  voidedAt?: string;
  createdBy: string;
  confirmedBy?: string;
  voidedBy?: string;
  voidReason?: string;
  businessName: string;
  note?: string;
};

export type OperatorSession = {
  businessId: string;
  businessName: string;
  operatorId: string;
  operatorCode: string;
  operatorName: string;
  plan: string;
  qrisStaticPayload: string;
  transferBankName: string;
  transferAccountNumber: string;
  transferAccountHolder: string;
};
