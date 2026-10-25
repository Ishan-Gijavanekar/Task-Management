package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Ishan-Gijavanekar/task-service/internal/domain"
	"github.com/Ishan-Gijavanekar/task-service/internal/policy"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type membershipReader interface {
	GetMembership(context.Context, primitive.ObjectID, string) (*domain.Membership, error)
}

type cachedMembership struct {
	membership *domain.Membership
	expiresAt  time.Time
}

type Authorizer struct {
	repo membershipReader
	ttl  time.Duration
	now  func() time.Time
	mu   sync.RWMutex
	data map[string]cachedMembership
}

func NewAuthorizer(repo membershipReader, ttl time.Duration) *Authorizer {
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	return &Authorizer{repo: repo, ttl: ttl, now: time.Now, data: map[string]cachedMembership{}}
}

func membershipKey(workspaceID primitive.ObjectID, userID string) string {
	return workspaceID.Hex() + ":" + userID
}

func (a *Authorizer) Membership(ctx context.Context, workspaceID primitive.ObjectID, userID string) (*domain.Membership, error) {
	key := membershipKey(workspaceID, userID)
	a.mu.RLock()
	entry, ok := a.data[key]
	a.mu.RUnlock()
	if ok && a.now().Before(entry.expiresAt) {
		copy := *entry.membership
		return &copy, nil
	}
	membership, err := a.repo.GetMembership(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	copy := *membership
	a.mu.Lock()
	a.data[key] = cachedMembership{membership: &copy, expiresAt: a.now().Add(a.ttl)}
	a.mu.Unlock()
	return membership, nil
}

func (a *Authorizer) Require(ctx context.Context, workspaceID primitive.ObjectID, userID string, action policy.Action) (*domain.Membership, error) {
	membership, err := a.Membership(ctx, workspaceID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrForbidden
		}
		return nil, err
	}
	if !policy.Allows(membership, action) {
		return nil, domain.ErrForbidden
	}
	return membership, nil
}

func (a *Authorizer) Invalidate(workspaceID primitive.ObjectID, userIDs ...string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, userID := range userIDs {
		delete(a.data, membershipKey(workspaceID, userID))
	}
}
