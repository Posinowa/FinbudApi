package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Posinowa/FinbudApp/internal/seed"
	"github.com/Posinowa/FinbudApp/pkg/config"
	"github.com/Posinowa/FinbudApp/pkg/database"
)

func main() {
	// CLI flags
	allFlag := flag.Bool("all", false, "Tüm test verilerini ekle (kullanıcı + transaction + budget)")
	categoriesFlag := flag.Bool("categories", false, "Sadece default kategorileri kontrol et/ekle")
	cleanFlag := flag.Bool("clean", false, "Test verilerini temizle")
	helpFlag := flag.Bool("help", false, "Yardım göster")

	flag.Parse()

	// Yardım
	if *helpFlag || (!*allFlag && !*categoriesFlag && !*cleanFlag) {
		printUsage()
		os.Exit(0)
	}

	// Config yükle
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Config yüklenemedi: %v", err)
	}

	// Veritabanına bağlan
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("❌ Veritabanına bağlanılamadı: %v", err)
	}
	defer db.Close()

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Seeder oluştur
	seeder := seed.NewSeeder(db)

	// İşlemi çalıştır
	switch {
	case *cleanFlag:
		if err := seeder.CleanTestData(ctx); err != nil {
			log.Fatalf("❌ Temizlik hatası: %v", err)
		}

	case *categoriesFlag:
		if err := seeder.SeedDefaultCategoriesOnly(ctx); err != nil {
			log.Fatalf("❌ Seed hatası: %v", err)
		}

	case *allFlag:
		fmt.Println("⚠️  DİKKAT: Bu işlem test verileri ekleyecek.")
		fmt.Println("   Production ortamında KULLANMAYIN!")
		fmt.Println()

		if err := seeder.SeedAll(ctx); err != nil {
			log.Fatalf("❌ Seed hatası: %v", err)
		}
	}
}

func printUsage() {
	fmt.Println(`
╔═══════════════════════════════════════════════════════════════╗
║                    FinbudApi Seed Tool                        ║
╠═══════════════════════════════════════════════════════════════╣
║  Kullanım:                                                    ║
║    go run cmd/seed/main.go [seçenek]                         ║
║                                                               ║
║  Seçenekler:                                                  ║
║    -all         Tüm test verilerini ekle                     ║
║                 (kullanıcı, kategoriler, transaction, budget) ║
║                                                               ║
║    -categories  Sadece default kategorileri kontrol et       ║
║                 (migration'da zaten varsa atlar)              ║
║                                                               ║
║    -clean       Test verilerini temizle                       ║
║                 (sadece test kullanıcısına ait veriler)       ║
║                                                               ║
║    -help        Bu yardım mesajını göster                     ║
║                                                               ║
║  Örnekler:                                                    ║
║    go run cmd/seed/main.go -all                              ║
║    go run cmd/seed/main.go -categories                       ║
║    go run cmd/seed/main.go -clean                            ║
║                                                               ║
║  ⚠️  UYARI: -all seçeneği sadece development/test            ║
║     ortamlarında kullanılmalıdır!                             ║
╚═══════════════════════════════════════════════════════════════╝
`)
}
