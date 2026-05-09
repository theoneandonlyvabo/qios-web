# QIOS Backend — Agent Reference

Baca ini sebelum menulis kode apapun. Ketika ragu, lihat `domain/operator/` sebagai referensi.

---

## 1. Architecture Overview

```
Request → Handler → Service → Repository → Database
                 ↓
           platform/response
```

Tiga layer per domain. Handler parse HTTP. Service punya business logic. Repository punya SQL.
Tidak ada layer yang menyentuh concern layer lain.

Layout per domain:
```
domain/<name>/
├── model.go        # tipe, sentinel errors, ToResponse()
├── handler.go      # HTTP handlers + validasi input + RegisterRoutes
├── service.go      # Service interface + implementasi
└── repository.go   # Repository interface + PostgresRepository
```

Domain yang ada saat ini:
```
domain/
├── auth/         # login, register, refresh, logout, Google OAuth (closure pattern, legacy)
├── dashboard/    # stubs — summary, trend, peak hours, top products (placeholder)
├── operator/     # CRUD operator + kasir login (canonical reference)
├── payment/      # POS orders, transaksi, Xendit integration, webhook
├── product/      # katalog produk + soft delete
└── user/         # profile owner + bisnis (closure pattern, legacy)
```

---

## 2. Layer Contract

**Handler:** parse request → validate → call service → map error → return response.
Tidak ada SQL. Tidak ada business logic.

**Service:** business rules dan orchestration. Error di-wrap dengan konteks:
`fmt.Errorf("payment service: create order: %w", err)`. Tidak ada `echo.Context`.
Didefinisikan sebagai interface.

**Repository:** semua SQL. Wrap error dengan konteks: `fmt.Errorf("payment: find order: %w", err)`.
Translate `sql.ErrNoRows` ke domain sentinel (`ErrNotFound`). Translate `pq.Error 23505` ke sentinel.
Tidak ada business logic.

---

## 3. Canonical Patterns

### 3.1 Error Handling

Sentinel di `model.go` atau `types.go`:
```go
var ErrNotFound      = errors.New("order not found")
var ErrInvalidStatus = errors.New("invalid order status transition")
```

Repository translate DB error ke sentinel:
```go
if errors.Is(err, sql.ErrNoRows) {
    return nil, ErrNotFound
}
```

Service wrap dengan konteks:
```go
return nil, fmt.Errorf("payment service: confirm: %w", err)
```

Handler punya `mapServiceError` — satu fungsi, semua kasus:
```go
func mapServiceError(c echo.Context, err error) error {
    switch {
    case errors.Is(err, ErrNotFound):
        return response.NotFoundMsg(c, "Order tidak ditemukan")
    case errors.Is(err, ErrInvalidStatus):
        return response.BadRequest(c, err.Error())
    default:
        return response.Internal(c)
    }
}
```

### 3.2 Response Wrapper

**Selalu pakai `platform/response` helpers. Jangan pernah raw `c.JSON()`.**

```go
response.OK(c, data)          // 200
response.Created(c, data)     // 201
response.NoContent(c)         // 200, data null
response.BadRequest(c, msg)   // 400
response.Unauthorized(c)      // 401
response.Forbidden(c)         // 403
response.NotFoundMsg(c, msg)  // 404
response.Conflict(c, msg)     // 409
response.Internal(c)          // 500
```

### 3.3 Handler Structure

Pakai struct + interface (bukan closure dengan raw `*sql.DB`):

```go
type Handler struct {
    service Service
}

func NewHandler(svc Service) *Handler {
    return &Handler{service: svc}
}

func (h *Handler) CreateOrder(c echo.Context) error { ... }

func RegisterRoutes(e *echo.Echo, h *Handler, authMw echo.MiddlewareFunc) {
    g := e.Group("/transactions", authMw, appmiddleware.RequireOperator)
    g.POST("", h.CreateOrder)
}
```

`domain/auth/` dan `domain/user/` masih pakai closure pattern (legacy).
Domain baru wajib pakai struct pattern. Refactor auth/user defer post-MVP.

### 3.4 Context Reading (JWT Claims)

Middleware inject claims sebagai `string`. Selalu parse eksplisit:

```go
func businessIDFromCtx(c echo.Context) (uuid.UUID, error) {
    raw, _ := c.Get("business_id").(string)
    id, err := uuid.Parse(raw)
    if err != nil {
        return uuid.Nil, errors.New("invalid business_id in token")
    }
    return id, nil
}
```

**JANGAN** `c.Get("business_id").(uuid.UUID)` — middleware inject `string`, ini akan panic.

### 3.5 Validation

Fungsi validator terpisah di `handler.go`, bukan inline:

```go
func validateCreateOrderRequest(r *CreateOrderRequest) string {
    if len(r.Items) == 0 {
        return "items wajib diisi minimal satu"
    }
    for _, item := range r.Items {
        if item.Quantity <= 0 {
            return "quantity harus lebih dari 0"
        }
    }
    return ""
}
```

Kembalikan empty string = valid. Panggil di awal handler:
```go
if msg := validateCreateOrderRequest(&req); msg != "" {
    return response.BadRequest(c, msg)
}
```

### 3.6 Transaction Management (Multi-table Write)

```go
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return nil, fmt.Errorf("payment service: begin tx: %w", err)
}
committed := false
defer func() {
    if !committed {
        _ = tx.Rollback()
    }
}()

// ... operasi ...

if err := tx.Commit(); err != nil {
    return nil, fmt.Errorf("payment service: commit: %w", err)
}
committed = true
```

---

## 4. Reference Implementation

**`domain/operator/`** adalah canonical reference untuk semua domain baru.

| File | Yang bisa dipelajari |
|------|----------------------|
| `model.go` | Sentinel errors, domain types, DTOs, `ToResponse()` |
| `repository.go` | Repository interface, `PostgresRepository`, `scanOperator` helper |
| `service.go` | Service interface, struct implementasi, external dependency via interface |
| `handler.go` | Handler struct, `validateXxx` functions, `mapServiceError`, `RegisterRoutes` |

`domain/payment/` skeleton (saat ini stub) mengikuti pola yang sama.

---

## 5. Inconsistency Registry

| # | Location | Issue | Disposition |
|---|----------|-------|-------------|
| 1 | ~~`domain/product/handler.go`~~ | ~~`c.Get("business_id").(uuid.UUID)` panic~~ | **Resolved 2026-05-09** |
| 2 | ~~`domain/transaction/`~~ | ~~Duplicate `/products` routes conflict~~ | **Resolved 2026-05-09 — folder deleted, merged ke `domain/payment/` + `domain/dashboard/`** |
| 3 | ~~`cmd/main.go`~~ | ~~`domain/product/` tidak ter-register~~ | **Resolved 2026-05-09** |
| 4 | ~~`domain/xendit/`~~ | ~~Payment concern terpecah~~ | **Resolved 2026-05-09 — folder deleted, merged ke `domain/payment/`** |
| 5 | `domain/auth/register.go` L153 | Raw `c.JSON(...)` instead of `response.Created()` | Defer — style only |
| 6 | `domain/auth/` `domain/user/` | Closure handler style, raw SQL dalam handler | Defer — working as-is, refactor post-MVP |
| 7 | `migrations/008_create_pos_orders.sql` | Status pakai lowercase, inkonsisten dengan `xendit_payments` uppercase | **Migration 014 approved** — case standardisasi defer (perlu data migration) |
| 8 | `businesses.xendit_api_key` / `xendit_secret_key` | Tidak diperlukan di MANAGED flow | **Resolved 2026-05-09** — Option C diadopsi, lihat section 8 (Xendit Integration Rules) |

---

## 6. Payment Domain Structure

**Status:** Konsolidasi selesai 2026-05-09. Semua payment concern di `domain/payment/`.

```
domain/payment/
├── model.go           # XenditStatus, ManagedAccountInput, ManagedAccountResult
├── xendit_service.go  # Xendit HTTP wrapper (Basic auth, /v2/accounts)
├── xendit_service_test.go
├── types.go           # PaymentMethod, OrderStatus, PosOrder, OrderItem, DTOs, sentinels
├── service.go         # Service interface + skeleton implementasi
├── repository.go      # Repository interface + PostgresRepository skeleton
├── handler.go         # Handler struct + transaction endpoints + Xendit connect/status
└── webhook.go         # WebhookHandler + verifyCallbackToken + qrWebhookPayload
```

Route ownership:
| Endpoint | Method | File |
|----------|--------|------|
| `/transactions` | POST/GET | `handler.go` |
| `/transactions/:id` | GET | `handler.go` |
| `/transactions/:id/complete` | POST | `handler.go` (cash flow) |
| `/payment/xendit/connect` | POST | `handler.go` |
| `/payment/xendit/status` | GET | `handler.go` |
| `/payment/xendit/webhook` | POST | `webhook.go` (belum di-register di main.go) |

**Webhook registration deferred** — butuh `XENDIT_WEBHOOK_TOKEN` env var. Wire saat actual Xendit integration sprint.

---

## 7. pos_orders Schema

**Migration 008 (existing):**
- `id`, `business_id`, `operator_id`, `order_id`, `total_amount`, `status`, `note`, `paid_at`, `created_at`

**Migration 014 (approved 2026-05-09, akan di-run saat startup berikutnya):**
- ADD `payment_method VARCHAR(20)` — required untuk distinguish cash vs digital
- ADD `updated_at TIMESTAMPTZ` — konsistensi dengan tabel lain
- UPDATE status CHECK constraint — tambah `'cancelled'`
- ADD index pada `payment_method`

**Status case standardisasi (lowercase vs uppercase):** Defer. Membutuhkan data migration karena `xendit_payments.status` pakai uppercase tapi `pos_orders.status` pakai lowercase. Mapping di service layer untuk MVP.

---

## 8. Xendit Integration Rules

**Status:** Approved 2026-05-09. Aturan ini dikunci untuk MVP.

### 8.1 Mode Aktif

**MANAGED.** OWNED pending approval Xendit Indonesia, paralel track post-MVP. Bukan blocker development.

### 8.2 Mekanisme API Call

Semua call ke Xendit atas nama sub-account menggunakan:
- **Header:** `for-user-id: <xendit_account_id>` (ambil dari `businesses.xendit_account_id`)
- **Auth:** master `XENDIT_SECRET_KEY` dari environment variable, Basic auth
- **Tidak ada per-merchant API key** untuk platform ops

Mekanisme code untuk MANAGED dan OWNED **identik** — perbedaan keduanya hanya di KYC handling dan fee structure, bukan di API call. Jangan buat interface `XenditClient` dengan implementasi MANAGED/OWNED terpisah — over-engineering.

### 8.3 Kolom Reserved

`businesses.xendit_api_key` dan `businesses.xendit_secret_key` ada di schema tapi **tidak diisi**.

- Schema dipertahankan untuk backward compat dan kandidat OWNED post-MVP
- `register.go` Step 6 hanya update `xendit_account_id` dan `xendit_status`
- Row baru selalu NULL di dua kolom itu — secara eksplisit menandakan "tidak dipakai"

**Jangan populate dua kolom ini sampai ada keputusan eksplisit untuk OWNED flow.**

### 8.4 Post-MVP Switch ke OWNED

- Tipe sub-account **tidak bisa diubah** setelah register — cutoff berlaku
- Merchant pre-OWNED approval → tetap MANAGED selamanya
- Merchant post-OWNED approval → bisa di-register sebagai OWNED
- Perlu tambah kolom `account_type VARCHAR(10) DEFAULT 'MANAGED'` di `businesses` saat transisi (migration baru, bukan sekarang)
- Transaksi flow tidak berubah saat switch — tetap master key + `for-user-id`
- KYC flow untuk OWNED harus dibangun terpisah (out of scope MVP)

---

## 9. Wiring Pattern (main.go)

Setiap domain mengikuti pattern wiring konsisten:

```go
// Domain dengan repo + service + handler:
xxxRepo    := xxx.NewPostgresRepository(db)  // atau NewRepository(db)
xxxSvc     := xxx.NewService(xxxRepo, ...deps)
xxxHandler := xxx.NewHandler(xxxSvc)
xxx.RegisterRoutes(e, xxxHandler, authMiddleware)

// Domain stub (dashboard):
dashboard.RegisterRoutes(e, dashboard.NewHandler(), authMiddleware)
```

Lihat `cmd/main.go` untuk urutan wiring lengkap.
