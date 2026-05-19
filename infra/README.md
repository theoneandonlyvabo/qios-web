# QIOS Infra Stack

Observability: Prometheus, Loki, Promtail, Grafana, Uptime Kuma.

## First-time setup

1. Copy `.env.example` to `.env` and set `APP_NAME` (default: `qios`).
2. Set the same `APP_NAME` in the root `.env` of your app stack.
3. Create the shared network (only once, host-level):

```bash
docker network create ${APP_NAME}-shared
# For QIOS with default APP_NAME:
docker network create qios-shared
```

## Start

```bash
cd infra
docker compose up -d
```

## Access

- Grafana:     http://localhost:3000  (admin/admin, change on first login)
- Prometheus:  http://localhost:9090
- Loki API:    http://localhost:3100
- Uptime Kuma: http://localhost:3001

## Notes

- App stack (`/docker-compose.yml`) must also be running for Promtail and Prometheus to capture data.
- All app containers must have label `logging: "promtail"` to be scraped.

## Using with your own app

1. Copy `infra/.env.example` to `infra/.env` and fill in your values.
2. Set the same `APP_NAME` in your root `.env`.
3. Create the shared network:
   ```bash
   docker network create ${APP_NAME}-shared
   ```
4. Add to every service in your app's `docker-compose.yml`:
   ```yaml
   labels:
     logging: "promtail"
   networks:
     - ${APP_NAME}-shared
   ```
5. Expose `/metrics` from your Go server (see Prerequisites).
6. Run `docker compose up -d` from `infra/`.
