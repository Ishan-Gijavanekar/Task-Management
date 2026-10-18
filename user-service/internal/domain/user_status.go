package domain

type UserStatus string

const (
	UserStatusPendingVerification UserStatus = "pending_verification"
	UserStatusActive              UserStatus = "active"
	UserStatusSuspended           UserStatus = "suspended"
	UserStatusDeleted             UserStatus = "deleted"
)

func (s UserStatus) IsValid() bool {
	switch s {
	case UserStatusPendingVerification, UserStatusActive, UserStatusSuspended, UserStatusDeleted:
		return true
	default:
		return false
	}
}
