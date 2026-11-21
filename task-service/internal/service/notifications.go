package service

import (
	"context"
	"errors"

	"github.com/Ishan-Gijavanekar/task-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Service) ListNotifications(ctx context.Context, actor string, unread bool, limit int64) ([]*domain.Notification, error) {
	if err := s.identities.Validate(ctx, actor); err != nil {
		return nil, err
	}
	return s.repo.ListNotifications(ctx, actor, unread, pageLimit(limit))
}

func (s *Service) MarkNotificationsRead(ctx context.Context, actor, notificationID string) error {
	if err := s.identities.Validate(ctx, actor); err != nil {
		return err
	}
	var id *primitive.ObjectID
	if notificationID != "" {
		parsed, err := parseID(notificationID, "notificationId")
		if err != nil {
			return err
		}
		id = &parsed
	}
	return s.repo.MarkNotificationRead(ctx, actor, id, s.now().UTC())
}

func (s *Service) NotificationPreferences(ctx context.Context, actor string) (*domain.NotificationPreferences, error) {
	preferences, err := s.repo.GetNotificationPreferences(ctx, actor)
	if errors.Is(err, domain.ErrNotFound) {
		return &domain.NotificationPreferences{UserID: actor, Mentions: true, TaskUpdates: true}, nil
	}
	return preferences, err
}

func (s *Service) UpdateNotificationPreferences(ctx context.Context, actor string, preferences domain.NotificationPreferences) (*domain.NotificationPreferences, error) {
	preferences.UserID = actor
	if err := s.repo.SetNotificationPreferences(ctx, &preferences); err != nil {
		return nil, err
	}
	return &preferences, nil
}
