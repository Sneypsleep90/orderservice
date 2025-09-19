package main

import (
	"context"
	"log"
	"myapp/internal/cache"
	"myapp/internal/config"
	"myapp/internal/database"
	"myapp/internal/handlers"
	"myapp/internal/kafka"
	"myapp/internal/repository"
	"myapp/internal/service"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load("config.env"); err != nil {
		log.Printf("Warning: config.env file not found: %v", err)
	}

	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(cfg, cfg.MigrationsPath); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	repo := repository.NewPostgresRepository(db)
	var orderCache cache.Cache
	if cfg.CacheType == "lru" {
		lruCache, err := cache.NewLRUCache(cfg.CacheLRUSize)
		if err != nil {
			log.Fatalf("Failed to create LRU cache: %v", err)
		}
		orderCache = lruCache
	} else {
		orderCache = cache.NewInMemoryCache()
	}
	statsCache := cache.NewStatsCache(orderCache)
	orderService := service.NewOrderService(repo, statsCache)

	if err := orderService.WarmupCache(); err != nil {
		log.Printf("Warning: Failed to warm up cache: %v", err)
	}

	log.Printf("Creating Kafka consumer with brokers: %s, group: %s, topic: %s",
		cfg.KafkaBrokers[0], cfg.KafkaGroupID, cfg.KafkaTopic)

	consumer, err := kafka.NewConsumer(
		cfg.KafkaBrokers[0],
		cfg.KafkaGroupID,
		cfg.KafkaTopic,
		orderService,
	)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer consumer.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("Starting Kafka consumer...")
	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Printf("Kafka consumer error: %v", err)
		}
	}()
	log.Println("Kafka consumer started successfully")

	router := mux.NewRouter()
	handler := handlers.NewHandler(orderService)
	handler.RegisterRoutes(router)

	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.ServerPort),
		Handler: router,
	}

	go func() {
		log.Printf("Starting HTTP server on port %d", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
