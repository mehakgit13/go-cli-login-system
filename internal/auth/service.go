package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/example/go-cli-login/internal/config"
	"github.com/example/go-cli-login/internal/db"
	"github.com/example/go-cli-login/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrLocked             = errors.New("account is temporarily locked")
	ErrMFARequired        = errors.New("two-factor authentication code required")
	ErrInvalidMFA         = errors.New("invalid two-factor authentication code")
	ErrWeakPassword       = errors.New("password does not meet minimum length")
)

type Service struct {
	repo *db.Repository
	cfg  config.Config
}

func NewService(repo *db.Repository, cfg config.Config) *Service {
	return &Service{repo: repo, cfg: cfg}
}

func (s *Service) Register(username, password string) error {
	username = strings.TrimSpace(username)
	if len(username) < 3 || len(username) > 64 {
		return errors.New("username must be 3-64 characters")
	}
	if len(password) < s.cfg.PasswordMinLength {
		return ErrWeakPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.CreateUser(username, string(hash), time.Now().UTC())
}

func (s *Service) Login(username, password, code string) (*models.User, error) {
	u, err := s.repo.GetUser(strings.TrimSpace(username))
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	now := time.Now().UTC()
	if u.LockedUntil != nil && u.LockedUntil.After(now) {
		return nil, fmt.Errorf("%w until %s", ErrLocked, u.LockedUntil.Local().Format(time.RFC1123))
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		s.recordFailure(u, now)
		return nil, ErrInvalidCredentials
	}
	if u.MFAEnabled {
		if code == "" {
			return nil, ErrMFARequired
		}
		secret, err := s.repo.GetMFASecret(u.ID)
		if err != nil {
			return nil, ErrInvalidMFA
		}
		if !validateTOTP(code, secret, now) {
			s.recordFailure(u, now)
			return nil, ErrInvalidMFA
		}
	}
	if err := s.repo.SetLoginSuccess(u.ID, now); err != nil {
		return nil, err
	}
	u.FailedAttempts = 0
	u.LockedUntil = nil
	u.LastLoginAt = &now
	return u, nil
}

func (s *Service) recordFailure(u *models.User, now time.Time) {
	attempts := u.FailedAttempts + 1
	var lock *time.Time
	if attempts >= s.cfg.MaxFailedAttempts {
		t := now.Add(s.cfg.LockoutDuration)
		lock = &t
	}
	_ = s.repo.SetLoginFailure(u.ID, attempts, lock)
}

func (s *Service) EnableMFA(userID int64, username string) (secret, uri string, err error) {
	b := make([]byte, 20)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	secret = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	q := url.Values{}
	q.Set("secret", secret)
	q.Set("issuer", "Go CLI Login")
	q.Set("algorithm", "SHA1")
	q.Set("digits", "6")
	q.Set("period", "30")
	uri = "otpauth://totp/Go%20CLI%20Login:" + url.PathEscape(username) + "?" + q.Encode()
	if err = s.repo.SetMFA(userID, true, &secret); err != nil {
		return "", "", err
	}
	return secret, uri, nil
}
func (s *Service) DisableMFA(userID int64) error { return s.repo.SetMFA(userID, false, nil) }

func RandomSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func validateTOTP(code, secret string, now time.Time) bool {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return false
	}
	if _, err := strconv.Atoi(code); err != nil {
		return false
	}
	raw, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(strings.TrimSpace(secret)))
	if err != nil {
		return false
	}
	for offset := int64(-1); offset <= 1; offset++ {
		counter := uint64(now.Unix()/30 + offset)
		msg := make([]byte, 8)
		for i := 7; i >= 0; i-- {
			msg[i] = byte(counter)
			counter >>= 8
		}
		mac := hmac.New(sha1.New, raw)
		_, _ = mac.Write(msg)
		sum := mac.Sum(nil)
		idx := sum[len(sum)-1] & 0x0f
		bin := (uint32(sum[idx])&0x7f)<<24 | uint32(sum[idx+1])<<16 | uint32(sum[idx+2])<<8 | uint32(sum[idx+3])
		expected := fmt.Sprintf("%06d", bin%1000000)
		if subtle.ConstantTimeCompare([]byte(expected), []byte(code)) == 1 {
			return true
		}
	}
	return false
}
