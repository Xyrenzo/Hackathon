package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"JumysTab/backend/internal/config"
	"JumysTab/backend/internal/handler"
	"JumysTab/backend/internal/middleware"
	"JumysTab/backend/internal/repository"
	"JumysTab/backend/internal/service"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	// Подключение к БД
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	// Инициализация слоев
	userRepo := repository.NewUserRepository(pool)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(authService)

	// Роутер
	r := mux.NewRouter()
	api := r.PathPrefix("/api/v1").Subrouter()

	// Публичные роуты
	api.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	api.HandleFunc("/auth/login", authHandler.Login).Methods("POST")
	api.HandleFunc("/auth/request-code", authHandler.RequestLoginCode).Methods("POST")

	// Защищенные роуты
	protected := api.PathPrefix("/").Subrouter()
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	protected.HandleFunc("/profile", authHandler.GetProfile).Methods("GET")

	// Graceful shutdown
	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Запускаем Telegram бота в горутине
	// telebot, _ := telegram.NewBot(cfg.TelegramBotToken, authService)
	// go telebot.Start(context.Background())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	server.Shutdown(context.Background())
}
