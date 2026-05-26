import { Product, Operator, Transaction, InsightCard } from '../types';

// ==========================================
// DATA SIMULASI UTAMA (Sesuai PRD & Kopi Senja)
// PT Skalar Solusi Digital - Developer: Alayavaro Rachmadia
// ==========================================

export const INITIAL_PRODUCTS: Product[] = [
  {
    id: 'prod-1',
    name: 'Kopi Susu Gula Aren',
    price: 22000,
    category: 'Kopi',
    description: 'Espresso blend khas Senja dengan susu segar UHT dan manisnya gula aren alami asli Sukabumi.',
    recipe: [
      { name: 'Susu UHT', qty: 150, unit: 'ml' },
      { name: 'Espresso', qty: 30, unit: 'ml' },
      { name: 'Gula Aren', qty: 20, unit: 'ml' }
    ],
    is_available: true,
    total_sold: 139
  },
  {
    id: 'prod-2',
    name: 'Americano',
    price: 18000,
    category: 'Kopi',
    description: 'Double shot espresso dari biji kopi arabika pilihan dengan air mineral berkualitas tinggi.',
    recipe: [
      { name: 'Espresso', qty: 60, unit: 'ml' }
    ],
    is_available: true,
    total_sold: 95
  },
  {
    id: 'prod-3',
    name: 'Matcha Latte',
    price: 25000,
    category: 'Non-Kopi',
    description: 'Bubuk matcha murni impor Jepang dengan susu UHT creamy dan sedikit sentuhan vanilla syrup.',
    recipe: [
      { name: 'Bubuk Matcha', qty: 15, unit: 'g' },
      { name: 'Susu UHT', qty: 180, unit: 'ml' }
    ],
    is_available: true,
    total_sold: 74
  },
  {
    id: 'prod-4',
    name: 'Croissant Butter',
    price: 28000,
    category: 'Makanan',
    description: 'Pastry mentega berlapis klasik, renyah di luar dan lembut berongga di bagian dalam.',
    recipe: [
      { name: 'Croissant Mentah', qty: 1, unit: 'pcs' }
    ],
    is_available: true,
    total_sold: 48
  },
  {
    id: 'prod-5',
    name: 'Cafe Latte',
    price: 24000,
    category: 'Kopi',
    description: 'Espresso shot seimbang dipadukan dengan foam susu halus tebal berstandar latte art.',
    recipe: [
      { name: 'Espresso', qty: 30, unit: 'ml' },
      { name: 'Susu UHT', qty: 160, unit: 'ml' }
    ],
    is_available: true,
    total_sold: 40
  }
];

export const INITIAL_OPERATORS: Operator[] = [
  {
    id: 'op-1',
    name: 'Siti Kasir',
    operator_code: 'kasir-siti',
    is_active: true,
    qr_token: 'QIOS-OP-QR-SITI-SEC-123',
    created_at: '2026-03-10T08:00:00Z'
  },
  {
    id: 'op-2',
    name: 'Budi Barista',
    operator_code: 'kasir-budi',
    is_active: true,
    qr_token: 'QIOS-OP-QR-BUDI-SEC-456',
    created_at: '2026-04-01T09:30:00Z'
  },
  {
    id: 'op-3',
    name: 'Andi Part Time',
    operator_code: 'kasir-andi',
    is_active: false,
    qr_token: 'QIOS-OP-QR-ANDI-SEC-789',
    created_at: '2026-05-15T14:15:00Z'
  }
];

export const INITIAL_TRANSACTIONS: Transaction[] = [
  {
    id: 'tx-1',
    order_id: 'QM-102526-20260525-001',
    total_amount: 62000,
    status: 'CONFIRMED',
    payment_method: 'CASH',
    created_by_operator_name: 'Siti Kasir',
    created_by_operator_id: 'op-1',
    confirmed_at: '2026-05-25T08:12:00Z',
    voided_at: null,
    void_reason: null,
    note: 'Bungkus rapi',
    created_at: '2026-05-25T08:10:00Z',
    items: [
      { product_id: 'prod-1', product_name: 'Kopi Susu Gula Aren', unit_price: 22000, quantity: 1, subtotal: 22000 },
      { product_id: 'prod-2', product_name: 'Americano', unit_price: 18000, quantity: 1, subtotal: 18000 },
      { product_id: 'prod-1', product_name: 'Kopi Susu Gula Aren', unit_price: 22000, quantity: 1, subtotal: 22000 }
    ]
  },
  {
    id: 'tx-2',
    order_id: 'QM-102526-20260525-002',
    total_amount: 50000,
    status: 'CONFIRMED',
    payment_method: 'QRIS_STATIC',
    created_by_operator_name: 'Siti Kasir',
    created_by_operator_id: 'op-1',
    confirmed_at: '2026-05-25T09:44:00Z',
    voided_at: null,
    void_reason: null,
    note: null,
    created_at: '2026-05-25T09:42:00Z',
    items: [
      { product_id: 'prod-3', product_name: 'Matcha Latte', unit_price: 25000, quantity: 2, subtotal: 50000 }
    ]
  },
  {
    id: 'tx-3',
    order_id: 'QM-102526-20260525-003',
    total_amount: 78000,
    status: 'CONFIRMED',
    payment_method: 'TRANSFER',
    created_by_operator_name: 'Budi Barista',
    created_by_operator_id: 'op-2',
    confirmed_at: '2026-05-25T11:21:00Z',
    voided_at: null,
    void_reason: null,
    note: 'Kirim bukti ke WA',
    created_at: '2026-05-25T11:18:00Z',
    items: [
      { product_id: 'prod-1', product_name: 'Kopi Susu Gula Aren', unit_price: 22000, quantity: 1, subtotal: 22000 },
      { product_id: 'prod-4', product_name: 'Croissant Butter', unit_price: 28000, quantity: 2, subtotal: 56000 }
    ]
  },
  {
    id: 'tx-4',
    order_id: 'QM-102526-20260525-004',
    total_amount: 40000,
    status: 'PENDING',
    payment_method: null,
    created_by_operator_name: 'Budi Barista',
    created_by_operator_id: 'op-2',
    confirmed_at: null,
    voided_at: null,
    void_reason: null,
    note: 'Customer masih cari kembalian',
    created_at: '2026-05-25T12:05:00Z',
    items: [
      { product_id: 'prod-2', product_name: 'Americano', unit_price: 18000, quantity: 1, subtotal: 18000 },
      { product_id: 'prod-1', product_name: 'Kopi Susu Gula Aren', unit_price: 22000, quantity: 1, subtotal: 22000 }
    ]
  },
  {
    id: 'tx-5',
    order_id: 'QM-102526-20260525-005',
    total_amount: 22000,
    status: 'VOIDED',
    payment_method: 'CASH',
    created_by_operator_name: 'Siti Kasir',
    created_by_operator_id: 'op-1',
    confirmed_at: null,
    voided_at: '2026-05-25T13:10:00Z',
    void_reason: 'Salah input menu oleh kasir baru',
    note: null,
    created_at: '2026-05-25T13:02:00Z',
    items: [
      { product_id: 'prod-1', product_name: 'Kopi Susu Gula Aren', unit_price: 22000, quantity: 1, subtotal: 22000 }
    ]
  },
  {
    id: 'tx-6',
    order_id: 'QM-102526-20260524-112',
    total_amount: 104000,
    status: 'CONFIRMED',
    payment_method: 'QRIS_STATIC',
    created_by_operator_name: 'Siti Kasir',
    created_by_operator_id: 'op-1',
    confirmed_at: '2026-05-24T15:30:00Z',
    voided_at: null,
    void_reason: null,
    note: null,
    created_at: '2026-05-24T15:28:00Z',
    items: [
      { product_id: 'prod-1', product_name: 'Kopi Susu Gula Aren', unit_price: 22000, quantity: 2, subtotal: 44000 },
      { product_id: 'prod-3', product_name: 'Matcha Latte', unit_price: 25000, quantity: 1, subtotal: 25000 },
      { product_id: 'prod-4', product_name: 'Croissant Butter', unit_price: 28000, quantity: 1, subtotal: 28000 }
    ]
  },
  {
    id: 'tx-7',
    order_id: 'QM-102526-20260523-098',
    total_amount: 48000,
    status: 'CONFIRMED',
    payment_method: 'CASH',
    created_by_operator_name: 'Andi Part Time',
    created_by_operator_id: 'op-3',
    confirmed_at: '2026-05-23T10:15:00Z',
    voided_at: null,
    void_reason: null,
    note: null,
    created_at: '2026-05-23T10:10:00Z',
    items: [
      { product_id: 'prod-5', product_name: 'Cafe Latte', unit_price: 24000, quantity: 2, subtotal: 48000 }
    ]
  }
];

export const INITIAL_INSIGHTS: InsightCard[] = [
  {
    id: 'in-1',
    type: 'trend',
    title: 'Selasa Dominasi Pekan Ini',
    narrative: 'Hari Selasa konsisten mencatatkan penjualan terkuat. Rata-rata revenue mencapai 38% di atas hari biasa dalam 30 hari terakhir.',
    updated_at: '2026-05-25T18:00:00Z',
    source_data_window: { start_date: '2026-04-25', end_date: '2026-05-25' }
  },
  {
    id: 'in-2',
    type: 'consumption',
    title: 'Lonjakan Konsumsi Susu UHT',
    narrative: 'Pemakaian susu UHT melonjak 22% minggu ini seiring larisnya varian Matcha Latte & Cafe Latte. Disarankan memesan restock bahan sebelum hari Kamis.',
    updated_at: '2026-05-25T19:15:00Z',
    source_data_window: { start_date: '2026-05-18', end_date: '2026-05-25' }
  },
  {
    id: 'in-3',
    type: 'opportunity',
    title: 'Sinergi Jam Makan Siang',
    narrative: 'Rentang jam 12.00 - 14.00 berkontribusi sebesar 41% dari total transaksi harian. Optimalkan promosi pairing menu Kopi + Croissant.',
    updated_at: '2026-05-25T14:30:00Z',
    source_data_window: { start_date: '2026-05-11', end_date: '2026-05-25' }
  },
  {
    id: 'in-4',
    type: 'warning',
    title: 'Produk Alami Churning',
    narrative: 'Menu Matcha Latte mencatat penurunan volume penjualan sebesar 15% dibanding minggu lalu. Evaluasi visibilitas menu atau tawarkan bundel promo.',
    updated_at: '2026-05-25T12:00:00Z',
    source_data_window: { start_date: '2026-05-18', end_date: '2026-05-25' }
  }
];