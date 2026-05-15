.PHONY: qios server client

qios:
	npx concurrently --names "server,client" --prefix-colors "cyan,magenta" \
		"cd apps/server && go run ./cmd/..." \
		"cd apps/client && npm run dev"

server:
	cd apps/server && go run ./cmd/...

client:
	cd apps/client && npm run dev