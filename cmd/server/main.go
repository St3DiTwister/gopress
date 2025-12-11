package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"gopress/internal/database"
	"gopress/internal/repository"
	"log"
	"net/http"
	"os"
	"time"

	httphandler "gopress/internal/handler/http"
	jwtpkg "gopress/pkg/jwt"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("error loading .env:", err)
	}

	ctx := context.Background()

	db, err := database.NewDB(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	pool := db.Pool()
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET not set")
	}
	jwtManager := jwtpkg.NewManager(secret, 24*time.Hour)
	userRepo := repository.NewUserRepo(pool)

	authHandler := httphandler.NewAuthHandler(userRepo, jwtManager)
	handlers := httphandler.Handlers{
		Auth: authHandler,
	}

	router := httphandler.NewRouter(handlers, jwtManager)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router.Handler(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Println("HTTP server is listening on", srv.Addr)

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}
