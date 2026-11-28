package messaging

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

const (
	DefaultStreamName = "DOMAIN_EVENTS"
	DomainSubject     = "events.>"
)

func EnsureDomainStream(ctx context.Context, js jetstream.JetStream, streamName string) error {
	_, err := js.CreateOrUpdateStream(
		ctx,
		jetstream.StreamConfig{
			Name:        streamName,
			Description: "Task Management Domain Services",
			Subjects: []string{
				DomainSubject,
			},
			Retention:  jetstream.LimitsPolicy,
			Storage:    jetstream.FileStorage,
			Discard:    jetstream.DiscardOld,
			MaxAge:     7 * 24 * time.Hour,
			MaxBytes:   1024 * 1024 * 1024,
			MaxMsgSize: 256 * 1024,
			Replicas:   1,
			Duplicates: 10 * time.Minute,
		},
	)

	if err != nil {
		return fmt.Errorf("initialize domain event stream: %w", err)
	}

	return nil
}
