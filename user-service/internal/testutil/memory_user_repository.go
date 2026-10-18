package testutil

import (
	"context"
	"sort"
	"sync"

	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"github.com/Ishan-Gijavanekar/user-service/internal/repository"
)

// MemoryUserRepository is a concurrency-safe repository used by transport and service tests.
type MemoryUserRepository struct {
	mu      sync.RWMutex
	byID    map[string]*domain.User
	byEmail map[string]*domain.User
}

func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		byID:    make(map[string]*domain.User),
		byEmail: make(map[string]*domain.User),
	}
}

func cloneUser(user *domain.User) *domain.User {
	if user == nil {
		return nil
	}
	copy := *user
	return &copy
}

func (r *MemoryUserRepository) Create(_ context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byEmail[user.Email]; exists {
		return domain.ErrUserAlreadyExsists
	}
	stored := cloneUser(user)
	r.byID[user.ID.Hex()] = stored
	r.byEmail[user.Email] = stored
	return nil
}

func (r *MemoryUserRepository) GetByID(_ context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, exists := r.byID[id]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return cloneUser(user), nil
}

func (r *MemoryUserRepository) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, exists := r.byEmail[email]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return cloneUser(user), nil
}

func (r *MemoryUserRepository) List(_ context.Context, params repository.ListUserParmas) ([]*domain.User, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	users := make([]*domain.User, 0, len(r.byID))
	for _, user := range r.byID {
		users = append(users, cloneUser(user))
	}
	sort.Slice(users, func(i, j int) bool { return users[i].CreatedAt.After(users[j].CreatedAt) })

	start := (params.Size - 1) * params.Limit
	if start < 0 {
		start = 0
	}
	if start >= len(users) {
		return []*domain.User{}, int64(len(users)), nil
	}
	end := start + params.Limit
	if end > len(users) {
		end = len(users)
	}
	return users[start:end], int64(len(users)), nil
}

func (r *MemoryUserRepository) Update(_ context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	old, exists := r.byID[user.ID.Hex()]
	if !exists {
		return domain.ErrUserNotFound
	}
	if current, exists := r.byEmail[user.Email]; exists && current.ID != user.ID {
		return domain.ErrUserAlreadyExsists
	}
	delete(r.byEmail, old.Email)
	stored := cloneUser(user)
	r.byID[user.ID.Hex()] = stored
	r.byEmail[user.Email] = stored
	return nil
}

func (r *MemoryUserRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, exists := r.byID[id]
	if !exists {
		return domain.ErrUserNotFound
	}
	delete(r.byID, id)
	delete(r.byEmail, user.Email)
	return nil
}

var _ repository.UserRespository = (*MemoryUserRepository)(nil)
