package domain

import "errors"

var (
	ErrInvalidArgument     = errors.New("invalid argument")
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("conflict")
	ErrForbidden           = errors.New("forbidden")
	ErrUnauthenticated     = errors.New("unauthenticated")
	ErrInvalidTransition   = errors.New("invalid task status transition")
	ErrDependencyCycle     = errors.New("task dependency would create a cycle")
	ErrIdempotencyConflict = errors.New("idempotency key was reused with different input")
	ErrRateLimited         = errors.New("rate limit exceeded")
	ErrInvitationExpired   = errors.New("invitation expired")
)
