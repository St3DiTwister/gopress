package main

import (
	"context"
	"errors"
	"fmt"
	articleSvc "gopress/internal/app/article"
	authSvc "gopress/internal/app/auth"
	"gopress/internal/infra/database"
	"gopress/internal/infra/repository"
	"gopress/internal/transport/grpc"
	httptransport "gopress/internal/transport/http"
	"gopress/internal/transport/http/handlers"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	jwtpkg "gopress/pkg/jwt"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("Warning: .env not loaded:", err)
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
		log.Fatal("JWT_SECRET not set in environment")
	}
	jwtManager := jwtpkg.NewManager(secret, 24*time.Hour)

	userRepo := repository.NewUserRepo(pool)
	articleRepo := repository.NewArticleRepo(pool)

	userService := authSvc.NewService(userRepo, jwtManager)
	articleService := articleSvc.NewService(articleRepo)

	authHandler := handlers.NewAuthHandler(userService)
	articleHandler := handlers.NewArticleHandler(articleService)
	httpHandlers := httptransport.Handlers{
		Auth:    authHandler,
		Article: articleHandler,
	}

	router := httptransport.NewRouter(httpHandlers, jwtManager)
	httpServer := &http.Server{
		Addr:         ":8080",
		Handler:      router.Handler(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = ":50051"
	} else if grpcPort[0] != ':' {
		grpcPort = ":" + grpcPort
	}

	grpcServer, err := grpc.NewServer(userRepo, articleRepo, jwtManager, grpcPort)
	if err != nil {
		log.Fatal("Failed to create gRPC server:", err)
	}

	go func() {
		if err := grpcServer.Start(); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	log.Println("HTTP server is listening on", httpServer.Addr)
	log.Println("gRPC server is listening on", grpcPort)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down servers...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctxShutdown); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	grpcServer.Stop()

	log.Println("Servers stopped gracefully.")
}
