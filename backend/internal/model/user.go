package model

import (
	"time"
)

type User struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Phone          string    `json:"phone"`
	City           string    `json:"city"`
	Skills         []string  `json:"skills"`
	Availability   string    `json:"availability"`
	TelegramChatID *string   `json:"telegramChatId"`
	Rating         float64   `json:"rating"`
	CreatedAt      time.Time `json:"createdAt"`
	IsVerified     bool      `json:"isVerified"`
}

type RegisterRequest struct {
	Name  string `json:"name" binding:"required"`
	Phone string `json:"phone" binding:"required"`
	City  string `json:"city" binding:"required"`
}

type LoginRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

type VerifyTelegramRequest struct {
	UserID       string `json:"userId"`
	TelegramCode string `json:"telegramCode"`
}
