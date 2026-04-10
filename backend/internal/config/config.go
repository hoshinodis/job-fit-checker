package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port               int
	Env                string
	SQLitePath         string
	OllamaBaseURL      string
	OllamaModel        string
	RequestTimeoutSec  int
	MaxJobTextLength   int
	MaxHTMLBytes       int64
	JobPollIntervalMS  int
	RateLimitPerMinute int
	StaticDir          string
}

func Load() *Config {
	return &Config{
		Port:               envInt("APP_PORT", 8080),
		Env:                envStr("APP_ENV", "development"),
		SQLitePath:         envStr("SQLITE_PATH", "data/job_fit.db"),
		OllamaBaseURL:      envStr("OLLAMA_BASE_URL", "http://localhost:11434"),
		OllamaModel:        envStr("OLLAMA_MODEL", "llama3.1:8b"),
		RequestTimeoutSec:  envInt("REQUEST_TIMEOUT_SECONDS", 120),
		MaxJobTextLength:   envInt("MAX_JOB_TEXT_LENGTH", 20000),
		MaxHTMLBytes:       int64(envInt("MAX_HTML_BYTES", 5*1024*1024)),
		JobPollIntervalMS:  envInt("JOB_POLL_INTERVAL_MS", 2000),
		RateLimitPerMinute: envInt("RATE_LIMIT_PER_MINUTE", 10),
		StaticDir:          envStr("STATIC_DIR", "static"),
	}
}

func envStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
