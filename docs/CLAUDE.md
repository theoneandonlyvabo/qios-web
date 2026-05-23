# CLAUDE.md — QIOS Project Bible

> Dokumen ini adalah sumber kebenaran tunggal untuk seluruh tim pengembang QIOS.
> Baca ini sebelum menyentuh satu baris kode pun. Jika ada konflik antara dokumen
> ini dengan kode yang ada, dokumen ini yang benar — kodenya yang perlu diperbaiki.

> Versi: 0.4 (post-pivot Mei 2026). Major changes dari versi sebelumnya:
> Xendit di-drop, full in-house transaction management. Recipe nested di product.
> Client di-split (dashboard / operator / admin). Plan + features merged ke business.
> Tidak ada self-serve registration; onboarding offline-first via admin Skalar.

---

## Apa itu QIOS

QIOS adalah **sistem manajemen keuangan dan operasional berbasis web untuk UMKM dan bisnis berkembang di Indonesia**.

QIOS bukan POS konvensional dan bukan sekadar tools pencatatan. QIOS menggantikan proses konvensional yang terfragmentasi dengan satu platform terpadu yang menghasilkan data akurat, real-time, dan actionable.

**Core value proposition:** Owner UMKM bisa melihat performa bisnis mereka secara real-time — produk terlaris, peak hours, tren revenue, pemakaian bahan baku, dan pola pembelian — tanpa harus paham akuntansi atau spreadsheet. Data yang masuk diolah menjadi insight bisnis yang actionable.

**Tiga lapisan solusi utama:**
- **Pencatatan transaksi in-house** — kasir mencatat transaksi langsung di QIOS dengan tiga metode pembayaran: QRIS statis (merchant pakai QR sendiri, kasir konfirmasi manual), tunai, dan transfer. Tidak ada dependency ke payment gateway eksternal di MVP.
- **Manajemen operasional terstruktur** — produk, harga, dan resep (komposisi bahan baku) dikelola lewat back office Skalar. Setiap penjualan otomatis ter-link ke konsumsi bahan baku, sehingga laporan periode bisa menampilkan pemakaian resource untuk restock dan insight.
- **Analytics dan business intelligence** — data diolah menjadi insight: tren pendapatan, analisis biaya, deteksi anomali, dan rekomendasi strategis berbasis pola transaksi.

**Yang QIOS lakukan:**
- Mencatat transaksi in-house dengan konfirmasi manual oleh operator
- Merekam item, payment method, operator, dan timestamp setiap transaksi
- Mengaitkan transaksi ke konsumsi bahan baku via resep produk
- Mengolah data transaksi menjadi visualisasi, laporan, dan insight
- Menyediakan PWA operator terpisah dari dashboard owner

**Yang QIOS tidak lakukan (MVP):**
- Manajemen stok real-time (consumption tracking ada, real-time deduction tidak)
- Sync dari marketplace eksternal (Tokopedia, GoBiz, dll)
- AI insight berbasis LLM (defer post-MVP; MVP menggunakan rule-based logic dengan schema AI-ready)
- Payment gateway integration (defer post-MVP; arsitektur menyiapkan extension point)
- Auto-forward laporan ke WA/Telegram (post-MVP add-on)
- Support iOS sebagai target utama (PWA Android first)
- Self-serve registration (offline-first 6 bulan pertama)

---

## Pengguna dan Role

**Owner** — pemilik bisnis, akses penuh ke dashboard, statistics, AI analytics, history transaksi, dan manajemen operator. Login via email/password atau Google OAuth. Tidak bisa edit produk/harga/resep — request lewat tim Skalar.

**Operator** — kasir/pegawai, akses terbatas ke PWA operator saja. Login via QR scan atau `operator_code` + password. Dibuat dan dikelola oleh owner dari dashboard. Tidak support Google OAuth.

**Admin (Skalar Staff)** — pengelola platform, akses ke admin panel untuk onboarding merchant, manage produk dan resep per merchant, assign plan, monitor transaksi cross-merchant, dan remove operator on request. Login via endpoint terpisah dengan JWT scope `admin`.

Satu user (owner) hanya bisa memiliki satu bisnis. Untuk bisnis berbeda, harus menggunakan akun berbeda.

---

## Arsitektur Sistem

### Overview

```
qios-web/
├── apps/
│   ├── client/
│   │   ├── admin/          # Next.js — admin panel Skalar staff
│   │   ├── dashboard/      # Next.js — interface owner, desktop-first
│   │   ├── operator/       # Next.js — PWA operator, mobile-first (Android)
│   │   └── Dockerfile      # Multistage build untuk tiga sub-app
│   └── server/
│       ├── api/            # Go + Echo — REST API backend
│       │   ├── cmd/
│       │   ├── config/
│       │   ├── core/       # Domain folders (auth, business, product, dll)
│       │   ├── pkg/        # Shared utilities (jwt, response, middleware, dll)
│       │   ├── .env
│       │   ├── .env.example
│       │   └── Dockerfile
│       ├── ai/             # AI service co-located, post-MVP
│       │   ├── cmd/
│       │   ├── config/
│       │   ├── core/       # AI logic core
│       │   ├── pipeline/   # Data processing pipeline
│       │   └── pkg/
│       └── bruno/          # Bruno API collection (committed)
├── docs/
│   ├── CLAUDE.md
│   └── qios-api.yml
└── infra/                  # Docker compose, deployment configs
```

Tiga client app independen yang share backend API. Dashboard, operator, dan admin di-deploy terpisah dengan domain/subdomain berbeda. AI service co-located di `apps/server/ai/` sebagai binary terpisah dari API — bisa di-deploy bareng atau pisah tergantung kebutuhan.

**Reasoning client split (vs monorepo 1 app dengan route group):**
- **Security surface jelas.** JWT scope berbeda per app. Token operator tidak bisa hit endpoint dashboard, dan sebaliknya. Tablet operator di counter tidak menjadi gateway ke data revenue owner.
- **Bundle size optimal.** Operator butuh ringan, fast load, offline-capable. Dashboard butuh chart library berat. Pisah = tidak saling membebani.
- **Deploy independent.** Fix bug operator tidak perlu redeploy dashboard, dan sebaliknya.
- **Mental model user split.** Owner di laptop, operator di tablet, admin Skalar di desktop kantor. Tiga konteks, tiga app.

### Client — Dashboard (Owner)

**Stack:** Next.js 16.2.6, TypeScript, Tailwind CSS v4, Recharts, npm

**Path:** `apps/client/dashboard/`

**Halaman:**

| Route | Halaman | Keterangan |
|-------|---------|------------|
| `/login` | Login | Email/password atau Google OAuth |
| `/dashboard` | Dashboard | Snapshot kondisi bisnis hari ini + 7 hari terakhir |
| `/statistics` | Statistics | Produk terlaris, breakdown transaksi, period comparison |
| `/analytics` | AI Analytics | Insight cards rule-based dari data transaksi |
| `/reports` | Reports | Laporan harian/bulanan, consumption, export PDF/CSV |
| `/history` | History | List semua transaksi raw dengan filter |
| `/operators` | Operators | CRUD akun operator milik bisnis |
| `/products` | Products | Read-only katalog produk + info "Edit via tim Skalar" |

**Auth flow:**
- Login: browser → `POST /api/auth/login` (Next.js route handler) → Go → Next.js set HttpOnly cookie server-side + return access token ke browser
- Access token disimpan di **memory** (runtime) + **localStorage** (untuk offline mode support — intentional, bukan bug)
- Refresh token disimpan di **HttpOnly cookie**, di-set server-side via Next.js route handler, tidak pernah terekspos ke JavaScript browser
- Logout: `POST /api/auth/logout` — hapus HttpOnly cookie server-side, clear in-memory token di client
- Middleware (`middleware.ts`) verifikasi HttpOnly cookie untuk semua route `/dashboard/*`, `/statistics/*`, dst
- Token rehydration: `hooks/useAuth.ts` baca access token dari localStorage ke memory saat app load

### Client — Operator (PWA)

**Stack:** Next.js 16.2.6, TypeScript, Tailwind CSS v4, npm. PWA service worker di-add saat polish phase.

**Path:** `apps/client/operator/`

**Halaman:**

| Route | Halaman | Keterangan |
|-------|---------|------------|
| `/` | Login Operator | QR scan (primary) atau `operator_code` + password (fallback) |
| `/order` | Order Aktif | Input order baru, pilih produk dari katalog |
| `/confirm` | Konfirmasi | Pilih payment method + slide-to-confirm (≥800ms) |
| `/history` | Riwayat Hari Ini | List transaksi yang dibuat operator ini hari ini |

**Auth flow:**
- Login via QR: scan → `POST /operator/auth/login/qr` → terima access token scope `operator`
- Login via credential: `POST /operator/auth/login` dengan `business_id` + `operator_code` + `password`
- Token disimpan di memory + localStorage untuk offline mode support
- Refresh token via HttpOnly cookie (sama pattern dashboard)
- JWT claim: `operator_id`, `business_id`, `scope: operator`

### Client — Admin (Skalar Staff Panel)

**Stack:** Next.js 16.2.6, TypeScript, Tailwind CSS v4, npm

**Path:** `apps/client/admin/`

**Halaman utama:**

| Route | Halaman | Keterangan |
|-------|---------|------------|
| `/login` | Admin Login | Email + password, scope `admin` |
| `/merchants` | Merchant List | List semua business dengan filter status/plan |
| `/merchants/new` | Onboard Merchant | Create business + owner account dalam satu form |
| `/merchants/{id}` | Merchant Detail | Edit plan/features/status, manage products, view transactions |
| `/merchants/{id}/products` | Products per Merchant | CRUD produk dan recipe untuk merchant terpilih |
| `/transactions` | Cross-Merchant Transactions | Read-only view dengan filter, force void capability |

**Catatan:** Admin tidak bisa create/edit operator merchant — hanya bisa REMOVE operator atas permintaan (lihat domain Admin di server). Operator tetap dikelola owner via dashboard.

### Security Headers

Semua response Next.js (di tiga client app) menyertakan header berikut:

`X-Frame-Options: DENY` · `X-Content-Type-Options: nosniff` · `Referrer-Policy: strict-origin-when-cross-origin` · `Strict-Transport-Security: max-age=31536000; includeSubDomains` · `Permissions-Policy: camera=(), microphone=(), geolocation=()`

PWA operator menambahkan `Permissions-Policy: camera=(self)` karena butuh akses kamera untuk QR scan.

---

### Server — API (Go + Echo)

**Stack:** Go 1.25, Echo v4, PostgreSQL 16, `lib/pq`, `golang-jwt`, `godotenv`

**Path:** `apps/server/api/`

**Struktur folder:**

```
apps/server/api/
├── cmd/                    # entry point — main.go
├── config/                 # config.go — load semua env vars ke struct Config
├── core/                   # business + view domains
│   ├── auth/               # owner login, Google OAuth, refresh, logout + operator login/QR
│   ├── user/               # profil owner + business info + operator CRUD (owner-side)
│   ├── product/            # read-only owner endpoint
│   ├── pos/                # orchestrator order kasir: cart, DRAFT→CONFIRMED flow, slide-to-confirm, sessions
│   ├── transaction/        # read-only log: owner history + filter
│   ├── dashboard/          # view: summary, trend, peak hours, top products
│   ├── analytics/          # view: deeper dive dengan custom timeframe + comparison
│   ├── report/             # view: daily/monthly sales, consumption, export
│   ├── insight/            # view: rule-based insight cards, AI-ready schema
│   └── admin/              # admin panel: onboard merchant, CRUD product+recipe,
│                           #              manage plan/features, audit, override void
└── pkg/                    # shared utilities
    ├── database/           # connect ke PostgreSQL, jalankan migrasi
    ├── jwt/                # issue dan verify JWT (scope owner/operator/admin)
    ├── middleware/         # auth middleware per scope, role guard
    ├── response/           # helper response JSON standar
    ├── qmid/               # generator QM-NNNNNN merchant ID
    └── encryption/         # AES-256 untuk data sensitif (placeholder)
```

> **Implementasi saat ini:** Domain `business/` di-merge ke dalam `user/` domain (owner satu bisnis). Domain `consumption/` di-handle sebagai background goroutine di dalam `pos/` domain saat CONFIRMED. Actual folder: `auth`, `user`, `product`, `pos`, `transaction`, `dashboard`, `analytics`, `report`, `insight`, `admin`.

**Konvensi domain:**

- Setiap domain di `core/` mengikuti pola `handler.go` → `service.go` → `repository.go`
- Handler tidak boleh menyentuh database langsung — semua lewat service dan repository
- Pattern canonical ada di `apps/server/api/AGENTS.md` — `core/user/` dijadikan referensi
- Domain bisnis (`auth`, `user`, `product`, `pos`, `transaction`, dst) punya tabel sendiri
- Domain view (`dashboard`, `analytics`, `report`, `insight`) tidak punya tabel — service inject repository dari beberapa domain bisnis untuk aggregation
- Dependency direction: view → bisnis. Bisnis tidak boleh depend on view.

### Server — AI Service

**Path:** `apps/server/ai/`

**Status:** Post-MVP. Selama MVP, insight rule-based dijalankan di `api/core/insight/` sebagai service biasa. AI service di-spin up saat sudah ada cukup data dan kebutuhan inferensi yang nyata.

**Struktur folder:**

```
apps/server/ai/
├── cmd/                    # binary entry point
├── config/                 # env loading
├── core/                   # AI logic core (model loading, inference)
├── pipeline/               # data processing pipeline (extract, transform, aggregate)
└── pkg/                    # shared utilities
```

**Komunikasi dengan API service:** AI service konsume `/reports` API sebagai input data, tidak query DB langsung. Output di-cache dan di-expose via internal endpoint yang dipanggil `core/insight/` di api server. Pemisahan ini bikin AI engine bisa di-iterate independent tanpa breaking API contract.

### Bruno Collection

**Path:** `apps/server/bruno/`

API collection di-commit ke repo untuk shared testing antar dev. Update saat ada endpoint baru atau request shape berubah.

### Database — PostgreSQL 16

Dijalankan via Docker. Migration dikelola secara manual menggunakan `migrate.go` berbasis file `.sql` bernomor urut di folder **`infra/database/migrations/`**. Migration bersifat **append-only** — file yang sudah ada tidak boleh diedit.

### Flow Transaksi (Operator)

```
1. Operator buka PWA, login via QR atau credential
2. Operator pilih produk dari katalog, set quantity
3. POST /transactions
   - Server generate order_id format: QM-NNNNNN-YYYYMMDD-{hex4}
   - INSERT transactions (status: PENDING, payment_method: null)
   - Snapshot product_name + unit_price ke transaction_items
   - Return transaction detail ke operator
4. Pembeli bayar (cash / scan QRIS static merchant / transfer)
5. Operator pilih payment method di PWA, slide-to-confirm (FE enforce ≥800ms hold)
6. POST /transactions/{id}/confirm body {payment_method}
   - Server set status: CONFIRMED, payment_method, confirmed_at, confirmed_by
   - Background job: parse product.recipe untuk tiap item, INSERT consumption_log
   - Update last_active_at di business
7. Selesai. Transaksi muncul di dashboard owner real-time (via polling atau SSE).
```

**Void flow:**

- Operator hanya bisa void transaksi PENDING yang dia sendiri buat (`created_by_operator_id` match)
- Owner bisa void status apapun di businessnya (PENDING atau CONFIRMED), tanpa grace period
- Admin Skalar bisa void cross-merchant via `/admin/transactions/{id}/void`
- `void_reason` wajib diisi untuk audit
- Status berubah ke `VOIDED`, transaksi tetap visible di history tapi tidak di-count di revenue/analytics

---

## API Contract

Kontrak lengkap ada di `docs/qios-api.yml` (OpenAPI 3.0.3, v0.4).

**Aturan:**

- Semua response menggunakan shape yang konsisten:
  ```json
  { "success": true/false, "data": ..., "error": "..." }
  ```
- Autentikasi menggunakan Bearer JWT di header `Authorization`
- Refresh token disimpan di httpOnly cookie, di-set server-side via Next.js route handler
- Access token disimpan di memory runtime + localStorage (offline mode support) — intentional, bukan bug
- Role-based access dijalankan di middleware — handler tidak perlu cek role sendiri
- JWT scope `owner`, `operator`, dan `admin` dipisah. Token satu scope tidak bisa hit endpoint scope lain.

**Endpoint per domain (ringkas):**

| Domain | Endpoint Utama | Keterangan |
|--------|----------------|------------|
| Auth (owner) | `/auth/login`, `/auth/google`, `/auth/refresh`, `/auth/logout` | Tidak ada public registration |
| User | `/users/me` (GET, PATCH) | Profil owner |
| Business | `/business` (GET, PATCH) | Owner hanya bisa edit name, location, qris_static_payload |
| Operator (owner-side) | `/business/operators` (CRUD owner) | Owner manage operator. Cap = `business.max_operators` |
| Operator Auth | `/operator/auth/login`, `/operator/auth/login/qr`, `/operator/me` | Public, untuk PWA operator |
| Product | `/products` (GET only) | Owner read-only. CRUD via admin domain. |
| Transaction | `/transactions` (POST, GET), `/transactions/{id}/confirm`, `/transactions/{id}/void` | Flow PENDING → CONFIRMED → VOIDED |
| Dashboard | `/dashboard/summary`, `/dashboard/transactions/trend`, `/dashboard/products/top`, `/dashboard/transactions/peak-hours` | Quick view aggregate |
| Analytics | `/analytics/overview` | Custom timeframe + period comparison |
| Report | `/reports/daily-sales`, `/reports/monthly-sales`, `/reports/consumption`, `/reports/export` | Laporan + export PDF/CSV |
| Insight | `/insight` | Rule-based MVP, AI-ready schema |
| Admin Auth | `/admin/auth/login`, `/admin/auth/refresh`, `/admin/auth/logout`, `/admin/me` | Surface terpisah |
| Admin Business | `/admin/businesses` (CRUD onboard), `/admin/businesses/{id}` | Onboard merchant + manage plan/features/status |
| Admin Product | `/admin/businesses/{id}/products`, `/admin/products/{id}` | Full CRUD product + recipe |
| Admin Operator | `/admin/businesses/{id}/operators/{id}` (DELETE only) | Remove operator on merchant request |
| Admin Transaction | `/admin/transactions`, `/admin/transactions/{id}/void` | Cross-merchant view + force void |

Setiap perubahan endpoint **harus diupdate di `docs/qios-api.yml` terlebih dahulu** sebelum implementasi.

---

## Database Schema

Target state v0.4 memerlukan 12 migration files baru. Migration reset dari schema lama (Xendit-era) ke schema v0.4 **belum dilakukan** — file di `infra/database/migrations/` saat ini masih berisi schema lama (includes `xendit_payments`, `webhook_events`, dll dari branch `old-dev-xendit`). Schema v0.4 di bawah adalah **target spec** untuk migration reset yang akan datang. Versi lama di-archive di branch `old-dev-xendit`.

| File | Tabel | Keterangan |
|------|-------|------------|
| 001 | `users` | Owner bisnis. Support email + Google OAuth. Tidak ada self-registration; user di-create oleh admin Skalar via `/admin/businesses`. |
| 002 | `refresh_tokens` | Multi-device session owner. Hash, bukan plain text. |
| 003 | `admin_users` | Akun staff Skalar. Role: `admin` atau `super_admin`. |
| 004 | `admin_refresh_tokens` | Session admin, dipisah dari owner agar JWT scope tidak bercampur. |
| 005 | `businesses` | Satu bisnis per owner. Menyimpan `qm_id` (format `QM-NNNNNN`), info bisnis, plan, features (JSONB array), max_operators, status (`ACTIVE`/`SUSPENDED`/`CHURNED`), lifecycle timestamps (`onboarded_at`, `suspended_at`, `churned_at`, `last_active_at`), `qris_static_payload`, dan `onboarded_by` (admin_id). |
| 006 | `operators` | Akun operator per bisnis. Field: `operator_code` (unik per business), `password_hash`, `qr_token` (statis, regenerate via owner), `is_active`. |
| 007 | `products` | Katalog produk. Field: `name`, `price`, `category`, `description`, `recipe` (JSONB array of ingredients), `is_available`, `total_sold`, soft delete via `deleted_at`. |
| 008 | `transactions` | Renamed dari `pos_orders`. Status: `PENDING`/`CONFIRMED`/`VOIDED`. Field: `order_id`, `total_amount`, `payment_method` (nullable saat PENDING), `created_by_operator_id`, `confirmed_by_type` + `confirmed_by_id`, `confirmed_at`, `voided_by_type` + `voided_by_id`, `voided_at`, `void_reason`, `note`. |
| 009 | `transaction_items` | Item per transaksi. Snapshot `product_name` + `unit_price` saat transaksi dibuat. Bukan FK ke produk untuk akurasi historis. |
| 010 | `consumption_log` | Auto-populated saat transaksi CONFIRMED. Field: `transaction_id`, `transaction_item_id`, `product_id`, `ingredient_name`, `qty`, `unit`, `recorded_at`. Source untuk `/reports/consumption` dan insight. |
| 011 | `admin_audit_logs` | Audit trail aksi admin. Field: `admin_id`, `action_type`, `resource_type`, `resource_id`, `before_state` (JSONB), `after_state` (JSONB), `reason`, `created_at`. |
| 012 | `business_status_log` (opsional, post-MVP) | Audit lengkap perubahan status business. Slot disiapkan untuk post-MVP. |

**Aturan penting:**

- `product_name` dan `unit_price` di `transaction_items` adalah **snapshot** — disimpan saat transaksi terjadi, bukan FK ke produk. Ini menjaga akurasi data historis jika produk diedit atau dihapus.
- `qm_id` di-generate di application layer (`pkg/qmid`) dengan format `QM-NNNNNN`. Generator wajib dipanggil di dalam tx yang sama dengan INSERT businesses, dan menggunakan `SELECT … FOR UPDATE` untuk row-level lock.
- `recipe` di `products` adalah JSONB array of objects: `[{ "name": "gula_aren", "qty": 20, "unit": "ml" }, ...]`. Validasi schema di service layer. Aggregation untuk consumption_log di-trigger saat transaksi CONFIRMED.
- Semua tabel menggunakan `UUID PRIMARY KEY DEFAULT gen_random_uuid()`. `qm_id` adalah field display terpisah dari primary key.
- Setiap tabel baru butuh index pada foreign key dan kolom yang sering di-query.
- Gunakan soft delete (`deleted_at`) untuk data yang punya histori transaksi (`products`, `operators`).
- `last_active_at` di `businesses` di-update via trigger atau service layer setiap transaksi CONFIRMED. Source untuk inactivity insight dan churn prediction.

### Onboarding Flow (Admin Skalar Onboards Merchant)

`POST /admin/businesses` mengeksekusi atomic transaction berikut:

1. `BEGIN` (isolation `SERIALIZABLE`)
2. Hash password owner dengan bcrypt
3. INSERT `users` (email, password_hash, name, phone, role='owner')
4. Generate `qm_id` via `pkg/qmid.Generate(tx)` — `SELECT qm_id FROM businesses ORDER BY qm_id DESC LIMIT 1 FOR UPDATE`
5. INSERT `businesses` (qm_id, name, location, category, plan, features, max_operators, status='ACTIVE', `onboarded_at` = NOW(), `onboarded_by` = admin.id, qris_static_payload, onboard_notes)
6. INSERT `admin_audit_logs` (action='onboard_merchant', resource='business', resource_id, after_state)
7. `COMMIT`

Tidak ada external network call. Onboarding pure DB transaction, gagal-nya local. Owner login pertama kali dengan password yang admin Skalar berikan tatap muka.

Lifecycle `status`: `ACTIVE` (default saat onboard) → `SUSPENDED` (admin set, suspended_at di-stamp) → `CHURNED` (admin set, churned_at di-stamp). Bisa reverse dari SUSPENDED → ACTIVE.

---

## User Requirements

### End Users (Business Owner)

#### Autentikasi & Akses

| ID | User Story | Requirement |
|----|------------|-------------|
| C-01 | Owner bisa masuk dengan aman dan cepat | Sistem mendukung login email/password dan Google OAuth. JWT digunakan untuk sesi autentikasi. Tidak ada self-registration di MVP. |
| C-02 | Sesi tetap aktif selama masih pakai aplikasi | Access token di-refresh otomatis selama sesi aktif. |
| C-03 | Lupa password ditangani manual | MVP: kontak tim Skalar untuk reset password (admin reset via `PATCH /admin/users/{id}`). Self-serve reset post-MVP. |

#### Dashboard & Visibilitas

| ID | User Story | Requirement |
|----|------------|-------------|
| C-04 | Lihat kondisi bisnis begitu buka aplikasi | Dashboard menampilkan ringkasan: total revenue (CONFIRMED), jumlah transaksi, AOV, transaksi pending dan voided. |
| C-05 | Tahu tren bisnis naik atau turun dibanding periode sebelumnya | Dashboard menampilkan perbandingan period-over-period dengan indikator visual. |
| C-06 | Lihat tren revenue 7 hari dan peak hours | Chart tren harian dan distribusi transaksi per jam tersedia di dashboard. |

#### Manajemen Produk

| ID | User Story | Requirement |
|----|------------|-------------|
| C-07 | Owner bisa lihat katalog produk | `GET /products` read-only untuk owner. Halaman `/products` dashboard tampilkan info "Untuk edit produk, hubungi tim Skalar." |
| C-08 | Produk yang dihapus tidak merusak data historis | Soft delete — produk tetap terbaca di transaksi lama meski sudah dihapus dari katalog. |
| C-09 | Filter produk berdasarkan kategori atau nama | Query parameter `category`, `q`, `is_available` tersedia di `GET /products`. |

#### Transaksi & History

| ID | User Story | Requirement |
|----|------------|-------------|
| C-10 | Lihat semua transaksi dengan filter fleksibel | Filter berdasarkan tanggal, status (PENDING/CONFIRMED/VOIDED), payment method, operator, pagination 20/page. |
| C-11 | Lihat detail satu transaksi termasuk item-itemnya | `GET /transactions/{id}` mengembalikan detail lengkap termasuk snapshot produk dan audit trail (confirmed_by, voided_by). |
| C-12 | Owner bisa void transaksi jika perlu | Owner bisa void status apapun di businessnya. `void_reason` wajib diisi. |

#### Statistics & Analytics

| ID | User Story | Requirement |
|----|------------|-------------|
| C-13 | Lihat produk terlaris dalam periode tertentu | `GET /dashboard/products/top` dengan filter period dan limit. |
| C-14 | Lihat tren transaksi harian dalam rentang waktu custom | `GET /dashboard/transactions/trend` dengan `start_date` dan `end_date`. |
| C-15 | Bandingkan performa dengan periode sebelumnya | `GET /analytics/overview` dengan `compare_with` query param. |

#### Reports & Consumption

| ID | User Story | Requirement |
|----|------------|-------------|
| C-16 | Owner bisa lihat laporan harian dan bulanan | `GET /reports/daily-sales`, `GET /reports/monthly-sales` dengan breakdown per payment method dan per produk. |
| C-17 | Owner tahu berapa banyak bahan baku yang terpakai per periode | `GET /reports/consumption` dengan filter `start_date`, `end_date`, opsional `ingredient`. |
| C-18 | Owner bisa download laporan ke PDF atau CSV | `POST /reports/export` body `{type, format, period}`. Return download URL temporary. |

#### AI Analytics

| ID | User Story | Requirement |
|----|------------|-------------|
| C-19 | Dapat insight otomatis dari data bisnis | `GET /insight` mengembalikan insight cards rule-based — MVP minimal 4-6 tipe insight termasuk consumption-based. |
| C-20 | Insight bisa di-expand untuk lihat data pendukung | Setiap insight card punya tombol "Lihat Data" yang expand ke breakdown relevan. |
| C-21 | Schema insight siap untuk AI engine post-MVP | Response include `model_version`, `confidence_score` (nullable di MVP). |

#### Operator Management

| ID | User Story | Requirement |
|----|------------|-------------|
| C-22 | Tambah dan kelola akun operator | Owner bisa CRUD operator via `/business/operators`. |
| C-23 | Jumlah operator dibatasi sesuai plan/add-on | Cap dari `business.max_operators`, default 3, admin update saat merchant beli add-on, -1 = unlimited. |
| C-24 | Operator login bisa pakai QR scan | Owner generate QR token dari dashboard, share ke device operator. `qr_token` statis, regenerate jika bocor. |

### Operator

| ID | User Story | Requirement |
|----|------------|-------------|
| K-01 | Operator login cepat dengan QR atau credential | `POST /operator/auth/login/qr` (primary), `POST /operator/auth/login` (fallback). |
| K-02 | Operator catat order baru | `POST /transactions` create status PENDING dengan items snapshot. |
| K-03 | Operator pilih payment method dan konfirmasi | Slide-to-confirm gesture ≥800ms di FE sebelum hit `POST /transactions/{id}/confirm`. |
| K-04 | Operator bisa batalin order PENDING yang dia bikin sendiri | `POST /transactions/{id}/void`, restricted ke `created_by_operator_id` match. |
| K-05 | Operator lihat riwayat transaksi hari ini | `GET /operator/transactions/today` (filter via JWT operator_id). |

### Administrator (Skalar Staff)

| ID | Story | Requirement |
|----|-------|-------------|
| A-01 | Admin onboard merchant baru tatap muka | `POST /admin/businesses` create business + owner account dalam satu transaction. |
| A-02 | Admin manage produk dan resep per merchant | Full CRUD via `/admin/businesses/{id}/products` dan `/admin/products/{id}`. |
| A-03 | Admin set plan dan features per merchant | `PATCH /admin/businesses/{id}` update plan (`normal`/`enterprise`/`white_label`/`max`), features array, max_operators. |
| A-04 | Admin remove operator merchant atas permintaan | `DELETE /admin/businesses/{id}/operators/{id}`, `reason` wajib. Admin tidak bisa create/edit operator. |
| A-05 | Admin lihat transaksi cross-merchant | `GET /admin/transactions` dengan filter. Read-only kecuali force void. |
| A-06 | Admin bisa force void transaksi | `POST /admin/transactions/{id}/void`, `void_reason` wajib, masuk audit log. |
| A-07 | Admin suspend atau churn merchant | `PATCH /admin/businesses/{id}` set `status` ke `SUSPENDED` atau `CHURNED`. Lifecycle timestamp auto-stamp. |
| A-08 | Semua aksi admin tercatat | Audit log dengan `before_state`, `after_state`, `reason`, timestamp, dan admin identity. |

---

## Detailed UI Requirements

### Dashboard App

#### Login Page

Pintu masuk tunggal ke dashboard owner. Tidak ada registrasi publik.

**UI Elements:**
- Logo QIOS
- Input: Email, Password
- Tombol: Login dengan Email, Login dengan Google
- State: Loading, Error (Credentials Salah, Business Suspended/Churned), Sukses Redirect
- Info copy: "Belum punya akun? Hubungi tim Skalar."

**Teknikal:**
- `POST /auth/login` atau `POST /auth/google`
- Redirect ke `/dashboard` setelah sukses
- Route guard: kalau sudah login, `/login` redirect ke `/dashboard`

#### Sidebar

Navigasi utama untuk owner.

**Menu Items:** Dashboard, Statistics, AI Analytics, Reports, History, Operators, Products

**UI Elements:**
- Logo QIOS
- Collapsible (Desktop), Drawer (Mobile)
- User Info (bottom): Nama Bisnis, Plan badge, Avatar, Logout
- Plan badge visual berbeda untuk plan `max` (premium UI treatment)

#### Dashboard

Snapshot kondisi bisnis. Tujuan: kasih owner gambaran kondisi bisnis hari ini dan 7 hari terakhir dalam hitungan detik.

**UI Elements:**
- Metric Cards: Total Revenue (delta % vs periode sebelumnya), Jumlah Transaksi CONFIRMED, Average Order Value, Transaksi PENDING saat ini, Transaksi VOIDED
- Chart: Revenue trend 7 hari terakhir (line chart), Peak hours (bar chart horizontal)
- List: Top 5 produk terlaris

**Teknikal:**
- Endpoints: `/dashboard/summary`, `/dashboard/transactions/trend`, `/dashboard/transactions/peak-hours`, `/dashboard/products/top?limit=5`
- Chart library: Recharts
- Timeframe filter: toggle "Hari Ini / 7 Hari" di header

#### Statistics

Deep dive performa produk dan tren transaksi.

**UI Elements:**
- Filter Bar: Date range picker dengan preset (7 hari, 30 hari, 3 bulan) + custom range
- Section 1 — Tren Transaksi: Line chart revenue over time + volume per hari
- Section 2 — Produk Terlaris: Tabel nama, jumlah terjual, kontribusi revenue (sortable)
- Section 3 — Perbandingan Periode: Toggle bandingkan dengan periode sebelumnya

**Teknikal:** `/analytics/overview`, `/dashboard/products/top`

#### AI Analytics

QIOS ngomong duluan berdasarkan data bisnis owner. Insight otomatis, bukan chatbot.

**Format: Insight Card.** Tiap card berisi:
- Icon kategori (trend, warning, opportunity, consumption)
- Judul singkat (maks 10 kata)
- Narasi 1-2 kalimat, bahasa natural, kontekstual
- Tombol "Lihat Data" yang expand ke mini chart/breakdown pendukung
- Timestamp: "Diperbarui X jam lalu"

**Contoh konten:**
- "Hari Selasa konsisten jadi hari terkuat bisnis lo. Rata-rata revenue 38% di atas hari lain dalam 30 hari terakhir."
- "Tiga produk tidak mencatat penjualan selama 14 hari terakhir."
- "Pemakaian susu UHT naik 22% bulan ini. Pertimbangkan untuk stock up sebelum minggu depan."
- "Waktu puncak transaksi 12.00–14.00 berkontribusi 41% dari total transaksi harian."

**Teknikal:**
- Endpoint: `/insight`
- MVP: rule-based logic di `core/insight/`, bukan LLM
- Schema sudah include `model_version` dan `confidence_score` untuk forward compatibility ke AI engine post-MVP
- State: loading skeleton per card, empty state ("Insight akan muncul setelah 7 hari transaksi")

#### Reports

Laporan structured untuk owner, exportable ke PDF atau CSV.

**Tab structure:**
- Daily Sales (default tanggal hari ini)
- Monthly Sales (default bulan berjalan)
- Consumption (default 30 hari terakhir)

**UI Elements per tab:**
- Filter periode (date picker per tab type)
- Summary card: total revenue, total transaksi (untuk sales), total qty per ingredient (untuk consumption)
- Tabel detail breakdown
- Tombol Export: pilih format PDF atau CSV

**Teknikal:** `/reports/daily-sales`, `/reports/monthly-sales`, `/reports/consumption`, `/reports/export`

#### History

List semua transaksi raw dengan filter.

**UI Elements:**
- Filter: Date range, status (PENDING/CONFIRMED/VOIDED), payment method, operator
- Tabel: order_id, tanggal, total, status, payment method, operator
- Pagination (20/page)
- Klik baris → detail transaksi (modal atau halaman baru) termasuk audit trail

**Teknikal:** `/transactions`, `/transactions/{id}`

#### Operators

CRUD akun operator milik bisnis.

**UI Elements:**
- List operator aktif: nama, operator_code, status (active/inactive), tanggal dibuat
- Tombol "Tambah Operator" → form (nama, operator_code, password)
- Tombol "QR Code" per operator → modal tampilkan QR untuk scan device operator, plus tombol "Regenerate QR"
- Tombol edit (rename, toggle active) dan hapus per operator
- Info slot: "X dari Y slot terpakai" dari `business.max_operators`
- Info copy untuk add-on operator: "Butuh operator lebih banyak? Hubungi tim Skalar untuk beli add-on."

**Teknikal:**
- `/business/operators` (GET, POST, PUT, DELETE)
- `/business/operators/{id}/regenerate-qr`

#### Products (Read-Only)

**UI Elements:**
- List produk: nama, kategori, harga, is_available toggle (visual indicator, tidak editable)
- Klik produk → detail modal termasuk recipe (read-only display)
- Banner di atas halaman: "Untuk edit produk, harga, atau resep, hubungi tim Skalar. Mereka bakal bantu update biar data lo tetap rapi."

**Teknikal:** `/products`, `/products/{id}`

### Operator App (PWA)

#### Login Operator

**UI Elements:**
- Logo QIOS Operator
- Primary action: tombol "Scan QR" (buka kamera)
- Secondary action: link "Login dengan Code" → form (business_id, operator_code, password)
- State: Loading, Error (Invalid QR, Operator Inactive, Business Suspended)

**Teknikal:** `/operator/auth/login/qr` atau `/operator/auth/login`

#### Order

**UI Elements:**
- Header: nama operator, nama business, logout
- Search bar produk
- Grid kategori
- List produk dengan tombol qty (+/-)
- Cart sticky di bawah: total amount, jumlah item, tombol "Lanjut ke Konfirmasi"

**Teknikal:** `/products` (GET dengan filter `is_available=true`), `/transactions` (POST saat lanjut ke konfirmasi)

#### Confirm

**UI Elements:**
- Summary order: list item + total
- Pilih payment method (radio): CASH, QRIS_STATIC, TRANSFER
- Jika QRIS_STATIC dipilih: tampilkan `business.qris_static_payload` sebagai QR code untuk pembeli scan
- Jika TRANSFER dipilih: tampilkan info rekening (di business settings, atau placeholder)
- Note field (opsional)
- **Slide-to-confirm button** (FE handle threshold ≥800ms hold)
- Setelah konfirmasi: transaksi CONFIRMED, kembali ke `/order` dengan toast sukses

**Teknikal:** `/transactions/{id}/confirm` dengan body `{payment_method, note}`

#### Riwayat Hari Ini

**UI Elements:**
- List transaksi yang operator ini buat hari ini
- Per row: waktu, order_id, total, status, payment method
- Tombol void untuk PENDING yang dia sendiri buat (opsional)

**Teknikal:** `/operator/transactions/today` (filter via JWT operator_id)

### Admin Panel App

#### Merchant Onboarding Form

**UI Elements:**
- Section Business: nama, lokasi, kategori, qris_static_payload
- Section Owner: nama, email, phone, password (admin set initial)
- Section Plan: pilih plan, set features (multi-select), set max_operators
- Section Notes: free-text untuk catatan deal khusus
- Tombol Submit → create everything in one transaction

**Teknikal:** `POST /admin/businesses`

#### Merchant Detail Page

**UI Elements:**
- Header: nama business, qm_id, status badge, plan badge
- Tab: Profile, Products, Operators, Transactions, Settings
- Profile: edit business info, owner info
- Products: CRUD produk + recipe (form structured ingredients[])
- Operators: read-only list, tombol remove dengan reason
- Transactions: cross-merchant view filtered to this business
- Settings: update plan, features, max_operators, status

#### Cross-Merchant Transactions

**UI Elements:**
- Filter: business, status, payment method, periode
- Tabel: business, order_id, tanggal, total, status, payment method
- Tombol force void per row (dengan modal reason)

---

## Environment Variables

Semua env vars dibaca via `config.Load()` di startup. Tidak ada `os.Getenv()` langsung di luar `config/config.go`.

| Variable | Keterangan | Default |
|----------|------------|---------|
| `APP_PORT` | Port server | `8080` |
| `DB_HOST` | Host PostgreSQL | `localhost` |
| `DB_PORT` | Port PostgreSQL | `5432` |
| `DB_USER` | Username DB | `postgres` |
| `DB_PASSWORD` | Password DB | — |
| `DB_NAME` | Nama database | `qios` |
| `JWT_SECRET` | Secret untuk sign JWT scope owner & operator | — |
| `JWT_ADMIN_SECRET` | Secret terpisah untuk sign JWT scope admin | — |
| `JWT_ACCESS_EXPIRY` | Durasi access token | `15m` |
| `JWT_REFRESH_EXPIRY` | Durasi refresh token | `720h` |
| `ENCRYPTION_KEY` | AES-256 key (64 hex chars / 32 bytes) untuk enkripsi data sensitif di DB | — |
| `REPORT_EXPORT_DIR` | Path penyimpanan sementara file PDF/CSV export | `/tmp/qios-reports` |
| `REPORT_EXPORT_TTL` | Durasi download URL valid | `1h` |
| `DASHBOARD_ORIGIN` | Allowed origin untuk dashboard app (CORS) | `http://localhost:3000` |
| `OPERATOR_ORIGIN` | Allowed origin untuk operator app (CORS) | `http://localhost:3001` |
| `ADMIN_ORIGIN` | Allowed origin untuk admin app (CORS) | `http://localhost:3002` |

**Wajib di-set saat startup:** `DB_PASSWORD`, `JWT_SECRET`, `JWT_ADMIN_SECRET`, `ENCRYPTION_KEY`. Startup gagal kalau kosong.

Buat file `.env` di `apps/server/api/` untuk local development. File ini tidak boleh di-commit — sudah ada di `.gitignore`.

---

## Setup Local Development

### Prasyarat

- Node.js v24.15.0
- Go 1.25
- Docker + Docker Compose

### Langkah

**1. Clone dan masuk ke repo:**
```bash
git clone https://github.com/theoneandonlyvabo/qios-web.git
cd qios-web
```

**2. Buat file `.env` untuk api server:**
```bash
cp apps/server/api/.env.example apps/server/api/.env
# Edit sesuai kebutuhan local
```

**3. Jalankan PostgreSQL via Docker:**
```bash
docker compose -f infra/docker-compose.yml up postgres -d
```

**4. Jalankan api server:**
```bash
cd apps/server/api
go run ./cmd/...
# Migration otomatis jalan saat startup
```

**5. Jalankan client app (jalanin yang lo butuh aja):**
```bash
# Dashboard owner
cd apps/client/dashboard && npm install && npm run dev   # http://localhost:3000

# Operator PWA
cd apps/client/operator && npm install && npm run dev    # http://localhost:3001

# Admin Skalar
cd apps/client/admin && npm install && npm run dev       # http://localhost:3002
```

Api server berjalan di `http://localhost:8080`.

**6. Seed initial data (admin Skalar account untuk testing):**
```bash
cd apps/server/api
go run ./cmd/seed
# Output: admin email + temporary password
```

**7. Bruno collection:**
```bash
# Buka Bruno desktop client, open collection di apps/server/bruno/
# Set environment local, mulai testing endpoint
```

---

## Sprint Roadmap (2 Bulan MVP)

> Legend: ✅ selesai · 🔄 in progress · ⏳ belum mulai

### Week 1 — Foundation

**Backend (api):**
1. ✅ Auth endpoints live: owner + operator + admin login, refresh, logout
2. ✅ Middleware JWT scope per role (RequireAuth, RequireAdmin)
3. ⏳ Finalisasi migration reset ke schema v0.4 (001-012 baru)
4. ⏳ Seed data: 1 admin Skalar account, 1 dummy business + owner + operator + 3 produk

**Frontend Dashboard (Dev 1):**
1. ⏳ Setup Next.js, folder structure, Tailwind config
2. ⏳ Design system: warna (warm dark/light dengan accent Ferrari red), font Plus Jakarta Sans, komponen atom
3. ⏳ Layout shell: sidebar + route guard
4. ⏳ Login page (static dulu)

**Frontend Operator (Dev 3):**
1. ⏳ Setup Next.js PWA-ready, mobile-first layout
2. ⏳ Login page (QR camera scan + credential fallback)

**Frontend Admin (Dev 2 part-time):**
1. ⏳ Setup Next.js, basic layout, login page

### Week 2 — Core Pages Static

**Backend:**
1. ✅ Business domain: GET/PATCH owner-side (di dalam `user/` domain)
2. ✅ Operator domain: full CRUD owner-side + login PWA operator
3. ✅ Product domain: GET owner-side, full CRUD admin-side (dengan recipe nested)
4. ✅ Transaction domain: POST (create PENDING), POST confirm, POST void
5. ✅ QRIS static payload: disimpan di businesses, dikembalikan saat confirm

**Frontend Dashboard (Dev 1):**
1. ⏳ Connect login ke API
2. ⏳ Dashboard page dengan mock data
3. ⏳ Operators page (CRUD)

**Frontend Operator (Dev 3):**
1. ⏳ Order page: list produk, cart logic
2. ⏳ Confirm page: pilih payment, slide-to-confirm gesture

**Frontend Admin (Dev 2):**
1. ⏳ Login + sidebar + merchant list page (mock data)

### Week 3 — Data Integration & Reports

**Backend:**
1. ✅ Dashboard endpoints: summary, trend, peak-hours, top products
2. ✅ Analytics endpoint: overview dengan period comparison
3. ✅ Report endpoints: daily-sales, monthly-sales, consumption
4. ✅ Consumption_log auto-populate saat transaksi CONFIRMED (di transaction service)
5. ✅ Admin onboarding endpoint live (POST /admin/businesses)
6. ✅ Admin CRUD product+recipe, manage plan/features/status, audit log
7. ✅ Insight endpoint live (rule-based MVP)

**Frontend Dashboard (Dev 1 + 2 cross):**
1. ⏳ Statistics page connect ke API
2. ⏳ History page connect ke API
3. ⏳ Reports page (3 tab + export trigger)

**Frontend Operator (Dev 3):**
1. ⏳ Connect ke real API
2. ⏳ Riwayat hari ini

**Frontend Admin (Dev 2):**
1. ⏳ Merchant onboarding form
2. ⏳ Merchant list page connect ke API

### Week 4 — AI Insight + Polish

**Backend:**
1. ✅ Insight endpoint live (rule-based, AI-ready schema)
2. ⏳ Report export endpoint (PDF + CSV generator)
3. ✅ Admin audit log integration di semua admin endpoint

**Frontend Dashboard (Dev 1):**
1. ⏳ AI Analytics page: render insight cards, empty state, loading
2. ⏳ Reports export connect ke backend (download URL flow)

**Frontend Operator (Dev 3):**
1. ⏳ Polish: error state, loading skeleton, offline indicator
2. ⏳ PWA manifest + service worker basic (caching shell only)

**Frontend Admin (Dev 2):**
1. ⏳ Merchant detail page: products CRUD + recipe form
2. ⏳ Cross-merchant transactions read-only view

### Week 5-6 — Integration Testing + Buffer

1. Full flow testing end-to-end untuk 3 surface
2. Cross-app integration (operator login muncul real-time di dashboard via polling)
3. Edge cases: empty data, error state, token expired, business SUSPENDED
4. Performance check: dashboard aggregation query, report generation
5. Staging deploy (Biznet GIO NEO Lite atau Vercel + Railway)

### Week 7-8 — QA + Demo Ready

1. Fix issues dari testing
2. Final QA round
3. Demo data prep
4. Documentation pass (CLAUDE.md, AGENTS.md, README per app)

---

## Version Control

### Branching

- `main` — production-ready, selalu stabil. Tidak ada yang push langsung ke sini.
- `dev` — integrasi semua feature sebelum naik ke main.
- `feature/<nama>` — satu branch per fitur atau task.

**Alur wajib:**
```
feature/<nama> → dev → main
```

### Naming Branch

```
feature/auth-owner-login
feature/operator-order-confirm
feature/dashboard-summary
feature/report-consumption
feature/admin-onboard-merchant
feature/transaction-void
fix/recipe-parse-edge-case
chore/update-dependencies
```

### Commit Message

Format: `<type>: <deskripsi singkat>`

```
feat: add transaction confirm endpoint with payment_method
feat: add consumption_log aggregation on transaction CONFIRMED
fix: enforce 800ms minimum on slide-to-confirm gesture
fix: prevent operator void on transactions not their own
chore: update go dependencies
docs: update qios-api.yml with admin domain endpoints
refactor: split dashboard service from analytics service
```

### Pull Request

- PR dari `feature/` selalu ke `dev`, bukan ke `main`
- Satu PR = satu fitur atau satu fix
- Deskripsi PR wajib menyebutkan endpoint atau komponen yang berubah
- Tidak ada self-merge — minimal satu reviewer

---

## Rules dan Konvensi

### Umum

- Bahasa kode dan komentar: **English**
- Bahasa commit message dan PR description: **English**
- Bahasa dokumentasi (CLAUDE.md, AGENTS.md, PRD): **Indonesia**

### Go (api server)

- Tidak ada `os.Getenv()` di luar `config/config.go`
- Tidak ada logic bisnis di handler — handler hanya terima request, panggil service, kembalikan response
- Error selalu di-wrap dengan konteks: `fmt.Errorf("auth: failed to find user: %w", err)`
- Semua response menggunakan helper dari `pkg/response/`
- Tidak ada raw SQL di luar layer repository
- Domain view (`dashboard`, `analytics`, `report`, `insight`) inject repository dari domain bisnis — tidak punya repository sendiri
- Dependency direction: view → bisnis. Pelanggaran direction = code review reject.

### TypeScript (client)

- Tidak ada `any` — gunakan type yang proper atau `unknown`
- Semua API call melalui satu HTTP client terpusat per app, bukan `fetch` langsung di komponen
- State management sesederhana mungkin — jangan tambah library state global sebelum benar-benar dibutuhkan
- Chart library: Recharts — jangan ganti tanpa diskusi tim
- Komponen UI yang sama di-share via `packages/ui` kalau dipakai di lebih dari satu app

### Database

- Migration bersifat append-only — tidak boleh edit file migration yang sudah ada
- Setiap tabel baru butuh index pada foreign key dan kolom yang sering di-query
- Gunakan soft delete (`deleted_at`) untuk data yang punya histori transaksi
- JSONB columns (misal `business.features`, `products.recipe`) divalidasi di service layer

### Security

- Password di-hash dengan bcrypt sebelum disimpan
- Refresh token disimpan sebagai hash, bukan plain text
- `ENCRYPTION_KEY` digunakan untuk enkripsi field sensitif di DB (kalau ada di masa depan)
- JWT scope owner/operator/admin di-validate di middleware. Cross-scope access = 403.
- Tidak ada secret atau credential yang di-commit ke repository
- Security headers dikonfigurasi di tiap `next.config.ts` per client app

---

## Yang Belum Final (Pending)

- **Halaman test:** Sebelum production deploy, hapus semua halaman mock data testing yang tidak di-guard middleware
- **Plan & features detail:** Daftar feature flag spesifik per plan masih placeholder. Board konfirmasi mapping plan ↔ features.
- **Payment gateway integration (post-MVP):** DOKU dipertimbangkan untuk QRIS dinamis, masih belum fix vendor-nya. Arsitektur in-house transaction sekarang menyiapkan extension point.
- **Inventory management real-time (post-MVP):** Consumption tracking sudah ada (aggregated), tapi stockout warning real-time belum di-design.
- **LLM integration untuk AI Analytics (post-MVP):** `apps/server/ai/` direncanakan sebagai service terpisah yang konsume `/reports` API. Schema insight sudah AI-ready (model_version, confidence_score nullable). Selama MVP, insight rule-based dijalankan di `api/core/insight/`.
- **Auto-forward report ke WA/Telegram (post-MVP add-on):** Vendor (Fonnte / WAHA / Telegram Bot API) belum diputuskan. Gated by features flag, endpoint belum ada di contract.
- **Self-serve registration (post-MVP):** Setelah 6 bulan offline-first onboarding atau ketika volume merchant mencapai threshold. Self-serve butuh KYC otomatis dan training material yang belum di-prep.
- **Business status history audit (post-MVP):** Tabel `business_status_log` untuk full audit perubahan status. Slot disiapkan, schema belum implement.
- **Owner sebagai operator:** Owner yang mau jadi operator = buat operator account untuk dirinya sendiri via dashboard. Tidak ada pseudo-operator otomatis di onboarding. Bisa diperhatikan UX-nya pasca MVP.
- **Password reset flow:** Self-serve reset belum ada. MVP: kontak Skalar untuk reset manual via admin endpoint.

---

## Dokumen Terkait

- `docs/qios-api.yml` — OpenAPI 3.0.3 contract lengkap v0.4
- `AGENTS.md` (root) — panduan umum untuk AI agents
- `apps/server/api/AGENTS.md` — panduan implementasi spesifik api server
- `apps/client/dashboard/AGENTS.md` — panduan implementasi dashboard (kalau dibuat)
- `apps/client/operator/AGENTS.md` — panduan implementasi operator PWA (kalau dibuat)
- `apps/client/admin/AGENTS.md` — panduan implementasi admin panel (kalau dibuat)
- `apps/server/bruno/` — Bruno API collection untuk testing
- PRD QIOS — dokumen product requirement lengkap (source of truth untuk product decisions)