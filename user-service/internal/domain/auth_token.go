package domain

import "time"

type AuthTokenType string

const (
	AuthTokenRefresh       AuthTokenType = "refresh"
	AuthTokenPasswordReset AuthTokenType = "password_reset"
	AuthTokenEmailVerify   AuthTokenType = "email_verify"
)

type AuthToken struct {
	ID               string        `bson:"_id"`
	UserID           string        `bson:"userId"`
	TokenHash        string        `bson:"tokenHash"`
	FamilyID         string        `bson:"familyId,omitempty"`
	Type             AuthTokenType `bson:"type"`
	Device           string        `bson:"device,omitempty"`
	IPAddress        string        `bson:"ipAddress,omitempty"`
	CreatedAt        time.Time     `bson:"createdAt"`
	ExpiresAt        time.Time     `bson:"expiresAt"`
	LastUsedAt       *time.Time    `bson:"lastUsedAt,omitempty"`
	RevokedAt        *time.Time    `bson:"revokedAt,omitempty"`
	RevocationReason string        `bson:"revocationReason,omitempty"`
}

func (t *AuthToken) Active(now time.Time) bool {
	return t != nil && t.RevokedAt == nil && now.Before(t.ExpiresAt)
}
