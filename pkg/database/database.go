package database

import (
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/Posinowa/FinbudApp/pkg/config"
)

func Connect(cfg *config.Config) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("veritabanina baglanılamadı: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("veritabani ping basarisiz: %w", err)
	}

	log.Println("Veritabani baglantisi basarili")
	return db, nil
}