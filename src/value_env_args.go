package main

import "os"

type EnvArgs struct {
	APIKey string
}

func NewEnvArgs() EnvArgs {
	return EnvArgs{
		APIKey: os.Getenv("DODO_API_KEY"),
	}
}
