@AGENTS.md

# CLAUDE.md — QIOS Project Bible

> Dokumen ini adalah sumber kebenaran tunggal untuk seluruh tim pengembang QIOS.
> Baca ini sebelum menyentuh satu baris kode pun. Jika ada konflik antara dokumen
> ini dengan kode yang ada, dokumen ini yang benar — kodenya yang perlu diperbaiki.

---

## Apa itu QIOS

QIOS adalah **business intelligence layer di atas sistem pembayaran Midtrans** untuk UMKM Indonesia.

QIOS bukan POS konvensional. QIOS tidak menggantikan kasir — QIOS adalah otak di balik kasir. Interface kasir yang dibangun bukan tujuan utama, melainkan pintu masuk data transaksi yang bersih dan terstruktur, yang kemudian diolah menjadi insight bisnis.

**Core value proposition:** Owner UMKM bisa melihat performa bisnis mereka secara real-time — produk terlaris, peak hours, tren revenue, dan pola pembelian — tanpa harus paham akuntansi atau spreadsheet.

**Yang QIOS lakukan:**
- Menerima pembayaran via QR Midtrans milik merchant
- Merekam setiap transaksi beserta item yang dibeli
- Mengolah data transaksi menjadi visualisasi dan insight
- Menyediakan interface kasir sederhana untuk operator di device Android

**Yang QIOS tidak lakukan (MVP):**
- Manajemen stok
- Sync dari marketplace eksternal (Tokopedia, GoBiz, dll)
- AI insight (defer post-MVP)
- Support iOS sebagai target utama

---

## Pengguna dan Role

**Owner** — pemilik bisnis, akses penuh ke dashboard, setting, dan semua data bisnis. Login via email/password atau Google OAuth.

**Operator** — kasir/pegawai, akses terbatas ke interface kasir saja. Login via email/password. Dibuat dan dikelola oleh owner. Tidak support Google OAuth.

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

**Stack:** Next.js 15, TypeScript, Tailwind CSS, npm

**Dua mode UI dalam satu codebase:**
- `(dashboard)` — interface owner, desktop-first, akses via `qios.id/dashboard`
- `(kasir)` — interface operator, mobile-first PWA, akses via `qios.id/kasir`

Pemisahan dilakukan via Next.js route groups. Layout, komponen, dan styling berbeda per mode. Logic bisnis dan API calls bisa dishare.

Interface kasir ditargetkan sebagai **PWA di Android**. iOS bukan prioritas MVP.

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
    └── payment/            # koneksi Midtrans, webhook handler
```

Setiap domain mengikuti pola: `handler.go` → `service.go` → `repository.go`. Handler tidak boleh menyentuh database langsung — semua lewat service dan repository.

### Database — PostgreSQL 16

Dijalankan via Docker. Migration dikelola secara manual menggunakan `migrate.go` berbasis file `.sql` bernomor urut di folder `migrations/`.

Migration bersifat **append-only** — file yang sudah ada tidak boleh diedit. Perubahan schema dilakukan dengan menambah file migration baru.

### Flow Transaksi

1. Operator pilih produk di interface kasir
2. Server generate `order_id` unik (format: `QIOS-YYYYMMDD-xxxx`) dan simpan ke `pos_orders`
3. QR static Midtrans milik merchant ditampilkan ke pembeli
4. Pembeli scan dan bayar — `order_id` dikirim sebagai payment reference ke Midtrans
5. Midtrans kirim webhook notifikasi ke server QIOS
6. Server cocokkan `order_id` dari webhook ke `pos_orders`, update status, increment `total_sold` di `products`

---

## API Contract

Kontrak lengkap ada di `docs/api.yaml` (OpenAPI 3.0.3).

**Aturan:**

- Semua response menggunakan shape yang konsisten:
  ```json
  { "success": true/false, "data": ..., "error": "..." }
  ```
- Autentikasi menggunakan Bearer JWT di header `Authorization`
- Refresh token disimpan di httpOnly cookie, bukan localStorage
- Role-based access dijalankan di middleware — handler tidak perlu cek role sendiri
- Endpoint webhook Midtrans (`POST /payment/midtrans/webhook`) tidak menggunakan Bearer auth, tapi diverifikasi via Midtrans signature key

Setiap perubahan endpoint **harus diupdate di `docs/api.yaml` terlebih dahulu** sebelum implementasi.

---

## Database Schema

12 migration files, urutan wajib dipertahankan:

| File | Tabel | Keterangan |
|------|-------|------------|
| 001 | `users` | Owner bisnis, support email + Google OAuth |
| 002 | `refresh_tokens` | Multi-device session |
| 003 | `password_reset_tokens` | Reset password via email |
| 004 | `plans`, `subscriptions` | Tier langganan QIOS — seed data pending konfirmasi board |
| 005 | `businesses` | Satu bisnis per owner, menyimpan Midtrans server key (encrypted) |
| 006 | `operators` | Akun kasir per bisnis |
| 007 | `products` | Katalog produk, soft delete |
| 008 | `pos_orders` | Order dari kasir, linked ke Midtrans via `order_id` |
| 009 | `pos_order_items` | Item per order, snapshot nama dan harga saat transaksi |
| 010 | `midtrans_payments` | Record pembayaran Midtrans |
| 011 | `webhook_events` | Log semua notifikasi masuk dari Midtrans |
| 012 | `admin_audit_logs` | Audit trail aksi admin |

**Aturan penting:**
- `product_name` dan `unit_price` di `pos_order_items` adalah snapshot — disimpan saat transaksi terjadi, bukan FK ke produk. Ini menjaga akurasi data historis jika produk diedit atau dihapus.
- `midtrans_server_key` di tabel `businesses` harus dienkripsi di level aplikasi sebelum disimpan.
- Semua tabel menggunakan `UUID PRIMARY KEY DEFAULT gen_random_uuid()`.

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
fix/webhook-signature-validation
chore/update-dependencies
```

### Commit Message

Format: `<type>: <deskripsi singkat>`

```
feat: tambah endpoint POST /auth/login
fix: perbaiki race condition di webhook handler
chore: update go dependencies
docs: update api.yaml dengan domain payment
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

### Database

- Migration bersifat append-only — tidak boleh edit file migration yang sudah ada
- Setiap tabel baru butuh index pada foreign key dan kolom yang sering di-query
- Gunakan soft delete (`deleted_at`) untuk data yang punya histori transaksi

### Security

- `midtrans_server_key` harus dienkripsi sebelum disimpan ke database
- Refresh token disimpan sebagai hash, bukan plain text
- Validasi signature Midtrans wajib dijalankan sebelum memproses webhook
- Tidak ada secret atau credential yang di-commit ke repository

---

## Yang Belum Final (Pending)

- Seed data `plans` dan `subscriptions` — tunggu konfirmasi pricing dari board
- Konfirmasi ke Midtrans: apakah item detail per transaksi tersedia via API atau webhook
- Flutter vs PWA untuk jangka panjang — MVP tetap PWA Android

---

## Dokumen Terkait

- `docs/api.yaml` — OpenAPI 3.0.3 contract lengkap
- `AGENTS.md` (root) — panduan umum untuk AI agents
- `apps/server/AGENTS.md` — panduan implementasi spesifik server