package postgres

import (
	"fmt"
	"time"

	"github.com/RX90/Chat/config"
	"github.com/RX90/Chat/internal/domain/entities"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	maxOpenConns    = 25
	maxIdleConns    = 5
	connMaxLifetime = 5 * time.Minute
	connMaxIdleTime = 1 * time.Minute
)

func NewPostgresDB(cfg *config.DBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open gorm connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get generic db: %w", err)
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

	if err = sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := db.AutoMigrate(&entities.User{}, &entities.Message{}, &entities.RefreshToken{}); err != nil {
		return nil, fmt.Errorf("failed to automigrate: %w", err)
	}

	if err := applyCustomIndexes(db); err != nil {
		return nil, fmt.Errorf("failed to apply custom indexes: %w", err)
	}

	return db, nil
}

func applyCustomIndexes(db *gorm.DB) error {
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique ON users (LOWER(email))`).Error; err != nil {
		return err
	}
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS users_username_unique ON users (LOWER(username))`).Error; err != nil {
		return err
	}
	return nil
}