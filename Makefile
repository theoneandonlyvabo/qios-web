SHELL := /bin/bash

GO := go
NPM := npm
CLIENT_APPS := apps/client/entry apps/client/operator apps/client/owner
SERVER_API := apps/server/api

# Local Postgres (no compose file — infra repo is separate)
PG_CONTAINER := qios-postgres
PG_IMAGE := postgres:16-alpine
PG_PORT := 5432
PG_USER := postgres
PG_PASSWORD := postgres
PG_DB := qios

.PHONY: help dev server api client entry operator owner install install-client lint fmt gofmt vet db db-stop db-rm db-reset db-logs db-psql seed-dev

help:
	@printf "\nUsage:\n"
	@printf "  make dev          Start backend and all client apps in parallel\n"
	@printf "  make server       Start the API server\n"
	@printf "  make api          Start the API server (alias of make server)\n"
	@printf "  make client       Start all client apps (entry, operator, owner)\n"
	@printf "  make entry        Start entry client (port 3000)\n"
	@printf "  make operator     Start operator client (port 3001)\n"
	@printf "  make owner        Start owner client (port 3002)\n"
	@printf "  make install      Install client dependencies\n"
	@printf "  make lint         Run client lint and Go vet\n"
	@printf "  make vet          Run Go vet on the API server\n"
	@printf "  make fmt          Format Go sources\n"
	@printf "  make seed-dev     Seed dev owner+business+operator (prints login credentials)\n"
	@printf "  make db           Start local Postgres in Docker (creates container if missing)\n"
	@printf "  make db-stop      Stop the Postgres container\n"
	@printf "  make db-rm        Stop and remove the Postgres container (keeps volume)\n"
	@printf "  make db-reset     Remove container + volume (DESTRUCTIVE — wipes data)\n"
	@printf "  make db-logs      Follow Postgres logs\n"
	@printf "  make db-psql      Open a psql shell inside the container\n\n"

# Start backend and all clients together
dev:
	npx concurrently --names "server,client" --prefix-colors "cyan,yellow" \
		"$(MAKE) server" \
		"$(MAKE) client"

server: api

api:
	cd $(SERVER_API) && $(GO) run ./cmd

# Seed dev owner + business + operator with known credentials.
# Idempotent — safe to re-run.
seed-dev:
	cd $(SERVER_API) && $(GO) run ./cmd/seed-dev

client:
	npx concurrently --names "entry,operator,owner" --prefix-colors "yellow,magenta,green" \
		"cd apps/client/entry && $(NPM) run dev" \
		"cd apps/client/operator && $(NPM) run dev" \
		"cd apps/client/owner && $(NPM) run dev"

entry:
	cd apps/client/entry && $(NPM) run dev

operator:
	cd apps/client/operator && $(NPM) run dev

owner:
	cd apps/client/owner && $(NPM) run dev

install: install-client

install-client:
	@for dir in $(CLIENT_APPS); do \
		printf "Installing dependencies in $$dir...\n"; \
		(cd $$dir && $(NPM) install); \
	done

lint:
	@printf "Linting client apps...\n"
	@for dir in $(CLIENT_APPS); do \
		printf "- $$dir\n"; \
		(cd $$dir && $(NPM) run lint); \
	done
	@printf "\nRunning Go vet...\n"
	cd $(SERVER_API) && $(GO) vet ./...

vet:
	cd $(SERVER_API) && $(GO) vet ./...

fmt: gofmt

gofmt:
	find $(SERVER_API) -name '*.go' | sort | xargs $(GO)fmt -w

# --- Local Postgres (Docker) ---------------------------------------------------
# Starts (or resumes) a single Postgres container named $(PG_CONTAINER).
# Credentials must match apps/server/api/.env.
db:
	@if [ -z "$$(docker ps -aq -f name=^/$(PG_CONTAINER)$$)" ]; then \
		printf "Creating $(PG_CONTAINER) ($(PG_IMAGE)) on port $(PG_PORT)...\n"; \
		docker run -d \
			--name $(PG_CONTAINER) \
			-e POSTGRES_USER=$(PG_USER) \
			-e POSTGRES_PASSWORD=$(PG_PASSWORD) \
			-e POSTGRES_DB=$(PG_DB) \
			-p $(PG_PORT):5432 \
			-v $(PG_CONTAINER)-data:/var/lib/postgresql/data \
			$(PG_IMAGE); \
	else \
		printf "Starting existing $(PG_CONTAINER)...\n"; \
		docker start $(PG_CONTAINER); \
	fi
	@printf "Waiting for Postgres to accept connections..."
	@until docker exec $(PG_CONTAINER) pg_isready -U $(PG_USER) -d $(PG_DB) >/dev/null 2>&1; do \
		printf "."; sleep 1; \
	done; \
	printf " ready.\n"

db-stop:
	-docker stop $(PG_CONTAINER)

db-rm: db-stop
	-docker rm $(PG_CONTAINER)

db-reset: db-rm
	-docker volume rm $(PG_CONTAINER)-data

db-logs:
	docker logs -f $(PG_CONTAINER)

db-psql:
	docker exec -it $(PG_CONTAINER) psql -U $(PG_USER) -d $(PG_DB)