# Vielish

Ứng dụng học tiếng Anh dành cho người Việt. Tập trung vào từ vựng (flashcard + SRS) và luyện nghe.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.25 + Gin |
| Frontend | Next.js 16 (App Router) + TypeScript + Tailwind |
| Database | PostgreSQL 16 + Redis 7 |
| Audio | Cloud TTS (Google/Azure) |
| Dev | Docker Compose |

## Yêu cầu cài đặt

- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [Go 1.22+](https://go.dev/dl/)
- [Node.js 18+](https://nodejs.org/)
- [golang-migrate](https://github.com/golang-migrate/migrate) — để chạy database migrations

```bash
# Cài golang-migrate (macOS)
brew install golang-migrate
```

## Chạy lần đầu (First-time Setup)

### 1. Khởi động database

```bash
docker-compose up -d
```

Kiểm tra đã chạy:

```bash
docker-compose ps
# postgres và redis đều ở trạng thái "Up"
```

### 2. Chạy database migration

```bash
migrate \
  -path server/migrations \
  -database "postgres://vielish:vielish_dev@localhost:5432/vielish?sslmode=disable" \
  up
# Expected: 1/u create_users (Xms)
```

### 3. Cấu hình môi trường (tuỳ chọn)

```bash
cp .env.example .env
# Chỉnh sửa .env nếu cần thay đổi cổng hoặc JWT_SECRET
```

### 4. Cài dependencies frontend

```bash
cd web && npm install
```

---

## Chạy hàng ngày (Daily Development)

Mở **3 terminal** riêng biệt:

### Terminal 1 — Database

```bash
docker-compose up -d
```

### Terminal 2 — Backend (Go)

```bash
cd server
go run cmd/api/main.go
# Server chạy tại http://localhost:8080
```

Khi thay đổi code, dừng (`Ctrl+C`) và chạy lại lệnh trên.

### Terminal 3 — Frontend (Next.js)

```bash
cd web
npm run dev
# App chạy tại http://localhost:3000
```

---

## Kiểm tra hoạt động

### Health check

```bash
curl http://localhost:8080/api/health
# {"status":"ok"}
```

### Đăng ký tài khoản

```bash
curl -s -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","display_name":"Test User"}'
# {"access_token":"...","refresh_token":"...","expires_in":3600}
```

### Đăng nhập

```bash
curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
# {"access_token":"...","refresh_token":"...","expires_in":3600}
```

### Đăng xuất

```bash
curl -s -X POST http://localhost:8080/api/auth/logout \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token_từ_login>"}'
# {"message":"logged out"}
```

---

## Chạy tests

### Tất cả tests backend

```bash
cd server
go test ./... -v
```

### Từng package

```bash
cd server
go test ./internal/config/ -v
go test ./internal/auth/ -v
go test ./pkg/response/ -v
```

> **Lưu ý:** Tests trong `internal/auth/` là integration tests, cần PostgreSQL và Redis đang chạy.

---

## Database Migrations

```bash
# Áp dụng tất cả migrations
migrate -path server/migrations \
  -database "postgres://vielish:vielish_dev@localhost:5432/vielish?sslmode=disable" up

# Rollback 1 migration
migrate -path server/migrations \
  -database "postgres://vielish:vielish_dev@localhost:5432/vielish?sslmode=disable" down 1

# Xem trạng thái hiện tại
migrate -path server/migrations \
  -database "postgres://vielish:vielish_dev@localhost:5432/vielish?sslmode=disable" version
```

---

## Cấu trúc dự án

```
vielish/
  docker-compose.yml        — PostgreSQL + Redis services
  .env.example              — Template biến môi trường
  server/                   — Go backend
    cmd/api/main.go         — Entrypoint
    internal/
      config/               — Đọc cấu hình từ env
      database/             — Kết nối PostgreSQL + Redis
      auth/                 — Auth: model, repo, service, handler, middleware
      router/               — Đăng ký routes + CORS
    pkg/response/           — JSON response helpers
    migrations/             — SQL migration files
  web/                      — Next.js frontend
    app/                    — Pages (App Router)
      page.tsx              — Landing page
      login/                — Trang đăng nhập
      register/             — Trang đăng ký
      dashboard/            — Dashboard (yêu cầu đăng nhập)
    components/
      auth-form.tsx         — Form đăng nhập/đăng ký
    lib/
      api.ts                — API client với auto token refresh
      auth-context.tsx      — React auth context provider
  docs/                     — Specs & implementation plans
```

---

## Biến môi trường

| Biến | Mặc định | Mô tả |
|------|---------|-------|
| `APP_ENV` | `development` | Môi trường (`development` / `production`) |
| `PORT` | `8080` | Cổng backend |
| `DATABASE_URL` | `postgres://vielish:vielish_dev@localhost:5432/vielish?sslmode=disable` | PostgreSQL URL |
| `REDIS_URL` | `redis://localhost:6379` | Redis URL |
| `JWT_SECRET` | _(dev default)_ | **Bắt buộc** khi `APP_ENV=production` |
| `CORS_ORIGINS` | `http://localhost:3000` | Các origin được phép (phân cách bằng dấu phẩy) |
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080` | URL backend (dùng cho frontend) |

---

## Roadmap

- [x] Design spec
- [x] Project scaffolding (Docker, Go, Next.js)
- [x] Auth (register, login, refresh token, logout)
- [ ] Vocabulary module (flashcard + SRS)
- [ ] Listening module
- [ ] Progress tracking & dashboard
- [ ] AI conversation practice (post-MVP)
- [ ] Mobile app (post-MVP)
