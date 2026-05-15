.PHONY: dev server client db down down-v compose

dev:
	npx concurrently --names "server,client" --prefix-colors "cyan,magenta" \
		"cd apps/server && go run ./cmd/..." \
		"cd apps/client && npm run dev"

server:
	cd apps/server && go run ./cmd/...

client:
	cd apps/client && npm run dev

# Jalankan PostgreSQL saja (untuk local dev sebelum `make server`).
db:
	docker compose up postgres -d

# Hentikan semua Docker service.
down:
	docker compose down

# Hentikan dan hapus volumes (reset DB).
down-v:
	docker compose down -v

# Jalankan semua via Docker Compose (production-like).
compose:
	docker compose up --build
