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

## Güncelleme Bildirimi

Yeni bir uygulama sürümü yayınlandığında, kullanıcılara dashboard'da güncelleme banner'ı gösterilir. Kullanıcı banner'dan doğrudan ilgili store'a yönlendirilebilir.

### Güncelleme Bildirimini Aç

```bash
gcloud run services update finbud-api \
  --region=europe-west1 \
  --update-env-vars="UPDATE_AVAILABLE=true"
```

### Güncelleme Bildirimini Kapat

```bash
gcloud run services update finbud-api \
  --region=europe-west1 \
  --update-env-vars="UPDATE_AVAILABLE=false"
```

### Store URL'lerini Ayarla

Android (Play Store) URL'si:

```bash
gcloud run services update finbud-api \
  --region=europe-west1 \
  --update-env-vars="ANDROID_STORE_URL=https://play.google.com/store/apps/details?id=com.finbud.finbud_app"
```

iOS (App Store) URL'si — App Store'a yüklendikten sonra uygulama ID'si ile ayarla:

```bash
gcloud run services update finbud-api \
  --region=europe-west1 \
  --update-env-vars="IOS_STORE_URL=https://apps.apple.com/app/idXXXXXXXXX"
```

Tüm parametreleri tek seferde ayarlamak için:

```bash
gcloud run services update finbud-api \
  --region=europe-west1 \
  --update-env-vars="UPDATE_AVAILABLE=true,ANDROID_STORE_URL=https://play.google.com/store/apps/details?id=com.finbud.finbud_app,IOS_STORE_URL=https://apps.apple.com/app/idXXXXXXXXX"
```

### Durum Kontrolü

```bash
curl https://finbud-api-197224562444.europe-west1.run.app/status
```

Örnek yanıt:
```json
{
  "maintenance": false,
  "update_available": true,
  "android_store_url": "https://play.google.com/store/apps/details?id=com.finbud.finbud_app",
  "ios_store_url": "https://apps.apple.com/app/idXXXXXXXXX"
}
```

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
