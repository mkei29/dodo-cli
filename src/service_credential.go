package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/log"
	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
)

const (
	keyringService           = "dodo-doc"
	keyringUser              = "access_token"
	proactiveRefreshLeadTime = 5 * time.Minute
)

type storedToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
	TokenURL     string    `json:"token_url"`
}

func saveCredentials(token *oauth2.Token, tokenURL string) error {
	stored := storedToken{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
		TokenURL:     tokenURL,
	}
	data, err := json.Marshal(stored)
	if err != nil {
		return fmt.Errorf("failed to serialize credentials: %w", err)
	}
	if err := keyring.Set(keyringService, keyringUser, string(data)); err != nil {
		return fmt.Errorf("failed to save credentials to keyring: %w", err)
	}
	return nil
}

func loadCredentials() (string, error) {
	data, err := keyring.Get(keyringService, keyringUser)
	if err != nil {
		return "", fmt.Errorf("failed to load credentials from keyring: %w", err)
	}

	var stored storedToken
	_ = json.Unmarshal([]byte(data), &stored)
	if stored.AccessToken == "" {
		// Legacy format: plain access token string
		return data, nil
	}

	token := &oauth2.Token{
		AccessToken:  stored.AccessToken,
		RefreshToken: stored.RefreshToken,
		Expiry:       stored.Expiry,
	}

	if ttl, ok := jwtTTL(stored.AccessToken); ok {
		log.Debugf("access token TTL: %s", ttl.Round(time.Second))
	}

	needsRefresh := !token.Valid() || jwtExpiresWithin(stored.AccessToken, proactiveRefreshLeadTime)

	if !needsRefresh {
		return token.AccessToken, nil
	}

	if stored.RefreshToken == "" {
		if !token.Valid() {
			return "", errors.New("access token expired, please run 'dodo login' again")
		}
		// Expiring soon but no refresh token — return as-is
		return token.AccessToken, nil
	}

	log.Debugf("refreshing access token")
	newToken, err := refreshToken(token, stored.TokenURL)
	if err != nil {
		if !token.Valid() {
			return "", fmt.Errorf("failed to refresh access token, please run 'dodo login' again: %w", err)
		}
		// Proactive refresh failed — return current token as fallback
		log.Debugf("proactive token refresh failed, using current token: %v", err)
		return token.AccessToken, nil
	}

	log.Debugf("access token refreshed successfully")
	// Best-effort save; non-fatal if it fails
	_ = saveCredentials(newToken, stored.TokenURL)

	return newToken.AccessToken, nil
}

// jwtTTL returns the remaining lifetime of the JWT access token derived
// from its exp claim. The second return value is false if the token is
// not a JWT or has no exp claim.
func jwtTTL(accessToken string) (time.Duration, bool) {
	parts := strings.SplitN(accessToken, ".", 3)
	if len(parts) != 3 {
		return 0, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, false
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Exp == 0 {
		return 0, false
	}
	return time.Until(time.Unix(claims.Exp, 0)), true
}

// jwtExpiresWithin reports whether the JWT access token's exp claim
// falls within the given duration from now.
// Returns false if the token is not a JWT or has no exp claim.
func jwtExpiresWithin(accessToken string, d time.Duration) bool {
	ttl, ok := jwtTTL(accessToken)
	return ok && ttl < d
}

func refreshToken(token *oauth2.Token, tokenURL string) (*oauth2.Token, error) {
	cfg := &oauth2.Config{
		ClientID: oauthClientID,
		Endpoint: oauth2.Endpoint{
			TokenURL: tokenURL,
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	newToken, err := cfg.TokenSource(ctx, token).Token()
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}
	return newToken, nil
}
