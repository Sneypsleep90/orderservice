package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBHost         string
	DBPort         int
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	KafkaBrokers   []string
	KafkaTopic     string
	KafkaGroupID   string
	ServerPort     int
	MigrationsPath string
}

func Load() Config {
	serverPort := 8081
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			serverPort = p
		}
	}

	dbPort := 5432
	if port := os.Getenv("DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			dbPort = p
		}
	}

	return Config{
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         dbPort,
		DBUser:         getEnv("DB_USER", "myapp_user"),
		DBPassword:     getEnv("DB_PASSWORD", "myapp_password"),
		DBName:         getEnv("DB_NAME", "myapp_db"),
		DBSSLMode:      getEnv("DB_SSLMODE", "disable"),
		KafkaBrokers:   []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		KafkaTopic:     getEnv("KAFKA_TOPIC", "orders"),
		KafkaGroupID:   getEnv("KAFKA_GROUP_ID", "order-service"),
		ServerPort:     serverPort,
		MigrationsPath: getEnv("MIGRATIONS_PATH", "./migrations"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
