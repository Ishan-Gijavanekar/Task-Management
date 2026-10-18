package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"
	"net/url"
)

type AuthNotificationSender interface {
	SendEmailVerification(ctx context.Context, email, token string) error
	SendPasswordReset(ctx context.Context, email, token string) error
}

type SMTPAuthNotifier struct {
	host, port, username, password, from, frontendURL string
}

func NewSMTPAuthNotifier(host, port, username, password, from, frontendURL string) *SMTPAuthNotifier {
	return &SMTPAuthNotifier{host: host, port: port, username: username, password: password, from: from, frontendURL: frontendURL}
}

func (n *SMTPAuthNotifier) send(to, subject, link string) error {
	if n.host == "" || n.port == "" || n.from == "" {
		return fmt.Errorf("SMTP notifier is not configured")
	}
	message := []byte("From: " + n.from + "\r\nTo: " + to + "\r\nSubject: " + subject +
		"\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n" + link + "\r\n")
	var smtpAuth smtp.Auth
	if n.username != "" {
		smtpAuth = smtp.PlainAuth("", n.username, n.password, n.host)
	}
	if err := smtp.SendMail(n.host+":"+n.port, smtpAuth, n.from, []string{to}, message); err != nil {
		return fmt.Errorf("send authentication email: %w", err)
	}
	return nil
}

func (n *SMTPAuthNotifier) SendEmailVerification(_ context.Context, email, token string) error {
	return n.send(email, "Verify your email", n.frontendURL+"/verify-email?token="+url.QueryEscape(token))
}

func (n *SMTPAuthNotifier) SendPasswordReset(_ context.Context, email, token string) error {
	return n.send(email, "Reset your password", n.frontendURL+"/reset-password?token="+url.QueryEscape(token))
}

// DevelopmentAuthNotifier logs local-development links. Production deployments
// should provide an email-provider implementation instead.
type DevelopmentAuthNotifier struct {
	logger      *slog.Logger
	frontendURL string
}

func NewDevelopmentAuthNotifier(logger *slog.Logger, frontendURL string) *DevelopmentAuthNotifier {
	return &DevelopmentAuthNotifier{logger: logger, frontendURL: frontendURL}
}

func (n *DevelopmentAuthNotifier) SendEmailVerification(_ context.Context, email, token string) error {
	n.logger.Info("development email verification", "email", email, "url", n.frontendURL+"/verify-email?token="+url.QueryEscape(token))
	return nil
}

func (n *DevelopmentAuthNotifier) SendPasswordReset(_ context.Context, email, token string) error {
	n.logger.Info("development password reset", "email", email, "url", n.frontendURL+"/reset-password?token="+url.QueryEscape(token))
	return nil
}
