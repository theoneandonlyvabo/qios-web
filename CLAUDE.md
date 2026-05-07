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
- **Manajemen order dan operasional** — menghubungkan alur order masuk, invoice, dan pembayaran dalam satu sistem via Midtrans Marketplace
- **Analytics dan business intelligence** — data diolah menjadi insight: tren pendapatan, analisis biaya, deteksi anomali, dan rekomendasi strategis

**Yang QIOS lakukan:**
- Menerima pembayaran via QR Midtrans milik merchant
- Merekam setiap transaksi beserta item yang dibeli
- Mengolah data transaksi menjadi visualisasi dan insight
- Membuat dan mengirim invoice dengan payment link terintegrasi
- Menyediakan interface kasir sederhana untuk operator di device Android

**Yang QIOS tidak lakukan (MVP):**
- Manajemen stok (TBD, dijadwalkan post-MVP)
- Sync dari marketplace eksternal (Tokopedia, GoBiz, dll)
- AI insight berbasis LLM (defer post-MVP — MVP menggunakan rule-based logic)
- Support iOS sebagai target utama

---

## Pengguna dan Role

**Owner** — pemilik bisnis, akses penuh ke dashboard, analytics, AI insight, order management, invoice, dan semua data bisnis. Login via email/password atau Google OAuth.

**Operator** — kasir/pegawai, akses terbatas ke interface kasir saja. Login via email/password. Dibuat dan dikelola oleh owner. Tidak support Google OAuth.

**Administrator (Internal Skalar Solutions)** — pengelola platform, akses ke admin panel untuk monitoring user, transaksi, kesehatan sistem, dan manajemen subscription.

Satu user (owner) hanya bisa memiliki satu bisnis. Untuk bisnis berbeda, harus menggunakan akun berbeda.

---

## Arsitektur Sistem

### Overview

```
apps/
├── client/     # Next.js 15 — frontend monorepo (dashboard + kasir)
└── server/     # Go + Echo — REST API backend
```

Satu monorepo, dua aplikasi. Client dan server dideploy secara terpisah tapi berada dalam satu repository untuk kemudahan koordinasi.

### Client — Next.js 15

**Stack:** Next.js 15, TypeScript, Tailwind CSS, Recharts, npm

**Dua mode UI dalam satu codebase:**
- `(dashboard)` — interface owner, desktop-first, akses via `qios.id/dashboard`
- `(kasir)` — interface operator, mobile-first PWA, akses via `qios.id/kasir`

Pemisahan dilakukan via Next.js route groups. Layout, komponen, dan styling berbeda per mode. Logic bisnis dan API calls bisa dishare.

Interface kasir ditargetkan sebagai **PWA di Android**. iOS bukan prioritas MVP.

**Halaman dalam route group `(dashboard)`:**

| Route | Halaman | Keterangan |
|-------|---------|------------|
| `/login` | Login | Pintu masuk ke QIOS |
| `/dashboard` | Dashboard | Snapshot kondisi bisnis |
| `/analytics` | Analytics | Deep dive performa bisnis |
| `/insight` | AI Insight | Insight rule-based dari data transaksi |
| `/orders` | Order Management | List semua order dan riwayat per pelanggan |
| `/settings` | Settings | Placeholder MVP |

**Halaman dalam route group `(kasir)`:**

| Route | Halaman | Keterangan |
|-------|---------|------------|
| `/kasir` | Interface Kasir | Input order, generate QR, cek status pembayaran |

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
    ├── auth/               # login, Google OAuth, refresh, logout
    ├── user/               # profil user dan bisnis
    ├── order/              # pos orders dan order items
    ├── invoice/            # invoice management
    ├── payment/            # koneksi Midtrans, webhook handler
    ├── analytics/          # summary dan breakdown transaksi
    ├── insight/            # rule-based business insight
    └── admin/              # admin panel, audit log, monitoring
```

Setiap domain mengikuti pola: `handler.go` → `service.go` → `repository.go`. Handler tidak boleh menyentuh database langsung — semua lewat service dan repository.

### Database — PostgreSQL 16

Dijalankan via Docker. Migration dikelola secara manual menggunakan `migrate.go` berbasis file `.sql` bernomor urut di folder `migrations/`.

Migration bersifat **append-only** — file yang sudah ada tidak boleh diedit. Perubahan schema dilakukan dengan menambah file migration baru.

### Flow Transaksi (Kasir)

1. Operator pilih produk di interface kasir
2. Server generate `order_id` unik (format: `QIOS-YYYYMMDD-xxxx`) dan simpan ke `pos_orders`
3. QR static Midtrans milik merchant ditampilkan ke pembeli
4. Pembeli scan dan bayar — `order_id` dikirim sebagai payment reference ke Midtrans
5. Midtrans kirim webhook notifikasi ke server QIOS
6. Server cocokkan `order_id` dari webhook ke `pos_orders`, update status, increment `total_sold` di `products`

### Flow Invoice

1. Owner buat invoice via form (nama pelanggan, item, jumlah, harga, due date)
2. Server simpan invoice dan panggil Midtrans API untuk generate payment link
3. Owner bagikan payment link ke pelanggan
4. Pelanggan bayar via Midtrans
5. Midtrans kirim webhook — server update status invoice ke `paid`
6. Notifikasi masuk ke owner bahwa pembayaran berhasil

---

## API Contract

Kontrak lengkap ada di `docs/qios-api.yaml` (OpenAPI 3.0.3).

**Aturan:**

- Semua response menggunakan shape yang konsisten:
  ```json
  { "success": true/false, "data": ..., "error": "..." }
  ```
- Autentikasi menggunakan Bearer JWT di header `Authorization`
- Refresh token disimpan di httpOnly cookie, bukan localStorage
- Role-based access dijalankan di middleware — handler tidak perlu cek role sendiri
- Endpoint webhook Midtrans (`POST /payment/midtrans/webhook`) tidak menggunakan Bearer auth, tapi diverifikasi via Midtrans signature key

**Endpoint penting per domain:**

| Domain | Endpoint | Keterangan |
|--------|----------|------------|
| Auth | `POST /auth/login` | Login email/password, Google OAuth |
| Auth | `POST /auth/refresh` | Refresh access token |
| Auth | `POST /auth/reset-password` | Reset password via email (token expire 1 jam) |
| Orders | `POST /orders` | Buat order baru |
| Orders | `GET /orders` | List semua order (filter by status, pelanggan) |
| Orders | `GET /orders/:id` | Detail order |
| Invoice | `POST /invoices` | Buat invoice baru |
| Invoice | `GET /invoices` | List invoice (filter status: pending/paid/overdue) |
| Invoice | `GET /invoices/:id` | Detail invoice |
| Payments | `POST /payments/initiate` | Generate payment link Midtrans |
| Payments | `GET /payments/:id/status` | Cek status pembayaran |
| Payments | `POST /payment/midtrans/webhook` | Terima notifikasi dari Midtrans |
| Analytics | `GET /analytics/summary` | Ringkasan metrik untuk dashboard |
| Analytics | `GET /analytics/transactions` | Breakdown transaksi dengan filter `?from=&to=` |
| Insight | `GET /insight` | List insight rule-based dari data transaksi |
| Transactions | `GET /transactions` | List semua transaksi dengan filter fleksibel |
| Transactions | `PUT /transactions/:id` | Edit transaksi |
| Transactions | `DELETE /transactions/:id` | Hapus transaksi (dengan audit trail) |

Setiap perubahan endpoint **harus diupdate di `docs/qios-api.yaml` terlebih dahulu** sebelum implementasi.

---

## Database Schema

12 migration files (+ penambahan ke depan), urutan wajib dipertahankan:

| File | Tabel | Keterangan |
|------|-------|------------|
| 001 | `users` | Owner bisnis, support email + Google OAuth |
| 002 | `refresh_tokens` | Multi-device session |
| 003 | `password_reset_tokens` | Reset password via email, expire 1 jam |
| 004 | `plans`, `subscriptions` | Tier langganan QIOS — seed data pending konfirmasi board |
| 005 | `businesses` | Satu bisnis per owner, menyimpan Midtrans server key (encrypted) |
| 006 | `operators` | Akun kasir per bisnis |
| 007 | `products` | Katalog produk, soft delete |
| 008 | `pos_orders` | Order dari kasir, linked ke Midtrans via `order_id` |
| 009 | `pos_order_items` | Item per order, snapshot nama dan harga saat transaksi |
| 010 | `midtrans_payments` | Record pembayaran Midtrans, menyimpan transaction ID, timestamp, status, dan nominal |
| 011 | `webhook_events` | Log semua notifikasi masuk dari Midtrans |
| 012 | `admin_audit_logs` | Audit trail aksi admin |

**Aturan penting:**
- `product_name` dan `unit_price` di `pos_order_items` adalah snapshot — disimpan saat transaksi terjadi, bukan FK ke produk. Ini menjaga akurasi data historis jika produk diedit atau dihapus.
- `midtrans_server_key` di tabel `businesses` harus dienkripsi di level aplikasi sebelum disimpan.
- Semua tabel menggunakan `UUID PRIMARY KEY DEFAULT gen_random_uuid()`.
- Setiap tabel baru butuh index pada foreign key dan kolom yang sering di-query.
- Gunakan soft delete (`deleted_at`) untuk data yang punya histori transaksi.
- Tidak ada data midtrans_payments yang boleh hilang meski webhook terlambat masuk.

---

## User Requirements

### End Users (Business Owner)

#### Autentikasi & Akses

| ID | User Story | Requirement |
|----|------------|-------------|
| C-01 | Owner bisa daftar dan masuk dengan aman dan cepat | Sistem mendukung registrasi email/password dan login Google OAuth. JWT digunakan untuk sesi autentikasi |
| C-02 | Sesi tetap aktif selama masih pakai aplikasi | Access token di-refresh otomatis selama sesi aktif, expired kalau idle terlalu lama |
| C-03 | Bisa reset password kalau lupa | Flow reset password via email dengan token yang expire dalam 1 jam |

#### Dashboard & Visibilitas

| ID | User Story | Requirement |
|----|------------|-------------|
| C-04 | Lihat kondisi bisnis begitu buka aplikasi | Dashboard menampilkan ringkasan: total pemasukan, pengeluaran, profit bersih, dan jumlah transaksi dalam periode yang bisa dipilih |
| C-05 | Tahu tren bisnis naik atau turun dibanding periode sebelumnya | Dashboard menampilkan perbandingan period-over-period dengan indikator visual (naik/turun/stabil) |
| C-06 | Lihat aktivitas transaksi terbaru tanpa buka halaman lain | Widget recent transactions di dashboard menampilkan 5-10 transaksi terakhir secara real-time |

#### Manajemen Transaksi

| ID | User Story | Requirement |
|----|------------|-------------|
| C-07 | Catat transaksi masuk dan keluar dengan cepat | Form input transaksi minimal: nominal, kategori, tanggal, keterangan opsional. Selesai dalam < 30 detik |
| C-08 | Transaksi otomatis terkategorisasi | Sistem menyediakan kategori default yang bisa dikustomisasi. AI pattern recognition untuk auto-kategorisasi berdasarkan histori |
| C-09 | Bisa edit atau hapus transaksi yang salah input | Setiap transaksi bisa diedit atau dihapus dengan audit trail yang tercatat |
| C-10 | Lihat semua transaksi dengan filter yang fleksibel | Filter berdasarkan tanggal, kategori, nominal range, dan status. Hasil bisa diexport |

#### Order & Invoice

| ID | User Story | Requirement |
|----|------------|-------------|
| C-11 | Buat invoice untuk pelanggan langsung dari QIOS | Form pembuatan invoice dengan field: nama pelanggan, item, jumlah, harga, dan due date |
| C-12 | Invoice bisa dibayar langsung via payment gateway | Setiap invoice bisa di-generate payment link yang terhubung ke Midtrans |
| C-13 | Tahu status pembayaran invoice secara real-time | Status invoice terupdate otomatis: pending, paid, overdue. Notifikasi masuk kalau ada pembayaran berhasil |
| C-14 | Lihat semua order dan riwayat transaksi pelanggan | Halaman order management menampilkan list semua order dengan status dan histori per pelanggan |

#### Inventory (TBD — Post-MVP)

| ID | User Story | Requirement |
|----|------------|-------------|
| C-18 | Stok terhubung ke transaksi penjualan | Setiap transaksi penjualan yang tercatat otomatis mengurangi stok item yang relevan |

### Payment Gateway (Midtrans)

Semua interaksi antara QIOS dan Midtrans harus reliable, aman, dan tidak butuh intervensi manual dari user maupun admin.

| ID | Story | Requirement |
|----|-------|-------------|
| PG-01 | Sistem harus bisa generate payment link untuk setiap invoice | QIOS memanggil Midtrans API untuk membuat transaksi baru dan mengembalikan payment URL ke user |
| PG-02 | Status pembayaran harus terupdate otomatis tanpa user harus refresh manual | Midtrans mengirim webhook notification ke QIOS setiap ada perubahan status transaksi. QIOS memproses dan mengupdate status invoice accordingly |
| PG-03 | Sistem harus handle payment yang gagal atau expired dengan benar | Jika transaksi expired atau gagal, status invoice diupdate ke `failed`/`expired`. User bisa generate ulang payment link |
| PG-04 | Semua data transaksi payment harus tersimpan dan bisa di-audit | Setiap transaksi Midtrans disimpan dengan transaction ID, timestamp, status, dan nominal. Tidak ada data yang hilang meski webhook terlambat masuk |
| PG-05 | Sistem harus aman dari manipulasi webhook palsu | Validasi signature key pada setiap webhook request dari Midtrans sebelum diproses |

### Administrator (Internal Skalar Solutions)

Pengelola platform butuh visibilitas dan kontrol penuh atas kesehatan sistem tanpa harus masuk langsung ke database atau server.

| ID | Story | Requirement |
|----|-------|-------------|
| A-01 | Admin harus bisa monitor jumlah user aktif dan pertumbuhan registrasi | Dashboard admin menampilkan metrik: total user, user aktif (DAU/MAU), registrasi baru per periode |
| A-02 | Admin harus bisa lihat status semua transaksi payment yang berjalan | Admin panel menampilkan log transaksi Midtrans across all users dengan filter status dan rentang waktu |
| A-03 | Admin harus bisa suspend atau nonaktifkan akun user yang bermasalah | Fungsi suspend/unsuspend user tersedia di admin panel dengan audit log siapa yang melakukan dan kapan |
| A-04 | Admin harus bisa monitor kesehatan server dan database secara real-time | Integrasi monitoring (uptime, response time, error rate) yang bisa diakses dari admin panel atau tools eksternal |
| A-05 | Admin harus bisa manage subscription dan plan user | Admin bisa lihat plan aktif tiap user, ubah plan, dan extend atau terminate subscription secara manual |
| A-06 | Semua aksi admin harus tercatat dalam audit log | Setiap aksi yang dilakukan admin (suspend, edit plan, akses data) tercatat dengan timestamp dan identity admin yang melakukan |

---

## Detailed UI Requirements

### Login Page

Pintu masuk tunggal ke seluruh ekosistem QIOS. Satu akun = satu bisnis. Tidak ada registrasi publik di MVP, onboarding dilakukan manual atau via invite.

**UI Elements:**
- Logo QIOS
- Input: Email, Password
- Tombol: Login
- State: Loading, Error (Credentials Salah), Sukses Redirect
- Tidak ada "Lupa Password" pada MVP

**Teknikal:**
- `POST /auth/login` dari API contract
- Response: JWT token, simpan di httpOnly cookie atau localStorage (ikut keputusan API contract)
- Redirect ke `/dashboard` setelah sukses
- Route guard: kalau sudah login, `/login` redirect ke `/dashboard`
- Next.js: `app/(auth)/login/page.tsx`

### Sidebar

Navigasi utama untuk owner/operator yang akses dashboard. Harus jelas, tidak overcrowded, mencerminkan hierarki fitur QIOS.

**UI Elements:**
- Logo QIOS
- Menu Items: Dashboard (icon: grid/home), Analytics (icon: bar chart), AI Insight (icon: spark/brain), Divider, Settings (icon: gear) — Placeholder MVP
- User Info (di bottom): Nama Bisnis, Avatar, Logout
- Collapsible (Desktop), Drawer (Mobile)

**Teknikal:**
- Komponen global di `app/(dashboard)/layout.tsx`
- Active state dari `usePathname()`
- Logout: clear token, redirect ke `/login`
- Tidak perlu role-based di MVP (satu user = satu bisnis)

### Dashboard

Halaman pertama setelah login. Tujuannya satu: kasih owner gambaran kondisi bisnis hari ini dan 7 hari terakhir dalam hitungan detik. Snapshot, bukan analisis mendalam.

**UI Elements:**
- Metric Cards: Total Revenue (dengan delta % vs periode sebelumnya), Jumlah Transaksi, Average Order Value, Transaksi Berhasil vs Gagal/Pending
- Chart: Revenue trend 7 hari terakhir (line chart), Jam tersibuk hari ini (bar chart horizontal — skip kalau data belum cukup)
- List: Top 5 produk terlaris (nama + jumlah terjual)

**Teknikal:**
- Endpoint: `GET /analytics/summary`
- Data fetching: Server Component atau `useEffect` + loading skeleton
- Chart library: **Recharts** (ringan, kompatibel Next.js)
- Timeframe filter: toggle "Hari Ini / 7 Hari" di header section, bukan full date picker
- Next.js: `app/(dashboard)/dashboard/page.tsx`

### Analytics

Detailed view untuk owner yang mau deep dive. User pilih timeframe dan breakdown sendiri. Target: evaluasi performa mingguan/bulanan sebelum ambil keputusan stok atau promo.

**UI Elements:**
- Filter Bar: Date range picker dengan preset (7 hari, 30 hari, 3 bulan) + custom range
- Section 1 — Revenue Overview: Line chart revenue over time, total revenue + jumlah transaksi
- Section 2 — Transaksi Breakdown: Bar chart volume per hari, status breakdown (berhasil/gagal/pending)
- Section 3 — Product Performance: Tabel nama produk, jumlah terjual, kontribusi revenue (sortable)
- Section 4 — Perbandingan Periode: Toggle bandingkan dengan periode sebelumnya, delta indicator per metric

**Teknikal:**
- Endpoint: `GET /analytics/transactions?from=&to=`
- Date range: `react-day-picker` atau native input, hindari heavy library
- Tabel: plain HTML table dengan Tailwind, tidak perlu data grid library di MVP
- Next.js: `app/(dashboard)/analytics/page.tsx`

### AI Insight

Bukan chatbot. QIOS ngomong duluan berdasarkan data bisnis owner. Tujuannya: kasih rekomendasi atau observasi yang actionable tanpa owner harus nanya.

**Format: Insight Card.** Tiap card berisi:
- Icon kategori (tren, peringatan, peluang)
- Judul singkat (maks 10 kata)
- Narasi 1-2 kalimat, bahasa natural, kontekstual
- Tombol "Lihat Data" yang expand ke mini chart/breakdown pendukung
- Timestamp: "Diperbarui X jam lalu"

**Contoh konten:**
- "Hari Selasa merupakan hari terkuat secara konsisten. Rata-rata revenue hari Selasa 38% di atas hari lain dalam 30 hari terakhir."
- "Tiga produk tidak mencatat penjualan selama 14 hari terakhir. Direkomendasikan untuk meninjau ulang stok atau penetapan harga item ini."
- "Waktu puncak transaksi adalah 12.00–14.00. Periode ini berkontribusi sebesar 41% dari total transaksi harian."

**Teknikal:**
- Endpoint: `GET /insight`
- MVP: insight semi-hardcoded logic di backend (rule-based), bukan LLM. LLM bisa dicolok nanti ke endpoint yang sama tanpa perubahan frontend
- Frontend hanya render response, tidak ada logika AI di client
- Expand "Lihat Data": fetch breakdown spesifik per insight type, atau sudah diinclude dalam response payload
- Refresh: manual button atau auto setiap buka halaman
- Next.js: `app/(dashboard)/insight/page.tsx`
- State: loading skeleton per card, empty state kalau data belum cukup ("Insight akan muncul setelah 7 hari transaksi")

### Kasir (Follow Through — Mobile PWA)

Interface untuk operator di lapangan. Mobile-first dari awal. PWA conversion dilakukan post-MVP.

**Scope MVP:**
- Input transaksi/order baru
- Generate QR Midtrans static dengan unique `order_id`
- Tampil status pembayaran (polling atau webhook-driven)
- Riwayat transaksi hari ini (list sederhana)

**API yang dicolok:**
- `POST /orders` — buat order baru
- `GET /orders/:id` — cek status
- `POST /payments/initiate` — generate QR Midtrans
- `GET /payments/:id/status` — cek status pembayaran

**Catatan untuk Dev:** Build `/kasir` sebagai route biasa dalam Next.js. Pastikan layout sudah mobile-first dan tidak ada dependency yang block PWA conversion nanti. Jangan implement service worker dulu.

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
| `MIDTRANS_SERVER_KEY` | Server key Midtrans global (fallback) | — |
| `MIDTRANS_ENV` | `sandbox` atau `production` | `sandbox` |

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
1. Order domain: `POST /orders`, `GET /orders`, `GET /orders/:id`
2. Payment domain: Midtrans QR initiation

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
1. Analytics endpoints: summary, transaksi breakdown
2. Insight logic: rule-based, minimal 3-5 insight types

**Frontend Developer 1 — Analytics:**
1. Analytics page: date filter, charts, tabel produk

**Frontend Developer 2 — Dashboard Live:**
1. Swap mock data ke API real
2. Loading skeleton, error state

**Frontend Developer 3 — Kasir Live:**
1. Connect ke `POST /orders` + Midtrans QR
2. Status payment polling

### Week 4 — Insight + Polish

**Backend:**
1. Insight endpoint live
2. Error handling, response consistency

**Frontend Developer 1 — AI Insight:**
1. Insight page: render insight cards
2. Empty state, loading skeleton

**Frontend Developer 2 & 3 — Cross-polish:**
1. Responsive check semua halaman
2. Edge cases: empty data, error state, token expired

### Week 5 — Integration Testing + Bug Fix

1. Full flow testing end-to-end
2. FE + BE sync untuk edge cases
3. Performance check (loading time, chart render)
4. Staging deploy (Hostinger/Biznet GioCloud)

### Week 6 — Buffer + MVP Finalization

1. Fix issues dari testing
2. Final QA
3. Demo-ready build
4. Dokumentasi minimal (README update, env setup)

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

Tidak ada pengecualian. Hotfix pun harus lewat branch sendiri.

### Naming Branch

```
feature/auth-login
feature/kasir-product-list
feature/dashboard-summary
feature/invoice-management
feature/analytics-breakdown
feature/insight-cards
fix/webhook-signature-validation
chore/update-dependencies
```

### Commit Message

Format: `<type>: <deskripsi singkat>`

```
feat: tambah endpoint POST /auth/login
feat: tambah invoice domain dengan Midtrans integration
fix: perbaiki race condition di webhook handler
fix: validasi signature key Midtrans di webhook endpoint
chore: update go dependencies
docs: update qios-api.yaml dengan domain payment dan invoice
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

- `midtrans_server_key` harus dienkripsi sebelum disimpan ke database
- Refresh token disimpan sebagai hash, bukan plain text
- Validasi signature Midtrans wajib dijalankan sebelum memproses webhook (requirement PG-05)
- Tidak ada secret atau credential yang di-commit ke repository

---

## Yang Belum Final (Pending)

- Seed data `plans` dan `subscriptions` — tunggu konfirmasi pricing dari board
- Konfirmasi ke Midtrans: apakah item detail per transaksi tersedia via API atau webhook
- Flutter vs PWA untuk jangka panjang — MVP tetap PWA Android
- Inventory management (C-18) — dijadwalkan post-MVP, schema belum final

---

## Dokumen Terkait

- `docs/qios-api.yaml` — OpenAPI 3.0.3 contract lengkap
- `AGENTS.md` (root) — panduan umum untuk AI agents
- `apps/server/AGENTS.md` — panduan implementasi spesifik server
- PRD QIOS — dokumen product requirement lengkap (source of truth untuk product decisions)