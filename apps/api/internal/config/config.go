package config

import (
	"context"
	"log"

	_ "github.com/joho/godotenv/autoload"
	envconfig "github.com/sethvargo/go-envconfig"
)

type Config struct {
	AuthUrl        string   `env:"AUTH_URL"`
	Port           int      `env:"API_PORT, default=8002"`
	JWKSUrl        string   `env:"AUTH_JWKS_URL"`
	DatabaseUrl    string   `env:"API_DB_URL"`
	TrustedOrigins []string `env:"AUTH_TRUSTED_ORIGINS"`
	InternalKey    string   `env:"API_INTERNAL_KEY, required"`

	S3Endpoint       string `env:"S3_ENDPOINT"`
	S3AccessKey      string `env:"S3_ACCESS_KEY"`
	S3SecretKey      string `env:"S3_SECRET_KEY"`
	S3Region         string `env:"S3_REGION, default=us-east-1"`
	S3ForcePathStyle bool   `env:"S3_FORCE_PATH_STYLE, default=true"`
	S3Bucket         string `env:"S3_BUCKET, default=domus-finance"`
}

func GetConfig() Config {
	var config Config

	if err := envconfig.Process(context.Background(), &config); err != nil {
		log.Fatalf("Error while parsing config: %s", err)
	}

	if len(config.InternalKey) < 32 {
		log.Fatalf("API_INTERNAL_KEY must be at least 32 characters")
	}

	return config
}
