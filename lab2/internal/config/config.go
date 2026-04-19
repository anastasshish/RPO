package config

import "os"

type Config struct {
	AppAddr        string
	DBPath         string
	JWTSecret      string
	MigrationsPath string
}

func FromEnv() Config {
	return Config{
		AppAddr:        getEnv("APP_ADDR", ":8080"),
		DBPath:         getEnv("DB_PATH", "./data/app.db"),
		JWTSecret:      getEnv("JWT_SECRET", "change-me-secret"),
		MigrationsPath: getEnv("MIGRATIONS_PATH", "./migrations"),
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
