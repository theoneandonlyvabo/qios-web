# CLAUDE.md — QIOS Project Bible

> Dokumen ini adalah sumber kebenaran tunggal untuk seluruh tim pengembang QIOS.
> Baca ini sebelum menyentuh satu baris kode pun. Jika ada konflik antara dokumen
> ini dengan kode yang ada, dokumen ini yang benar — kodenya yang perlu diperbaiki.

---

## Apa itu QIOS

QIOS adalah **sistem manajemen keuangan dan operasional berbasis web untuk UMKM dan bisnis berkembang di Indonesia**.

QIOS bukan POS konvensional dan bukan sekadar tools pencatatan. QIOS menggantikan proses konvensional yang terfragmentasi dengan satu platform terpadu yang menghasilkan data akurat, real-time, dan actionable.

**Core value proposition:** Owner UMKM bisa melihat performa bisnis mereka secara real-time — produk terlaris, peak hours, tren revenue, dan pola pembelian — tanpa harus paham akuntansi atau spreadsheet. Data yang masuk diolah menjadi insight bisnis yang actionable.

**Tiga lapisan solusi utama:**
- **Otomasi pencatatan dan akuntansi** — transaksi dicatat, dikategorikan, dan dikelompokkan otomatis ke struktur akun sesuai standar
- **Manajemen order dan operasional** — menghubungkan alur order masuk, invoice, dan pembayaran dalam satu sistem via xenPlatform
- **Analytics dan business intelligence** — data diolah menjadi insight: tren pendapatan, analisis biaya, deteksi anomali, dan rekomendasi strategis

**Yang QIOS lakukan:**
- Menerima pembayaran via QRIS Xendit milik merchant
- Merekam setiap transaksi beserta item yang dibeli
- Mengolah data transaksi menjadi visualisasi dan insight
- Menyediakan interface kasir sederhana untuk operator di device Android

**Yang QIOS tidak lakukan (MVP):**
- Manajemen stok (TBD, dijadwalkan post-MVP)
- Sync dari marketplace eksternal (Tokopedia, GoBiz, dll)
- AI insight berbasis LLM (defer post-MVP — MVP menggunakan rule-based logic)
- Support iOS sebagai target utama

---

## Pengguna dan Role

**Owner** — pemilik bisnis, akses penuh ke dashboard, statistics, AI analytics, history transaksi, dan manajemen operator. Login via email/password atau Google OAuth.

**Operator** — kasir/pegawai, akses terbatas ke interface kasir saja. Login via email/password. Dibuat dan dikelola oleh owner. Tidak support Google OAuth.

**Administrator (Internal Skalar Solutions)** — pengelola platform, akses ke admin panel untuk monitoring user, transaksi, kesehatan sistem, dan manajemen subscription.

Satu user (owner) hanya bisa memiliki satu bisnis. Untuk bisnis berbeda, harus menggunakan akun berbeda.

---

## Arsitektur Sistem

### Overview

```
apps/
├── client/     # Next.js 16.2.6 — frontend monorepo (dashboard + kasir)
└── server/     # Go + Echo — REST API backend
```

Satu monorepo, dua aplikasi. Client dan server dideploy secara terpisah tapi berada dalam satu repository untuk kemudahan koordinasi.

### Client — Next.js 16.2.6

**Stack:** Next.js 16.2.6, TypeScript, Tailwind CSS v4, Recharts, npm

**Dua mode UI dalam satu codebase:**
- `(dashboard)` — interface owner, desktop-first, akses via `qios.id/dashboard`
- `(kasir)` — interface operator, mobile-first PWA, akses via `qios.id/kasir`

Pemisahan dilakukan via Next.js route groups. Layout, komponen, dan styling berbeda per mode. Logic bisnis dan API calls bisa dishare.

Interface kasir ditargetkan sebagai **PWA di Android**. iOS bukan prioritas MVP.

**Halaman dalam route group `(dashboard)`:**

| Route | Halaman | Keterangan |
|-------|---------|------------|
| `/login` | Login | Pintu masuk ke QIOS |
| `/dashboard` | Dashboard | Snapshot kondisi bisnis hari ini |
| `/statistics` | Statistics | Produk terlaris, breakdown transaksi, performa produk |
| `/analytics` | AI Analytics | Insight rule-based dari data transaksi |
| `/history` | History | List semua transaksi raw dengan filter |
| `/operators` | Operators | CRUD akun operator (kasir) milik bisnis |

**Halaman dalam route group `(kasir)`:**

| Route | Halaman | Keterangan |
|-------|---------|------------|
| `/kasir` | Interface Kasir | Input order, generate QR, cek status pembayaran |

**Auth Architecture (Next.js client layer):**
- Login: browser → `POST /api/auth/login` (Next.js route handler) → Go → Next.js route set HttpOnly cookie server-side + return access token ke browser
- Access token disimpan di **memory** (runtime) + **localStorage** untuk offline mode support — ini keputusan desain, bukan bug
- Refresh token disimpan di **HttpOnly cookie**, di-set server-side via Next.js route handler, tidak pernah terekspos ke JavaScript browser
- Logout: `POST /api/auth/logout` (Next.js route handler) — hapus HttpOnly cookie server-side, clear in-memory token di client
- Middleware (`middleware.ts`) verifikasi HttpOnly cookie untuk semua route `/dashboard/*`, `/profile/*`, `/settings/*`
- Token rehydration: `hooks/useAuth.ts` baca access token dari localStorage ke memory saat app load (untuk offline mode)

**Security Headers (`next.config.ts`):**
Semua response Next.js menyertakan header berikut:
`X-Frame-Options: DENY` · `X-Content-Type-Options: nosniff` · `Referrer-Policy: strict-origin-when-cross-origin` · `Strict-Transport-Security: max-age=31536000; includeSubDomains` · `Permissions-Policy: camera=(), microphone=(), geolocation=()`

**Admin Test Page (`/admin`):**
`app/(dashboard)/admin/page.tsx` — halaman mock data untuk testing semua UI state (data / loading / error / empty). **Tidak memerlukan auth — tidak ada di matcher middleware.** Harus dihapus sebelum production deploy.

---

### Server — Go + Echo

**Stack:** Go 1.26.2, Echo v4, PostgreSQL 16, `lib/pq`, `golang-jwt`, `godotenv`

**Struktur folder:**
```
apps/server/
├── cmd/                    # entry point — main.go
├── config/                 # config.go — load semua env vars ke struct Config
├── migrations/             # file .sql bernomor urut, dijalankan via migrate.go
├── platform/
│   ├── database/           # connect ke PostgreSQL, jalankan migrasi
│   ├── jwt/                # issue dan verify JWT
│   ├── middleware/         # auth middleware, role guard
│   └── response/           # helper response JSON standar
└── domain/
    ├── auth/               # register, login, Google OAuth, refresh, logout
    ├── dashboard/          # summary, tren transaksi, peak hours, top produk (placeholder)
    ├── operator/           # CRUD akun operator + login kasir (canonical reference)
    ├── payment/            # POS orders, transaksi, Xendit integration, webhook
    ├── product/            # katalog produk, soft delete
    └── user/               # profil user dan bisnis

Domain post-MVP yang belum di-implement: admin/, analytic/, statistic/.
```

Setiap domain mengikuti pola: `handler.go` → `service.go` → `repository.go`. Handler tidak boleh menyentuh database langsung — semua lewat service dan repository. Pattern canonical ada di `apps/server/AGENTS.md` — `domain/operator/` dijadikan referensi.

**Konsolidasi (2026-05-09):** `domain/transaction/` dan `domain/xendit/` digabung ke `domain/payment/` supaya semua payment concern (POS orders, Xendit integration, webhook) ada di satu tempat. Dashboard endpoint dipindah ke `domain/dashboard/` baru.

### Database — PostgreSQL 16

Dijalankan via Docker. Migration dikelola secara manual menggunakan `migrate.go` berbasis file `.sql` bernomor urut di folder `migrations/`.

Migration bersifat **append-only** — file yang sudah ada tidak boleh diedit. Perubahan schema dilakukan dengan menambah file migration baru.

### Flow Transaksi (Kasir)

1. Operator pilih produk di interface kasir
2. Server generate `order_id` unik (format: `QIOS-YYYYMMDD-xxxx`) dan simpan ke `pos_orders`
3. QR static Xendit milik merchant ditampilkan ke pembeli
4. Pembeli scan dan bayar — `order_id` dikirim sebagai payment reference ke Xendit
5. Xendit kirim webhook notifikasi ke server QIOS
6. Server cocokkan `order_id` dari webhook ke `pos_orders`, update status, increment `total_sold` di `products`

---

## API Contract

Kontrak lengkap ada di `docs/qios-api.yml` (OpenAPI 3.0.3).

**Aturan:**

- Semua response menggunakan shape yang konsisten:
  ```json
  { "success": true/false, "data": ..., "error": "..." }
  ```
- Autentikasi menggunakan Bearer JWT di header `Authorization`
- Refresh token disimpan di httpOnly cookie, di-set server-side via Next.js route handler
- Access token disimpan di memory runtime + localStorage (offline mode) — intentional, bukan bug
- Role-based access dijalankan di middleware — handler tidak perlu cek role sendiri
- Endpoint webhook Xendit (`POST /payment/xendit/webhook`) tidak menggunakan Bearer auth, tapi diverifikasi via Xendit signature key

**Endpoint per domain:**

| Domain | Endpoint | Method | Keterangan |
|--------|----------|--------|------------|
| Auth | `/auth/register` | POST | Registrasi owner + bisnis + sub-account Xendit dalam satu tx atomic. Body: `email`, `password`, `full_name`, `phone`, `business_name`, `address`, `city`, `country`. Response: `access_token`, `user_id`, `business_id`, `qm_id`, `xendit_status`. |
| Auth | `/auth/login` | POST | Login email/password |
| Auth | `/auth/google` | POST | Login Google OAuth |
| Auth | `/auth/refresh` | POST | Refresh access token |
| Auth | `/auth/logout` | POST | Logout, hapus cookie |
| User | `/users/me` | GET | Ambil profil + bisnis |
| User | `/users/me` | PATCH | Update profil |
| Business | `/business` | GET | Detail bisnis |
| Business | `/business` | PATCH | Update info bisnis |
| Operator | `/business/operators` | GET | List semua operator |
| Operator | `/business/operators` | POST | Tambah operator baru |
| Operator | `/business/operators/{id}` | DELETE | Hapus operator |
| Product | `/products` | GET | List produk (filter + search) |
| Product | `/products` | POST | Tambah produk baru |
| Product | `/products/{id}` | PATCH | Update produk |
| Product | `/products/{id}` | DELETE | Soft delete produk |
| Payment | `/transactions` | GET | List transaksi (filter + pagination) |
| Payment | `/transactions` | POST | Buat order baru dari kasir |
| Payment | `/transactions/{id}` | GET | Detail transaksi |
| Payment | `/transactions/{id}/complete` | POST | Selesaikan order cash (tidak via Xendit) |
| Payment | `/payment/xendit/connect` | POST | Hubungkan akun Xendit |
| Payment | `/payment/xendit/status` | GET | Cek status koneksi Xendit |
| Payment | `/payment/xendit/webhook` | POST | Terima notifikasi dari Xendit |
| Dashboard | `/dashboard/summary` | GET | Ringkasan performa bisnis |
| Dashboard | `/dashboard/transactions/trend` | GET | Tren transaksi harian |
| Dashboard | `/dashboard/transactions/peak-hours` | GET | Distribusi transaksi per jam |
| Statistic | `/dashboard/products/top` | GET | Produk terlaris |
| Analytic | `/insight` | GET | List insight rule-based |

Setiap perubahan endpoint **harus diupdate di `docs/qios-api.yml` terlebih dahulu** sebelum implementasi.

---

## Database Schema

14 migration files (+ penambahan ke depan), urutan wajib dipertahankan:

| File | Tabel | Keterangan |
|------|-------|------------|
| 001 | `users` | Owner bisnis, support email + Google OAuth |
| 002 | `refresh_tokens` | Multi-device session |
| 003 | `password_reset_tokens` | Reset password via email, expire 1 jam |
| 004 | `plans`, `subscriptions` | Tier langganan QIOS — seed data pending konfirmasi board (slot disiapkan, belum diisi) |
| 005 | `businesses` | Satu bisnis per owner. Menyimpan `qm_id` (format `QM-NNNNNN`), profil bisnis (name/phone/address/city/country), serta `xendit_account_id`, `xendit_api_key`, `xendit_secret_key`, dan `xendit_status` (`PENDING`/`REGISTERED`/`ACTIVE`/`SUSPENDED`). Credentials Xendit harus dienkripsi sebelum disimpan. |
| 006 | `operators` | Akun kasir per bisnis |
| 007 | `products` | Katalog produk, soft delete |
| 008 | `pos_orders` | Order dari kasir, linked ke Xendit via `order_id` |
| 009 | `pos_order_items` | Item per order, snapshot nama dan harga saat transaksi |
| 010 | `xendit_payments` | Record pembayaran Xendit. Menyimpan `xendit_account_id` (sub-account `for-user-id`), `xendit_invoice_id`, `xendit_charge_id`, `payment_method`, `amount`, `status`, dan `raw_payload` JSONB. Migration ini juga drop tabel lama `midtrans_payments` (peninggalan iterasi pre-Xendit). |
| 011 | `webhook_events` | Log semua notifikasi masuk dari Xendit |
| 012 | `admin_audit_logs` | Audit trail aksi admin |
| 013 | `operators` (alter) | Tambah `operator_code` untuk login kasir |
| 014 | `pos_orders` (alter) | Tambah `payment_method` (`CASH`/`QRIS`/`EWALLET`/`VIRTUAL_ACCOUNT`), `updated_at`, dan status `cancelled`. Approved 2026-05-09. |

**Aturan penting:**
- `product_name` dan `unit_price` di `pos_order_items` adalah snapshot — disimpan saat transaksi terjadi, bukan FK ke produk. Ini menjaga akurasi data historis jika produk diedit atau dihapus.
- `xendit_secret_key` di tabel `businesses` harus dienkripsi di level aplikasi sebelum disimpan.
- `qm_id` di-generate di application layer (`platform/qmid`) dengan format `QM-NNNNNN`. Generator wajib dipanggil di dalam tx yang sama dengan INSERT businesses, dan menggunakan `SELECT … FOR UPDATE` untuk row-level lock.
- Semua tabel menggunakan `UUID PRIMARY KEY DEFAULT gen_random_uuid()`.
- Setiap tabel baru butuh index pada foreign key dan kolom yang sering di-query.
- Gunakan soft delete (`deleted_at`) untuk data yang punya histori transaksi.
- Tidak ada data xendit_payments yang boleh hilang meski webhook terlambat masuk.

### Onboarding Flow (Register)

`POST /auth/register` mengeksekusi atomic transaction berikut:

1. `BEGIN` (isolation `SERIALIZABLE`)
2. INSERT `users`
3. Generate `qm_id` via `platform/qmid.Generate(tx)` — `SELECT qm_id FROM businesses ORDER BY qm_id DESC LIMIT 1 FOR UPDATE`
4. INSERT `businesses` dengan `xendit_status = 'PENDING'`
5. Call `payment.XenditService.CreateManagedAccount` → `POST {XENDIT_BASE_URL}/v2/accounts` (Basic auth, `type: "MANAGED"`, `public_profile.business_name`)
6. UPDATE `businesses` set `xendit_account_id`, `xendit_api_key`, `xendit_secret_key`, `xendit_status = 'REGISTERED'`
7. `COMMIT`

Kalau step 5 atau 6 gagal, seluruh tx di-rollback. Step 5 adalah external network call di tengah tx — orphaned Xendit sub-account dimungkinkan kalau commit DB gagal *setelah* call sukses; reconcile via job, bukan inline.

Lifecycle `xendit_status`: `PENDING` → `REGISTERED` (post-account creation, KYC belum selesai) → `ACTIVE` (webhook `account.activated`) → opsional `SUSPENDED`.

---

## User Requirements

### End Users (Business Owner)

#### Autentikasi & Akses

| ID | User Story | Requirement |
|----|------------|-------------|
| C-01 | Owner bisa daftar dan masuk dengan aman dan cepat | Sistem mendukung login email/password dan Google OAuth. JWT digunakan untuk sesi autentikasi |
| C-02 | Sesi tetap aktif selama masih pakai aplikasi | Access token di-refresh otomatis selama sesi aktif, expired kalau idle terlalu lama |
| C-03 | Bisa reset password kalau lupa | Flow reset password via email dengan token yang expire dalam 1 jam |

#### Dashboard & Visibilitas

| ID | User Story | Requirement |
|----|------------|-------------|
| C-04 | Lihat kondisi bisnis begitu buka aplikasi | Dashboard menampilkan ringkasan: total revenue, jumlah transaksi, AOV, dan transaksi gagal/pending dalam periode yang bisa dipilih |
| C-05 | Tahu tren bisnis naik atau turun dibanding periode sebelumnya | Dashboard menampilkan perbandingan period-over-period dengan indikator visual (naik/turun/stabil) |
| C-06 | Lihat tren revenue 7 hari dan peak hours | Chart tren harian dan distribusi transaksi per jam tersedia di dashboard |

#### Manajemen Produk

| ID | User Story | Requirement |
|----|------------|-------------|
| C-07 | Tambah dan kelola katalog produk | Form tambah produk: nama, harga, kategori, deskripsi. Edit dan hapus tersedia |
| C-08 | Produk yang dihapus tidak merusak data historis | Soft delete — produk tetap terbaca di transaksi lama meski sudah dihapus dari katalog |
| C-09 | Filter produk berdasarkan kategori atau nama | Query parameter `category` dan `q` tersedia di `GET /products` |

#### Transaksi & History

| ID | User Story | Requirement |
|----|------------|-------------|
| C-10 | Lihat semua transaksi dengan filter fleksibel | Filter berdasarkan tanggal, status (pending/paid/failed/expired), pagination 20/page |
| C-11 | Lihat detail satu transaksi termasuk item-itemnya | `GET /transactions/{id}` mengembalikan detail lengkap termasuk snapshot produk |

#### Statistics

| ID | User Story | Requirement |
|----|------------|-------------|
| C-12 | Lihat produk terlaris dalam periode tertentu | `GET /dashboard/products/top` dengan filter period dan limit |
| C-13 | Lihat tren transaksi harian dalam rentang waktu custom | `GET /dashboard/transactions/trend` dengan `start_date` dan `end_date` |

#### AI Analytics

| ID | User Story | Requirement |
|----|------------|-------------|
| C-14 | Dapat insight otomatis dari data bisnis | `GET /insight` mengembalikan insight cards rule-based — MVP minimal 3-5 tipe insight |
| C-15 | Insight bisa di-expand untuk lihat data pendukung | Setiap insight card punya tombol "Lihat Data" yang expand ke breakdown relevan |

#### Operator Management

| ID | User Story | Requirement |
|----|------------|-------------|
| C-16 | Tambah dan kelola akun kasir | Owner bisa CRUD operator via `/business/operators` |
| C-17 | Jumlah operator dibatasi sesuai plan | Slot operator dinamis dari plan, fallback 3, -1 = unlimited |

### Payment Gateway (Xendit)

| ID | Story | Requirement |
|----|-------|-------------|
| PG-01 | Sistem generate QR untuk setiap transaksi kasir | QIOS buat `order_id` unik, QR static Xendit merchant ditampilkan ke pembeli |
| PG-02 | Status pembayaran terupdate otomatis | Xendit kirim webhook ke QIOS, server update status transaksi accordingly |
| PG-03 | Sistem handle payment gagal atau expired dengan benar | Status diupdate ke `failed`/`expired`. Operator bisa buat order baru |
| PG-04 | Semua data transaksi payment tersimpan dan bisa di-audit | Setiap transaksi Xendit disimpan lengkap di `xendit_payments` |
| PG-05 | Sistem aman dari manipulasi webhook palsu | Validasi signature key Xendit wajib sebelum webhook diproses |

### Administrator (Internal Skalar Solutions)

| ID | Story | Requirement |
|----|-------|-------------|
| A-01 | Monitor jumlah user aktif dan pertumbuhan registrasi | Dashboard admin: total user, DAU/MAU, registrasi baru per periode |
| A-02 | Lihat status semua transaksi payment | Log transaksi Xendit across all users dengan filter |
| A-03 | Suspend atau nonaktifkan akun bermasalah | Fungsi suspend/unsuspend dengan audit log |
| A-04 | Monitor kesehatan server dan database | Integrasi monitoring: uptime, response time, error rate |
| A-05 | Manage subscription dan plan user | Admin bisa lihat, ubah, extend, atau terminate plan |
| A-06 | Semua aksi admin tercatat | Audit log dengan timestamp dan identity admin |

---

## Detailed UI Requirements

### Login Page

Pintu masuk tunggal ke seluruh ekosistem QIOS. Tidak ada registrasi publik di MVP, onboarding dilakukan manual atau via invite.

**UI Elements:**
- Logo QIOS
- Input: Email, Password
- Tombol: Login
- State: Loading, Error (Credentials Salah), Sukses Redirect
- Tidak ada "Lupa Password" pada MVP

**Teknikal:**
- `POST /auth/login`
- Redirect ke `/dashboard` setelah sukses
- Route guard: kalau sudah login, `/login` redirect ke `/dashboard`
- Next.js: `app/(dashboard)/login/page.tsx`

### Sidebar

Navigasi utama untuk owner yang akses dashboard.

**UI Elements:**
- Logo QIOS
- Menu Items: Dashboard, Statistics, AI Analytics, History, Operators
- Collapsible (Desktop), Drawer (Mobile)
- User Info (bottom): Nama Bisnis, Avatar, Logout

**Teknikal:**
- Komponen global di `app/(dashboard)/layout.tsx`
- Active state dari `usePathname()`
- Logout: clear token, redirect ke `/login`

### Dashboard

Snapshot kondisi bisnis. Tujuan: kasih owner gambaran kondisi bisnis hari ini dan 7 hari terakhir dalam hitungan detik.

**UI Elements:**
- Metric Cards: Total Revenue (delta % vs periode sebelumnya), Jumlah Transaksi, Average Order Value, Transaksi Gagal/Pending
- Chart: Revenue trend 7 hari terakhir (line chart), Peak hours (bar chart horizontal)
- List: Top 5 produk terlaris (nama + jumlah terjual)

**Teknikal:**
- Endpoints: `GET /dashboard/summary`, `GET /dashboard/transactions/trend`, `GET /dashboard/transactions/peak-hours`, `GET /dashboard/products/top?limit=5`
- Chart library: Recharts
- Timeframe filter: toggle "Hari Ini / 7 Hari" di header
- Next.js: `app/(dashboard)/dashboard/page.tsx`

### Statistics

Deep dive performa produk dan tren transaksi. Target: evaluasi mingguan/bulanan sebelum ambil keputusan.

**UI Elements:**
- Filter Bar: Date range picker dengan preset (7 hari, 30 hari, 3 bulan) + custom range
- Section 1 — Tren Transaksi: Line chart revenue over time + volume per hari
- Section 2 — Produk Terlaris: Tabel nama produk, jumlah terjual, kontribusi revenue (sortable)
- Section 3 — Perbandingan Periode: Toggle bandingkan dengan periode sebelumnya

**Teknikal:**
- Endpoints: `GET /dashboard/transactions/trend`, `GET /dashboard/products/top`
- Next.js: `app/(dashboard)/statistics/page.tsx`

### AI Analytics

QIOS ngomong duluan berdasarkan data bisnis owner. Insight otomatis, bukan chatbot.

**Format: Insight Card.** Tiap card berisi:
- Icon kategori (tren, peringatan, peluang)
- Judul singkat (maks 10 kata)
- Narasi 1-2 kalimat, bahasa natural, kontekstual
- Tombol "Lihat Data" yang expand ke mini chart/breakdown pendukung
- Timestamp: "Diperbarui X jam lalu"

**Contoh konten:**
- "Hari Selasa merupakan hari terkuat secara konsisten. Rata-rata revenue hari Selasa 38% di atas hari lain dalam 30 hari terakhir."
- "Tiga produk tidak mencatat penjualan selama 14 hari terakhir."
- "Waktu puncak transaksi adalah 12.00–14.00. Berkontribusi 41% dari total transaksi harian."

**Teknikal:**
- Endpoint: `GET /insight`
- MVP: rule-based logic di backend, bukan LLM. LLM bisa dicolok post-MVP ke endpoint yang sama
- Frontend hanya render response, tidak ada logika AI di client
- State: loading skeleton per card, empty state kalau data belum cukup ("Insight akan muncul setelah 7 hari transaksi")
- Next.js: `app/(dashboard)/analytics/page.tsx`

### History

List semua transaksi raw dengan filter. Reference point untuk owner kalau perlu trace transaksi spesifik.

**UI Elements:**
- Filter: Date range, status (pending/paid/failed/expired)
- Tabel: order_id, tanggal, total, status, operator
- Pagination (20/page)
- Klik baris → detail transaksi (modal atau halaman baru)

**Teknikal:**
- Endpoints: `GET /transactions`, `GET /transactions/{id}`
- Next.js: `app/(dashboard)/history/page.tsx`

### Operators

CRUD akun kasir milik bisnis. Owner kelola dari sini, operator tidak punya akses ke halaman ini.

**UI Elements:**
- List operator aktif: nama, email, tanggal dibuat
- Tombol "Tambah Operator" → form (nama, email, password)
- Tombol hapus per operator → konfirmasi dialog
- Info slot: "X dari Y slot terpakai" sesuai plan

**Teknikal:**
- Endpoints: `GET /business/operators`, `POST /business/operators`, `DELETE /business/operators/{id}`
- Next.js: `app/(dashboard)/operators/page.tsx`

### Kasir (Mobile PWA)

Interface untuk operator di lapangan. Mobile-first dari awal.

**Scope MVP:**
- Input order baru, pilih produk dari katalog
- Generate QR Xendit static dengan unique `order_id`
- Tampil status pembayaran
- Riwayat transaksi hari ini

**API yang dicolok:**
- `POST /transactions` — buat order baru
- `GET /transactions/{id}` — cek status
- `GET /products` — load katalog produk

**Catatan untuk Dev:** Build `/kasir` sebagai route biasa dalam Next.js. Layout mobile-first, tidak ada dependency yang block PWA conversion nanti. Jangan implement service worker dulu.

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
| `JWT_SECRET` | Secret untuk sign JWT | — |
| `JWT_ACCESS_EXPIRY` | Durasi access token | `15m` |
| `JWT_REFRESH_EXPIRY` | Durasi refresh token | `720h` |
| `XENDIT_SECRET_KEY` | Master secret key Xendit (xenPlatform). Dipakai untuk Basic auth saat membuat sub-account dan operasi platform-level. **Wajib di-set** — startup gagal kalau kosong. | — |
| `XENDIT_ENV` | `sandbox` atau `production` | `sandbox` |
| `XENDIT_BASE_URL` | Override base URL Xendit (untuk testing/staging) | `https://api.xendit.io` |

Buat file `.env` di `apps/server/` untuk local development. File ini tidak boleh di-commit — sudah ada di `.gitignore`.

---

## Setup Local Development

### Prasyarat

- Node.js v24.15.0
- Go 1.26.2
- Docker + Docker Compose

### Langkah

**1. Clone dan masuk ke repo:**
```bash
git clone https://github.com/theoneandonlyvabo/qios-web.git
cd qios-web
```

**2. Buat file `.env` untuk server:**
```bash
cp apps/server/.env.example apps/server/.env
# Edit sesuai kebutuhan local
```

**3. Jalankan PostgreSQL via Docker:**
```bash
docker compose up postgres -d
```

**4. Jalankan server:**
```bash
cd apps/server
go run ./cmd/...
# Migration otomatis jalan saat startup
```

**5. Jalankan client:**
```bash
cd apps/client
npm install
npm run dev
```

Client berjalan di `http://localhost:3000`, server di `http://localhost:8080`.

**Catatan npm audit:**
Satu moderate vulnerability tersisa (postcss, nested di dalam Next.js internals). Tidak bisa di-fix dari project ini — harus menunggu patch dari upstream Next.js. Bukan dependency langsung project. Pantau update Next.js untuk resolusi.

---

## Sprint Roadmap

### Week 1 — Foundation

**Backend:**
1. Finalisasi semua migration (pastikan 001-012 clean)
2. Auth endpoint live: `POST /auth/login`, JWT middleware
3. Seed data awal untuk testing FE

**Frontend:**
1. Setup monorepo Next.js, folder structure, Tailwind config
2. Design system: warna, font, komponen atom (Button, Input, Card, Badge)
3. Layout shell: sidebar + route group `(dashboard)` dan `(kasir)`
4. Login page (static dulu, belum connect API)

### Week 2 — Core Pages Static

**Backend:**
1. Transaction domain: `POST /transactions`, `GET /transactions`, `GET /transactions/{id}`
2. Product domain: `GET /products`, `POST /products`, `PATCH /products/{id}`, `DELETE /products/{id}`
3. Operator domain: `GET /business/operators`, `POST /business/operators`, `DELETE /business/operators/{id}`
4. Payment domain: Xendit connect + webhook skeleton

**Frontend Developer 1 — Login + Auth:**
1. Connect login ke `POST /auth/login`
2. Route guard, token handling, redirect logic

**Frontend Developer 2 — Dashboard:**
1. Dashboard page dengan mock data
2. Metric cards + line chart (Recharts)
3. Top produk list

**Frontend Developer 3 — Kasir:**
1. Layout kasir (mobile-first)
2. Form input order, static UI

### Week 3 — Data Integration

**Backend:**
1. Dashboard endpoints: summary, trend, peak-hours
2. Statistic endpoint: products/top
3. Analytic logic: rule-based, minimal 3-5 insight types

**Frontend Developer 1 — Statistics + History:**
1. Statistics page: date filter, charts, tabel produk
2. History page: tabel transaksi + filter

**Frontend Developer 2 — Dashboard Live:**
1. Swap mock data ke API real
2. Loading skeleton, error state

**Frontend Developer 3 — Kasir Live:**
1. Connect ke `POST /transactions` + load produk
2. Status payment polling

### Week 4 — AI Analytics + Polish

**Backend:**
1. Insight endpoint live (`GET /insight`)
2. Error handling, response consistency

**Frontend Developer 1 — AI Analytics:**
1. AI Analytics page: render insight cards
2. Empty state, loading skeleton

**Frontend Developer 2 & 3 — Cross-polish:**
1. Operators page: CRUD full
2. Responsive check semua halaman
3. Edge cases: empty data, error state, token expired

### Week 5-6 — Integration Testing + Buffer

1. Full flow testing end-to-end
2. FE + BE sync untuk edge cases
3. Performance check
4. Staging deploy (Vercel + Railway)
5. Fix issues dari testing
6. Final QA + demo-ready build

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
feature/auth-login
feature/kasir-product-list
feature/dashboard-summary
feature/statistics-breakdown
feature/analytics-insight-cards
feature/history-transactions
feature/operators-crud
fix/webhook-signature-validation
chore/update-dependencies
```

### Commit Message

Format: `<type>: <deskripsi singkat>`

```
feat: tambah endpoint POST /auth/login
feat: tambah transaction domain
fix: perbaiki race condition di webhook handler
fix: validasi signature key Xendit di webhook endpoint
chore: update go dependencies
docs: update qios-api.yml dengan domain product dan transaction
refactor: pisahkan auth service dari handler
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
- Bahasa dokumentasi (CLAUDE.md, AGENTS.md): **Indonesia**

### Go (Server)

- Tidak ada `os.Getenv()` di luar `config/config.go`
- Tidak ada logic bisnis di handler — handler hanya terima request, panggil service, kembalikan response
- Error selalu di-wrap dengan konteks: `fmt.Errorf("auth: failed to find user: %w", err)`
- Semua response menggunakan helper dari `platform/response/`
- Tidak ada raw SQL di luar layer repository

### TypeScript (Client)

- Tidak ada `any` — gunakan type yang proper atau `unknown`
- Semua API call melalui satu HTTP client terpusat, bukan `fetch` langsung di komponen
- State management sesederhana mungkin — jangan tambah library state global sebelum benar-benar dibutuhkan
- Chart library: Recharts — jangan ganti tanpa diskusi tim

### Database

- Migration bersifat append-only — tidak boleh edit file migration yang sudah ada
- Setiap tabel baru butuh index pada foreign key dan kolom yang sering di-query
- Gunakan soft delete (`deleted_at`) untuk data yang punya histori transaksi

### Security

- `xendit_secret_key` harus dienkripsi sebelum disimpan ke database
- Refresh token disimpan sebagai hash, bukan plain text
- Validasi signature Xendit wajib dijalankan sebelum memproses webhook (requirement PG-05)
- Tidak ada secret atau credential yang di-commit ke repository
- Security headers dikonfigurasi di `apps/client/next.config.ts` — berlaku untuk semua response: X-Frame-Options, X-Content-Type-Options, HSTS, Referrer-Policy, Permissions-Policy

---

## Yang Belum Final (Pending)

- **Hapus sebelum prod:** `app/(dashboard)/admin/page.tsx` — halaman test-only tanpa auth guard, akses via `/admin`
- Seed data `plans` dan `subscriptions` — tunggu konfirmasi pricing dari board
- Konfirmasi ke Xendit: apakah item detail per transaksi tersedia via API atau webhook
- Flutter vs PWA untuk jangka panjang — MVP tetap PWA Android
- Inventory management (C-18) — dijadwalkan post-MVP, schema belum final
- LLM integration untuk AI Analytics — dijadwalkan post-MVP
- Feature flag per `qm_id` — enforcement per-feature atau per-plan, granularitas, middleware vs service layer, edge case (expired/downgrade/grandfathering). Defer post-MVP. Seed data `plans` dan `subscriptions` (migration 004) tetap pending konfirmasi board sebelum ini bisa dilanjut.

---

## Dokumen Terkait

- `docs/qios-api.yml` — OpenAPI 3.0.3 contract lengkap
- `AGENTS.md` (root) — panduan umum untuk AI agents
- `apps/server/AGENTS.md` — panduan implementasi spesifik server
- PRD QIOS — dokumen product requirement lengkap (source of truth untuk product decisions)