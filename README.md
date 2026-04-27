# RPLibrary Backend API

## 1. Setup Project

### 1.1 Prasyarat

- **Go 1.25+** (lihat [go.mod](go.mod)) - [Download](https://golang.org/dl/)
- **PostgreSQL 12 or newer** - [Download](https://www.postgresql.org/download/)
- **Git** - [Download](https://git-scm.com/)
- **Docker & Docker Compose** (opsional, direkomendasikan)

### 1.2 Konfigurasi Environment

1. Clone Repository

```bash
git clone https://github.com/Shabrinashsf/ADRIFT-backend.git
cd ADRIFT-backend
```

2. Salin `.env.example` ke `.env`

```bash
cp .env.example .env
```

3. Isi variabel sesuai environment lokal.

Daftar variabel utama:

- Database: `DB_HOST`, `DB_USER`, `DB_PASS`, `DB_NAME`, `DB_PORT`
- SMTP: `SMTP_HOST`, `SMTP_PORT`, `SMTP_SENDER_NAME`, `SMTP_SENDER_EMAIL`, `SMTP_AUTH_EMAIL`, `SMTP_AUTH_PASSWORD`
- Security: `JWT_SECRET`, `AES_KEY`
- App: `APP_PORT`, `APP_ENV`, `BE_URL`, `APP_URL`
- Storage cloud (opsional): `AWS_ACCESS_KEY`, `AWS_SECRET_KEY`, `AWS_REGION`, `S3_BUCKET`

Referensi lengkap: [.env.example](.env.example)

##### Cara mendapatkan akun SMTP gratis menggunakan akun google:

1. Buka [Google Account](https://myaccount.google.com/)
2. Pilih menu [Security](https://myaccount.google.com/security) dan aktifkan 2-Step Verification.
3. Setelah itu, pilih menu [App Passwords](https://myaccount.google.com/apppasswords)
4. Buat nama aplikasi (ex: RPLibrary)
5. Salin password yang diberikan, lalu masukkan ke variabel `SMTP_AUTH_PASSWORD` di file `.env`

### 1.3 Menjalankan Dengan Docker

```bash
docker-compose up --build -d
docker-compose exec app go run main.go --migrate
docker-compose exec app go run main.go --seed
docker-compose logs -f app
```

### 1.4 Menjalankan Lokal (Tanpa Docker App)

```bash
go mod download
go run main.go --migrate
go run main.go --seed
go run main.go
```

Server default berjalan di: `http://localhost:8888`

Fait avec amour by [Shabrinashsf](https://github.com/Shabrinashsf)
