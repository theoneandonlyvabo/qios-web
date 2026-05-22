.PHONY: dev server client db down down-v compose

dev:
	npx concurrently --names "server,client" --prefix-colors "cyan,red" \
		"make server" \
		"make client"

server:
	npx concurrently --names "ai,api" --prefix-colors "cyan,cyan" \
		"cd app/server/ai && go run ./cmd/..." \
		"cd app/server/api && go run ./cmd/..."

client:
	npx concurrently --names "dashboard,operator" --prefix-colors "red,red" \
		"cd app/dashboard && npm run dev" \
		"cd app/operator && npm run dev"

# Jalankan PostgreSQL saja (untuk local dev sebelum `make server`).
db:
	docker compose up postgres -d

# Hentikan semua Docker service.
stop:
	docker compose down

# Hentikan dan hapus volumes (reset DB).
reset:
	docker compose down -v

# Jalankan semua via Docker Compose (production-like).
compose:
	docker compose up --build