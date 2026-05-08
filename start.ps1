# FinbudApi başlatma scripti
# Kullanım: .\start.ps1

# .env dosyasını oku
if (Test-Path ".env") {
    Get-Content ".env" | ForEach-Object {
        if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
            [System.Environment]::SetEnvironmentVariable($matches[1].Trim(), $matches[2].Trim())
        }
    }
} else {
    Write-Error ".env dosyası bulunamadı"
    exit 1
}

$DB_USER     = [System.Environment]::GetEnvironmentVariable("DB_USER")
$DB_PASSWORD = [System.Environment]::GetEnvironmentVariable("DB_PASSWORD")
$DB_NAME     = [System.Environment]::GetEnvironmentVariable("DB_NAME")
$DB_PORT     = [System.Environment]::GetEnvironmentVariable("DB_PORT")
if (-not $DB_PORT) { $DB_PORT = "5432" }

$DB_URL = "postgresql://${DB_USER}:${DB_PASSWORD}@localhost:${DB_PORT}/${DB_NAME}?sslmode=disable"

# 1. Veritabanı container'ını başlat (volume varsa veri korunur)
Write-Host "▶ Veritabanı başlatılıyor..."
docker-compose up -d db

# 2. Veritabanı hazır olana kadar bekle
Write-Host "⏳ Veritabanı bağlantısı bekleniyor..."
$ready = $false
for ($i = 0; $i -lt 20; $i++) {
    Start-Sleep -Seconds 2
    $result = docker-compose exec -T db pg_isready -U $DB_USER -d $DB_NAME 2>&1
    if ($LASTEXITCODE -eq 0) {
        $ready = $true
        break
    }
}

if (-not $ready) {
    Write-Error "Veritabanına bağlanılamadı"
    exit 1
}
Write-Host "✅ Veritabanı hazır"

# 3. Migration'ları çalıştır (idempotent — zaten uygulanmışsa atlar)
Write-Host "▶ Migration'lar çalıştırılıyor..."
migrate -path migrations -database $DB_URL up
if ($LASTEXITCODE -ne 0) {
    Write-Error "Migration hatası"
    exit 1
}
Write-Host "✅ Migration'lar tamam"

# 4. Default kategorileri yükle (idempotent — varsa atlar, yoksa ekler)
Write-Host "▶ Seed kontrol ediliyor..."
go run cmd/seed/main.go -categories
if ($LASTEXITCODE -ne 0) {
    Write-Error "Seed hatası"
    exit 1
}
Write-Host "✅ Seed tamam"

# 5. Sunucuyu başlat
Write-Host "🚀 Sunucu başlatılıyor..."
go run cmd/main.go
