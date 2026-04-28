# QIOS

<img src="https://skillicons.dev/icons?i=nextjs,go,postgres,aws" />

<br />
<br />

[![Next.js](https://img.shields.io/badge/Next.js-15-black?style=for-the-badge&logo=nextdotjs)](https://nextjs.org/)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Echo](https://img.shields.io/badge/Echo-v4-00ADD8?style=for-the-badge)](https://echo.labstack.com/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=for-the-badge&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Node.js](https://img.shields.io/badge/Node.js-v20+-339933?style=for-the-badge&logo=nodedotjs&logoColor=white)](https://nodejs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5-3178C6?style=for-the-badge&logo=typescript&logoColor=white)](https://www.typescriptlang.org/)
[![JWT](https://img.shields.io/badge/Auth-JWT-000000?style=for-the-badge&logo=jsonwebtokens)](https://jwt.io/)
[![Midtrans](https://img.shields.io/badge/Payment-Midtrans-003399?style=for-the-badge)](https://midtrans.com/)
[![License](https://img.shields.io/badge/License-Private-red?style=for-the-badge)]()

<br />

[![PRD](https://img.shields.io/badge/📄%20Baca%20PRD%20Lengkap-Google%20Docs-4285F4?style=for-the-badge&logo=googledocs&logoColor=white)](https://docs.google.com/document/d/10RzXPqzt4TTMYUx1XPSTBdMS4jRiizxS0h3nTYEVRIM/edit?usp=sharing)

<br />

QIOS adalah platform SaaS berbasis web untuk manajemen pembayaran, inventori, dan AI insights khusus UMKM Indonesia. Dibangun dengan arsitektur monorepo yang memisahkan tanggung jawab secara jelas antara frontend, backend, dan database.

---

## Daftar Isi

- [Gambaran Arsitektur](#gambaran-arsitektur)
- [Kenapa Next.js, Go, dan PostgreSQL?](#kenapa-nextjs-go-dan-postgresql)
- [Struktur Repository](#struktur-repository)
- [Frontend Guide](#-frontend-guide)
- [Backend Guide](#-backend-guide)

---

## Gambaran Arsitektur

QIOS dibangun dengan tiga layer utama:

| Layer | Teknologi | Tanggung Jawab |
|---|---|---|
| Client | <img src="https://skillicons.dev/icons?i=nextjs" height="20" /> Next.js | Tampilan dan gateway request |
| Server | <img src="https://skillicons.dev/icons?i=go" height="20" /> Go + Echo | Logika bisnis dan keputusan |
| Database | <img src="https://skillicons.dev/icons?i=postgres" height="20" /> PostgreSQL | Penyimpanan data permanen |

**Cara kerjanya:**

```
Browser (User)
      ↓
apps/client     → Next.js   — UI dan mid-end API Routes
      ↓
apps/server     → Go        — business logic, auth, payment
      ↓
PostgreSQL                  — penyimpanan data permanen
```

> Browser tidak pernah ngobrol langsung dengan Go. Semua request dari browser harus lewat API Routes di Next.js terlebih dahulu.

---

## Kenapa Next.js, Go, dan PostgreSQL?

<img src="https://skillicons.dev/icons?i=nextjs" height="20" /> **Next.js** dipilih karena bisa handle frontend dan mid-end sekaligus dalam satu framework. API Routes-nya jalan di server, jadi bisa jadi jembatan aman antara browser dan backend Go.

<img src="https://skillicons.dev/icons?i=go" height="20" /> **Go** dipilih untuk backend karena performanya tinggi, concurrency-nya native, dan deployment-nya simpel. Sangat cocok untuk handle banyak request paralel seperti payment webhook.

<img src="https://skillicons.dev/icons?i=postgres" height="20" /> **PostgreSQL** dipilih karena mature, stabil, dan mendukung ACID — artinya data transaksi tidak akan corrupt meskipun terjadi error di tengah proses.

---

## Struktur Repository

```
/
├── apps/
│   ├── client/                 → Next.js (frontend + mid-end)
│   └── server/                 → Go + Echo (backend)
├── docs/
│   └── api.yaml                → OpenAPI spec, kontrak FE dan BE
├── docker-compose.yml
└── .env.example
```

---

---

# 🖥 Frontend Guide

> Bagian ini khusus untuk developer yang handle `apps/client` (Next.js).
> Jika kamu adalah developer backend, loncat ke [Backend Guide](#-backend-guide).

[![Next.js](https://img.shields.io/badge/Next.js-15-black?style=for-the-badge&logo=nextdotjs)](https://nextjs.org/)
[![Node.js](https://img.shields.io/badge/Node.js-v20+-339933?style=for-the-badge&logo=nodedotjs&logoColor=white)](https://nodejs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5-3178C6?style=for-the-badge&logo=typescript&logoColor=white)](https://www.typescriptlang.org/)
[![Tailwind](https://img.shields.io/badge/Tailwind-v4-06B6D4?style=for-the-badge&logo=tailwindcss&logoColor=white)](https://tailwindcss.com/)

---

## Apa yang Kamu Handle

- Semua tampilan UI — halaman, komponen, layout
- API Routes sebagai jembatan antara browser dan backend Go
- Auth guard — cek apakah user sudah login sebelum masuk halaman tertentu
- Token management — akses token disimpan di memory, bukan localStorage

Kamu **tidak** perlu menyentuh `apps/server`, `migrations/`, atau konfigurasi database.

---

## Prerequisites

### <img src="https://skillicons.dev/icons?i=git" height="20" /> 1. Install Git

- Download di: https://git-scm.com/downloads
- Verifikasi setelah install:

```bash
git -v
```

### <img src="https://skillicons.dev/icons?i=nodejs" height="20" /> 2. Install Node.js

- Download versi **LTS** terbaru di: https://nodejs.org/en
- Verifikasi setelah install:

```bash
node -v
npm -v
```

Output yang diharapkan: node `v20.x.x` ke atas, npm `v10.x.x` ke atas

### <img src="https://skillicons.dev/icons?i=vscode" height="20" /> 3. Install VS Code (Direkomendasikan)

- Download di: https://code.visualstudio.com/
- Install extension berikut:
  - **ESLint** — deteksi error di kode
  - **Prettier** — format kode otomatis
  - **Tailwind CSS IntelliSense** — autocomplete class Tailwind

---

## Minta Akses Repository

1. Buat akun GitHub di https://github.com jika belum punya
2. Kirim username GitHub kamu ke project lead
3. Terima email undangan dari GitHub, klik **Accept invitation**
4. Repository bisa diakses di https://github.com/theoneandonlyvabo/qios-web

---

## Clone Repository

```bash
# 1. masuk ke folder tempat kamu mau simpan project
cd Documents

# 2. clone repository
git clone https://github.com/theoneandonlyvabo/qios-web.git

# 3. masuk ke folder project
cd qios-web

# 4. verifikasi isi folder
ls
```

Kamu harus melihat folder `apps`, `docs`, dan file `docker-compose.yml`.

---

## Setup Project

```bash
# 1. masuk ke folder client
cd apps/client

# 2. install semua dependencies
npm install

# 3. buat file environment dari template
cp .env.example .env.local

# 4. buka dan isi nilai environment
code .env.local
```

Tanyakan nilai `.env.local` ke project lead. Jangan pernah commit file ini.

---

## Menjalankan Development Server

```bash
npm run dev
```

Buka browser dan akses `http://localhost:3000`. Jika muncul tampilan aplikasi, setup berhasil.

---

## Struktur Folder Client

```
apps/client/
├── app/
│   ├── (public)/           # halaman tanpa login — landing, login, register
│   ├── (dashboard)/        # halaman dengan login — dashboard, profil, transaksi
│   └── api/                # API Routes — jembatan ke backend Go
│       ├── auth/           # login, logout, refresh token
│       ├── user/           # data profil user
│       ├── order/          # buat dan lihat order
│       └── payment/        # inisiasi pembayaran
├── components/             # komponen UI yang bisa dipakai ulang
├── lib/
│   ├── api.ts              # fetch wrapper untuk request ke Go
│   └── auth.ts             # manajemen token
└── middleware.ts            # auth guard untuk halaman dashboard
```

---

## Environment Variables

```
NEXT_PUBLIC_APP_URL=http://localhost:3000
API_BASE_URL=http://localhost:8080
```

- `NEXT_PUBLIC_` — bisa dibaca browser. Jangan taruh data sensitif di sini
- Tanpa prefix — server-side only, tidak bisa diakses dari browser

---

## Git Workflow

```
main            → production-ready, jangan push langsung ke sini
dev             → branch aktif pengembangan, semua feature branch merge ke sini
feature/nama    → branch kerjamu, selalu dari dev
```

```bash
# mulai fitur baru
git checkout dev
git pull origin dev
git checkout -b feature/nama-fitur

# setelah selesai
git add .
git commit -m "feat: deskripsi singkat"
git push origin feature/nama-fitur
```

Buka Pull Request ke `dev` di GitHub. Jangan merge sendiri. Project lead yang review dan merge ke `main` kalau sudah siap.

---

## Aturan Penulisan Kode

- Komponen UI masuk ke `components/` — fokus tampilan, tidak boleh ada logika bisnis
- Jangan panggil Go langsung dari browser — selalu lewat `app/api/`
- Jangan simpan token di `localStorage` — sudah dihandle di `lib/auth.ts`
- Jangan buat folder baru tanpa diskusi dengan project lead

---

## Common Issues

**`npm install` gagal** — pastikan Node.js versi v20 ke atas. Cek dengan `node -v`.

**Environment variable tidak terbaca** — pastikan nama file persis `.env.local` dan letaknya di `apps/client/`.

**Error 401 Unauthorized** — token expired. Coba logout dan login ulang.

**Tidak bisa connect ke backend** — pastikan server Go berjalan di `localhost:8080`. Hubungi developer backend.

**Perubahan tidak muncul** — hard refresh dengan `Ctrl + Shift + R` (Windows) atau `Cmd + Shift + R` (Mac).

---

---

# ⚙️ Backend Guide

> Bagian ini khusus untuk developer yang handle `apps/server` (Go + Echo).
> Jika kamu adalah developer frontend, kembali ke [Frontend Guide](#-frontend-guide).

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Echo](https://img.shields.io/badge/Echo-v4-00ADD8?style=for-the-badge)](https://echo.labstack.com/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=for-the-badge&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![JWT](https://img.shields.io/badge/Auth-JWT-000000?style=for-the-badge&logo=jsonwebtokens)](https://jwt.io/)
[![Midtrans](https://img.shields.io/badge/Payment-Midtrans-003399?style=for-the-badge)](https://midtrans.com/)

---

## Apa yang Kamu Handle

- Semua logika bisnis — validasi, kalkulasi, aturan domain
- Auth — issue dan verifikasi JWT
- Integrasi Midtrans — inisiasi pembayaran dan handle webhook
- Akses database — query dan mutasi PostgreSQL

Kamu **tidak** perlu menyentuh `apps/client` atau file UI apapun.

---

## Prerequisites

### <img src="https://skillicons.dev/icons?i=git" height="20" /> 1. Install Git

- Download di: https://git-scm.com/downloads
- Verifikasi:

```bash
git -v
```

### <img src="https://skillicons.dev/icons?i=go" height="20" /> 2. Install Go

- Download versi terbaru di: https://go.dev/dl/
- Ikuti proses instalasi sesuai sistem operasi
- Verifikasi:

```bash
go version
```

Output yang diharapkan: `go version go1.22.x` ke atas

### <img src="https://skillicons.dev/icons?i=postgres" height="20" /> 3. Install PostgreSQL

- Download di: https://www.postgresql.org/download/
- Ikuti proses instalasi, catat username dan password yang kamu set
- Verifikasi:

```bash
psql --version
```

### <img src="https://skillicons.dev/icons?i=vscode" height="20" /> 4. Install VS Code (Direkomendasikan)

- Download di: https://code.visualstudio.com/
- Install extension:
  - **Go** (by Google) — syntax highlighting, autocomplete, dan tooling Go

---

## Minta Akses Repository

1. Buat akun GitHub di https://github.com jika belum punya
2. Kirim username GitHub kamu ke project lead
3. Terima email undangan dari GitHub, klik **Accept invitation**
4. Repository bisa diakses di https://github.com/theoneandonlyvabo/qios-web

---

## Clone Repository

```bash
# 1. masuk ke folder tempat kamu mau simpan project
cd Documents

# 2. clone repository
git clone https://github.com/theoneandonlyvabo/qios-web.git

# 3. masuk ke folder project
cd qios-web

# 4. verifikasi isi folder
ls
```

---

## Setup Project

```bash
# 1. masuk ke folder server
cd apps/server

# 2. install semua dependencies Go
go mod tidy

# 3. buat file environment dari template
cp .env.example .env

# 4. buka dan isi nilai environment
code .env
```

Tanyakan nilai `.env` ke project lead. Jangan pernah commit file ini.

---

## Menjalankan Development Server

```bash
go run cmd/main.go
```

Server akan berjalan di `http://localhost:8080`. Verifikasi dengan:

```bash
curl http://localhost:8080/health
```

Output yang diharapkan: `{"status":"ok"}`

---

## Struktur Folder Server

```
apps/server/
├── cmd/
│   └── main.go             # entry point — jalanin server di sini
├── domain/                 # semua logika bisnis per area
│   ├── auth/               # login, register, issue dan refresh JWT
│   │   ├── handler.go      # terima HTTP request, validasi input
│   │   ├── service.go      # logika bisnis
│   │   └── repository.go   # query database
│   ├── user/               # data user, update profil
│   ├── order/              # buat order, cek status
│   └── payment/            # integrasi Midtrans, handle webhook
├── platform/               # utility yang dipakai lintas domain
│   ├── jwt/                # generate dan verify token
│   ├── middleware/         # auth middleware, rate limiter
│   ├── database/           # koneksi postgres
│   └── response/           # format standar semua API response
├── migrations/             # SQL schema — jangan edit manual
├── config/
│   └── config.go           # load environment variables
├── go.mod
└── go.sum
```

**Aturan layer di dalam setiap domain:**

```
handler.go    → terima request, validasi input, return response
service.go    → semua keputusan bisnis ada di sini
repository.go → query database saja, tidak ada logika
```

Tidak boleh skip layer. Handler tidak boleh langsung query database. Service tidak boleh tau soal HTTP.

---

## Environment Variables

```
APP_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=qios
JWT_SECRET=your_jwt_secret
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=720h
MIDTRANS_SERVER_KEY=your_midtrans_key
MIDTRANS_ENV=sandbox
```

---

## Format API Response

Semua endpoint wajib menggunakan format response berikut. Jangan return format berbeda:

```json
{
  "success": true,
  "data": {},
  "error": null,
  "meta": {}
}
```

Format ini sudah dihandle di `platform/response/`. Gunakan helper yang tersedia, jangan buat sendiri.

---

## Auth Flow

```
POST /auth/login
  → verifikasi email + password di DB
  → issue access token (15 menit) + refresh token (30 hari)
  → return ke Next.js

Next.js simpan refresh token di httpOnly cookie
Browser simpan access token di memory

Setiap request berikutnya:
  → browser kirim access token di header Authorization
  → middleware Go verifikasi token sebelum request diproses
```

---

## Payment Webhook Flow

```
User bayar di Midtrans
  → Midtrans kirim POST ke /payment/webhook
  → handler verifikasi signature Midtrans
  → update status order di database
```

Webhook harus selalu verifikasi signature sebelum proses apapun. Jangan pernah skip langkah ini.

---

## Git Workflow

```
main            → production-ready, jangan push langsung ke sini
dev             → branch aktif pengembangan, semua feature branch merge ke sini
feature/nama    → branch kerjamu, selalu dari dev
```

```bash
# mulai fitur baru
git checkout dev
git pull origin dev
git checkout -b feature/nama-fitur

# setelah selesai
git add .
git commit -m "feat: deskripsi singkat"
git push origin feature/nama-fitur
```

Buka Pull Request ke `dev` di GitHub. Jangan merge sendiri. Project lead yang review dan merge ke `main` kalau sudah siap.

**Format commit:**

```
feat: tambah endpoint order
fix: perbaiki validasi JWT expired
refactor: pisah service dan repository di domain payment
chore: update go dependencies
```

---

## Aturan Penulisan Kode

- Setiap domain punya tiga file — handler, service, repository. Jangan gabung
- Antar domain tidak boleh saling import langsung — komunikasi lewat interface
- Semua response pakai helper dari `platform/response/`
- Jangan taruh logika bisnis di handler atau repository
- Jangan buat folder baru tanpa diskusi dengan project lead

---

## Common Issues

**`go mod tidy` error** — pastikan Go versi 1.22 ke atas. Cek dengan `go version`.

**Environment variable tidak terbaca** — pastikan nama file persis `.env` dan letaknya di `apps/server/`.

**Tidak bisa connect ke PostgreSQL** — pastikan PostgreSQL sedang berjalan dan nilai `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` di `.env` sudah benar.

**Port 8080 sudah dipakai** — ada proses lain yang pakai port yang sama. Matikan prosesnya atau ganti `APP_PORT` di `.env`.

---

## Kontak

Jika ada pertanyaan tentang codebase ini, hubungi project lead sebelum membuat asumsi sendiri.