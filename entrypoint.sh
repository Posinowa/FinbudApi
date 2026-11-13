#!/bin/sh
set -e

# SSL modu: varsayılan "disable" (Cloud SQL socket / local docker), Neon için "require"
DB_SSLMODE="${DB_SSLMODE:-disable}"

# Cloud SQL Unix socket veya normal TCP bağlantısı
if echo "$DB_HOST" | grep -q "^/"; then
  # Unix socket (Cloud Run / Cloud SQL)
  DB_URL="postgresql://${DB_USER}:${DB_PASSWORD}@/${DB_NAME}?host=${DB_HOST}&sslmode=${DB_SSLMODE}"
else
  # TCP bağlantısı (local / docker)
  DB_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"
fi

echo "▶ Migration'lar çalıştırılıyor..."
migrate -path ./migrations -database "$DB_URL" up
echo "✅ Migration'lar tamam"

echo "▶ Seed kontrol ediliyor..."
./seeder -categories
echo "✅ Seed tamam"

echo "🚀 Sunucu başlatılıyor..."
exec ./server
