package database

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPostgresDB(ctx context.Context) (*gorm.DB, error) {
	dsn := os.Getenv("CONNECTION_STRING")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		return nil, fmt.Errorf("no database DSN provided in CONNECTION_STRING or DATABASE_URL")
	}

	os.Setenv("GODEBUG", "netdns=cgo")

	gormLogger := logger.Default.LogMode(logger.Info)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:      gormLogger,
		PrepareStmt: false, // Disable prepared statements to avoid cached plan issues
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	_ = db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto;")

	am := strings.ToLower(strings.TrimSpace(os.Getenv("AUTO_MIGRATE")))
	if am == "1" || am == "true" || am == "yes" {
		if err := db.AutoMigrate(
			&domain.User{},
			&domain.Category{},
			&domain.EmployerProfile{},
			&domain.WorkerProfile{},
			&domain.WorkerCV{},
			&domain.Availability{},
			&domain.Job{},
			&domain.JobSchedule{},
			&domain.Application{},
		); err != nil {
			return nil, fmt.Errorf("auto migrate failed: %w", err)
		}
	}

	return db, nil
}
