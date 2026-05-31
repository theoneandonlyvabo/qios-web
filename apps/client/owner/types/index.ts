// ==========================================
// TYPES & INTERFACES (Sesuai API Contract v0.4)
// PT Skalar Solusi Digital - Developer: Alayavaro Rachmadia
// ==========================================

export interface RecipeIngredient {
  name: string;
  qty: number;
  unit: 'ml' | 'l' | 'g' | 'kg' | 'pcs';
}

export interface Product {
  id: string;
  name: string;
  price: number;
  category: string;
  description: string;
  recipe: RecipeIngredient[];
  is_available: boolean;
  total_sold: number;
}

export interface Operator {
  id: string;
  name: string;
  operator_code: string;
  is_active: boolean;
  qr_token: string;
  created_at: string;
}

export type TransactionStatus = 'PENDING' | 'CONFIRMED' | 'VOIDED';

export type PaymentMethod = 'CASH' | 'QRIS_STATIC' | 'TRANSFER';

export interface TransactionItem {
  product_id: string;
  product_name: string;
  unit_price: number;
  quantity: number;
  subtotal: number;
}

export interface Transaction {
  id: string;
  order_id: string;
  total_amount: number;
  status: TransactionStatus;
  payment_method: PaymentMethod | null;
  created_by_operator_name: string;
  created_by_operator_id: string | null;
  confirmed_at: string | null;
  voided_at: string | null;
  void_reason: string | null;
  note: string | null;
  created_at: string;
  items: TransactionItem[];
}

export interface InsightCard {
  id: string;
  type: 'trend' | 'warning' | 'opportunity' | 'consumption';
  title: string;
  narrative: string;
  updated_at: string;
  source_data_window: { start_date: string; end_date: string };
}