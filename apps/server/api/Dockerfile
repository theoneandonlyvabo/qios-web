FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY apps/server/api/go.mod apps/server/api/go.sum ./
RUN go mod download

COPY apps/server/api/ .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/...

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /build/server .

# Merge both migration directories into one — migrate.go globs migrations/*.sql from CWD
COPY infra/database/migrations/ ./migrations/
COPY apps/server/api/migrations/ ./migrations/

EXPOSE 8080

CMD ["./server"]
