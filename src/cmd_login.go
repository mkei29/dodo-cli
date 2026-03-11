package main

import (
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

//go:embed templates/login_success.html
var loginSuccessHTML string

//go:embed templates/login_error.html
var loginErrorHTML string

const (
	oauthClientID     = "cli"
	oauthAuthURL      = "https://api.dodo-doc.com/oauth2/auth"
	oauthTokenURL     = "https://api.dodo-doc.com/oauth2/token"
	oauthCallbackPath = "/callback"
	oauthTimeout      = 5 * time.Minute
)

var oauthEndpoint = oauth2.Endpoint{ //nolint:gochecknoglobals
	AuthURL:  oauthAuthURL,
	TokenURL: oauthTokenURL,
}

type LoginArgs struct {
	debug   bool
	noColor bool
}

func (opts *LoginArgs) DisableLogging() bool  { return false }
func (opts *LoginArgs) EnableDebugMode() bool { return opts.debug }
func (opts *LoginArgs) EnableColor() bool     { return !opts.noColor }

func CreateLoginCmd() *cobra.Command {
	opts := LoginArgs{}
	cmd := &cobra.Command{
		Use:           "login",
		Short:         "Login to dodo-doc using OAuth2",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, _ []string) error {
			printer := NewErrorPrinter(ErrorLevel)
			if err := InitLogger(&opts); err != nil {
				return printer.HandleError(err)
			}
			if err := loginCmdEntrypoint(); err != nil {
				return printer.HandleError(err)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&opts.debug, "debug", false, "Enable debug mode")
	cmd.Flags().BoolVar(&opts.noColor, "no-color", false, "Disable color output")
	return cmd
}

func loginCmdEntrypoint() error {
	// Generate PKCE verifier (oauth2 package handles base64url encoding)
	verifier := oauth2.GenerateVerifier()

	// Generate state for CSRF protection
	state, err := generateOAuthState()
	if err != nil {
		return fmt.Errorf("failed to generate state: %w", err)
	}

	// Find an available random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to find available port: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://localhost:%d%s", port, oauthCallbackPath)

	cfg := &oauth2.Config{
		ClientID:    oauthClientID,
		Scopes:      []string{"read", "write"},
		Endpoint:    oauthEndpoint,
		RedirectURL: redirectURI,
	}

	// Build authorization URL with PKCE S256 challenge
	authURL := cfg.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))
	log.Infof("opening browser to authorize...")
	log.Infof("if browser does not open, please visit: %s", authURL)
	if err := openBrowser(authURL); err != nil {
		log.Debugf("failed to open browser: %v", err)
	}

	// Wait for the OAuth2 callback
	code, receivedState, err := waitForCallback(listener, oauthTimeout)
	if err != nil {
		return fmt.Errorf("authorization failed: %w", err)
	}
	if receivedState != state {
		return errors.New("state mismatch: possible CSRF attack")
	}

	// Exchange authorization code for token (oauth2 package handles PKCE verifier)
	log.Debugf("exchanging authorization code for token")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	token, err := cfg.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Save credentials to disk
	if err := saveCredentials(token); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	log.Infof("successfully logged in!")
	return nil
}

func generateOAuthState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to read random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

type callbackResult struct {
	code  string
	state string
	err   error
}

func waitForCallback(listener net.Listener, timeout time.Duration) (string, string, error) {
	resultCh := make(chan callbackResult, 1)

	mux := http.NewServeMux()
	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}

	successTmpl := template.Must(template.New("success").Parse(loginSuccessHTML))
	errorTmpl := template.Must(template.New("error").Parse(loginErrorHTML))

	mux.HandleFunc(oauthCallbackPath, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if errMsg := q.Get("error"); errMsg != "" {
			desc := q.Get("error_description")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			errorTmpl.Execute(w, map[string]string{"Error": errMsg, "Description": desc}) //nolint:errcheck
			resultCh <- callbackResult{err: fmt.Errorf("%s: %s", errMsg, desc)}
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		successTmpl.Execute(w, nil) //nolint:errcheck
		resultCh <- callbackResult{code: q.Get("code"), state: q.Get("state")}
	})

	go func() {
		srv.Serve(listener) //nolint:errcheck
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case result := <-resultCh:
		srv.Shutdown(context.Background()) //nolint:errcheck
		return result.code, result.state, result.err
	case <-ctx.Done():
		srv.Shutdown(context.Background()) //nolint:errcheck
		return "", "", fmt.Errorf("timed out waiting for authorization (timeout: %s)", timeout)
	}
}

type credentials struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
	RefreshToken string    `json:"refresh_token,omitempty"`
}

func credentialsPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config dir: %w", err)
	}
	return filepath.Join(configDir, "dodo", "credentials.json"), nil
}

func saveCredentials(token *oauth2.Token) error {
	creds := credentials{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		ExpiresAt:    token.Expiry,
		RefreshToken: token.RefreshToken,
	}

	path, err := credentialsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("failed to create credentials directory: %w", err)
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}
	return nil
}
