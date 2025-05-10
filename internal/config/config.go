package config

import "os"

type Config struct {
    TemporalAddress string
    Namespace       string
    Port            string
}

func Load() Config {
    return Config{
        TemporalAddress: getenv("TEMPORAL_ADDRESS", "localhost:7233"),
        Namespace:       getenv("TEMPORAL_NAMESPACE", "default"),
        Port:            getenv("PORT", "8080"),
    }
}

func getenv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}