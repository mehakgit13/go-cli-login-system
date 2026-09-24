package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabasePath      string
	SessionTimeout    time.Duration
	MaxFailedAttempts int
	LockoutDuration   time.Duration
	PasswordMinLength int
	CookieSecret      string
}

func Load() Config {
	return Config{
		DatabasePath:      getenv("DB_PATH", "data/app.db"),
		SessionTimeout:    durationEnv("SESSION_TIMEOUT", 30*time.Minute),
		MaxFailedAttempts: intEnv("MAX_FAILED_ATTEMPTS", 5),
		LockoutDuration:   durationEnv("LOCKOUT_DURATION", 15*time.Minute),
		PasswordMinLength: intEnv("PASSWORD_MIN_LENGTH", 8),
		CookieSecret:      getenv("SESSION_SECRET", "change-me-in-production"),
	}
}

func getenv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
func intEnv(k string, fallback int) int {
	v, err := strconv.Atoi(os.Getenv(k))
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}
func durationEnv(k string, fallback time.Duration) time.Duration {
	v, err := time.ParseDuration(os.Getenv(k))
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}
