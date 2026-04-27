package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/config"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/database"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/handler"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/repository"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/routes"
	"github.com/TEAM-4-FP-RPL/PartWorks-BE/internal/usecase"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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

	jobRepo := repository.NewJobRepository(db)
	userRepo := repository.NewUserRepository(db)

	jobUsecase := usecase.NewJobUsecase(jobRepo)
	authUC := usecase.NewAuthUsecase(userRepo)

	jobHandler := handler.NewJobHandler(jobUsecase)
	authH := handler.NewAuthHandler(authUC)

	router := routes.NewRouter(authH, jobHandler)

	port := config.GetPort()
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	
	log.Printf("server starting on port :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}