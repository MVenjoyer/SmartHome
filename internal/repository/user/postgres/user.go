package postgres

import (
	"context"
	"errors"
	"homework/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserIsNil    = errors.New("user is nil")
)

const (
	saveUserRequest = `INSERT INTO users (name) VALUES ($1)`
	getUserRequest  = `SELECT id,name FROM users WHERE id = $1`
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}

func (r *UserRepository) SaveUser(ctx context.Context, user *domain.User) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if user == nil {
		return ErrUserIsNil
	}
	_, err := r.pool.Exec(ctx, saveUserRequest, user.Name)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	row := r.pool.QueryRow(ctx, getUserRequest, id)
	user := domain.User{}
	if err := row.Scan(&user.ID, &user.Name); err != nil {
		return nil, ErrUserNotFound
	}
	return &user, nil
}
