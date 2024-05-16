// Package config provides configuration for the chat service.
package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config is the configuration for the chat service.
type Config struct {
	Listen                 string
	AllowAnon              bool
	CookieName             string
	RedisHost              string
	RedisPassword          string
	RabbitMQHost           string
	JWTSecret              string
	JWTExpirationInSeconds int64
}

// Envs is the configuration singleton for the chat service.
var Envs = initConfig()

func initConfig() Config {
	godotenv.Load()

	return Config{
		Listen:                 getEnv("LISTEN", ":8080"),
		AllowAnon:              getEnv("ALLOW_ANON", "true") == "true",
		CookieName:             getEnv("COOKIE_NAME", ""),
		RedisHost:              fmt.Sprintf("%s:%s", getEnv("REDIS_HOST", "localhost"), getEnv("REDIS_PORT", "6379")),
		RedisPassword:          getEnv("REDIS_PASSWORD", ""),
		RabbitMQHost:           fmt.Sprintf("amqp://%s:%s@%s:%s/", getEnv("RABBITMQ_USER", "guest"), getEnv("RABBITMQ_PASSWORD", "guest"), getEnv("RABBITMQ_HOST", "localhost"), getEnv("RABBITMQ_PORT", "5672")),
		JWTSecret:              getEnv("JWT_SECRET", "ChANg3-Th1S-$ecR3T"),
		JWTExpirationInSeconds: getEnvInSeconds("JWT_EXPIRATION_IN_SECONDS", 3600*24*7),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInSeconds(key string, fallback int64) int64 {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fallback
		}
		return i
	}
	return fallback
}
