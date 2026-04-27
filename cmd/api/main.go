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
	jobUsecase := usecase.NewJobUsecase(jobRepo)
	jobHandler := handler.NewJobHandler(jobUsecase)

	authH := handler.NewAuthHandler() 

	router := routes.NewRouter(authH, jobHandler)

	port := config.GetPort()
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{Addr: ":" + port, Handler: router}
	
	log.Printf("listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}