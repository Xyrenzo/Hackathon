package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"JumysTab/backend/internal/model"
	"JumysTab/backend/internal/repository"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	userRepo   *repository.UserRepository
	jwtSecret  []byte
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (*model.User, string, error) {
	// Проверяем, существует ли пользователь
	existingUser, _ := s.userRepo.GetUserByPhone(ctx, req.Phone)
	if existingUser != nil {
		return nil, "", fmt.Errorf("user with this phone already exists")
	}

	// Создаем пользователя
	user, err := s.userRepo.CreateUser(ctx, req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	// Генерируем код для телеграм верификации
	telegramCode, err := generateCode(8)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate telegram code: %w", err)
	}

	// Сохраняем код
	err = s.userRepo.StoreTelegramCode(ctx, user.ID, telegramCode)
	if err != nil {
		return nil, "", fmt.Errorf("failed to store telegram code: %w", err)
	}

	return user, telegramCode, nil
}

func (s *AuthService) Login(ctx context.Context, phone string, code string) (string, error) {
	// Проверяем код
	valid, err := s.userRepo.VerifyLoginCode(ctx, phone, code)
	if err != nil || !valid {
		return "", fmt.Errorf("invalid or expired code")
	}

	// Получаем пользователя
	user, err := s.userRepo.GetUserByPhone(ctx, phone)
	if err != nil {
		return "", fmt.Errorf("user not found")
	}

	// Проверяем верификацию телеграм
	if !user.IsVerified {
		return "", fmt.Errorf("account not activated. Please verify your Telegram account")
	}

	// Генерируем JWT токен
	token, err := s.generateJWT(user)
	if err != nil {
		return "", fmt.Errorf("failed to generate token")
	}

	return token, nil
}

func (s *AuthService) RequestLoginCode(ctx context.Context, phone string) error {
	// Проверяем существование пользователя
	_, err := s.userRepo.GetUserByPhone(ctx, phone)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Генерируем код
	code, err := generateCode(6)
	if err != nil {
		return fmt.Errorf("failed to generate login code")
	}

	// Сохраняем код
	err = s.userRepo.StoreLoginCode(ctx, phone, code)
	if err != nil {
		return fmt.Errorf("failed to store login code")
	}

	// TODO: Отправить код через SMS или Telegram
	fmt.Printf("Login code for %s: %s\n", phone, code)

	return nil
}

func (s *AuthService) VerifyTelegram(ctx context.Context, userID string, telegramCode string) error {
	// Проверяем код
	storedCode, err := s.userRepo.GetTelegramCode(ctx, userID)
	if err != nil {
		return fmt.Errorf("invalid or expired verification code")
	}

	if storedCode != telegramCode {
		return fmt.Errorf("verification code mismatch")
	}

	return nil
}

func (s *AuthService) ConfirmTelegramVerification(ctx context.Context, userID string, chatID string) error {
	return s.userRepo.UpdateTelegramChatID(ctx, userID, chatID)
}

func (s *AuthService) generateJWT(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"phone":   user.Phone,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func generateCode(length int) (string, error) {
	const charset = "0123456789"
	code := make([]byte, length)
	
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		code[i] = charset[n.Int64()]
	}
	
	return string(code), nil
}
