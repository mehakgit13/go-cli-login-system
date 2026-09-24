package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/example/go-cli-login/internal/auth"
	"github.com/example/go-cli-login/internal/config"
	"github.com/example/go-cli-login/internal/models"
	"github.com/example/go-cli-login/internal/session"
	"golang.org/x/term"
)

type App struct {
	auth        *auth.Service
	sessions    *session.Manager
	cfg         config.Config
	currentUser *models.User
	sessionID   string
}

func NewApp(a *auth.Service, s *session.Manager, c config.Config) *App {
	return &App{
		auth:     a,
		sessions: s,
		cfg:      c,
	}
}

func (a *App) Run() error {
	fmt.Println("==============================================")
	fmt.Println("  Go CLI Login System")
	fmt.Println("  Secure authentication + optional TOTP 2FA")
	fmt.Println("==============================================")
	fmt.Println("Type 'help' for commands.")

	for {
		prompt := "login> "
		if a.currentUser != nil {
			prompt = "app> "
		}

		rl, err := readline.NewEx(&readline.Config{
			Prompt:          prompt,
			HistoryFile:     "data/.history",
			InterruptPrompt: "^C",
			EOFPrompt:       "exit",
			AutoComplete:    commandCompleter(a.currentUser != nil),
		})
		if err != nil {
			return err
		}

		line, err := rl.Readline()
		rl.Close()

		if err != nil {
			if errors.Is(err, readline.ErrInterrupt) {
				fmt.Println()
				continue
			}
			return nil
		}

		cmd, args := split(line)

		if cmd == "" {
			continue
		}

		if err := a.handle(cmd, args); err != nil {
			fmt.Println("Error:", err)
		}

		if cmd == "exit" {
			return nil
		}
	}
}

// commandCompleterFunc implements readline.AutoCompleter.
//
// readline v1.5.1 expects:
//
//	Do([]rune, int) ([][]rune, int)
//
// The first return value contains possible completion candidates.
type commandCompleterFunc struct {
	loggedIn bool
}

func (c commandCompleterFunc) Do(line []rune, pos int) (newLines [][]rune, length int) {
	commands := []string{
		"help",
		"exit",
		"register",
		"login",
	}

	if c.loggedIn {
		commands = append(
			commands,
			"whoami",
			"enable-2fa",
			"disable-2fa",
			"logout",
		)
	}

	if pos < 0 || pos > len(line) {
		return nil, 0
	}

	prefix := string(line[:pos])

	for _, cmd := range commands {
		if strings.HasPrefix(cmd, prefix) {
			return [][]rune{
				[]rune(cmd),
			}, len(prefix)
		}
	}

	return nil, 0
}

func commandCompleter(loggedIn bool) readline.AutoCompleter {
	return commandCompleterFunc{
		loggedIn: loggedIn,
	}
}

func split(line string) (string, []string) {
	p := strings.Fields(line)

	if len(p) == 0 {
		return "", nil
	}

	return strings.ToLower(p[0]), p[1:]
}

func (a *App) handle(cmd string, args []string) error {
	if cmd == "exit" {
		fmt.Println("Goodbye.")
		return nil
	}

	// Check whether the current session is still valid.
	if a.currentUser != nil && cmd != "logout" {
		if _, err := a.sessions.Get(a.sessionID); err != nil {
			a.currentUser = nil
			a.sessionID = ""

			return errors.New("session expired; please login again")
		}
	}

	// Commands available before login.
	if a.currentUser == nil {
		switch cmd {
		case "help":
			fmt.Println("register, login, help, exit")

		case "register":
			return a.register()

		case "login":
			return a.login()

		default:
			fmt.Println("Unknown command. Type 'help'.")
		}

		return nil
	}

	// Commands available after login.
	switch cmd {
	case "help":
		fmt.Println("whoami, enable-2fa, disable-2fa, logout, help, exit")

	case "whoami":
		a.showUser()

	case "enable-2fa":
		return a.enable2FA()

	case "disable-2fa":
		return a.disable2FA()

	case "logout":
		return a.logout()

	default:
		fmt.Println("Unknown command. Type 'help'.")
	}

	return nil
}

func readLine(label string) (string, error) {
	fmt.Print(label)

	var v string

	_, err := fmt.Scanln(&v)

	return v, err
}

func readPassword(label string) (string, error) {
	fmt.Print(label)

	b, err := term.ReadPassword(int(os.Stdin.Fd()))

	fmt.Println()

	return string(b), err
}

func (a *App) register() error {
	u, err := readLine("Username: ")
	if err != nil {
		return err
	}

	p, err := readPassword("Password: ")
	if err != nil {
		return err
	}

	p2, err := readPassword("Confirm password: ")
	if err != nil {
		return err
	}

	if p != p2 {
		return errors.New("passwords do not match")
	}

	if err := a.auth.Register(u, p); err != nil {
		return err
	}

	fmt.Println("Registration successful.")

	return nil
}

func (a *App) login() error {
	u, err := readLine("Username: ")
	if err != nil {
		return err
	}

	p, err := readPassword("Password: ")
	if err != nil {
		return err
	}

	user, err := a.auth.Login(u, p, "")

	// If MFA is enabled, the first authentication attempt intentionally
	// requests the TOTP code.
	if errors.Is(err, auth.ErrMFARequired) {
		code, er := readLine("TOTP code: ")
		if er != nil {
			return er
		}

		user, err = a.auth.Login(
			u,
			p,
			strings.TrimSpace(code),
		)
	}

	if err != nil {
		return err
	}

	s, err := a.sessions.Create(user.ID)
	if err != nil {
		return err
	}

	a.currentUser = user
	a.sessionID = s.ID

	fmt.Println("Login successful.")
	a.showUser()

	return nil
}

func (a *App) showUser() {
	if a.currentUser == nil {
		return
	}

	fmt.Println("----- User Details -----")
	fmt.Println("Username:", a.currentUser.Username)

	fmt.Println(
		"Registration date:",
		a.currentUser.RegisteredAt.Local().Format(time.RFC1123),
	)

	if a.currentUser.MFAEnabled {
		fmt.Println("MFA status: enabled")
	} else {
		fmt.Println("MFA status: disabled")
	}

	if s, err := a.sessions.Get(a.sessionID); err == nil {
		fmt.Println(
			"Session expiration time:",
			s.ExpiresAt.Local().Format(time.RFC1123),
		)
	}

	if a.currentUser.LastLoginAt != nil {
		fmt.Println(
			"Last login time:",
			a.currentUser.LastLoginAt.Local().Format(time.RFC1123),
		)
	}

	fmt.Println("------------------------")
}

func (a *App) enable2FA() error {
	if a.currentUser.MFAEnabled {
		return errors.New("MFA is already enabled")
	}

	secret, uri, err := a.auth.EnableMFA(
		a.currentUser.ID,
		a.currentUser.Username,
	)
	if err != nil {
		return err
	}

	a.currentUser.MFAEnabled = true

	fmt.Println("2FA enabled.")
	fmt.Println(
		"Secret (enter this in Google Authenticator):",
		secret,
	)
	fmt.Println("OTP provisioning URI:", uri)
	fmt.Println("Log out and log back in to verify your TOTP code.")

	return nil
}

func (a *App) disable2FA() error {
	if !a.currentUser.MFAEnabled {
		return errors.New("MFA is already disabled")
	}

	if err := a.auth.DisableMFA(a.currentUser.ID); err != nil {
		return err
	}

	a.currentUser.MFAEnabled = false

	fmt.Println("2FA disabled.")

	return nil
}

func (a *App) logout() error {
	if a.sessionID != "" {
		_ = a.sessions.Delete(a.sessionID)
	}

	a.currentUser = nil
	a.sessionID = ""

	fmt.Println("Logged out.")

	return nil
}