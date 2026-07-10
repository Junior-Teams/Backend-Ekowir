# ApiGo

REST API sederhana untuk mengelola data **User** (auth berbasis JWT) dan **Apk** (entri game/aplikasi dengan upload file), dibangun dengan Go, [Gin](https://github.com/gin-gonic/gin), dan [GORM](https://gorm.io/) + PostgreSQL.

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

| Method | Endpoint                     | Auth   | Keterangan                     |
|--------|-------------------------------|--------|---------------------------------|
| GET    | `/api/`                       | Publik | Health check                    |
| POST   | `/api/token`                  | Publik | Login → dapat JWT               |
| POST   | `/api/user/register`          | Publik | Registrasi user baru            |
| GET    | `/api/apk`                    | Publik | List semua apk                  |
| GET    | `/api/auth/google/login`      | Publik | Redirect ke Google buat login   |
| GET    | `/api/auth/google/callback`   | Publik | Callback Google, redirect balik ke frontend bawa JWT |
| GET    | `/api/secured/ping`           | JWT    | Test endpoint secured           |
| POST   | `/api/secured/apk`            | JWT    | Upload apk baru (footage+cover) |
| GET    | `/api/secured/users`          | JWT    | List semua user                 |
| GET    | `/api/secured/users/:id`      | JWT    | Detail user                     |
| PUT    | `/api/secured/users/:id`      | JWT    | Update user                      |
| DELETE | `/api/secured/users/:id`      | JWT    | Hapus user                       |

Untuk endpoint yang butuh JWT, kirim header:
```
Authorization: Bearer <token>
```
(token mentah tanpa prefix `Bearer` juga tetap didukung)
