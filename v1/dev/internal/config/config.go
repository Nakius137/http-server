package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
)

type Config struct {
	DatabaseURL string
}

func Load() (*Config, error) {
	err := requiredEnvs("DB_USER", "DB_HOST", "DB_PORT", "DB_PASSWORD", "DB_DATABASE")
	if err != nil {
		return nil, err
	}

	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD")),
		Host:   net.JoinHostPort(os.Getenv("DB_HOST"), os.Getenv("DB_PORT")),
		Path:   os.Getenv("DB_DATABASE"),
	}

	q := u.Query()
	q.Set("sslmode", "disable")

	u.RawQuery = q.Encode()
	databaseURL := u.String()

	return &Config{DatabaseURL: databaseURL}, nil
}

func requiredEnvs(envs ...string) error {
	for _, env := range envs {
		ok := os.Getenv(env)
		if len(ok) < 1 {
			return fmt.Errorf("missing env var: %s", env)
		}
	}

	return nil
}
