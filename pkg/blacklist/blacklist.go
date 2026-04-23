package blacklist

import (
	"context"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

// Init DB bağlantısını blacklist paketine atar. main.go'da çağrılmalı.
func Init(database *sqlx.DB) {
	db = database
}

// Add token jti'sini blacklist'e ekler.
func Add(jti string) {
	if db == nil {
		log.Println("blacklist: DB başlatılmamış, token blacklist'e eklenemedi")
		return
	}
	expiresAt := time.Now().Add(time.Hour)
	_, err := db.ExecContext(
		context.Background(),
		`INSERT INTO token_blacklist (jti, expires_at) VALUES ($1, $2) ON CONFLICT (jti) DO NOTHING`,
		jti, expiresAt,
	)
	if err != nil {
		log.Printf("blacklist: token eklenemedi: %v", err)
	}
}

// IsBlacklisted token jti'sinin blacklist'te olup olmadığını kontrol eder.
func IsBlacklisted(jti string) bool {
	if db == nil {
		return false
	}
	var exists bool
	err := db.QueryRowContext(
		context.Background(),
		`SELECT EXISTS(SELECT 1 FROM token_blacklist WHERE jti = $1 AND expires_at > NOW())`,
		jti,
	).Scan(&exists)
	if err != nil {
		log.Printf("blacklist: kontrol hatası: %v", err)
		return false
	}
	return exists
}
