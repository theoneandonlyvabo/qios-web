# QIOS v0.5 Migration Guide

> For FE teams and infra. Backend code changes are already live on `dev` branch.
> Migration 014 must be run manually against each environment — see section below.

---

## Endpoint Changes

### Order Domain (formerly POS)

| Old Path | New Path |
|----------|----------|
| `POST /pos/orders` | `POST /orders` |
| `GET /pos/orders` | `GET /orders` |
| `PATCH /pos/orders/:order_id/items` | `PATCH /orders/:order_id/items` |
| `POST /pos/orders/:order_id/checkout/begin` | `POST /orders/:order_id/checkout/begin` |
| `POST /pos/orders/:order_id/checkout/confirm` | `POST /orders/:order_id/checkout/confirm` |
| `POST /pos/orders/:order_id/void` | `POST /orders/:order_id/void` |
| `GET /pos/sessions` | `GET /orders/sessions` |
| `DELETE /pos/sessions/:session_id` | `DELETE /orders/sessions/:session_id` |

### Metrics Domain (consolidated from 4 domains)

| Old Path | New Path |
|----------|----------|
| `GET /dashboard/summary` | `GET /metrics/summary` |
| `GET /dashboard/transactions/trend` | `GET /metrics/trend` |
| `GET /dashboard/products/top` | `GET /metrics/top-products` |
| `GET /dashboard/transactions/peak-hours` | `GET /metrics/peak-hours` |
| `GET /analytics/overview` | `GET /metrics/overview` |
| `GET /reports/daily-sales` | `GET /metrics/reports/daily-sales` |
| `GET /reports/monthly-sales` | `GET /metrics/reports/monthly-sales` |
| `GET /reports/consumption` | `GET /metrics/reports/consumption` |
| `POST /reports/export` | `POST /metrics/reports/export` |
| `GET /insight` | `GET /metrics/insight` |

### Unchanged Paths

`/kasir/auth/*`, `/admin/*`, `/auth/*`, `/users/*`, `/business/*`, `/products/*`, `/transactions/*`, `/operator/*`

---

## Bug Fix — Silent Zero Data (dashboard + reports)

Previous `/dashboard/*` and `/reports/*` endpoints returned **zero data silently** because their SQL used `status = 'paid'`. Migration 007 defines no such status — the valid terminal status is `CONFIRMED`. This has been fixed in the consolidated `metrics` domain. All metrics endpoints now correctly count `CONFIRMED` transactions.

**Impact:** Any cached dashboard/report data showing zeros should be treated as stale. First request after deploy will return correct data.

---

## Database Table Renames (Migration 014)

File: `apps/server/api/migrations/014_rename_pos_tables.sql`

| Old Name | New Name |
|----------|----------|
| `pos_orders` | `orders` |
| `pos_order_items` | `order_items` |
| `pos_sessions` | `order_sessions` |
| `order_items.pos_order_id` (FK column) | `order_items.order_id` |

### Run Instructions

**Always run on staging first. Verify row counts before and after.**

```bash
# 1. Verify row counts before
psql $DATABASE_URL -c "SELECT COUNT(*) FROM pos_orders;"
psql $DATABASE_URL -c "SELECT COUNT(*) FROM pos_order_items;"
psql $DATABASE_URL -c "SELECT COUNT(*) FROM pos_sessions;"

# 2. Run migration (it is wrapped in BEGIN/COMMIT — safe to run)
psql $DATABASE_URL -f apps/server/api/migrations/014_rename_pos_tables.sql

# 3. Verify row counts after (same numbers expected)
psql $DATABASE_URL -c "SELECT COUNT(*) FROM orders;"
psql $DATABASE_URL -c "SELECT COUNT(*) FROM order_items;"
psql $DATABASE_URL -c "SELECT COUNT(*) FROM order_sessions;"
```

**Production:** Run during low-traffic window. The migration uses `ALTER TABLE … RENAME` which takes an `ACCESS EXCLUSIVE` lock briefly per table. No data is moved.

---

## Expected Future Change

`/metrics/insight` may be moved to a top-level `/insight` endpoint (or a separate domain) post-MVP when the AI service at `apps/server/ai/` goes live. The handler will be refactored from a DB-querying rule engine into an HTTP client to the internal AI service. **FE team: plan for this path to change again post-MVP.**

---

## Bruno Collection

`apps/server/bruno/` does not currently exist in the repo. If your team maintains a local Bruno collection, apply the path changes from the tables above.
