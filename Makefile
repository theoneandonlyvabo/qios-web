SHELL := /bin/bash

GO := go
NPM := npm
DC := docker compose
COMPOSE_FILE := infra/docker-compose.yml
CLIENT_APPS := apps/client/admin apps/client/operator apps/client/owner
SERVER_AI := apps/server/ai
SERVER_API := apps/server/api

.PHONY: help dev server api ai client admin operator owner install install-client lint fmt gofmt db down down-v compose logs

help:
	@printf "\nUsage:\n"
	@printf "  make dev          Start backend and frontend in parallel\n"
	@printf "  make server       Start API and AI backend in parallel\n"
	@printf "  make api          Start only the API server\n"
	@printf "  make ai           Start only the AI backend\n"
	@printf "  make client       Start all client apps (admin, operator, owner)\n"
	@printf "  make admin        Start admin client\n"
	@printf "  make operator     Start operator client\n"
	@printf "  make owner        Start owner client\n"
	@printf "  make install      Install client dependencies\n"
	@printf "  make lint         Run client lint and Go vet\n"
	@printf "  make fmt          Format Go sources\n"
	@printf "  make db           Start PostgreSQL via Docker Compose\n"
	@printf "  make down         Stop Docker Compose services\n"
	@printf "  make down-v       Stop Docker Compose and remove volumes\n"
	@printf "  make compose      Build and start Docker Compose\n"
	@printf "  make logs         Follow Docker Compose logs\n\n"

# Start backend and frontend together
dev:
	npx concurrently --names "server,client" --prefix-colors "cyan,yellow" \
		"make server" \
		"make client"

server:
	npx concurrently --names "ai,api" --prefix-colors "cyan,cyan" \
		"cd $(SERVER_AI) && $(GO) run ./cmd/..." \
		"cd $(SERVER_API) && $(GO) run ./cmd/..."

api:
	cd $(SERVER_API) && $(GO) run ./cmd/...

ai:
	cd $(SERVER_AI) && $(GO) run ./cmd/...

client:
	npx concurrently --names "admin,operator,owner" --prefix-colors "yellow,yellow,yellow" \
		"cd apps/client/admin && $(NPM) run dev" \
		"cd apps/client/operator && $(NPM) run dev" \
		"cd apps/client/owner && $(NPM) run dev"

admin:
	cd apps/client/admin && $(NPM) run dev

operator:
	cd apps/client/operator && $(NPM) run dev

owner:
	cd apps/client/owner && $(NPM) run dev

install: install-client

install-client:
	@for dir in $(CLIENT_APPS); do \
		printf "Installing dependencies in $$dir...\n"; \
		cd $$dir && $(NPM) install; \
	done

lint:
	@printf "Linting client apps...\n"
	@for dir in $(CLIENT_APPS); do \
		printf "- $$dir\n"; \
		cd $$dir && $(NPM) run lint; \
	done
	@printf "\nRunning Go vet...\n"
	cd apps/server && $(GO) vet ./...

fmt: gofmt

gofmt:
	find apps/server -name '*.go' | sort | xargs $(GO)fmt -w

# Docker Compose helpers
# `docker compose -f infra/docker-compose.yml` is used because root docker-compose.yml is intentionally empty.
db:
	$(DC) -f $(COMPOSE_FILE) up postgres -d

down:
	$(DC) -f $(COMPOSE_FILE) down

down-v:
	$(DC) -f $(COMPOSE_FILE) down -v

compose:
	$(DC) -f $(COMPOSE_FILE) up --build

logs:
	$(DC) -f $(COMPOSE_FILE) logs -f