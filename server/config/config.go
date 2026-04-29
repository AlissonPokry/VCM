package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	Port                           string
	DBPath                         string
	UploadDir                      string
	ThumbnailDir                   string
	MaxFileSizeMB                  int64
	N8NWebhookSecret               string
	CORSOrigin                     string
	Env                            string
	PlatformDispatchTimeoutSeconds int
	ProjectRoot                    string
	FFmpegEnabled                  bool
}

// Load reads environment variables and returns typed application configuration.
func Load() (Config, error) {
	env := firstNonEmpty(os.Getenv("APP_ENV"), os.Getenv("NODE_ENV"), "development")
	if env != "production" {
		_ = godotenv.Load("../.env", ".env")
	}

	root, err := projectRoot()
	if err != nil {
		return Config{}, err
	}

	maxFileSize, err := intFromEnv("MAX_FILE_SIZE_MB", 500)
	if err != nil {
		return Config{}, err
	}

	dispatchTimeout, err := intFromEnv("PLATFORM_DISPATCH_TIMEOUT_SECONDS", 60)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Port:                           firstNonEmpty(os.Getenv("PORT"), "3001"),
		DBPath:                         resolvePath(root, firstNonEmpty(os.Getenv("DB_PATH"), "./server/db/reel_queue.sqlite")),
		UploadDir:                      resolvePath(root, firstNonEmpty(os.Getenv("UPLOAD_DIR"), "./server/uploads")),
		ThumbnailDir:                   resolvePath(root, firstNonEmpty(os.Getenv("THUMBNAIL_DIR"), "./server/thumbnails")),
		MaxFileSizeMB:                  int64(maxFileSize),
		N8NWebhookSecret:               firstNonEmpty(os.Getenv("N8N_WEBHOOK_SECRET"), "change_me_before_production"),
		CORSOrigin:                     firstNonEmpty(os.Getenv("CORS_ORIGIN"), "http://localhost:5173"),
		Env:                            firstNonEmpty(os.Getenv("APP_ENV"), env),
		PlatformDispatchTimeoutSeconds: dispatchTimeout,
		ProjectRoot:                    root,
	}

	return cfg, nil
}

// DispatchTimeout returns the configured per-platform posting timeout.
func (c Config) DispatchTimeout() time.Duration {
	return time.Duration(c.PlatformDispatchTimeoutSeconds) * time.Second
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func intFromEnv(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}

func projectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	if filepath.Base(wd) == "server" {
		return filepath.Dir(wd), nil
	}
	return wd, nil
}

func resolvePath(root string, value string) string {
	if filepath.IsAbs(value) {
		return filepath.Clean(value)
	}
	return filepath.Clean(filepath.Join(root, value))
}
