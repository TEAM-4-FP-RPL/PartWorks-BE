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

	userRepo := repository.NewUserRepository(db)
	authUC := usecase.NewAuthUsecase(userRepo)
	authH := handler.NewAuthHandler(authUC)

	jobRepo := repository.NewJobRepository(db)
	jobUC := usecase.NewJobUsecase(jobRepo)
	jobH := handler.NewJobHandler(jobUC)

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
