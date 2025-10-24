package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {

	os.Setenv("DISCORD_TOKEN", "test-token")
	os.Setenv("BOT_PREFIX", "!")
	defer func() {
		os.Unsetenv("DISCORD_TOKEN")
		os.Unsetenv("BOT_PREFIX")
	}()

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if config.DiscordToken != "test-token" {
		t.Errorf("Load() DiscordToken = %v, want %v", config.DiscordToken, "test-token")
	}

	if config.BotPrefix != "!" {
		t.Errorf("Load() BotPrefix = %v, want %v", config.BotPrefix, "!")
	}
}

func TestLoad_MissingToken(t *testing.T) {
	// clear environment to simulate missing token
	os.Unsetenv("DISCORD_TOKEN")
	os.Unsetenv("BOT_PREFIX")

	_, err := Load()
	if err != ErrMissingDiscordToken {
		t.Errorf("Load() error = %v, want %v", err, ErrMissingDiscordToken)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				DiscordToken: "test-token",
				BotPrefix:    "!",
			},
			wantErr: false,
		},
		{
			name: "missing token",
			config: &Config{
				DiscordToken: "",
				BotPrefix:    "!",
			},
			wantErr: true,
		},
		{
			name: "empty prefix",
			config: &Config{
				DiscordToken: "test-token",
				BotPrefix:    "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.config.BotPrefix == "" {
				if tt.config.BotPrefix != "!" {
					t.Errorf("Validate() should set default prefix to '!', got %v", tt.config.BotPrefix)
				}
			}
		})
	}
}
