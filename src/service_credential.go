package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
)

const (
	keyringService = "dodo-doc"
	keyringUser    = "access_token"
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

	if token.Valid() {
		return token.AccessToken, nil
	}

	if stored.RefreshToken == "" {
		return "", errors.New("access token expired, please run 'dodo login' again")
	}

	newToken, err := refreshToken(token, stored.TokenURL)
	if err != nil {
		return "", fmt.Errorf("failed to refresh access token, please run 'dodo login' again: %w", err)
	}

	// Best-effort save; non-fatal if it fails
	_ = saveCredentials(newToken, stored.TokenURL)

	return newToken.AccessToken, nil
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
