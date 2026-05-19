package config

import "os"

type Config struct {
	JWTSecret string
	DBPath    string
	Port      string
}

func Load() Config {
	return Config{
		JWTSecret: os.Getenv("JWT_SECRET"),
		DBPath:    getEnv("DB_PATH", "./auth.db"),
		Port:      getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
