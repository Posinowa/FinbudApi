package seed

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// Seeder veritabanına seed işlemi yapar
type Seeder struct {
	db *sqlx.DB
}

// NewSeeder yeni bir Seeder oluşturur
func NewSeeder(db *sqlx.DB) *Seeder {
	return &Seeder{db: db}
}

// SeedAll tüm test verilerini ekler (idempotent)
func (s *Seeder) SeedAll(ctx context.Context) error {
	log.Println("📦 Seed işlemi başlatılıyor...")

	// 1. Default kategorileri kontrol et (migration'da zaten var, bu extra güvenlik)
	if err := s.ensureDefaultCategories(ctx); err != nil {
		return fmt.Errorf("default categories hatası: %w", err)
	}

	// 2. Test kullanıcısı oluştur
	userID, err := s.seedTestUser(ctx)
	if err != nil {
		return fmt.Errorf("test user hatası: %w", err)
	}

	// 3. Test kategorileri (kullanıcıya özel)
	if err := s.seedTestCategories(ctx, userID); err != nil {
		return fmt.Errorf("test categories hatası: %w", err)
	}

	// 4. Test transaction'ları
	if err := s.seedTestTransactions(ctx, userID); err != nil {
		return fmt.Errorf("test transactions hatası: %w", err)
	}

	// 5. Test budget'ları
	if err := s.seedTestBudgets(ctx, userID); err != nil {
		return fmt.Errorf("test budgets hatası: %w", err)
	}

	log.Println("✅ Seed işlemi tamamlandı!")
	return nil
}

// SeedDefaultCategoriesOnly sadece default kategorileri ekler
func (s *Seeder) SeedDefaultCategoriesOnly(ctx context.Context) error {
	log.Println("📦 Default kategoriler kontrol ediliyor...")
	if err := s.ensureDefaultCategories(ctx); err != nil {
		return fmt.Errorf("default categories hatası: %w", err)
	}
	log.Println("✅ Default kategoriler hazır!")
	return nil
}

// ensureDefaultCategories migration'da eklenen kategorilerin varlığını doğrular
// Eğer yoksa ekler (idempotent)
func (s *Seeder) ensureDefaultCategories(ctx context.Context) error {
	for _, cat := range DefaultCategories {
		query := `
			INSERT INTO categories (name, icon, type, is_default, user_id)
			VALUES ($1, $2, $3, true, NULL)
			ON CONFLICT DO NOTHING
		`
		_, err := s.db.ExecContext(ctx, query, cat.Name, cat.Icon, cat.Type)
		if err != nil {
			return fmt.Errorf("kategori eklenemedi (%s): %w", cat.Name, err)
		}
	}

	// Kaç kategori var kontrol et
	var count int
	err := s.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM categories WHERE is_default = true")
	if err != nil {
		return err
	}
	log.Printf("   ✓ %d default kategori mevcut", count)

	return nil
}

// seedTestUser test kullanıcısını oluşturur (idempotent)
func (s *Seeder) seedTestUser(ctx context.Context) (string, error) {
	testUser := GetTestUser()

	// Önce var mı kontrol et
	var existingID string
	err := s.db.GetContext(ctx, &existingID,
		"SELECT id FROM users WHERE email = $1", testUser.Email)

	if err == nil && existingID != "" {
		log.Printf("   ✓ Test kullanıcı zaten mevcut (ID: %s)", existingID)
		return existingID, nil
	}

	// Password hash'le
	hash, err := bcrypt.GenerateFromPassword([]byte(testUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("password hash hatası: %w", err)
	}

	// Kullanıcı oluştur
	var userID string
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO users (full_name, email, password_hash)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO UPDATE SET full_name = EXCLUDED.full_name
		RETURNING id
	`, testUser.FullName, testUser.Email, string(hash)).Scan(&userID)

	if err != nil {
		return "", fmt.Errorf("kullanıcı oluşturulamadı: %w", err)
	}

	log.Printf("   ✓ Test kullanıcı oluşturuldu (ID: %s)", userID)
	log.Printf("     Email: %s", testUser.Email)
	log.Printf("     Şifre: %s", testUser.Password)

	return userID, nil
}

// seedTestCategories kullanıcıya özel kategorileri ekler (idempotent)
func (s *Seeder) seedTestCategories(ctx context.Context, userID string) error {
	categories := GetTestCategories()
	var count int

	for _, cat := range categories {
		query := `
			INSERT INTO categories (user_id, name, icon, type, is_default)
			VALUES ($1, $2, $3, $4, false)
			ON CONFLICT DO NOTHING
		`
		result, err := s.db.ExecContext(ctx, query, userID, cat.Name, cat.Icon, cat.Type)
		if err != nil {
			return fmt.Errorf("kategori eklenemedi (%s): %w", cat.Name, err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			count++
		}
	}

	log.Printf("   ✓ %d kullanıcı kategorisi eklendi", count)
	return nil
}

// seedTestTransactions test transaction'larını ekler
func (s *Seeder) seedTestTransactions(ctx context.Context, userID string) error {
	transactions := GetTestTransactions()

	// Önce mevcut test transaction'larını temizle (opsiyonel, idempotent için)
	// Bu sayede her çalıştırmada aynı veriler olur
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM transactions 
		WHERE user_id = $1 
		AND description LIKE '%' || 'seed' || '%' OR user_id = $1
	`, userID)
	// Hata olsa bile devam et (ilk çalıştırmada tablo boş olabilir)
	_ = err

	// Kategori ID'lerini çek (cache)
	categoryMap, err := s.getCategoryMap(ctx, userID)
	if err != nil {
		return fmt.Errorf("kategori map alınamadı: %w", err)
	}

	var count int
	for _, tx := range transactions {
		categoryID, ok := categoryMap[tx.CategoryName]
		if !ok {
			log.Printf("   ⚠ Kategori bulunamadı: %s, atlanıyor", tx.CategoryName)
			continue
		}

		date := time.Now().AddDate(0, 0, -tx.DaysAgo)

		query := `
			INSERT INTO transactions (user_id, category_id, amount, type, description, date)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err := s.db.ExecContext(ctx, query,
			userID, categoryID, tx.Amount, tx.Type, tx.Description, date)
		if err != nil {
			return fmt.Errorf("transaction eklenemedi: %w", err)
		}
		count++
	}

	log.Printf("   ✓ %d transaction eklendi", count)
	return nil
}

// seedTestBudgets test budget'larını ekler (idempotent)
func (s *Seeder) seedTestBudgets(ctx context.Context, userID string) error {
	budgets := GetTestBudgets()

	categoryMap, err := s.getCategoryMap(ctx, userID)
	if err != nil {
		return fmt.Errorf("kategori map alınamadı: %w", err)
	}

	var count int
	for _, b := range budgets {
		categoryID, ok := categoryMap[b.CategoryName]
		if !ok {
			log.Printf("   ⚠ Kategori bulunamadı: %s, atlanıyor", b.CategoryName)
			continue
		}

		// Aynı kategori+ay+yıl için budget varsa güncelle, yoksa ekle
		query := `
			INSERT INTO budgets (user_id, category_id, amount, month, year)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT DO NOTHING
		`
		result, err := s.db.ExecContext(ctx, query,
			userID, categoryID, b.Amount, b.Month, b.Year)
		if err != nil {
			// Unique constraint yoksa normal insert dene
			// Duplicate kontrolü için alternatif yöntem
			var existing int
			checkErr := s.db.GetContext(ctx, &existing, `
				SELECT COUNT(*) FROM budgets 
				WHERE user_id = $1 AND category_id = $2 AND month = $3 AND year = $4
			`, userID, categoryID, b.Month, b.Year)

			if checkErr == nil && existing > 0 {
				continue // zaten var, atla
			}

			return fmt.Errorf("budget eklenemedi: %w", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			count++
		}
	}

	log.Printf("   ✓ %d budget eklendi", count)
	return nil
}

// getCategoryMap kullanıcının erişebildiği tüm kategorileri map olarak döner
func (s *Seeder) getCategoryMap(ctx context.Context, userID string) (map[string]string, error) {
	type catRow struct {
		ID   string `db:"id"`
		Name string `db:"name"`
	}

	var categories []catRow
	query := `
		SELECT id, name FROM categories 
		WHERE user_id = $1 OR is_default = true
	`
	err := s.db.SelectContext(ctx, &categories, query, userID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, c := range categories {
		result[c.Name] = c.ID
	}

	return result, nil
}

// CleanTestData test verilerini temizler (dikkatli kullan!)
func (s *Seeder) CleanTestData(ctx context.Context) error {
	log.Println("🧹 Test verileri temizleniyor...")

	testUser := GetTestUser()

	// Test kullanıcısının ID'sini bul
	var userID string
	err := s.db.GetContext(ctx, &userID,
		"SELECT id FROM users WHERE email = $1", testUser.Email)
	if err != nil {
		log.Println("   Test kullanıcı bulunamadı, temizlenecek veri yok")
		return nil
	}

	// Sırayla temizle (foreign key bağımlılıkları nedeniyle)
	queries := []string{
		"DELETE FROM transactions WHERE user_id = $1",
		"DELETE FROM budgets WHERE user_id = $1",
		"DELETE FROM refresh_tokens WHERE user_id = $1",
		"DELETE FROM categories WHERE user_id = $1 AND is_default = false",
		"DELETE FROM users WHERE id = $1",
	}

	for _, q := range queries {
		_, err := s.db.ExecContext(ctx, q, userID)
		if err != nil {
			log.Printf("   ⚠ Temizlik hatası: %v", err)
		}
	}

	log.Println("✅ Test verileri temizlendi!")
	return nil
}
