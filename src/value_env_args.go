package main

import (
	"os"

	"github.com/caarlos0/log"
)

type EnvArgs struct {
	APIKey string
}

func NewEnvArgs() EnvArgs {
	if key := os.Getenv("DODO_API_KEY"); key != "" {
		return EnvArgs{APIKey: key}
	}

	token, err := loadCredentials()
	if err != nil {
		log.Debugf("no credentials found in keyring: %v", err)
		return EnvArgs{}
	}
	return EnvArgs{APIKey: token}
}
