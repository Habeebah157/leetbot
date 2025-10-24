package config

import "errors"

var (
	ErrMissingDiscordToken = errors.New("DISCORD_TOKEN environment variable is required")
)
