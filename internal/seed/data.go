package seed

import "time"

// TestUser test kullanıcı bilgileri
type TestUser struct {
	Email        string
	FullName     string
	Password     string // plain text - hash'lenecek
	PasswordHash string // runtime'da set edilecek
}

// TestCategory kullanıcıya özel test kategorisi
type TestCategory struct {
	Name string
	Icon string
	Type string // "income" veya "expense"
}

// TestTransaction test transaction verisi
type TestTransaction struct {
	CategoryName string // kategori adına göre eşleştirilecek
	Amount       float64
	Type         string // "income" veya "expense"
	Description  string
	DaysAgo      int // bugünden kaç gün önce
}

// TestBudget test budget verisi
type TestBudget struct {
	CategoryName string // kategori adına göre eşleştirilecek
	Amount       float64
	Month        int
	Year         int
}

// GetTestUser test kullanıcısını döndürür
func GetTestUser() TestUser {
	return TestUser{
		Email:    "test@finbud.dev",
		FullName: "Test Kullanıcı",
		Password: "Test123!",
	}
}

// GetTestCategories kullanıcıya özel test kategorilerini döndürür
func GetTestCategories() []TestCategory {
	return []TestCategory{
		{Name: "Kripto", Icon: "🪙", Type: "income"},
		{Name: "Hobi", Icon: "🎨", Type: "expense"},
	}
}

// GetTestTransactions test transaction'larını döndürür
func GetTestTransactions() []TestTransaction {
	return []TestTransaction{
		// Gelirler (default kategoriler kullanılıyor)
		{CategoryName: "Maaş", Amount: 45000.00, Type: "income", Description: "Nisan maaşı", DaysAgo: 5},
		{CategoryName: "Freelance", Amount: 5000.00, Type: "income", Description: "Logo tasarım projesi", DaysAgo: 10},
		{CategoryName: "Kripto", Amount: 1500.00, Type: "income", Description: "BTC satışı", DaysAgo: 3},

		// Giderler (default kategoriler kullanılıyor)
		{CategoryName: "Yemek", Amount: 250.00, Type: "expense", Description: "Haftalık market", DaysAgo: 2},
		{CategoryName: "Yemek", Amount: 180.00, Type: "expense", Description: "Restoran", DaysAgo: 4},
		{CategoryName: "Ulaşım", Amount: 500.00, Type: "expense", Description: "Benzin", DaysAgo: 6},
		{CategoryName: "Faturalar", Amount: 1200.00, Type: "expense", Description: "Elektrik + Doğalgaz", DaysAgo: 8},
		{CategoryName: "Faturalar", Amount: 350.00, Type: "expense", Description: "İnternet", DaysAgo: 8},
		{CategoryName: "Alışveriş", Amount: 2500.00, Type: "expense", Description: "Elektronik", DaysAgo: 12},
		{CategoryName: "Sağlık", Amount: 400.00, Type: "expense", Description: "Eczane", DaysAgo: 7},
		{CategoryName: "Eğlence", Amount: 150.00, Type: "expense", Description: "Netflix + Spotify", DaysAgo: 1},
		{CategoryName: "Hobi", Amount: 800.00, Type: "expense", Description: "Resim malzemeleri", DaysAgo: 9},
	}
}

// GetTestBudgets test budget'larını döndürür
func GetTestBudgets() []TestBudget {
	now := time.Now()
	currentMonth := int(now.Month())
	currentYear := now.Year()

	return []TestBudget{
		{CategoryName: "Yemek", Amount: 3000.00, Month: currentMonth, Year: currentYear},
		{CategoryName: "Ulaşım", Amount: 1500.00, Month: currentMonth, Year: currentYear},
		{CategoryName: "Faturalar", Amount: 2000.00, Month: currentMonth, Year: currentYear},
		{CategoryName: "Alışveriş", Amount: 3000.00, Month: currentMonth, Year: currentYear},
		{CategoryName: "Eğlence", Amount: 500.00, Month: currentMonth, Year: currentYear},
	}
}

// DefaultCategories - Eğer migration'da yoksa eklenecek default kategoriler
// NOT: Bunlar zaten migration'da var, bu liste sadece referans için
var DefaultCategories = []struct {
	Name string
	Icon string
	Type string
}{
	// Gelir kategorileri
	{Name: "Maaş", Icon: "assets/icons/para.png", Type: "income"},
	{Name: "Freelance", Icon: "assets/icons/laptop.png", Type: "income"},
	{Name: "Yatırım", Icon: "assets/icons/para_akis.png", Type: "income"},
	{Name: "Diğer Gelir", Icon: "assets/icons/gelir_cüzdan.png", Type: "income"},
	// Gider kategorileri
	{Name: "Yemek", Icon: "assets/icons/restorant.png", Type: "expense"},
	{Name: "Ulaşım", Icon: "assets/icons/araba.png", Type: "expense"},
	{Name: "Alışveriş", Icon: "assets/icons/market_arabasi.png", Type: "expense"},
	{Name: "Faturalar", Icon: "assets/icons/fatura.png", Type: "expense"},
	{Name: "Sağlık", Icon: "assets/icons/saglik.png", Type: "expense"},
	{Name: "Eğlence", Icon: "assets/icons/sinema.png", Type: "expense"},
	{Name: "Diğer Gider", Icon: "assets/icons/koli.png", Type: "expense"},
}
