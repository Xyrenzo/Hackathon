package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"JumysTab/internal/model"
)

var ErrNotFound = errors.New("not found")
var ErrAlreadyExists = errors.New("already exists")

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, u *model.User) error {
	const q = `
		INSERT INTO users (id, name, phone, city, skills, availability, telegram_chat_id, rating, tg_verified, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`
	_, err := r.pool.Exec(ctx, q,
		u.ID, u.Name, u.Phone, u.City,
		u.Skills, u.Availability,
		u.TelegramChatID, u.Rating, u.TGVerified, u.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyExists
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	const q = `SELECT id,name,phone,city,skills,availability,telegram_chat_id,rating,tg_verified,created_at FROM users WHERE id=$1`
	return r.scanOne(ctx, q, id)
}

func (r *UserRepository) FindByName(ctx context.Context, name string) (*model.User, error) {
	const q = `SELECT id,name,phone,city,skills,availability,telegram_chat_id,rating,tg_verified,created_at FROM users WHERE LOWER(name)=LOWER($1)`
	return r.scanOne(ctx, q, name)
}

func (r *UserRepository) FindByTelegramToken(ctx context.Context, token string) (*model.User, error) {
	const q = `
		SELECT u.id,u.name,u.phone,u.city,u.skills,u.availability,u.telegram_chat_id,u.rating,u.tg_verified,u.created_at
		FROM users u
		JOIN pending_registrations pr ON pr.user_id = u.id
		WHERE pr.token = $1
	`
	return r.scanOne(ctx, q, token)
}

func (r *UserRepository) SetTelegramVerified(ctx context.Context, userID string, chatID int64) error {
	const q = `UPDATE users SET telegram_chat_id=$1, tg_verified=true WHERE id=$2`
	_, err := r.pool.Exec(ctx, q, chatID, userID)
	if err != nil {
		return fmt.Errorf("set tg verified: %w", err)
	}
	return nil
}

func (r *UserRepository) GetProfile(ctx context.Context, userID string) (*model.User, error) {
	return r.FindByID(ctx, userID)
}

// scanOne — shared scan helper
func (r *UserRepository) scanOne(ctx context.Context, q string, args ...any) (*model.User, error) {
	row := r.pool.QueryRow(ctx, q, args...)
	u := &model.User{}
	err := row.Scan(
		&u.ID, &u.Name, &u.Phone, &u.City,
		&u.Skills, &u.Availability,
		&u.TelegramChatID, &u.Rating, &u.TGVerified, &u.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return u, nil
}

func isUniqueViolation(err error) bool {
	return err != nil && (containsCode(err, "23505"))
}

func containsCode(err error, code string) bool {
	type pgErr interface{ SQLState() string }
	if e, ok := err.(pgErr); ok {
		return e.SQLState() == code
	}
	return false
}
