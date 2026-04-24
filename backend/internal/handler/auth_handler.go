package handler

import (
	"encoding/json"
	"net/http"

	"JumysTab/backend/internal/model"
	"JumysTab/backend/internal/service"
	"fmt"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Phone == "" || req.City == "" {
		http.Error(w, "Name, phone and city are required", http.StatusBadRequest)
		return
	}

	user, telegramCode, err := h.authService.Register(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	// Формируем ссылку для верификации
	verificationLink := fmt.Sprintf("https://t.me/your_bot?start=%s", telegramCode)

	response := map[string]interface{}{
		"user":             user,
		"verificationLink": verificationLink,
		"telegramCode":     telegramCode,
		"message":          "Please verify your account via Telegram bot",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) RequestLoginCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone string `json:"phone"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.authService.RequestLoginCode(r.Context(), req.Phone); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Verification code sent",
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.authService.Login(r.Context(), req.Phone, req.Code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token":   token,
		"message": "Login successful",
	})
}

func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	
	// Получаем профиль через сервис или репозиторий
	// Для примера, добавим метод в хендлер
	
	response := map[string]string{
		"user_id": userID,
		"message": "Profile data would be here",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
