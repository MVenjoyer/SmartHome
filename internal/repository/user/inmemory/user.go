package inmemory

import (
	"context"
	"errors"
	"homework/internal/domain"
	"sync"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserIsNil    = errors.New("user is nil")
)

type UserRepository struct {
	users map[int64]*domain.User
	mu    sync.RWMutex
}

func NewUserRepository() *UserRepository {
	return &UserRepository{users: make(map[int64]*domain.User)}
}

func (r *UserRepository) SaveUser(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if user == nil {
		return ErrUserIsNil
	}
	user.ID = int64(len(r.users) + 1)
	r.users[user.ID] = user
	return nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	user, ok := r.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}
