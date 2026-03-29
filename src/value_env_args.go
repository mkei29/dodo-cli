package main

import (
	"os"

	"github.com/caarlos0/log"
)

type EnvArgs struct {
	APIKey      string // from DODO_API_KEY env var
	AccessToken string // from OAuth2 login (keyring)
}

func NewEnvArgs() EnvArgs {
	env := EnvArgs{
		APIKey: os.Getenv("DODO_API_KEY"),
	}
	token, err := loadCredentials()
	if err != nil {
		log.Warnf("keyring credentials unavailable: %v", err)
		if env.APIKey != "" {
			log.Debugf("using auth: API key (DODO_API_KEY)")
		}
		return env
	}
	env.AccessToken = token
	log.Debugf("using auth: OAuth2 access token (keyring)")
	return env
}

// BearerToken returns the token to use for API authentication.
// AccessToken (from OAuth2 login) takes precedence over APIKey.
func (e EnvArgs) BearerToken() string {
	if e.AccessToken != "" {
		return e.AccessToken
	}
	return e.APIKey
}

// IsAuthenticated reports whether any auth credential is available.
func (e EnvArgs) IsAuthenticated() bool {
	return e.BearerToken() != ""
}
