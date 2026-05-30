SHELL := /bin/bash

GO := go
NPM := npm
CLIENT_APPS := apps/client/entry apps/client/operator apps/client/owner
SERVER_API := apps/server/api

.PHONY: help dev server api client entry operator owner install install-client lint fmt gofmt vet

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
	@printf "  make fmt          Format Go sources\n\n"
	@printf "Note: Infra (PostgreSQL, docker-compose) lives in a separate repo.\n"
	@printf "Start Postgres there before running 'make server'.\n\n"

# Start backend and all clients together
dev:
	npx concurrently --names "server,client" --prefix-colors "cyan,yellow" \
		"$(MAKE) server" \
		"$(MAKE) client"

server: api

api:
	cd $(SERVER_API) && $(GO) run ./cmd/...

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