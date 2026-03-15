# 📱 Kişisel Finans Uygulaması — API Dokümantasyonu

Bu döküman, Flutter tabanlı kişisel finans uygulamasının Go backend API'si için öngörülen endpoint'leri, request/response field'larını ve veri modellerini açıklar.

---

## 📋 İçindekiler

- [Genel Bilgiler](#genel-bilgiler)
- [Authentication](#authentication)
- [Veri Modelleri](#veri-modelleri)
- [Endpoint'ler](#endpointler)
  - [Auth](#-auth)
  - [User / Profil](#-user--profil)
  - [Transactions](#-transactions)
  - [Budgets](#-budgets)
  - [Categories](#-categories)
  - [Dashboard](#-dashboard)
- [HTTP Durum Kodları](#http-durum-kodları)
- [Hata Formatı](#hata-formatı)

---

## Genel Bilgiler

| Alan | Değer |
|------|-------|
| Base URL (Local) | `http://localhost:8080/api/v1` |
| Base URL (Production) | `https://api.finance.app/v1` |
| İçerik Tipi | `application/json` |
| API Versiyonu | `1.0.0` |
| Dokümantasyon | `localhost:8080/docs` (Scalar UI) |

---

## Authentication

Korumalı endpoint'ler **Bearer JWT token** gerektirir.

`/auth/login` endpoint'inden alınan `access_token` her istekte header'a eklenir:

```
Authorization: Bearer <access_token>
```

- 🔓 **Public** — Token gerekmez: `POST /auth/register`, `POST /auth/login`, `POST /auth/refresh`
- 🔒 **Protected** — Token zorunludur: Diğer tüm endpoint'ler

Token süresi dolduğunda `POST /auth/refresh` ile yenilenebilir.

---

## Veri Modelleri

### User

Kullanıcı hesap bilgilerini temsil eder.

| Field | Tip | Zorunlu | Açıklama | Örnek |
|-------|-----|---------|----------|-------|
| `id` | integer | ✅ | Benzersiz kullanıcı ID'si | `42` |
| `name` | string | ✅ | Kullanıcının tam adı | `"Ali Yılmaz"` |
| `email` | string (email) | ✅ | Giriş için kullanılan e-posta | `"ali@example.com"` |
| `currency` | string | ❌ | Tercih edilen para birimi (ISO 4217) | `"TRY"` |
| `created_at` | string (date-time) | ✅ | Hesap oluşturulma zamanı | `"2026-01-15T10:00:00Z"` |
| `updated_at` | string (date-time) | ✅ | Son güncelleme zamanı | `"2026-03-01T08:30:00Z"` |

> `password_hash` alanı hiçbir zaman response'ta dönmez.

---

### Transaction

Bir gelir veya gider işlemini temsil eder.

| Field | Tip | Zorunlu | Açıklama | Örnek |
|-------|-----|---------|----------|-------|
| `id` | integer | ✅ | İşlem ID'si | `101` |
| `user_id` | integer | ✅ | Hangi kullanıcıya ait | `42` |
| `amount` | number (double) | ✅ | İşlem tutarı — her zaman **pozitif**, yön `type` ile belirlenir | `250.00` |
| `type` | enum | ✅ | `income` veya `expense` | `"expense"` |
| `category_id` | integer | ✅ | Bağlı kategori ID'si | `3` |
| `category` | Category object | ✅ | Kategori detayı (nested) | `{...}` |
| `description` | string | ❌ | Kullanıcı notu | `"Market alışverişi"` |
| `date` | string (date) | ✅ | İşlem tarihi (`YYYY-MM-DD`) | `"2026-03-10"` |
| `created_at` | string (date-time) | ✅ | Kayıt zamanı | `"2026-03-10T14:22:00Z"` |

---

### Budget

Kategori bazlı aylık harcama limitini temsil eder.

| Field | Tip | Zorunlu | Açıklama | Örnek |
|-------|-----|---------|----------|-------|
| `id` | integer | ✅ | Bütçe ID'si | `7` |
| `user_id` | integer | ✅ | Hangi kullanıcıya ait | `42` |
| `category_id` | integer | ✅ | Hangi kategori için | `3` |
| `category` | Category object | ✅ | Kategori detayı (nested) | `{...}` |
| `limit` | number (double) | ✅ | Aylık harcama limiti | `1500.00` |
| `spent` | number (double) | ✅ (hesaplanır) | O ay harcanan tutar | `870.00` |
| `remaining` | number (double) | ✅ (hesaplanır) | Kalan bütçe (`limit - spent`) | `630.00` |
| `percent_used` | number (double) | ✅ (hesaplanır) | Kullanım yüzdesi (0–100) | `58.0` |
| `month` | string (`YYYY-MM`) | ✅ | Geçerli olduğu ay | `"2026-03"` |
| `created_at` | string (date-time) | ✅ | Oluşturulma zamanı | `"2026-03-01T08:00:00Z"` |

---

### Category

İşlem kategorisini temsil eder.

| Field | Tip | Zorunlu | Açıklama | Örnek |
|-------|-----|---------|----------|-------|
| `id` | integer | ✅ | Kategori ID'si | `3` |
| `name` | string | ✅ | Kategori adı | `"Market"` |
| `icon` | string | ❌ | Flutter icon adı veya emoji | `"shopping_cart"` |
| `type` | enum | ✅ | `income` veya `expense` | `"expense"` |
| `is_default` | boolean | ✅ | Sistem varsayılan kategorisi mi? | `true` |

---

### AuthResponse

`/auth/login` ve `/auth/refresh` endpoint'lerinin response'u.

| Field | Tip | Açıklama | Örnek |
|-------|-----|----------|-------|
| `access_token` | string | Korumalı endpoint'lerde kullanılacak JWT | `"eyJhbGci..."` |
| `refresh_token` | string | Access token yenilemek için kullanılır | `"eyJhbGci..."` |
| `token_type` | string | Token türü | `"Bearer"` |
| `expires_in` | integer | Saniye cinsinden geçerlilik süresi | `3600` |

---

### DashboardSummary

Ana sayfa özet verisini temsil eder.

| Field | Tip | Açıklama | Örnek |
|-------|-----|----------|-------|
| `month` | string | Özet ayı | `"2026-03"` |
| `balance` | number | seçili ay için net bakiye | `4250.00` |
| `total_income` | number | Seçili ayın toplam geliri | `8500.00` |
| `total_expense` | number | Seçili ayın toplam gideri | `4250.00` |
| `budget_summary` | array | seçili aya ait bütçe özetleri | `[...]` |
| `recent_transactions` | array | son işlemler | `[...]` |

`budget_summary` her elemanı:

| Field | Tip | Açıklama |
|-------|-----|----------|
| `category_id` | integer | Kategori ID'si |
| `category_name` | string | Kategori adı |
| `category_icon` | string | İkon |
| `limit` | number | Aylık limit |
| `spent` | number | Harcanan tutar |
| `percent_used` | number | Kullanım yüzdesi |

---

## Endpoint'ler

---

### 🔐 Auth

#### `POST /auth/register` — Yeni kullanıcı kaydı

🔓 Public

**Request Body:**

```json
{
  "name": "Ali Yılmaz",
  "email": "ali@example.com",
  "password": "Gizli123!"
}
```

| Field | Tip | Zorunlu | Kural |
|-------|-----|---------|-------|
| `name` | string | ✅ | Min 2, max 100 karakter |
| `email` | string | ✅ | Geçerli e-posta formatı |
| `password` | string | ✅ | Min 8 karakter |

**Response `201 Created`:**

```json
{
  "message": "Kayıt başarılı",
  "user_id": 42
}
```

**Olası Hatalar:** `400 Bad Request`, `409 Conflict` (e-posta zaten kayıtlı)

---

#### `POST /auth/login` — Kullanıcı girişi

🔓 Public

**Request Body:**

```json
{
  "email": "ali@example.com",
  "password": "Gizli123!"
}
```

| Field | Tip | Zorunlu |
|-------|-----|---------|
| `email` | string | ✅ |
| `password` | string | ✅ |

**Response `200 OK`:**

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

**Olası Hatalar:** `400 Bad Request`, `401 Unauthorized`

---

#### `POST /auth/refresh` — Token yenile

🔓 Public

**Request Body:**

```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response `200 OK`:** → `AuthResponse` (yukarıdaki login ile aynı yapı)

**Olası Hatalar:** `401 Unauthorized`

---

#### `POST /auth/logout` — Oturumu sonlandır

🔒 Protected

**Request Body:** Yok

**Response `200 OK`:**

```json
{
  "message": "Çıkış başarılı"
}
```

---

### 👤 User / Profil

#### `GET /users/me` — Profil bilgilerini getir

🔒 Protected

**Request:** Yok (token'dan kullanıcı belirlenir)

**Response `200 OK`:**

```json
{
  "id": 42,
  "name": "Ali Yılmaz",
  "email": "ali@example.com",
  "currency": "TRY",
  "created_at": "2026-01-15T10:00:00Z",
  "updated_at": "2026-03-01T08:30:00Z"
}
```

---

#### `PUT /users/me` — Profil güncelle

🔒 Protected

**Request Body:**

```json
{
  "name": "Ali Yılmaz",
  "currency": "USD"
}
```

| Field | Tip | Zorunlu | Açıklama |
|-------|-----|---------|----------|
| `name` | string | ❌ | Min 2, max 100 karakter |
| `currency` | string | ❌ | ISO 4217 para birimi kodu |

**Response `200 OK`:** → `User` objesi

---

#### `PUT /users/me/password` — Şifre değiştir

🔒 Protected

**Request Body:**

```json
{
  "old_password": "EskiSifre123!",
  "new_password": "YeniSifre456!"
}
```

| Field | Tip | Zorunlu | Açıklama |
|-------|-----|---------|----------|
| `old_password` | string | ✅ | Mevcut şifre |
| `new_password` | string | ✅ | Min 8 karakter |

**Response `200 OK`:**

```json
{
  "message": "Şifre başarıyla değiştirildi"
}
```

---

#### `DELETE /users/me` — Hesabı sil

🔒 Protected

**Response `200 OK`:**

```json
{
  "message": "Hesap silindi"
}
```

---

### 💸 Transactions

#### `GET /transactions` — İşlem listesi

🔒 Protected

**Query Parametreleri:**

| Parametre | Tip | Zorunlu | Açıklama | Örnek |
|-----------|-----|---------|----------|-------|
| `type` | string | ❌ | `income` veya `expense` | `?type=expense` |
| `category_id` | integer | ❌ | Kategori filtresi | `?category_id=3` |
| `month` | string | ❌ | `YYYY-MM` formatında | `?month=2026-03` |
| `page` | integer | ❌ | Sayfa numarası (varsayılan: `1`) | `?page=2` |
| `limit` | integer | ❌ | Sayfa başına kayıt (varsayılan: `20`) | `?limit=10` |

**Response `200 OK`:**

```json
{
  "data": [
    {
      "id": 101,
      "user_id": 42,
      "amount": 250.00,
      "type": "expense",
      "category_id": 3,
      "category": {
        "id": 3,
        "name": "Market",
        "icon": "shopping_cart",
        "type": "expense",
        "is_default": true
      },
      "description": "Market alışverişi",
      "date": "2026-03-10",
      "created_at": "2026-03-10T14:22:00Z"
    }
  ],
  "meta": {
    "total": 87,
    "page": 1,
    "limit": 20,
    "total_pages": 5
  }
}
```

---

#### `POST /transactions` — Yeni işlem oluştur

🔒 Protected

**Request Body:**

```json
{
  "amount": 250.00,
  "type": "expense",
  "category_id": 3,
  "description": "Market alışverişi",
  "date": "2026-03-10"
}
```

| Field | Tip | Zorunlu | Açıklama |
|-------|-----|---------|----------|
| `amount` | number | ✅ | Min `0.01` — her zaman pozitif |
| `type` | enum | ✅ | `income` veya `expense` |
| `category_id` | integer | ✅ | Geçerli bir kategori ID'si |
| `description` | string | ❌ | Max 255 karakter |
| `date` | string (date) | ✅ | `YYYY-MM-DD` formatı |

**Response `201 Created`:** → `Transaction` objesi

---

#### `GET /transactions/{id}` — İşlem detayı

🔒 Protected

**Path Parametresi:** `id` — İşlem ID'si

**Response `200 OK`:** → `Transaction` objesi

**Olası Hatalar:** `404 Not Found`

---

#### `PUT /transactions/{id}` — İşlem güncelle

🔒 Protected

**Path Parametresi:** `id` — İşlem ID'si

**Request Body:** (Tüm field'lar opsiyonel — sadece değiştirilecek olanlar gönderilir)

```json
{
  "amount": 300.00,
  "description": "Market ve fırın"
}
```

| Field | Tip | Zorunlu | Açıklama |
|-------|-----|---------|----------|
| `amount` | number | ❌ | Min `0.01` |
| `category_id` | integer | ❌ | — |
| `description` | string | ❌ | Max 255 karakter |
| `date` | string (date) | ❌ | `YYYY-MM-DD` formatı |

**Response `200 OK`:** → güncellenmiş `Transaction` objesi

---

#### `DELETE /transactions/{id}` — İşlem sil

🔒 Protected

**Path Parametresi:** `id` — İşlem ID'si

**Response `204 No Content`**

**Olası Hatalar:** `404 Not Found`

---

### 💰 Budgets

#### `GET /budgets` — Bütçe listesi

🔒 Protected

**Query Parametreleri:**

| Parametre | Tip | Zorunlu | Açıklama | Örnek |
|-----------|-----|---------|----------|-------|
| `month` | string | ❌ | `YYYY-MM` — boş bırakılırsa mevcut ay | `?month=2026-03` |

**Response `200 OK`:**

```json
{
  "month": "2026-03",
  "data": [
    {
      "id": 7,
      "user_id": 42,
      "category_id": 3,
      "category": {
        "id": 3,
        "name": "Market",
        "icon": "shopping_cart",
        "type": "expense",
        "is_default": true
      },
      "limit": 1500.00,
      "spent": 870.00,
      "remaining": 630.00,
      "percent_used": 58.0,
      "month": "2026-03",
      "created_at": "2026-03-01T08:00:00Z"
    }
  ]
}
```

---

#### `POST /budgets` — Yeni bütçe oluştur

🔒 Protected

**Request Body:**

```json
{
  "category_id": 3,
  "limit": 1500.00,
  "month": "2026-03"
}
```

| Field | Tip | Zorunlu | Açıklama |
|-------|-----|---------|----------|
| `category_id` | integer | ✅ | Geçerli bir kategori ID'si |
| `limit` | number | ✅ | Min `1.00` |
| `month` | string | ✅ | `YYYY-MM` formatı |

**Response `201 Created`:** → `Budget` objesi

**Olası Hatalar:** `409 Conflict` (aynı kategori + ay için bütçe zaten var)

---

#### `PUT /budgets/{id}` — Bütçe güncelle

🔒 Protected

**Path Parametresi:** `id` — Bütçe ID'si

**Request Body:**

```json
{
  "limit": 2000.00
}
```

| Field | Tip | Zorunlu | Açıklama |
|-------|-----|---------|----------|
| `limit` | number | ✅ | Min `1.00` |

**Response `200 OK`:** → güncellenmiş `Budget` objesi

---

#### `DELETE /budgets/{id}` — Bütçe sil

🔒 Protected

**Path Parametresi:** `id` — Bütçe ID'si

**Response `204 No Content`**

---

### 🏷️ Categories

#### `GET /categories` — Kategori listesi

🔒 Protected

**Query Parametreleri:**

| Parametre | Tip | Zorunlu | Açıklama |
|-----------|-----|---------|----------|
| `type` | string | ❌ | `income` veya `expense` |

**Response `200 OK`:**

```json
{
  "data": [
    {
      "id": 1,
      "name": "Maaş",
      "icon": "payments",
      "type": "income",
      "is_default": true
    },
    {
      "id": 3,
      "name": "Market",
      "icon": "shopping_cart",
      "type": "expense",
      "is_default": true
    }
  ]
}
```

---

#### `POST /categories` — Özel kategori oluştur

🔒 Protected

**Request Body:**

```json
{
  "name": "Spor",
  "icon": "fitness_center",
  "type": "expense"
}
```

| Field | Tip | Zorunlu | Açıklama |
|-------|-----|---------|----------|
| `name` | string | ✅ | Max 50 karakter |
| `icon` | string | ❌ | Flutter icon adı veya emoji |
| `type` | enum | ✅ | `income` veya `expense` |

**Response `201 Created`:** → `Category` objesi

---

#### `PUT /categories/{id}` — Kategori güncelle

🔒 Protected

**Path Parametresi:** `id` — Kategori ID'si

**Request Body:**

```json
{
  "name": "Spor & Sağlık",
  "icon": "sports_gymnastics"
}
```

| Field | Tip | Zorunlu |
|-------|-----|---------|
| `name` | string | ❌ |
| `icon` | string | ❌ |

**Response `200 OK`:** → güncellenmiş `Category` objesi

---

#### `DELETE /categories/{id}` — Kategori sil

🔒 Protected

**Path Parametresi:** `id` — Kategori ID'si

**Response `204 No Content`**

> ⚠️ Varsayılan (`is_default: true`) kategoriler silinemez.

---

### 📊 Dashboard

#### `GET /dashboard/summary` — Ana sayfa özet verileri

🔒 Protected

**Query Parametreleri:**

| Parametre | Tip | Zorunlu | Açıklama |
|-----------|-----|---------|----------|
| `month` | string | ❌ | `YYYY-MM` — boş bırakılırsa mevcut ay |

**Response `200 OK`:**

```json
{
  "month": "2026-03",
  "balance": 4250.00,
  "total_income": 8500.00,
  "total_expense": 4250.00,
  "budget_summary": [
    {
      "category_id": 3,
      "category_name": "Market",
      "category_icon": "shopping_cart",
      "limit": 1500.00,
      "spent": 870.00,
      "percent_used": 58.0
    }
  ],
  "recent_transactions": [
    {
      "id": 101,
      "amount": 250.00,
      "type": "expense",
      "category": { "id": 3, "name": "Market", "icon": "shopping_cart", "type": "expense", "is_default": true },
      "description": "Market alışverişi",
      "date": "2026-03-10",
      "created_at": "2026-03-10T14:22:00Z"
    }
  ]
}
```

---

## HTTP Durum Kodları

| Kod | Anlam | Ne zaman döner? |
|-----|-------|-----------------|
| `200 OK` | Başarılı | GET ve PUT istekleri başarılı olduğunda |
| `201 Created` | Oluşturuldu | POST ile yeni kayıt oluşturulduğunda |
| `204 No Content` | İçerik yok | DELETE başarılı olduğunda |
| `400 Bad Request` | Geçersiz istek | Eksik veya hatalı field gönderildiğinde |
| `401 Unauthorized` | Yetkisiz | Token yoksa, geçersizse veya süresi dolmuşsa |
| `404 Not Found` | Bulunamadı | İstenen kayıt mevcut değilse |
| `409 Conflict` | Çakışma | Kayıt zaten mevcutsa (örn: aynı e-posta) |

---

## Hata Formatı

Tüm hatalar aynı JSON yapısında döner:

```json
{
  "error": "validation_error",
  "message": "Geçersiz e-posta formatı",
  "details": {
    "email": "Geçerli bir e-posta adresi girin",
    "password": "Şifre en az 8 karakter olmalıdır"
  }
}
```

| Field | Tip | Açıklama |
|-------|-----|----------|
| `error` | string | Makine tarafından okunabilir hata kodu |
| `message` | string | İnsan tarafından okunabilir açıklama |
| `details` | object | Field bazlı hata detayları (sadece validation hatalarında) |

---

## 📁 İlgili Dosyalar

| Dosya | Açıklama |
|-------|----------|
| [`openapi.yaml`](./openapi.yaml) | Tam OpenAPI 3.0.3 spesifikasyonu |

---

*Son güncelleme: Mart 2026 — v1.0.0*
