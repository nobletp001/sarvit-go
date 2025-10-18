package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all environment variables for the app.
type Config struct {
	AppEnv        string
	Port          string
	MongoURI      string
	MongoDB       string
	JWTSecret     string
	JWTTtlMinutes int
	DBDriver      string // "mongo" or "memory" (future switch)
}

// Load reads .env file and system environment variables into Config.
func Load() Config {
	// Load .env file if present (safe even if missing)
	_ = godotenv.Load()

	cfg := Config{
		AppEnv:    getEnv("APP_ENV", "development"),
		Port:      getEnv("PORT", "3000"),
		MongoURI:  getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:   getEnv("MONGO_DB", "todoapp"),
		JWTSecret: getEnv("JWT_SECRET", "super-secret-change-me"),
		DBDriver:  getEnv("DB_DRIVER", "mongo"), // Default: mongo
	}

	// Convert JWT_TTL_MINUTES to int safely
	if ttl, err := strconv.Atoi(getEnv("JWT_TTL_MINUTES", "30")); err == nil {
		cfg.JWTTtlMinutes = ttl
	} else {
		cfg.JWTTtlMinutes = 30
	}

	log.Printf("✅ Config loaded: PORT=%s | DB=%s | DRIVER=%s | ENV=%s",
		cfg.Port, cfg.MongoDB, cfg.DBDriver, cfg.AppEnv)
	return cfg
}

// Helper to safely get env with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
