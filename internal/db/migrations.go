package db

import "database/sql"

func Migrate(db *sql.DB) error {
	const schema = `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE COLLATE NOCASE,
    password_hash TEXT NOT NULL,
    registered_at DATETIME NOT NULL,
    last_login_at DATETIME,
    mfa_enabled INTEGER NOT NULL DEFAULT 0,
    mfa_secret TEXT,
    failed_attempts INTEGER NOT NULL DEFAULT 0,
    locked_until DATETIME
);

CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL,
    expires_at DATETIME NOT NULL,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);
`
	_, err := db.Exec(schema)
	return err
}
