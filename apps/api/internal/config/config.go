package config

import (
	"context"
	"log"

	_ "github.com/joho/godotenv/autoload"
	envconfig "github.com/sethvargo/go-envconfig"
)

type Config struct {
	Port           int      `env:"API_PORT, default=8002"`
	JWKSUrl        string   `env:"AUTH_JWKS_URL"`
	DatabaseUrl    string   `env:"API_DB_URL"`
	TrustedOrigins []string `env:"AUTH_TRUSTED_ORIGINS"`
}

func GetConfig() Config {
	var config Config

	if err := envconfig.Process(context.Background(), &config); err != nil {
		log.Fatalf("Error while parsing config: %s", err)
	}

	return config
}
