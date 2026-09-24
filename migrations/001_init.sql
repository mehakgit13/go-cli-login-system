CREATE TABLE users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  registered_at DATETIME NOT NULL,
  last_login_at DATETIME,
  mfa_enabled INTEGER NOT NULL DEFAULT 0,
  mfa_secret TEXT,
  failed_attempts INTEGER NOT NULL DEFAULT 0,
  locked_until DATETIME
);

CREATE TABLE sessions (
  id TEXT PRIMARY KEY,
  user_id INTEGER NOT NULL,
  created_at DATETIME NOT NULL,
  expires_at DATETIME NOT NULL
);
