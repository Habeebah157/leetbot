package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)


type Config struct {
	DiscordToken string
	BotPrefix    string
}


func Load() (*Config, error) {

	_ = godotenv.Load()

	config := &Config{
		DiscordToken: getEnvVar("DISCORD_TOKEN", ""),
		BotPrefix:    getEnvVar("BOT_PREFIX", "!"),
	}

	if config.DiscordToken == "" {
		return nil, ErrMissingDiscordToken
	}

	return config, nil
}


func getEnvVar(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}


func (c *Config) Validate() error {
	if c.DiscordToken == "" {
		return ErrMissingDiscordToken
	}

	if c.BotPrefix == "" {
		log.Println("Warning: BOT_PREFIX is empty, using default '!'")
		c.BotPrefix = "!"
	}

	return nil
}

