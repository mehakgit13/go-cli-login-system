package db

import (
	"database/sql"
	"errors"
	"time"

	"github.com/example/go-cli-login/internal/models"
)

var ErrNotFound = errors.New("not found")

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) CreateUser(username, passwordHash string, registeredAt time.Time) error {
	_, err := r.db.Exec(`INSERT INTO users(username,password_hash,registered_at) VALUES(?,?,?)`,
		username, passwordHash, registeredAt)
	return err
}

func (r *Repository) GetUser(username string) (*models.User, error) {
	var u models.User
	var last, locked sql.NullTime
	var mfa int
	err := r.db.QueryRow(`SELECT id,username,password_hash,registered_at,last_login_at,mfa_enabled,failed_attempts,locked_until
		FROM users WHERE username = ?`, username).
		Scan(&u.ID, &u.Username, &u.PasswordHash, &u.RegisteredAt, &last, &mfa, &u.FailedAttempts, &locked)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if last.Valid {
		u.LastLoginAt = &last.Time
	}
	if locked.Valid {
		u.LockedUntil = &locked.Time
	}
	u.MFAEnabled = mfa != 0
	return &u, nil
}

func (r *Repository) SetLoginSuccess(userID int64, t time.Time) error {
	_, err := r.db.Exec(`UPDATE users SET failed_attempts=0, locked_until=NULL, last_login_at=? WHERE id=?`, t, userID)
	return err
}
func (r *Repository) SetLoginFailure(userID int64, attempts int, lockedUntil *time.Time) error {
	_, err := r.db.Exec(`UPDATE users SET failed_attempts=?, locked_until=? WHERE id=?`, attempts, lockedUntil, userID)
	return err
}
func (r *Repository) SetMFA(userID int64, enabled bool, secret *string) error {
	v := 0
	if enabled {
		v = 1
	}
	_, err := r.db.Exec(`UPDATE users SET mfa_enabled=?,mfa_secret=? WHERE id=?`, v, secret, userID)
	return err
}
func (r *Repository) GetMFASecret(userID int64) (string, error) {
	var s sql.NullString
	err := r.db.QueryRow(`SELECT mfa_secret FROM users WHERE id=?`, userID).Scan(&s)
	if err != nil {
		return "", err
	}
	if !s.Valid {
		return "", ErrNotFound
	}
	return s.String, nil
}

func (r *Repository) CreateSession(s models.Session) error {
	_, err := r.db.Exec(`INSERT INTO sessions(id,user_id,created_at,expires_at) VALUES(?,?,?,?)`,
		s.ID, s.UserID, s.CreatedAt, s.ExpiresAt)
	return err
}
func (r *Repository) GetSession(id string) (*models.Session, error) {
	var s models.Session
	err := r.db.QueryRow(`SELECT id,user_id,created_at,expires_at FROM sessions WHERE id=?`, id).
		Scan(&s.ID, &s.UserID, &s.CreatedAt, &s.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}
func (r *Repository) DeleteSession(id string) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE id=?`, id)
	return err
}
func (r *Repository) DeleteExpiredSessions(now time.Time) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE expires_at <= ?`, now)
	return err
}
