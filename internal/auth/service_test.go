package auth

import (
	"database/sql"
	"os"
	"testing"

	"github.com/example/go-cli-login/internal/config"
	"github.com/example/go-cli-login/internal/db"
	_ "github.com/mattn/go-sqlite3"
)

func testService(t *testing.T) *Service {
	t.Helper()
	f, err := os.CreateTemp("", "auth-*.db")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.Close()
	t.Cleanup(func() { os.Remove(path) })
	d, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	if err := db.Migrate(d); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{DatabasePath: path, PasswordMinLength: 8, MaxFailedAttempts: 3, LockoutDuration: 100000000}
	return NewService(db.NewRepository(d), cfg)
}
func TestRegisterAndLogin(t *testing.T) {
	s := testService(t)
	if err := s.Register("alice", "password123"); err != nil {
		t.Fatal(err)
	}
	u, err := s.Login("alice", "password123", "")
	if err != nil {
		t.Fatal(err)
	}
	if u.Username != "alice" {
		t.Fatalf("unexpected user %q", u.Username)
	}
}
func TestWeakPassword(t *testing.T) {
	s := testService(t)
	if err := s.Register("alice", "short"); err != ErrWeakPassword {
		t.Fatalf("got %v", err)
	}
}
