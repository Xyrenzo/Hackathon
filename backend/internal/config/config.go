package config

import (
	"os"
)

type Config struct {
	DatabaseURL      string
	JWTSecret        string
	TelegramBotToken string
	ServerPort       string
}

func Load() *Config {
	return &Config{
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:32772/yourdb?sslmode=disable"),
		JWTSecret:        getEnv("JWT_SECRET", "your-secret-key"),
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", "your-bot-token"),
		ServerPort:       getEnv("SERVER_PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
