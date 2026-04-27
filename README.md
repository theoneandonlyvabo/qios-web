# QIOS — Panduan Developer Frontend

Dokumen ini khusus untuk developer yang handle layer `client` (Next.js). Baca sampai habis sebelum mulai nulis kode apapun.

---

## Daftar Isi

1. [Gambaran Arsitektur](#gambaran-arsitektur)
2. [Kenapa Next.js, Go, dan PostgreSQL?](#kenapa-nextjs-go-dan-postgresql)
3. [Apa yang Kamu Handle](#apa-yang-kamu-handle)
4. [Prerequisites](#prerequisites)
5. [Minta Akses Repository](#minta-akses-repository)
6. [Clone Repository](#clone-repository)
7. [Setup Project](#setup-project)
8. [Menjalankan Development Server](#menjalankan-development-server)
9. [Struktur Folder](#struktur-folder)
10. [Environment Variables](#environment-variables)
11. [Git Workflow](#git-workflow)
12. [Aturan Penulisan Kode](#aturan-penulisan-kode)
13. [Common Issues](#common-issues)

---

## Gambaran Arsitektur

QIOS dibangun dengan tiga layer utama yang punya tanggung jawab berbeda:

```
Browser (User)
      ↓
client/         → Next.js    — tampilan dan gateway request
      ↓
server/         → Go         — semua logika bisnis dan keputusan
      ↓
Database        → PostgreSQL — penyimpanan data permanen
```

**Cara kerjanya:**

1. User buka browser, melihat dan berinteraksi dengan UI yang dibangun di `client`
2. Ketika user melakukan sesuatu (login, bayar, lihat data), request dikirim ke API Route di `client` — bukan langsung ke backend
3. API Route di `client` meneruskan request ke `server` (Go)
4. Go memproses logika, mengambil atau menyimpan data ke PostgreSQL
5. Hasilnya dikembalikan ke `client`, lalu ditampilkan ke user

> Browser tidak pernah ngobrol langsung dengan Go. Semua harus lewat `client` dulu.

---

## Kenapa Next.js, Go, dan PostgreSQL?

Ini bukan pilihan random. Ada alasan di balik setiap teknologi yang dipilih.

**Next.js** dipilih karena dia bisa handle dua hal sekaligus: frontend (UI) dan mid-end (API Routes yang jalan di server). Artinya satu framework cukup untuk semua yang berhubungan dengan layer user. Next.js juga punya fitur server-side rendering, file-based routing, dan built-in middleware yang bikin development lebih cepat dan terstruktur.

**Go** dipilih untuk backend karena performanya tinggi, concurrency-nya native, dan deployment-nya simpel. Go sangat cocok untuk service yang perlu handle banyak request secara paralel — seperti payment webhook yang bisa datang bersamaan dari banyak user. Go juga strongly-typed, jadi lebih aman dan mudah di-maintain jangka panjang.

**PostgreSQL** dipilih karena dia relational database yang mature dan stabil. Untuk aplikasi SaaS yang handle transaksi dan data user, relational database adalah pilihan paling aman karena mendukung ACID — artinya data tidak akan corrupt meskipun terjadi error di tengah proses.

---

## Apa yang Kamu Handle

Sebagai developer frontend, tanggung jawabmu ada di folder `apps/client`. Secara spesifik:

- Semua tampilan UI (halaman, komponen, layout)
- API Routes yang jadi jembatan antara browser dan backend Go
- Auth guard — middleware yang cek apakah user sudah login sebelum masuk halaman tertentu
- Token management — akses token disimpan di memory, bukan localStorage

Kamu tidak perlu menyentuh folder `apps/server` (Go), `migrations/`, atau konfigurasi database.

---

## Prerequisites

Pastikan semua tools berikut sudah terinstall di komputermu sebelum melanjutkan.

### 1. Install Git

Git digunakan untuk clone repository dan mengelola kode.

- Download di: https://git-scm.com/downloads
- Pilih sistem operasi kamu, ikuti proses instalasi
- Setelah selesai, buka terminal dan verifikasi:

```bash
git -v
```

Output yang diharapkan: `git version 2.x.x`

### 2. Install Node.js

Node.js dibutuhkan untuk menjalankan Next.js.

- Download di: https://nodejs.org/en
- Pilih versi **LTS** yang paling baru
- Ikuti proses instalasi
- Setelah selesai, verifikasi:

```bash
node -v
npm -v
```

Output yang diharapkan: node `v20.x.x` ke atas, npm `v10.x.x` ke atas

### 3. Install VS Code (Direkomendasikan)

- Download di: https://code.visualstudio.com/
- Setelah terbuka, install extension berikut:
  - **ESLint** — deteksi error di kode
  - **Prettier** — format kode otomatis
  - **Tailwind CSS IntelliSense** — autocomplete class Tailwind

---

## Minta Akses Repository

Sebelum bisa clone, kamu harus punya akses ke repository.

1. Buat akun GitHub jika belum punya di https://github.com
2. Kirim username GitHub kamu ke project lead
3. Kamu akan mendapat email undangan dari GitHub — buka email dan klik **Accept invitation**
4. Setelah diterima, kamu bisa mengakses repository di https://github.com/theoneandonlyvabo/qios-web

---

## Clone Repository

Setelah punya akses, buka terminal di komputermu.

### 1. Pilih folder tempat menyimpan project

```bash
# contoh: masuk ke folder Documents
cd Documents
```

### 2. Clone repository

```bash
git clone https://github.com/theoneandonlyvabo/qios-web.git
```

Perintah ini akan mendownload semua file project ke folder `qios-web`.

### 3. Masuk ke folder project

```bash
cd qios-web
```

### 4. Verifikasi isi folder

```bash
ls
```

Kamu harus melihat folder `apps`, `docs`, dan file seperti `docker-compose.yml`.

---

## Setup Project

### 1. Masuk ke folder client

```bash
cd apps/client
```

### 2. Install semua dependencies

```bash
npm install
```

Perintah ini mengunduh semua library yang dibutuhkan Next.js. Tunggu sampai selesai.

### 3. Buat file environment

```bash
cp .env.example .env.local
```

Perintah ini menyalin template environment ke file `.env.local`.

### 4. Isi nilai environment

Buka file `.env.local`:

```bash
code .env.local
```

Isi nilai yang diperlukan. Tanyakan ke project lead untuk mendapatkan nilai yang benar. Jangan pernah share atau commit file ini.

---

## Menjalankan Development Server

Pastikan kamu masih di dalam folder `apps/client`, lalu jalankan:

```bash
npm run dev
```

Buka browser dan akses:

```
http://localhost:3000
```

Jika muncul tampilan aplikasi, setup berhasil.

> Jika ada error, baca pesan error di terminal dengan teliti. Sebagian besar error di tahap ini disebabkan oleh environment variable yang salah atau dependencies yang belum terinstall.

---

## Struktur Folder

```
apps/client/
├── app/
│   ├── (public)/           # halaman yang bisa diakses tanpa login
│   │                       # contoh: landing page, login, register
│   ├── (dashboard)/        # halaman yang butuh login
│   │                       # contoh: dashboard utama, profil, transaksi
│   └── api/                # API Routes — jembatan ke backend Go
│       ├── auth/           # login, logout, refresh token
│       ├── user/           # data profil user
│       ├── order/          # buat dan lihat order
│       └── payment/        # inisiasi pembayaran
├── components/             # komponen UI yang bisa dipakai ulang
├── lib/                    # helper functions
│   ├── api.ts              # fetch wrapper untuk request ke Go
│   └── auth.ts             # manajemen token
└── middleware.ts            # cek auth sebelum user masuk halaman dashboard
```

**Aturan sederhana:**

- Halaman baru masuk ke `app/(public)/` atau `app/(dashboard)/` tergantung apakah butuh login
- Komponen yang dipakai lebih dari satu tempat masuk ke `components/`
- Jangan buat folder baru di luar struktur ini tanpa diskusi dengan project lead

---

## Environment Variables

File `.env.local` berisi konfigurasi berikut:

```
NEXT_PUBLIC_APP_URL=http://localhost:3000
API_BASE_URL=http://localhost:8080
```

Penjelasan:

- `NEXT_PUBLIC_APP_URL` — URL aplikasi Next.js, bisa dibaca browser
- `API_BASE_URL` — URL backend Go, hanya dibaca di server-side. Tidak boleh diakses dari browser

> Variabel yang diawali `NEXT_PUBLIC_` bisa dibaca oleh browser. Variabel lainnya hanya bisa dibaca di server. Jangan pernah taruh data sensitif di variabel `NEXT_PUBLIC_`.

---

## Git Workflow

Kita menggunakan branching strategy sederhana:

```
main            → kode production-ready, jangan push langsung ke sini
dev             → branch aktif pengembangan, ini base branch kamu
feature/nama    → branch kerjamu untuk setiap fitur baru
```

### Mulai mengerjakan fitur baru

```bash
# pastikan branch dev kamu up to date
git checkout dev
git pull origin dev

# buat branch baru dari dev
git checkout -b feature/nama-fitur-kamu
```

Contoh nama branch yang baik:
- `feature/login-page`
- `feature/dashboard-layout`
- `feature/payment-initiation`

### Setelah fitur selesai

```bash
# simpan perubahan
git add .
git commit -m "feat: deskripsi singkat apa yang kamu buat"

# push ke GitHub
git push origin feature/nama-fitur-kamu
```

### Buka Pull Request

1. Buka https://github.com/theoneandonlyvabo/qios-web
2. Klik tombol **Compare & pull request** yang muncul
3. Pastikan base branch-nya adalah `dev`, bukan `main`
4. Tulis deskripsi singkat apa yang kamu kerjakan
5. Klik **Create pull request**
6. Tunggu review dari project lead — jangan merge sendiri

### Format pesan commit

```
feat: tambah halaman login
fix: perbaiki redirect setelah token expired
style: rapikan padding di header dashboard
refactor: sederhanakan fungsi fetch di lib/api.ts
chore: update dependencies
```

---

## Aturan Penulisan Kode

- Semua komponen UI masuk ke folder `components/`
- Komponen harus fokus pada tampilan — tidak boleh berisi logika bisnis
- Jangan panggil backend Go langsung dari browser — selalu lewat `app/api/`
- Jangan simpan token di `localStorage` — ini sudah dihandle di `lib/auth.ts`
- Jika tidak yakin sesuatu harus ditaruh di mana, tanya dulu ke project lead sebelum membuat folder baru

---

## Common Issues

**`npm install` gagal atau error**

Pastikan Node.js sudah versi terbaru. Cek dengan `node -v`. Jika versinya di bawah v20, update Node.js terlebih dahulu.

**Environment variable tidak terbaca**

Pastikan nama file-nya persis `.env.local` dan letaknya di dalam folder `apps/client/`, bukan di root project.

**Halaman menampilkan error 401 Unauthorized**

Access token kamu sudah expired atau tidak ada. Coba logout dan login ulang.

**Tidak bisa connect ke backend**

Pastikan backend Go sedang berjalan di `localhost:8080`. Hubungi developer backend untuk memastikan server mereka aktif.

**Perubahan kode tidak muncul di browser**

Coba hard refresh di browser dengan `Ctrl + Shift + R` (Windows) atau `Cmd + Shift + R` (Mac).

---

## Kontak

Jika ada pertanyaan tentang codebase ini, hubungi project lead sebelum membuat asumsi sendiri.