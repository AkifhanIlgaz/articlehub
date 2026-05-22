package config

import (
	"fmt"
	"os"
)

// Config; uygulama genelinde kullanılan ortam değişkenlerini tutar.
// TODO: go get github.com/joho/godotenv && godotenv.Load() ekle
type Config struct {
	ESAddress string
	IndexName string
	Port      string
}

func Load() (*Config, error) {
	cfg := &Config{
		ESAddress: getEnvOrDefault("ELASTICSEARCH_URL", "http://localhost:9200"),
		IndexName: getEnvOrDefault("ES_INDEX_NAME", "articles"),
		Port:      getEnvOrDefault("PORT", "8080"),
	}

	if cfg.ESAddress == "" {
		return nil, fmt.Errorf("config: ELASTICSEARCH_URL is required")
	}
	return cfg, nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
