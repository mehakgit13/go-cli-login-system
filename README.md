# Containerized CLI Login System with Optional 2FA

A Go command-line authentication system implementing the assignment requirements: registration, bcrypt password hashing, optional TOTP MFA compatible with Google Authenticator, account lockout, session timeout, persistent SQLite storage, Docker containerization, and an interactive CLI.

## Requirements

- Go 1.22+ for local development
- Docker + Docker Compose

## Run with Docker

```bash
docker compose build
docker compose run --rm cli
```

The SQLite database is stored in the named Docker volume `app-data`, so it survives container recreation.

## Commands

Before login:

```text
register
login
help
exit
```

After login:

```text
whoami
enable-2fa
disable-2fa
logout
help
exit
```

## Example

1. Run the application.
2. `register`
3. Create a username and password of at least 8 characters.
4. `login`
5. `enable-2fa`
6. Add the displayed secret/URI to Google Authenticator.
7. `logout`
8. `login` again and provide the six-digit TOTP code.

## Security

- Passwords are stored using bcrypt hashes, never plaintext.
- Failed logins are counted.
- After 5 failed attempts, the account is locked for 15 minutes by default.
- Sessions use cryptographically random opaque IDs.
- Sessions expire after 30 minutes by default.
- TOTP uses the standard 30-second period and SHA-1 six-digit TOTP codes used by common authenticator applications; verification accepts the adjacent time windows to tolerate normal clock drift.
- SQLite foreign keys are enabled.

Configuration can be changed through environment variables:

| Variable | Default |
|---|---|
| DB_PATH | data/app.db |
| SESSION_TIMEOUT | 30m |
| MAX_FAILED_ATTEMPTS | 5 |
| LOCKOUT_DURATION | 15m |
| PASSWORD_MIN_LENGTH | 8 |
| SESSION_SECRET | change-me-in-production |

## Local development

Because SQLite uses CGO:

```bash
go mod tidy
go run ./cmd/cli
```

## Tests

Run:

```bash
go test ./...
```

## Project structure

```text
.
├── cmd/cli/main.go
├── internal/
│   ├── auth/
│   ├── cli/
│   ├── config/
│   ├── db/
│   ├── models/
│   └── session/
├── migrations/001_init.sql
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── README.md
```

## Notes

The CLI stores readline history in `data/.history`. The application is intentionally organized into authentication, persistence, sessions, configuration, models, and CLI layers so the security-sensitive logic is not mixed into the user interface.

## Submission checklist

- [x] Go source code
- [x] Registration and login
- [x] bcrypt password hashing
- [x] Optional TOTP 2FA
- [x] Failed-attempt lockout
- [x] Configurable session timeout
- [x] SQLite persistence
- [x] Dockerfile and Docker Compose
- [x] Interactive CLI with history and tab completion
- [x] Help/error/success feedback
- [x] Schema/migration
- [x] Unit tests
- [x] README
