package testutil

import (
	"context"
	"sync"
	"time"

	"github.com/Ishan-Gijavanekar/user-service/internal/domain"
	"github.com/Ishan-Gijavanekar/user-service/internal/repository"
)

type MemoryAuthTokenRepository struct {
	mu     sync.RWMutex
	byHash map[string]*domain.AuthToken
}

func NewMemoryAuthTokenRepository() *MemoryAuthTokenRepository {
	return &MemoryAuthTokenRepository{byHash: make(map[string]*domain.AuthToken)}
}

func cloneAuthToken(token *domain.AuthToken) *domain.AuthToken {
	if token == nil {
		return nil
	}
	copy := *token
	return &copy
}

func tokenKey(hash string, tokenType domain.AuthTokenType) string {
	return string(tokenType) + ":" + hash
}

func (r *MemoryAuthTokenRepository) Create(_ context.Context, token *domain.AuthToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byHash[tokenKey(token.TokenHash, token.Type)] = cloneAuthToken(token)
	return nil
}

func (r *MemoryAuthTokenRepository) FindByHash(_ context.Context, hash string, tokenType domain.AuthTokenType) (*domain.AuthToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	token, ok := r.byHash[tokenKey(hash, tokenType)]
	if !ok {
		return nil, domain.ErrInvalidToken
	}
	return cloneAuthToken(token), nil
}

func (r *MemoryAuthTokenRepository) Revoke(_ context.Context, id, reason string, usedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, token := range r.byHash {
		if token.ID == id {
			if token.RevokedAt != nil {
				return domain.ErrTokenRevoked
			}
			token.LastUsedAt = &usedAt
			token.RevokedAt = &usedAt
			token.RevocationReason = reason
			return nil
		}
	}
	return domain.ErrInvalidToken
}

func (r *MemoryAuthTokenRepository) RevokeFamily(_ context.Context, familyID, reason string, revokedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, token := range r.byHash {
		if token.FamilyID == familyID && token.RevokedAt == nil {
			token.RevokedAt = &revokedAt
			token.RevocationReason = reason
		}
	}
	return nil
}

func (r *MemoryAuthTokenRepository) RevokeUser(_ context.Context, userID, reason string, revokedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, token := range r.byHash {
		if token.UserID == userID && token.RevokedAt == nil {
			token.RevokedAt = &revokedAt
			token.RevocationReason = reason
		}
	}
	return nil
}

var _ repository.AuthTokenRepository = (*MemoryAuthTokenRepository)(nil)
