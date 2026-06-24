package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AntonioForYou/wallet-api/internal/config"
	httpHandler "github.com/AntonioForYou/wallet-api/internal/handler/http"
	"github.com/AntonioForYou/wallet-api/internal/repository/postgres"
	"github.com/AntonioForYou/wallet-api/internal/worker"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dsn := postgres.GenerateDSN(cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode)
	pool, err := postgres.InitPool(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	repo := postgres.NewWalletRepo(pool)

	workerPool := worker.NewPool(cfg.WorkerPoolSize, cfg.WorkerBufferSize, repo)
	workerPool.Start(context.Background())

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	handler := httpHandler.NewHandler(workerPool, repo)
	handler.RegisterRoutes(router)

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	go func() {
		log.Printf("Server is running on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
