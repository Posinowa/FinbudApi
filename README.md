# Finbud API

Go ile yazılmış kişisel finans yönetimi REST API'si. Google Cloud Run üzerinde çalışır, PostgreSQL (Cloud SQL) kullanır.

---

## Canlı API

```
https://finbud-api-197224562444.europe-west1.run.app
```

---

## Bakım Modu

Bakım modunu açmak veya kapatmak için Google Cloud CLI kullanılır.

### Bakım Modunu Aç

```bash
gcloud run services update finbud-api \
  --region=europe-west1 \
  --update-env-vars="MAINTENANCE_MODE=true"
```

### Bakım Modunu Kapat

```bash
gcloud run services update finbud-api \
  --region=europe-west1 \
  --update-env-vars="MAINTENANCE_MODE=false"
```

### Bakım Modu Durumunu Kontrol Et

```bash
curl https://finbud-api-197224562444.europe-west1.run.app/status
```

Yanıt:
- `{"maintenance": true}` → Bakım modu aktif
- `{"maintenance": false}` → Uygulama normal çalışıyor

---

## Yerel Geliştirme Ortamı

### Gereksinimler

- Go 1.21+
- Docker & Docker Compose
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI

### Başlatma

```bash
powershell -ExecutionPolicy Bypass -File .\start.ps1
```

Bu komut sırasıyla şunları yapar:
1. PostgreSQL container'ını başlatır
2. Migration'ları çalıştırır
3. Seed verilerini yükler
4. API sunucusunu başlatır

### Docker Compose ile Tüm Servisleri Başlat

```bash
docker-compose up
```

---

## Deployment

### Yeni Versiyon Deploy Et

```bash
docker build -t europe-west1-docker.pkg.dev/finbud-app-2026/finbud-repo/api:latest .
docker push europe-west1-docker.pkg.dev/finbud-app-2026/finbud-repo/api:latest
gcloud run deploy finbud-api \
  --image=europe-west1-docker.pkg.dev/finbud-app-2026/finbud-repo/api:latest \
  --region=europe-west1
```

---

## Sağlık Kontrolü

```bash
curl https://finbud-api-197224562444.europe-west1.run.app/health
```
