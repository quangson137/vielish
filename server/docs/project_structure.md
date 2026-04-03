# Vielish Server — Kiến trúc tổng thể

> **Go module:** `github.com/sonpham/vielish/server`
> **Architecture:** Hexagonal (Ports & Adapters) + Clean Architecture + DDD
> **DI Framework:** Uber Fx

---

## Sơ đồ các layer

```
                    ┌──────────────────────────────────────────────┐
                    │              Driving Adapters                │
                    │         (internal/driving/httpui/)           │
                    │                                              │
                    │   Handler ─── Middleware ─── Presenter       │
                    └──────────────────┬───────────────────────────┘
                                       │
                                       ▼
                    ┌──────────────────────────────────────────────┐
                    │           Application Layer                  │
                    │           (internal/appcore/)                │
                    │                                              │
                    │              Use Cases                       │
                    └──────────────────┬───────────────────────────┘
                                       │
                                       ▼
                    ┌──────────────────────────────────────────────┐
                    │             Domain Layer                     │
                    │           (internal/domain/)                 │
                    │                                              │
                    │   Entity ─── Service ─── Repository (Port)   │
                    └──────────────────┬───────────────────────────┘
                                       ▲
                                       │ implements
                    ┌──────────────────┴───────────────────────────┐
                    │            Driven Adapters                   │
                    │           (internal/driven/)                 │
                    │                                              │
                    │       PostgreSQL ─── Redis ─── ...           │
                    └──────────────────────────────────────────────┘
```

---

## Nhiệm vụ từng layer

### 1. Domain Layer — `internal/domain/`

> Lõi nghiệp vụ. **Không phụ thuộc bất kỳ framework hay infrastructure nào.**

| Thành phần | Nhiệm vụ |
|------------|-----------|
| **Entity** | Định nghĩa business objects (ví dụ: `User`) |
| **Service** | Pure business logic — hash password, tạo/validate JWT, tạo refresh token |
| **Repository (Port)** | **Interface** mô tả các operations cần thiết — domain chỉ khai báo, không implement |
| **Errors** | Sentinel errors mang ngữ nghĩa business (`ErrEmailExists`, `ErrInvalidCredentials`...) |

```
internal/domain/auth/
├── entity.go        # User struct
├── service.go       # HashPassword, GenerateAccessToken, ValidateAccessToken...
├── repository.go    # Repository interface (Port)
└── errors.go        # Domain errors
```

**Nguyên tắc:** Domain layer là trung tâm, các layer khác phụ thuộc vào nó — **không bao giờ ngược lại**.

---

### 2. Application Layer — `internal/appcore/`

> Orchestrate các domain objects để thực hiện một use case cụ thể.

| Thành phần | Nhiệm vụ |
|------------|-----------|
| **UseCase** | Điều phối luồng nghiệp vụ: gọi Service → gọi Repository → trả kết quả |
| **DTO** | Input/Output objects cho từng use case (`RegisterInput`, `TokenOutput`...) |

```
internal/appcore/auth/
├── usecase.go       # Register, Login, Refresh, Logout
└── dto.go           # RegisterInput, LoginInput, TokenOutput
```

**Ví dụ flow `Login`:**
1. Gọi `repo.GetByEmail()` → lấy user
2. Gọi `service.CheckPassword()` → xác minh mật khẩu
3. Gọi `service.GenerateAccessToken()` + `service.GenerateRefreshToken()`
4. Gọi `repo.StoreRefreshToken()` → lưu vào Redis
5. Trả `TokenOutput`

---

### 3. Driving Adapters — `internal/driving/`

> **Primary adapters** — nhận request từ bên ngoài, chuyển đổi thành lời gọi use case.

| Thành phần | Nhiệm vụ |
|------------|-----------|
| **Handler** | Nhận HTTP request → validate input → gọi UseCase → xử lý domain errors thành HTTP status |
| **Middleware** | Cross-cutting concerns: JWT authentication, extract userID vào context |
| **Presenter** | Format UseCase output thành JSON response |

```
internal/driving/httpui/
├── handler/         # auth_handler.go — Register, Login, Refresh, Logout
├── middleware/      # auth.go — Bearer token validation
├── presenter/       # auth_presenter.go — Format token response
└── server.go        # Gin engine, CORS, route registration
```

**Hướng phụ thuộc:** Handler → UseCase → Domain. Handler **không bao giờ** gọi trực tiếp Repository hay Infrastructure.

---

### 4. Driven Adapters — `internal/driven/`

> **Secondary adapters** — implement các Port (interface) mà Domain layer khai báo.

| Thành phần | Nhiệm vụ |
|------------|-----------|
| **Repository impl** | Triển khai `domain.Repository` interface bằng PostgreSQL (GORM) + Redis |
| **Database** | Khởi tạo và quản lý connections (GORM, Redis) |
| **Model** | GORM model + mapping sang domain entity |

```
internal/driven/
├── auth/            # Repository implementation (PostgreSQL + Redis)
│   ├── repository.go
│   └── gorm_model.go
└── database/        # Connection providers (GORM, Redis)
```

**Dependency Inversion:** Driven adapter implement interface của Domain layer. Fx inject concrete type qua `fx.As(new(domain.Repository))` — UseCase chỉ biết interface, không biết implementation.

---

### 5. Shared Packages — `pkg/`

> Utilities dùng chung, không chứa business logic. Có thể tái sử dụng giữa các service.

| Package | Nhiệm vụ |
|---------|-----------|
| `config` | Load config từ YAML + env vars (Viper) |
| `log` | Zap logger factory |
| `tracing` | OpenTelemetry setup |
| `ctxbase` | Context value helpers (userID, requestID) |
| `httpbase` | Chuẩn hóa JSON response |

---

## Dependency Rule

```
Driving Adapters ──▶ Application Layer ──▶ Domain Layer ◀── Driven Adapters
    (httpui)            (appcore)            (domain)          (driven)
```

- Mũi tên chỉ **hướng phụ thuộc** (import)
- Domain layer **không import** bất kỳ layer nào khác
- Driven adapters **phụ thuộc vào Domain** (implement interface), không phụ thuộc vào Driving hay Appcore
- **Dependency Inversion Principle**: UseCase phụ thuộc vào `domain.Repository` (interface), không phụ thuộc vào concrete PostgreSQL/Redis implementation

---

## Wiring (Uber Fx)

Mỗi layer/feature expose một `fx.Module`. Entry point (`cmd/api/main.go`) compose tất cả:

```go
fx.New(
    config.Module,       // pkg
    log.Module,          // pkg
    tracing.Module,      // pkg
    database.Module,     // driven/database

    fx.Module("auth",
        authDomain.Module,   // domain
        authAppcore.Module,  // appcore
        authDriven.Module,   // driven (implements domain.Repository)
    ),

    httpui.Module,       // driving
).Run()
```

Khi thêm feature mới (ví dụ: `vocab`), chỉ cần tạo thêm các package tương ứng trong `domain/`, `appcore/`, `driven/`, `driving/` và thêm `fx.Module("vocab", ...)` vào `main.go`.
