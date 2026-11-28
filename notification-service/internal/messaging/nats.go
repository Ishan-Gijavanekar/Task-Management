package messaging

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Client struct {
	Connection *nats.Conn
	JetStream  jetstream.JetStream
}

func Connect(ctx context.Context, url string, logger *slog.Logger) (*Client, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	nc, err := nats.Connect(
		url,
		nats.Name("notification-service"),
		nats.Timeout(5*time.Second),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
		nats.DrainTimeout(5*time.Second),
		nats.DisconnectErrHandler(
			func(_ *nats.Conn, err error) {
				logger.Warn("nats disconnected", "error", err)
			},
		),
		nats.ReconnectHandler(
			func(conn *nats.Conn) {
				logger.Info("nats reconnected")
			},
		),
		nats.ClosedHandler(
			func(conn *nats.Conn) {
				logger.Info("nats connection closed", "error", conn.LastError())
			},
		),

		nats.ErrorHandler(
			func(_ *nats.Conn, _ *nats.Subscription, err error) {
				logger.Error("nats asynchronous error", "error", err)
			},
		),
	)

	if err != nil {
		return nil, fmt.Errorf("connect to nats: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("initialize jetstream %w", err)
	}

	client := Client{
		Connection: nc,
		JetStream:  js,
	}

	if err := client.Ping(ctx); err != nil {
		nc.Close()
		return nil, err
	}

	return &client, nil
}

func (c *Client) Ping(ctx context.Context) error {
	if !c.Connection.IsConnected() {
		return fmt.Errorf("nats is not connected")
	}

	if err := c.Connection.FlushWithContext(ctx); err != nil {
		return fmt.Errorf("nats ping: %w", err)
	}

	if _, err := c.JetStream.AccountInfo(ctx); err != nil {
		return fmt.Errorf("jetstream unavailable: %w", err)
	}

	return nil
}

func (c *Client) Close() error {
	err := c.Connection.Drain()

	c.Connection.Close()

	return err
}
