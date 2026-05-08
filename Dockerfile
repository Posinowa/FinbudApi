# ── Build aşaması ──────────────────────────────────────────────
FROM golang:alpine AS builder

WORKDIR /app

# Bağımlılıkları önce indir (cache için)
COPY go.mod go.sum ./
RUN go mod download

# Kaynak kodunu kopyala
COPY . .

# Ana sunucu binary'sini derle
RUN go build -o server ./cmd/main.go

# Seed binary'sini derle
RUN go build -o seeder ./cmd/seed/main.go

# ── Çalışma aşaması ────────────────────────────────────────────
FROM alpine:latest

WORKDIR /app

# Binary'leri kopyala
COPY --from=builder /app/server .
COPY --from=builder /app/seeder .

# Migration dosyalarını kopyala
COPY migrations ./migrations

# Entrypoint scriptini kopyala
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

# migrate CLI'ı kur
RUN apk add --no-cache curl tar && \
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.1/migrate.linux-amd64.tar.gz \
    | tar xvz -C /usr/local/bin migrate && \
    apk del curl tar

EXPOSE 8080

ENTRYPOINT ["./entrypoint.sh"]
