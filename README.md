# ApiGo

REST API sederhana untuk mengelola data **User** (auth berbasis JWT) dan **Apk** (entri game/aplikasi dengan upload file), dibangun dengan Go, [Gin](https://github.com/gin-gonic/gin), dan [GORM](https://gorm.io/) + PostgreSQL.

## Alur Aplikasi (Flowchart)

![Flowchart Ekowir](Ekowir%20-flowchart.png)

Flowchart di atas menggambarkan alur utama aplikasi Ekowir dari sisi user:

1. **Autentikasi** — User memulai dengan **Register / Login**. Sistem melakukan proses **Authentication**; kalau gagal, user dikembalikan ke proses autentikasi untuk mencoba lagi. Kalau berhasil, user masuk ke **Beranda**.
2. **Beranda** — Dari beranda, user bisa memilih empat menu utama: **Belajar**, **Leaderboard**, **Forum**, dan **Profile**. User juga bisa **Logout** untuk mengakhiri sesi.
3. **Alur Belajar** — Menu belajar adalah alur inti aplikasi:
   - User melihat **daftar course**, lalu **memilih course** yang diinginkan.
   - Di dalam course, user **memilih modul** dan mulai **belajar materi**.
   - Setelah materi selesai, user mengerjakan **Quiz**.
   - Hasil quiz dicek terhadap nilai kelulusan (**Passing**): kalau **tidak lulus**, user kembali mengerjakan quiz; kalau **lulus**, user mendapat reward berupa **XP + Badge** (gamifikasi) dan alur selesai.
4. **Alur Forum** — Dari menu forum, user bisa **melihat forum** dan masuk ke **daftar thread**. Dari sini user bisa **membuat thread baru** atau **membuka thread** yang sudah ada lalu **menambahkan komentar**.
5. **Leaderboard & Profile** — Leaderboard menampilkan peringkat user berdasarkan XP yang dikumpulkan dari quiz, sedangkan Profile menampilkan data serta pencapaian (badge/tier) milik user.

## Prasyarat

- [Docker](https://docs.docker.com/get-docker/) & Docker Compose (cara termudah), **atau**
- Go 1.22+ dan PostgreSQL kalau mau jalan lokal tanpa Docker

## Setup

1. Clone repo:
   ```bash
   git clone https://github.com/ALZEE23/ApiGo
   cd ApiGo
   ```
2. Copy `.env.example` menjadi `.env` dan isi sesuai kebutuhan:
   ```bash
   cp .env.example .env
   ```
   | Variabel      | Keterangan                                                   |
   |---------------|---------------------------------------------------------------|
   | `DB_USER`     | Username database Postgres                                   |
   | `DB_PASSWORD` | Password database Postgres                                   |
   | `DB_NAME`     | Nama database                                                 |
   | `DB_HOST`     | Host Postgres. Pakai `db` kalau lewat Docker Compose (default), atau `localhost` kalau Postgres jalan lokal |
   | `JWT_SECRET`  | Secret untuk sign/verify JWT — wajib diisi, ganti dengan string acak |
   | `GOOGLE_CLIENT_ID` | Client ID OAuth Google (lihat [Setup Google OAuth](#setup-google-oauth) di bawah) |
   | `GOOGLE_CLIENT_SECRET` | Client Secret OAuth Google |
   | `GOOGLE_REDIRECT_URL` | URL callback backend, harus sama persis dengan yang didaftarkan di Google Cloud Console (default `http://localhost:3000/api/auth/google/callback`) |
   | `FRONTEND_URL` | Base URL frontend, dipakai untuk redirect balik setelah login Google berhasil/gagal (mis. `http://localhost:5173`) |

## Setup Google OAuth

1. Buka [Google Cloud Console → Credentials](https://console.cloud.google.com/apis/credentials), buat **OAuth client ID** dengan tipe **Web application**.
2. Di **Authorized redirect URIs**, tambahkan persis nilai `GOOGLE_REDIRECT_URL` di `.env` (contoh: `http://localhost:3000/api/auth/google/callback`).
3. Copy **Client ID** & **Client Secret** yang dihasilkan ke `.env` (`GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET`).
4. Di frontend, cukup arahkan tombol "Login with Google" ke `GET {API_URL}/api/auth/google/login` (full-page redirect biasa, bukan lewat `fetch`/AJAX). Setelah user login & consent di Google, backend akan redirect balik ke `{FRONTEND_URL}/oauth/callback?token=<jwt>` (atau `?error=...` kalau gagal) — frontend perlu punya route di path itu untuk membaca query param `token` dan menyimpannya (mis. di `localStorage`), sama seperti token yang didapat dari `/api/token`.

## Menjalankan project

### Opsi A — Docker Compose (rekomendasi)

```bash
docker compose up --build
```

Ini akan menjalankan service `web` (API, hot-reload pakai [air](https://github.com/air-verse/air)) dan `db` (Postgres) sekaligus. API akan tersedia di `http://localhost:3000`.

### Opsi B — Lokal (tanpa Docker)

1. Pastikan PostgreSQL sudah jalan dan set `DB_HOST=localhost` (atau host lain) di `.env`.
2. Install dependency:
   ```bash
   go mod tidy
   ```
3. Jalankan:
   ```bash
   go run ./cmd/main.go
   ```
   Atau pakai hot-reload dengan `air` (config sudah ada di `.air.toml`):
   ```bash
   go install github.com/air-verse/air@v1.52.3
   air
   ```

API akan tersedia di `http://localhost:3000`.

## Seeding data dummy

Untuk membuat user dummy (pakai [faker](https://github.com/bxcodec/faker)):

```bash
go run ./cmd/seed -seed=users
```

## Ringkasan endpoint

### Publik (tanpa auth)

| Method | Endpoint                     | Keterangan                     |
|--------|-------------------------------|---------------------------------|
| GET    | `/api/`                       | Health check                    |
| POST   | `/api/token`                  | Login → dapat JWT               |
| POST   | `/api/user/register`          | Registrasi user baru            |
| GET    | `/api/auth/google/login`      | Redirect ke Google buat login   |
| GET    | `/api/auth/google/callback`   | Callback Google, redirect balik ke frontend bawa JWT |
| GET    | `/api/apk`                    | List semua apk                  |
| GET    | `/api/modules`                | List semua modul (course)       |
| GET    | `/api/modules/:id`            | Detail modul                    |
| GET    | `/api/materis`                | List semua materi               |
| GET    | `/api/materis/:id`            | Detail materi                   |
| GET    | `/api/quizzes`                | List semua quiz                 |
| GET    | `/api/quizzes/:id`            | Detail quiz                     |
| GET    | `/api/questions`              | List semua pertanyaan quiz (tanpa kunci jawaban) |
| GET    | `/api/questions/:id`          | Detail pertanyaan (tanpa kunci jawaban) |
| GET    | `/api/forums`                 | List semua thread forum         |
| GET    | `/api/forums/:id`             | Detail thread forum             |
| GET    | `/api/comments`               | List semua komentar             |
| GET    | `/api/comments/:id`           | Detail komentar                 |
| GET    | `/api/tiers`                  | List semua tier                 |
| GET    | `/api/tiers/:id`              | Detail tier                     |
| GET    | `/api/leaderboard`            | Leaderboard user berdasarkan XP |
| GET    | `/api/rewards`                | List semua reward               |
| GET    | `/api/rewards/:id`            | Detail reward                   |

### User (butuh JWT)

| Method | Endpoint                              | Keterangan                                  |
|--------|----------------------------------------|----------------------------------------------|
| GET    | `/api/secured/ping`                    | Test endpoint secured                        |
| GET    | `/api/secured/me`                      | Profil user yang sedang login                |
| PUT    | `/api/secured/me`                      | Update profil sendiri                        |
| PUT    | `/api/secured/me/password`             | Ganti password sendiri                       |
| POST   | `/api/secured/logout`                  | Logout                                       |
| GET    | `/api/secured/me/courses`              | Riwayat course milik user                    |
| GET    | `/api/secured/me/activity`             | Riwayat aktivitas user                       |
| POST   | `/api/secured/apk`                     | Upload apk baru (footage+cover)              |
| POST   | `/api/secured/quizzes/:id/submit`      | Submit jawaban quiz → dapat XP kalau lulus   |
| POST   | `/api/secured/materis/:id/complete`    | Tandai materi selesai                        |
| GET    | `/api/secured/modules/:id/progress`    | Progress user di sebuah modul                |
| GET    | `/api/secured/rewards`                 | List reward milik user                       |
| POST   | `/api/secured/rewards/:id/claim`       | Klaim reward                                 |
| POST   | `/api/secured/forums`                  | Buat thread forum baru                       |
| PUT    | `/api/secured/forums/:id`              | Update thread milik sendiri                  |
| DELETE | `/api/secured/forums/:id`              | Hapus thread milik sendiri                   |
| POST   | `/api/secured/comments`                | Buat komentar                                |
| PUT    | `/api/secured/comments/:id`            | Update komentar milik sendiri                |
| DELETE | `/api/secured/comments/:id`            | Hapus komentar milik sendiri                 |

### Admin (butuh JWT + role `admin`)

| Method | Endpoint                          | Keterangan                       |
|--------|------------------------------------|-----------------------------------|
| GET    | `/api/secured/dashboard`           | Statistik dashboard admin         |
| GET    | `/api/secured/users`               | List semua user                   |
| GET    | `/api/secured/users/:id`           | Detail user                       |
| POST   | `/api/secured/users`               | Buat user baru                    |
| PUT    | `/api/secured/users/:id`           | Update user                       |
| DELETE | `/api/secured/users/:id`           | Hapus user                        |
| POST   | `/api/secured/modules`             | Buat modul                        |
| PUT    | `/api/secured/modules/:id`         | Update modul                      |
| DELETE | `/api/secured/modules/:id`         | Hapus modul                       |
| POST   | `/api/secured/materis`             | Buat materi                       |
| PUT    | `/api/secured/materis/:id`         | Update materi                     |
| DELETE | `/api/secured/materis/:id`         | Hapus materi                      |
| POST   | `/api/secured/quizzes`             | Buat quiz                         |
| PUT    | `/api/secured/quizzes/:id`         | Update quiz                       |
| DELETE | `/api/secured/quizzes/:id`         | Hapus quiz                        |
| GET    | `/api/secured/questions`           | List pertanyaan (versi admin, termasuk kunci jawaban) |
| GET    | `/api/secured/questions/:id`       | Detail pertanyaan (versi admin)   |
| POST   | `/api/secured/questions`           | Buat pertanyaan                   |
| PUT    | `/api/secured/questions/:id`       | Update pertanyaan                 |
| DELETE | `/api/secured/questions/:id`       | Hapus pertanyaan                  |
| POST   | `/api/secured/tiers`               | Buat tier                         |
| PUT    | `/api/secured/tiers/:id`           | Update tier                       |
| DELETE | `/api/secured/tiers/:id`           | Hapus tier                        |
| POST   | `/api/secured/rewards`             | Buat reward                       |
| PUT    | `/api/secured/rewards/:id`         | Update reward                     |
| DELETE | `/api/secured/rewards/:id`         | Hapus reward                      |

File yang di-upload (mis. cover/footage) bisa diakses lewat path statis `/storage/...`.

Untuk endpoint yang butuh JWT, kirim header:
```
Authorization: Bearer <token>
```
(token mentah tanpa prefix `Bearer` juga tetap didukung)
