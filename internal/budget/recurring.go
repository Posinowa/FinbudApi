package budget

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
)

// StartRecurringJob her ayın 1'inde tekrarlayan bütçeleri otomatik oluşturur.
// Sunucu başlarken mevcut ay için de kontrol yapar.
func StartRecurringJob(repo *Repository) {
	// Sunucu açılışında mevcut ay için hemen çalıştır
	go func() {
		createRecurringBudgets(repo)

		for {
			// Bir sonraki ayın 1'ine kadar bekle
			now := time.Now()
			next := time.Date(now.Year(), now.Month()+1, 1, 0, 1, 0, 0, now.Location())
			time.Sleep(time.Until(next))

			createRecurringBudgets(repo)
		}
	}()
}

// createRecurringBudgets mevcut ay için tekrarlayan bütçeleri oluşturur
func createRecurringBudgets(repo *Repository) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	// Tüm tekrarlayan bütçe şablonlarını getir (her kullanıcı+kategori için en güncel)
	templates, err := repo.GetRecurringTemplates(ctx)
	if err != nil {
		log.Printf("⚠️  Recurring bütçe şablonları alınamadı: %v", err)
		return
	}

	created := 0
	for _, tmpl := range templates {
		// Bu ay için zaten var mı kontrol et
		exists, err := repo.Exists(ctx, tmpl.UserID, tmpl.CategoryID, currentYear, currentMonth)
		if err != nil || exists {
			continue
		}

		// Yoksa oluştur
		newBudget := &Budget{
			ID:          uuid.New().String(),
			UserID:      tmpl.UserID,
			CategoryID:  tmpl.CategoryID,
			Amount:      tmpl.Amount,
			Month:       currentMonth,
			Year:        currentYear,
			IsRecurring: true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := repo.Create(ctx, newBudget); err != nil {
			log.Printf("⚠️  Recurring bütçe oluşturulamadı (user:%s category:%s): %v",
				tmpl.UserID, tmpl.CategoryID, err)
			continue
		}
		created++
	}

	if created > 0 {
		log.Printf("✅ %d tekrarlayan bütçe otomatik oluşturuldu (%d-%02d)", created, currentYear, currentMonth)
	}
}
