package main

import (
	"fmt"

	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
)

const (
	keyringService = "dodo-doc"
	keyringUser    = "access_token"
)

func saveCredentials(token *oauth2.Token) error {
	if err := keyring.Set(keyringService, keyringUser, token.AccessToken); err != nil {
		return fmt.Errorf("failed to save credentials to keyring: %w", err)
	}
	return nil
}

func loadCredentials() (string, error) {
	token, err := keyring.Get(keyringService, keyringUser)
	if err != nil {
		return "", fmt.Errorf("failed to load credentials from keyring: %w", err)
	}
	return token, nil
}
