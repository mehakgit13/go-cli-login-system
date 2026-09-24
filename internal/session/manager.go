package session

import (
	"errors"
	"time"

	"github.com/example/go-cli-login/internal/auth"
	"github.com/example/go-cli-login/internal/config"
	"github.com/example/go-cli-login/internal/db"
	"github.com/example/go-cli-login/internal/models"
)

var ErrExpired = errors.New("session expired")

type Manager struct {
	repo *db.Repository
	cfg  config.Config
}

func NewManager(repo *db.Repository, cfg config.Config) *Manager {
	return &Manager{repo: repo, cfg: cfg}
}

func (m *Manager) Create(userID int64) (*models.Session, error) {
	now := time.Now().UTC()
	id, err := auth.RandomSessionID()
	if err != nil {
		return nil, err
	}
	s := &models.Session{ID: id, UserID: userID, CreatedAt: now, ExpiresAt: now.Add(m.cfg.SessionTimeout)}
	if err := m.repo.CreateSession(*s); err != nil {
		return nil, err
	}
	return s, nil
}
func (m *Manager) Get(id string) (*models.Session, error) {
	s, err := m.repo.GetSession(id)
	if err != nil {
		return nil, err
	}
	if !s.ExpiresAt.After(time.Now().UTC()) {
		_ = m.repo.DeleteSession(id)
		return nil, ErrExpired
	}
	return s, nil
}
func (m *Manager) Delete(id string) error { return m.repo.DeleteSession(id) }
