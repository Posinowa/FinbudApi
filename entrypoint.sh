#!/bin/sh
set -e

DB_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

echo "▶ Migration'lar çalıştırılıyor..."
migrate -path ./migrations -database "$DB_URL" up
echo "✅ Migration'lar tamam"

echo "▶ Seed kontrol ediliyor..."
./seeder -categories
echo "✅ Seed tamam"

echo "🚀 Sunucu başlatılıyor..."
exec ./server
