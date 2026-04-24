package repository

import (
	"context"
	"fmt"
	"time"

	"JumysTab/backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) CreateUser(ctx context.Context, req model.RegisterRequest) (*model.User, error) {
	id := uuid.New().String()
	now := time.Now()

	user := &model.User{
		ID:        id,
		Name:      req.Name,
		Phone:     req.Phone,
		City:      req.City,
		CreatedAt: now,
	}

	query := `
		INSERT INTO users (id, name, phone, city, created_at, is_verified)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, phone, city, skills, availability, telegram_chat_id, rating, created_at, is_verified
	`

	err := r.pool.QueryRow(ctx, query,
		user.ID, user.Name, user.Phone, user.City, user.CreatedAt, false,
	).Scan(
		&user.ID, &user.Name, &user.Phone, &user.City,
		&user.Skills, &user.Availability, &user.TelegramChatID,
		&user.Rating, &user.CreatedAt, &user.IsVerified,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetUserByPhone(ctx context.Context, phone string) (*model.User, error) {
	user := &model.User{}

	query := `
		SELECT id, name, phone, city, COALESCE(skills, '{}'), 
		       COALESCE(availability, ''), telegram_chat_id, 
		       COALESCE(rating, 0), created_at, is_verified
		FROM users
		WHERE phone = $1
	`

	err := r.pool.QueryRow(ctx, query, phone).Scan(
		&user.ID, &user.Name, &user.Phone, &user.City,
		&user.Skills, &user.Availability, &user.TelegramChatID,
		&user.Rating, &user.CreatedAt, &user.IsVerified,
	)

	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	user := &model.User{}

	query := `
		SELECT id, name, phone, city, COALESCE(skills, '{}'), 
		       COALESCE(availability, ''), telegram_chat_id, 
		       COALESCE(rating, 0), created_at, is_verified
		FROM users
		WHERE id = $1
	`

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Name, &user.Phone, &user.City,
		&user.Skills, &user.Availability, &user.TelegramChatID,
		&user.Rating, &user.CreatedAt, &user.IsVerified,
	)

	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}

func (r *UserRepository) UpdateTelegramChatID(ctx context.Context, userID string, chatID string) error {
	query := `
		UPDATE users 
		SET telegram_chat_id = $2, is_verified = true
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, userID, chatID)
	return err
}

func (r *UserRepository) StoreTelegramCode(ctx context.Context, userID string, code string) error {
	query := `
		INSERT INTO telegram_verification (user_id, code, created_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.pool.Exec(ctx, query, userID, code, time.Now())
	return err
}

func (r *UserRepository) GetTelegramCode(ctx context.Context, userID string) (string, error) {
	var code string
	query := `
		SELECT code FROM telegram_verification
		WHERE user_id = $1 AND created_at > $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	err := r.pool.QueryRow(ctx, query, userID, time.Now().Add(-24*time.Hour)).Scan(&code)
	if err != nil {
		return "", fmt.Errorf("code not found or expired: %w", err)
	}

	return code, nil
}

func (r *UserRepository) StoreLoginCode(ctx context.Context, phone string, code string) error {
	query := `
		INSERT INTO login_codes (phone, code, created_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.pool.Exec(ctx, query, phone, code, time.Now())
	return err
}

func (r *UserRepository) VerifyLoginCode(ctx context.Context, phone string, code string) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM login_codes
		WHERE phone = $1 AND code = $2 
		AND created_at > $3 AND used = false
	`

	err := r.pool.QueryRow(ctx, query, phone, code, time.Now().Add(-15*time.Minute)).Scan(&count)
	if err != nil {
		return false, err
	}

	if count > 0 {
		_, err = r.pool.Exec(ctx, "UPDATE login_codes SET used = true WHERE phone = $1 AND code = $2", phone, code)
		return true, err
	}

	return false, nil
}
