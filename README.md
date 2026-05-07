<div align="center">

# QIOS — by Skalar Solutions

**Platform manajemen keuangan dan operasional untuk UMKM Indonesia.**
Bukan POS — QIOS adalah BI layer di atas payment flow.

<br />

[![Next.js](https://img.shields.io/badge/Next.js-15-black?style=for-the-badge&logo=nextdotjs)](https://nextjs.org/)
[![Go](https://img.shields.io/badge/Go-1.26.2-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=for-the-badge&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5-3178C6?style=for-the-badge&logo=typescript&logoColor=white)](https://www.typescriptlang.org/)
[![Xendit](https://img.shields.io/badge/Payment-Xendit-003399?style=for-the-badge)](https://xendit.co/)
[![License](https://img.shields.io/badge/License-Private-red?style=for-the-badge)]()

<br />

[![PRD](https://img.shields.io/badge/📄%20Baca%20PRD%20Lengkap-Google%20Docs-4285F4?style=for-the-badge&logo=googledocs&logoColor=white)](https://docs.google.com/document/d/10RzXPqzt4TTMYUx1XPSTBdMS4jRiizxS0h3nTYEVRIM/edit?usp=sharing)

</div>

---

## Daftar Isi

- [Arsitektur](#arsitektur)
- [Struktur Repository](#struktur-repository)
- [Frontend Guide](#-frontend-guide)
- [Backend Guide](#-backend-guide)

---

## Arsitektur

```
Browser (Owner / Operator)
         │
         ▼
  ┌─────────────────┐
  │   apps/client   │  Next.js 15 — UI, route groups, API Routes
  └────────┬────────┘
           │  HTTP (internal)
           ▼
  ┌─────────────────┐
  │   apps/server   │  Go + Echo — business logic, auth, webhook
  └────────┬────────┘
           │
           ▼
  ┌─────────────────┐
  │   PostgreSQL    │  Docker — persistent storage, 12 migrations
  └─────────────────┘
```

> Browser tidak pernah ngobrol langsung dengan Go. Semua request wajib lewat API Routes Next.js terlebih dahulu.

### Tech Stack

| Layer | Teknologi | Alasan |
|---|---|---|
| Frontend | Next.js 15 + TypeScript | Handle UI dan mid-end dalam satu framework |
| Backend | Go 1.26.2 + Echo v4 | Performa tinggi, concurrency native untuk webhook |
| Database | PostgreSQL 16 | ACID compliance — data transaksi tidak boleh corrupt |
| Payment | Xendit xenPlatform | Sub-account per merchant, split rule otomatis |
| Auth | JWT (httpOnly cookie) | Refresh token aman, access token di memory |

---

## Struktur Repository

```
qios-web/
├── apps/
│   ├── client/          → Next.js (frontend + mid-end)
│   └── server/          → Go + Echo (backend)
├── docs/
│   └── qios-api.yaml    → OpenAPI 3.0.3 — kontrak antara FE dan BE
├── docker-compose.yml
└── .env.example
```

---

<br />

<div align="center">

# 🖥 Frontend Guide

*Bagian ini untuk developer yang handle `apps/client`.*
*Developer backend? Loncat ke [Backend Guide](#-backend-guide).*

[![Next.js](https://img.shields.io/badge/Next.js-15-black?style=flat-square&logo=nextdotjs)](https://nextjs.org/)
[![Node.js](https://img.shields.io/badge/Node.js-v24.15.0-339933?style=flat-square&logo=nodedotjs&logoColor=white)](https://nodejs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5-3178C6?style=flat-square&logo=typescript&logoColor=white)](https://www.typescriptlang.org/)
[![Tailwind](https://img.shields.io/badge/Tailwind-v4-06B6D4?style=flat-square&logo=tailwindcss&logoColor=white)](https://tailwindcss.com/)

</div>

### Tanggung Jawabmu

| ✅ Kamu handle | ❌ Jangan disentuh |
|---|---|
| Semua UI — halaman, komponen, layout | `apps/server/` |
| API Routes sebagai jembatan ke backend Go | `migrations/` |
| Auth guard dan redirect | Konfigurasi database |
| Token management (memory, bukan localStorage) | File `.env` server |

---

### Prerequisites

**1. Git** — https://git-scm.com/downloads

```bash
git -v
```

**2. Node.js v24.15.0** — https://nodejs.org/en

```bash
node -v   # harus v24.15.0
npm -v    # harus v10.x ke atas
```

**3. VS Code** (rekomendasi) — https://code.visualstudio.com/
Install extension: `ESLint`, `Prettier`, `Tailwind CSS IntelliSense`

---

### Akses Repository

1. Buat akun GitHub di https://github.com jika belum punya
2. Kirim username GitHub ke project lead
3. Terima email undangan → klik **Accept invitation**
4. Repo: https://github.com/theoneandonlyvabo/qios-web

---

### Quickstart

```bash
# Clone
git clone https://github.com/theoneandonlyvabo/qios-web.git
cd qios-web/apps/client

# Install dependencies
npm install

# Setup environment
cp .env.example .env.local
# → Minta nilai .env.local ke project lead. Jangan commit file ini.

# Jalankan
npm run dev
```

Buka `http://localhost:3000`. Tampilan muncul = setup berhasil.

---

### Struktur Folder

```
apps/client/
├── app/
│   ├── (dashboard)/        # owner — desktop-first
│   │   ├── dashboard/      # snapshot bisnis hari ini
│   │   ├── statistics/     # produk terlaris, tren transaksi
│   │   ├── analytics/      # AI analytics — insight rule-based
│   │   ├── history/        # list semua transaksi
│   │   ├── operators/      # CRUD akun kasir
│   │   └── login/          # pintu masuk
│   └── (kasir)/            # operator — mobile-first PWA
│       └── kasir/          # input order, QR Xendit, status bayar
├── components/             # komponen UI reusable
├── lib/
│   ├── api.ts              # fetch wrapper ke Go
│   └── auth.ts             # manajemen token
└── middleware.ts            # auth guard
```

---

### Environment Variables

```bash
NEXT_PUBLIC_APP_URL=http://localhost:3000
API_BASE_URL=http://localhost:8080
```

> `NEXT_PUBLIC_*` — terbaca di browser. Jangan taruh data sensitif.

---

### Git Workflow

```
main ──────── production-ready, jangan push langsung
  └── dev ─── integrasi semua fitur
        └── feature/<nama> ── branch kerjamu
```

```bash
# Mulai fitur baru
git checkout dev && git pull origin dev
git checkout -b feature/nama-fitur

# Setelah selesai
git add .
git commit -m "feat: deskripsi singkat"
git push origin feature/nama-fitur
```

Buka Pull Request ke `dev`. Jangan self-merge.

---

### Aturan Kode

- Komponen di `components/` — tampilan saja, tanpa logika bisnis
- Jangan panggil Go langsung dari browser — selalu lewat `app/api/`
- Token tidak boleh di `localStorage` — sudah dihandle di `lib/auth.ts`
- Chart library: **Recharts** — jangan ganti tanpa diskusi tim
- Jangan buat folder baru tanpa diskusi project lead

---

<details>
<summary><strong>Common Issues (klik untuk expand)</strong></summary>

<br />

| Gejala | Solusi |
|---|---|
| `npm install` gagal | Cek versi Node: `node -v` harus `v24.15.0` |
| Environment variable tidak terbaca | Nama file harus persis `.env.local` di `apps/client/` |
| Error 401 Unauthorized | Token expired — logout dan login ulang |
| Tidak bisa connect ke backend | Pastikan Go server jalan di `localhost:8080` |
| Perubahan tidak muncul | Hard refresh: `Ctrl+Shift+R` / `Cmd+Shift+R` |

</details>

---

<br />

<div align="center">

# ⚙️ Backend Guide

*Bagian ini untuk developer yang handle `apps/server`.*
*Developer frontend? Kembali ke [Frontend Guide](#-frontend-guide).*

[![Go](https://img.shields.io/badge/Go-1.26.2-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev/)
[![Echo](https://img.shields.io/badge/Echo-v4-00ADD8?style=flat-square)](https://echo.labstack.com/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat-square&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Xendit](https://img.shields.io/badge/Xendit-xenPlatform-003399?style=flat-square)](https://xendit.co/)

</div>

### Tanggung Jawabmu

| ✅ Kamu handle | ❌ Jangan disentuh |
|---|---|
| Semua logika bisnis per domain | `apps/client/` |
| Auth — issue dan verifikasi JWT | File UI apapun |
| Xendit xenPlatform — sub-account, QRIS, webhook | Styling atau komponen |
| Akses database — query dan mutasi PostgreSQL | |

---

### Prerequisites

**1. Git** — https://git-scm.com/downloads

```bash
git -v
```

**2. Go 1.26.2** — https://go.dev/dl/

```bash
go version   # harus go1.26.2
```

**3. Docker** — https://www.docker.com/products/docker-desktop/

```bash
docker --version
```

**4. VS Code** (rekomendasi) + extension **Go** (by Google)

---

### Akses Repository

1. Buat akun GitHub di https://github.com jika belum punya
2. Kirim username GitHub ke project lead
3. Terima email undangan → klik **Accept invitation**
4. Repo: https://github.com/theoneandonlyvabo/qios-web

---

### Quickstart

```bash
# Clone
git clone https://github.com/theoneandonlyvabo/qios-web.git
cd qios-web

# Jalankan PostgreSQL
docker compose up postgres -d

# Setup server
cd apps/server
go mod tidy
cp .env.example .env
# → Minta nilai .env ke project lead. Jangan commit file ini.

# Jalankan (migration otomatis saat startup)
go run ./cmd/...
```

Server jalan di `http://localhost:8080`.

---

### Struktur Folder

```
apps/server/
├── cmd/main.go             # entry point
├── config/config.go        # load semua env vars ke struct Config
├── domain/                 # logika bisnis per area
│   ├── admin/              # admin panel, audit log
│   ├── analytic/           # rule-based insight engine
│   ├── auth/               # login, refresh, logout
│   ├── dashboard/          # summary, tren, peak hours
│   ├── operator/           # CRUD akun kasir
│   ├── payment/            # Xendit, webhook handler
│   ├── product/            # katalog produk, soft delete
│   ├── statistic/          # produk terlaris, breakdown
│   ├── transaction/        # pos orders, order items
│   └── user/               # profil user dan bisnis
├── platform/
│   ├── database/           # koneksi PostgreSQL + migrasi
│   ├── jwt/                # issue dan verify token
│   ├── middleware/         # auth guard, role check
│   └── response/           # helper JSON response standar
└── migrations/             # file .sql bernomor urut (001–012)
```

**Status implementasi:**

| Domain | Status |
|---|---|
| auth | ✅ Done |
| user | ✅ Done |
| product | ✅ Done |
| operator | ✅ Done |
| transaction | 🔄 In progress |
| payment / xendit | 🔄 In progress |
| dashboard | ⏳ Pending |
| statistic | ⏳ Pending |
| analytic | ⏳ Pending |
| admin | ⏳ Pending |

**Aturan layer dalam setiap domain:**

```
handler.go    → terima request, validasi input, return response
    ↓
service.go    → semua keputusan bisnis di sini
    ↓
repository.go → query database saja, tanpa logika
```

Handler tidak boleh sentuh DB. Service tidak boleh tau soal HTTP.

---

### Environment Variables

| Variable | Nilai Default | Keterangan |
|---|---|---|
| `APP_PORT` | `8080` | Port server |
| `DB_HOST` | `localhost` | Host PostgreSQL |
| `DB_PORT` | `5432` | Port PostgreSQL |
| `DB_USER` | `postgres` | Username DB |
| `DB_PASSWORD` | — | Password DB |
| `DB_NAME` | `qios` | Nama database |
| `JWT_SECRET` | — | **Wajib diisi** — tidak boleh kosong |
| `JWT_ACCESS_EXPIRY` | `15m` | Durasi access token |
| `JWT_REFRESH_EXPIRY` | `720h` | Durasi refresh token |
| `XENDIT_SECRET_KEY` | — | Secret key Xendit |
| `XENDIT_ENV` | `sandbox` | `sandbox` atau `production` |

---

### Format API Response

Semua endpoint wajib pakai format ini — jangan buat format sendiri:

```json
{
  "success": true,
  "data": {},
  "error": null,
  "meta": {}
}
```

Gunakan helper dari `platform/response/`.

---

### Auth Flow

```
POST /auth/login
  → verifikasi email + password
  → issue access token (15m) + refresh token (720h)

Next.js:  refresh token → httpOnly cookie
Browser:  access token → memory

Setiap request:
  → Authorization: Bearer <access_token>
  → middleware verifikasi sebelum handler dipanggil
```

---

### Payment Flow — Xendit + xenPlatform

QIOS pakai **xenPlatform**: satu master account QIOS menaungi banyak sub-account merchant. Payment masuk ke sub-account masing-masing, fee QIOS diambil otomatis via split rule.

```
Operator pilih produk di kasir
  ↓
Server buat order_id unik (QIOS-YYYYMMDD-xxxx) → simpan ke pos_orders
  ↓
QR static Xendit merchant ditampilkan ke pembeli
  ↓
Pembeli scan dan bayar — order_id jadi payment reference
  ↓
Xendit → POST /payment/xendit/webhook
  ↓
Server cocokkan order_id → update status transaksi
```

> Setiap request ke Xendit wajib include header `for-user-id: {xendit_account_id}`.
>
> **Webhook wajib verifikasi signature Xendit sebelum proses apapun. Jangan skip.**

---

### Git Workflow

```
main ──────── production-ready, jangan push langsung
  └── dev ─── integrasi semua fitur
        └── feature/<nama> ── branch kerjamu
```

```bash
# Mulai fitur baru
git checkout dev && git pull origin dev
git checkout -b feature/nama-fitur

# Setelah selesai
git add .
git commit -m "feat: deskripsi singkat"
git push origin feature/nama-fitur
```

**Format commit:**

```
feat: tambah endpoint POST /transactions
fix: perbaiki type assertion panic di product handler
refactor: pisah service dan repository di domain payment
chore: update go dependencies
docs: update qios-api.yaml dengan domain transaction
```

---

### Aturan Kode

- Tidak ada `os.Getenv()` di luar `config/config.go`
- Semua domain wajib pakai interface (bukan concrete struct) agar bisa di-mock
- Semua method service dan repository harus terima `context.Context`
- Antar domain tidak boleh saling import langsung
- Semua response via helper `platform/response/`
- Migration bersifat append-only — jangan edit file yang sudah ada
- Jangan buat folder baru tanpa diskusi project lead

---

<details>
<summary><strong>Common Issues (klik untuk expand)</strong></summary>

<br />

| Gejala | Solusi |
|---|---|
| `go mod tidy` error | Cek versi Go: `go version` harus `go1.26.2` |
| Env var tidak terbaca | Nama file harus persis `.env` di `apps/server/` |
| Tidak bisa connect ke PostgreSQL | Cek Docker: `docker compose ps` — pastikan container aktif |
| Port 8080 dipakai proses lain | Matikan proses atau ganti `APP_PORT` di `.env` |
| Migration gagal | Jalankan server dari direktori `apps/server/`, bukan root repo |

</details>

---

<div align="center">

*Pertanyaan tentang codebase? Hubungi project lead sebelum membuat asumsi sendiri.*

</div>