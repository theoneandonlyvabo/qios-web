<div align="center">

# QIOS

**Sistem manajemen keuangan dan operasional berbasis web untuk UMKM Indonesia**

[![Next.js](https://img.shields.io/badge/Next.js-16.2.6-black?style=flat-square&logo=nextdotjs)](https://nextjs.org)
[![TypeScript](https://img.shields.io/badge/TypeScript-5-3178c6?style=flat-square&logo=typescript)](https://www.typescriptlang.org)
[![Go](https://img.shields.io/badge/Go-1.25-00add8?style=flat-square&logo=go)](https://go.dev)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat-square&logo=postgresql)](https://www.postgresql.org)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ed?style=flat-square&logo=docker)](https://docker.com)
[![Bruno](https://img.shields.io/badge/Bruno-API_Collection-FF6C37?style=flat-square)](https://www.usebruno.com)

> Project Bible lengkap ada di [`docs/CLAUDE.md`](docs/CLAUDE.md). Baca sebelum menyentuh kode.

</div>

---

## Daftar Isi

- [Arsitektur](#arsitektur)
- [Dashboard App](#%EF%B8%8F-dashboard-app-owner)
- [Operator App PWA](#-operator-app-pwa-kasir)
- [Admin App](#-admin-app-skalar-staff)
- [Backend API](#%EF%B8%8F-backend-api)
- [Common Issues](#common-issues)

---

## Arsitektur

### Tech Stack

| Layer | Teknologi |
|---|---|
| Dashboard App (owner) | Next.js 16, TypeScript, Tailwind CSS v4, Recharts |
| Operator App (kasir) | Next.js 16, TypeScript, Tailwind CSS v4, PWA |
| Admin App (Skalar) | Next.js 16, TypeScript, Tailwind CSS v4 |
| API Server | Go 1.25, Echo v4, lib/pq, golang-jwt |
| Database | PostgreSQL 16 (Docker) |
| API Spec | OpenAPI 3.0.3 (`docs/qios-api.yml`) |
| API Testing | Bruno (committed ke repo) |

### Struktur Repository

```
qios-web/
├── app/
│   ├── client/
│   │   ├── dashboard/      # Next.js — interface owner (desktop-first)
│   │   ├── operator/       # Next.js — PWA kasir (mobile-first, Android)
│   │   └── admin/          # Next.js — panel Skalar staff
│   └── server/
│       ├── api/            # Go + Echo — REST API
│       └── bruno/          # Bruno API collection
├── docs/
│   ├── CLAUDE.md           # Project Bible (baca ini dulu)
│   └── qios-api.yml        # OpenAPI contract v0.4
└── infra/
    ├── database/
    │   └── migrations/     # SQL migration files (append-only)
    └── docker-compose.yml
```

Tiga client app independen yang share satu backend API. Di-deploy ke subdomain berbeda. Tidak ada shared component library antar app di MVP.

---

## 🖥️ Dashboard App (Owner)

**Path:** `app/client/dashboard/`  
**Port dev:** `http://localhost:3000`

### Tanggung Jawabmu

| ✅ Kamu handle | ❌ Jangan disentuh |
|---|---|
| Semua UI — halaman, komponen, layout | `app/server/` |
| API Routes sebagai jembatan ke backend Go | `infra/database/migrations/` |
| Auth guard dan redirect middleware | Konfigurasi database |
| Token management (memory + localStorage) | File `.env` server |

### Prerequisites

- Node.js v24+
- npm
- Backend API running di `http://localhost:8080`

### Quickstart

```bash
cd app/client/dashboard
npm install
cp .env.example .env.local
# Minta nilai .env.local ke project lead
npm run dev
```

Buka `http://localhost:3000`. Login page muncul = setup berhasil.

### Struktur Folder

```
app/client/dashboard/
├── app/
│   ├── (auth)/
│   │   └── login/          # email/password + Google OAuth
│   ├── dashboard/          # snapshot bisnis hari ini
│   ├── statistics/         # tren transaksi + produk terlaris
│   ├── analytics/          # AI insight cards (rule-based MVP)
│   ├── reports/            # laporan harian/bulanan/consumption + export
│   ├── history/            # list semua transaksi dengan filter
│   ├── operators/          # CRUD akun kasir
│   └── products/           # katalog produk (read-only)
├── components/             # komponen UI reusable
├── hooks/                  # useAuth, dll
├── lib/
│   ├── api.ts              # HTTP client terpusat
│   └── auth.ts             # token management
└── middleware.ts            # auth guard semua route
```

### Environment Variables

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_GOOGLE_CLIENT_ID=your-google-client-id
```

### Git Workflow

```
main ──────── production-ready, jangan push langsung
  └── dev ─── integrasi semua fitur
        └── feature/<nama> ── branch kerjamu
```

```bash
git checkout dev && git pull origin dev
git checkout -b feature/dashboard-statistics-chart

# setelah selesai
git push origin feature/dashboard-statistics-chart
# buat PR ke dev
```

### Aturan Kode

- Tidak ada `any` — gunakan type yang proper atau `unknown`
- Semua API call melalui `lib/api.ts`, bukan `fetch` langsung di komponen
- State management sesederhana mungkin — jangan tambah library global sebelum dibutuhkan
- Chart library: **Recharts** — jangan ganti tanpa diskusi tim
- Komponen naming: PascalCase. File hooks: camelCase dengan prefix `use`

---

## 📱 Operator App PWA (Kasir)

**Path:** `app/client/operator/`  
**Port dev:** `http://localhost:3001`

### Tanggung Jawabmu

| ✅ Kamu handle | ❌ Jangan disentuh |
|---|---|
| UI mobile-first untuk alur kasir | `app/server/` |
| QR scan login (kamera device) | `infra/database/migrations/` |
| Slide-to-confirm gesture (≥800ms hold) | Konfigurasi database |
| Offline indicator + error states | File `.env` server |

### Prerequisites

- Node.js v24+
- npm
- Backend API running di `http://localhost:8080`
- Chrome/Android untuk test kamera QR

### Quickstart

```bash
cd app/client/operator
npm install
cp .env.example .env.local
npm run dev
```

Buka `http://localhost:3001`. Login QR atau credential muncul = setup berhasil.

### Struktur Folder

```
app/client/operator/
├── app/
│   ├── login/              # QR scan (primary) atau operator_code + password
│   ├── order/              # pilih produk, set qty, cart
│   ├── confirm/            # pilih payment method + slide-to-confirm
│   └── history/            # riwayat transaksi hari ini
├── components/
├── hooks/
│   ├── useCamera.ts        # kamera untuk QR scan
│   └── useSlideConfirm.ts  # threshold ≥800ms gesture
└── lib/
    ├── api.ts
    └── qr.ts               # QR decode helper
```

### Auth Flow Operator

- **Login QR:** Operator scan QR yang owner generate dari dashboard → `POST /operator/auth/login/qr`
- **Login credential:** Input `operator_code` + password → `POST /operator/auth/login`
- JWT claim: `operator_id`, `business_id`, `scope: operator`
- Token di memory + localStorage (offline mode — intentional, bukan bug)

### Payment Methods

| Method | Flow |
|---|---|
| `CASH` | Kasir pilih Cash, slide-to-confirm |
| `QRIS_STATIC` | App tampilkan `business.qris_static_payload` sebagai QR untuk pembeli scan, kasir konfirmasi manual |
| `TRANSFER` | Tampilkan info rekening, kasir konfirmasi manual |

Tidak ada webhook atau payment gateway. Semua konfirmasi oleh kasir.

### Environment Variables

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### Aturan Kode

- **Mobile-first.** Semua layout dirancang untuk 360–430px dulu, baru tablet
- **Slide-to-confirm threshold ≥800ms** — jangan kurangi tanpa diskusi
- Tidak ada `any`
- Semua API call via `lib/api.ts`

---

## 🛡 Admin App (Skalar Staff)

**Path:** `app/client/admin/`  
**Port dev:** `http://localhost:3002`

### Tanggung Jawabmu

| ✅ Kamu handle | ❌ Jangan disentuh |
|---|---|
| UI onboarding merchant baru | `app/server/` |
| CRUD produk + recipe per merchant | `infra/database/migrations/` |
| Manage plan, features, status merchant | Konfigurasi database |
| Cross-merchant transaction view | File `.env` server |

### Prerequisites

- Node.js v24+
- npm
- Backend API running di `http://localhost:8080`

### Quickstart

```bash
cd app/client/admin
npm install
cp .env.example .env.local
npm run dev
```

Buka `http://localhost:3002`. Login admin muncul = setup berhasil.

### Struktur Folder

```
app/client/admin/
├── app/
│   ├── login/              # admin email/password, scope admin
│   ├── merchants/
│   │   ├── page.tsx        # list semua business dengan filter status/plan
│   │   ├── new/            # form onboard merchant baru
│   │   └── [id]/           # merchant detail (profile, products, operators, transactions, settings)
│   └── transactions/       # cross-merchant transactions read-only
├── components/
└── lib/
    └── api.ts
```

### Catatan Penting

Admin **tidak bisa** create/edit operator merchant — hanya bisa remove atas permintaan owner. Operator dikelola owner via dashboard.

Onboarding merchant = satu form yang create `users` + `businesses` dalam satu atomic DB transaction. Tidak ada external API call saat onboarding.

### Environment Variables

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

---

## ⚙️ Backend API

**Path:** `app/server/api/`  
**Port dev:** `http://localhost:8080`

### Tanggung Jawabmu

| ✅ Kamu handle | ❌ Jangan disentuh |
|---|---|
| Domain handler, service, repository | `app/client/` |
| SQL migration files (append-only) | File `.env` client |
| JWT issue dan verify per scope | Konfigurasi server client |
| Bruno collection update saat ada endpoint baru | — |

### Prerequisites

- Go 1.25+
- Docker + Docker Compose

### Quickstart

**1. Clone dan setup env:**

```bash
git clone https://github.com/theoneandonlyvabo/qios-web.git
cd qios-web

cp app/server/api/.env.example app/server/api/.env
# Edit .env: set DB_PASSWORD, JWT_SECRET, JWT_ADMIN_SECRET, ENCRYPTION_KEY
```

**2. Jalankan PostgreSQL:**

```bash
docker compose -f infra/docker-compose.yml up postgres -d
```

**3. Jalankan API server:**

```bash
cd app/server/api
go run ./cmd/...
# Migration otomatis jalan saat startup
# Output: "Server running on :8080"
```

**4. Seed data awal (opsional):**

```bash
go run ./cmd/seed
# Output: admin email + temporary password
```

**5. Buka Bruno collection:**

```
Buka Bruno desktop → open collection di app/server/bruno/
Set environment "local" → mulai testing endpoint
```

### Struktur Folder

```
app/server/api/
├── cmd/                    # entry point — main.go
├── config/                 # config.go — load semua env vars ke struct Config
├── core/                   # domain-driven business + view logic
│   ├── auth/               # owner login, Google OAuth, refresh, logout
│   ├── user/               # profil owner + business info (GET/PATCH /business)
│   ├── operator/           # CRUD operator (owner-side) + login PWA operator
│   ├── product/            # read-only owner, full CRUD via admin domain
│   ├── transaction/        # PENDING → CONFIRMED → VOIDED + consumption_log
│   ├── dashboard/          # view: summary, trend, peak hours, top products
│   ├── analytics/          # view: overview dengan period comparison
│   ├── report/             # view: daily/monthly sales, consumption, export
│   ├── insight/            # view: rule-based insight cards
│   └── admin/              # onboard merchant, CRUD product+recipe, audit, void
└── pkg/                    # shared utilities
    ├── database/           # connect PostgreSQL, jalankan migrasi
    ├── jwt/                # issue dan verify JWT (owner/operator/admin)
    ├── middleware/         # auth middleware per scope + RequireAdmin guard
    ├── response/           # helper response JSON standar
    ├── qmid/               # generator format QM-NNNNNN
    └── encryption/         # AES-256 placeholder
```

### Environment Variables

| Variable | Keterangan | Default |
|---|---|---|
| `APP_PORT` | Port server | `8080` |
| `DB_HOST` | Host PostgreSQL | `localhost` |
| `DB_PORT` | Port PostgreSQL | `5432` |
| `DB_USER` | Username DB | `postgres` |
| `DB_PASSWORD` | Password DB (**wajib**) | — |
| `DB_NAME` | Nama database | `qios` |
| `JWT_SECRET` | Secret JWT scope owner & operator (**wajib**) | — |
| `JWT_ADMIN_SECRET` | Secret terpisah JWT scope admin (**wajib**) | — |
| `JWT_ACCESS_EXPIRY` | Durasi access token | `15m` |
| `JWT_REFRESH_EXPIRY` | Durasi refresh token | `720h` |
| `ENCRYPTION_KEY` | AES-256 key 64 hex chars (**wajib**) | — |
| `REPORT_EXPORT_DIR` | Path sementara PDF/CSV export | `/tmp/qios-reports` |
| `REPORT_EXPORT_TTL` | Durasi download URL valid | `1h` |
| `DASHBOARD_ORIGIN` | Allowed CORS origin dashboard | `http://localhost:3000` |
| `OPERATOR_ORIGIN` | Allowed CORS origin operator | `http://localhost:3001` |
| `ADMIN_ORIGIN` | Allowed CORS origin admin | `http://localhost:3002` |

Startup **gagal** kalau `DB_PASSWORD`, `JWT_SECRET`, `JWT_ADMIN_SECRET`, atau `ENCRYPTION_KEY` kosong.

### Format API Response

```json
{ "success": true, "data": { ... }, "error": null }
{ "success": false, "data": null, "error": "pesan error" }
```

### Auth Flow (API Level)

```
Owner:   POST /auth/login              → access token (15m) + refresh token cookie (720h)
Operator: POST /operator/auth/login/qr (QR scan) atau /operator/auth/login (credential)
Admin:    POST /admin/auth/login        → JWT_ADMIN_SECRET terpisah, scope: admin
```

JWT scope `owner`, `operator`, `admin` dipisah — cross-scope access = 403.

### Domain Pattern

```
handler.go → service.go → repository.go
```

- **Handler:** terima request, validasi, panggil service, kembalikan response. Tidak menyentuh DB.
- **Service:** business logic. Tidak ada raw SQL.
- **Repository:** semua SQL. Tidak ada business logic.

Domain view (`dashboard`, `analytics`, `report`, `insight`) inject repository dari domain bisnis — tidak punya tabel sendiri. Dependency direction: **view → bisnis**. Arah sebaliknya = code review reject.

### Git Workflow

```bash
git checkout dev && git pull origin dev
git checkout -b feature/nama-endpoint

git commit -m "feat: add POST /transactions/{id}/void"
git push origin feature/nama-endpoint
# buat PR ke dev
```

**Format commit:**

```
feat: add transaction confirm endpoint with payment_method
fix: prevent operator void on others' transactions
refactor: split dashboard service from analytics service
chore: update go dependencies
docs: update qios-api.yml with admin endpoints
```

### Aturan Kode

- Tidak ada `os.Getenv()` di luar `config/config.go`
- Tidak ada logic bisnis di handler
- Error selalu di-wrap: `fmt.Errorf("auth: failed to find user: %w", err)`
- Semua response via `pkg/response/`
- Tidak ada raw SQL di luar layer repository
- Migration **append-only** — jangan edit file di `infra/database/migrations/`
- Update `docs/qios-api.yml` **sebelum** implementasi endpoint baru

---

## Common Issues

<details>
<summary><strong>Klik untuk expand</strong></summary>

<br />

| Gejala | Solusi |
|---|---|
| `go mod tidy` error | Cek versi Go: `go version` harus `1.25+` |
| Env var tidak terbaca | Nama file harus persis `.env` di `app/server/api/` |
| Tidak bisa connect ke PostgreSQL | Cek Docker: `docker compose -f infra/docker-compose.yml ps` |
| Port 8080 dipakai proses lain | Ganti `APP_PORT` di `.env` |
| Migration gagal saat startup | Jalankan dari direktori `app/server/api/`, bukan root repo |
| `401 Unauthorized` terus | Cek apakah token sudah expired — hit `/auth/refresh` |
| `403 Forbidden` di endpoint admin | Butuh token scope `admin` dari `POST /admin/auth/login` |
| QR scan tidak jalan di dev | Browser butuh permission kamera — pastikan `localhost` atau HTTPS |
| QRIS QR tidak muncul di confirm | Cek `business.qris_static_payload` sudah di-set via PATCH /business |

</details>

---

> Untuk implementasi detail: [`docs/CLAUDE.md`](docs/CLAUDE.md)  
> Update API contract sebelum buat endpoint baru: [`docs/qios-api.yml`](docs/qios-api.yml)
