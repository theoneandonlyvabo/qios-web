# apps/server/Dockerfile

FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server ./cmd/main.go

# Production runner — image sekecil mungkin
FROM alpine:latest AS runner
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]