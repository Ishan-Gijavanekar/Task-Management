package domain

type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

func (r UserRole) IsValid() bool {
	switch r {
	case UserRoleUser, UserRoleAdmin:
		return true
	default:
		return false
	}
}
