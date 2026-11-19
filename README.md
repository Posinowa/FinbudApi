# Finbud API

Go ile yazılmış kişisel finans yönetimi REST API'si. Google Cloud Run üzerinde çalışır, veritabanı olarak Neon (PostgreSQL 18) kullanır.

---

## Canlı API

```
https://finbud-api-197224562444.europe-west1.run.app
```

---

## Veritabanı (Neon)

| Ayar | Değer |
|---|---|
| Sağlayıcı | [Neon](https://neon.com) (ücretsiz plan) |
| Proje / veritabanı | `finbud` / `finbud` |
| Bölge | AWS eu-central-1 (Frankfurt) |
| PostgreSQL sürümü | 18 |

Cloud Run'daki bağlantı ayarları:

| Ortam değişkeni | Kaynak |
|---|---|
| `DB_HOST` | Neon'un doğrudan (pooler olmayan) adresi |
| `DB_SSLMODE` | `require` |
| `DB_USER` | Secret Manager → `NEON_DB_USER` |
| `DB_PASSWORD` | Secret Manager → `NEON_DB_PASSWORD` |
| `DB_NAME` | Secret Manager → `NEON_DB_NAME` |

Notlar:
- Neon, 5 dakika kullanılmayınca uykuya geçer. Uzun bir aradan sonraki ilk istek birkaç yüz milisaniye gecikebilir.
- Uygulama boştaki bağlantıları 2 dakika sonra kapatır (`pkg/database/database.go`). Bu sayede Neon uykuya geçebilir ve ücretsiz kota korunur.
- Yerel geliştirmede `DB_SSLMODE` boş bırakılır, varsayılan olarak `disable` kullanılır.

### Yedek Alma

Neon'un ücretsiz planı yalnızca son 6 saate geri dönmeye izin verir. Düzenli olarak (örneğin ayda bir) Cloud Shell'de yedek alınmalıdır:

```bash
# Cloud Shell'deki pg_dump 16 sürümünde; Neon 18 olduğu için önce 18 sürümünü kur
# (Cloud Shell her oturumda sıfırlandığı için her yedekten önce gerekir)
sudo apt-get install -y postgresql-client-18

export PGPASSWORD="$(gcloud secrets versions access latest --secret=NEON_DB_PASSWORD)"
/usr/lib/postgresql/18/bin/pg_dump "host=<NEON_HOST> user=finbud_user dbname=finbud sslmode=require" -Fc -f finbud-yedek-$(date +%Y%m%d).dump
unset PGPASSWORD
```

Geri yüklemek için de aynı sürümdeki `/usr/lib/postgresql/18/bin/pg_restore` kullanılır.

`<NEON_HOST>` yerine Neon panelindeki host adresi yazılır. Yedek dosyaları repoya eklenmez (`.gitignore` içinde `*.dump`).

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

- Go 1.26+
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

Alternatif olarak, image'ı Cloud Build'e derletmek için:

```bash
git clone https://github.com/Posinowa/FinbudApi.git finbud-deploy
cd finbud-deploy
gcloud run deploy finbud-api --source . --region europe-west1
```

Notlar:
- Mevcut ortam değişkenleri ve secret'lar yeni revizyona aynen taşınır.
- Önce test etmek için `--no-traffic --tag test` eklenir, test adresinde kontrol edildikten sonra `gcloud run services update-traffic finbud-api --region europe-west1 --to-latest` ile trafik verilir.
- `.gitattributes` ve Dockerfile, `entrypoint.sh`'ın Windows'ta CRLF'ye dönüşüp container'ı bozmasını engeller.
---

## Sağlık Kontrolü

```bash
curl https://finbud-api-197224562444.europe-west1.run.app/health
```
