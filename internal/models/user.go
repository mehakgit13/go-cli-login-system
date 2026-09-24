package models

import "time"

type User struct {
	ID             int64
	Username       string
	PasswordHash   string
	RegisteredAt   time.Time
	LastLoginAt    *time.Time
	MFAEnabled     bool
	FailedAttempts int
	LockedUntil    *time.Time
}

type Session struct {
	ID        string
	UserID    int64
	CreatedAt time.Time
	ExpiresAt time.Time
}
