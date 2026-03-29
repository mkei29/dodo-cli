package main

import (
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/http"
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
	oauthClientID       = "cli"
	oauthAuthPath       = "/oauth2/auth"
	oauthTokenPath      = "/oauth2/token"
	oauthCallbackPath   = "/callback"
	oauthTimeout        = 1 * time.Minute
	oauthDefaultBaseURL = "https://api.dodo-doc.com"
)

type LoginArgs struct {
	endpoint string
	debug    bool
	noColor  bool
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			printer := NewErrorPrinter(ErrorLevel)
			if err := InitLogger(&opts); err != nil {
				return printer.HandleError(err)
			}
			if err := loginCmdEntrypoint(cmd.Context(), opts); err != nil {
				return printer.HandleError(err)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&opts.debug, "debug", false, "Enable debug mode")
	cmd.Flags().BoolVar(&opts.noColor, "no-color", false, "Disable color output")
	cmd.Flags().StringVar(&opts.endpoint, "endpoint", oauthDefaultBaseURL, "Base URL for OAuth2 endpoints (for testing only)")
	return cmd
}

func loginCmdEntrypoint(ctx context.Context, args LoginArgs) error {
	if args.endpoint != oauthDefaultBaseURL {
		log.Warnf("using non-default endpoint: %s", args.endpoint)
	}

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
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d%s", port, oauthCallbackPath)

	cfg := &oauth2.Config{
		ClientID: oauthClientID,
		Scopes:   []string{"read", "write", "offline_access"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  args.endpoint + oauthAuthPath,
			TokenURL: args.endpoint + oauthTokenPath,
		},
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
	code, receivedState, err := waitForCallback(ctx, listener, oauthTimeout)
	if err != nil {
		return fmt.Errorf("authorization failed: %w", err)
	}
	if receivedState != state {
		return errors.New("state mismatch: possible CSRF attack")
	}

	// Exchange authorization code for token (oauth2 package handles PKCE verifier)
	log.Debugf("exchanging authorization code for token")
	exchangeCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	token, err := cfg.Exchange(exchangeCtx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Save credentials to keyring
	if err := saveCredentials(token, cfg.Endpoint.TokenURL); err != nil {
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

func waitForCallback(ctx context.Context, listener net.Listener, timeout time.Duration) (string, string, error) {
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

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case result := <-resultCh:
		srv.Shutdown(context.Background()) //nolint:errcheck,contextcheck
		return result.code, result.state, result.err
	case <-timeoutCtx.Done():
		srv.Shutdown(context.Background()) //nolint:errcheck,contextcheck
		return "", "", fmt.Errorf("timed out waiting for authorization (timeout: %s)", timeout)
	}
}
