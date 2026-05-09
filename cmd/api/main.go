package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/config"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/database"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/handler"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/repository"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/routes"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/storage"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/usecase"
	"github.com/joho/godotenv"
)

func main() {
	// Try load .env from current working dir, then fallback to executable dir.
	_ = godotenv.Load()
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		_ = godotenv.Load(filepath.Join(exeDir, ".env"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Log DSN (without password for security)
	dsn := os.Getenv("CONNECTION_STRING")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	// Mask password if present
	maskedDSN := dsn
	if idx := strings.Index(dsn, "password="); idx >= 0 {
		if end := strings.Index(dsn[idx+9:], " "); end >= 0 {
			maskedDSN = dsn[:idx+9] + "****" + dsn[idx+9+end:]
		}
	}
	log.Printf("Attempting to connect to database: %s", maskedDSN)

	db, err := database.NewPostgresDB(ctx)
	if err != nil {
		log.Fatalf("database initialization failed: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("database ping failed: %v", err)
	}

	log.Println("database connected")

	userRepo := repository.NewUserRepository(db)
	authUC := usecase.NewAuthUsecase(userRepo)
	authH := handler.NewAuthHandler(authUC)

	jobRepo := repository.NewJobRepository(db)
	jobUC := usecase.NewJobUsecase(jobRepo, userRepo)

	st, err := storage.NewFromEnv()
	if err != nil {
		log.Fatalf("storage initialization failed: %v", err)
	}

	jobH := handler.NewJobHandler(jobUC, st)

	router := routes.NewRouter(authH, jobH)

	port := config.GetPort()
	srv := &http.Server{Addr: ":" + port, Handler: router}
	go func() {
		log.Printf("listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	select {}
}
