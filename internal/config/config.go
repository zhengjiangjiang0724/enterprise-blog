package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Log      LogConfig
	Upload   UploadConfig
}

type ServerConfig struct {
	Host string
	Port string
	Mode string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	// 连接池配置
	MaxOpenConns           int
	MaxIdleConns           int
	ConnMaxLifetimeMinutes int
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret      string
	ExpireHours int
}

type LogConfig struct {
	Level string
	File  string
}

type UploadConfig struct {
	Dir         string
	MaxSize     int64
	AllowedExts []string
}

var AppConfig *Config

func Load() error {
	// 加载 .env 文件（如果存在）
	_ = godotenv.Load()

	AppConfig = &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("SERVER_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "enterprise_blog"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			// 默认值适用于中小规模部署，可通过环境变量覆盖
			MaxOpenConns:           getEnvAsInt("DB_MAX_OPEN_CONNS", 50),
			MaxIdleConns:           getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetimeMinutes: getEnvAsInt("DB_CONN_MAX_LIFETIME_MINUTES", 60),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			ExpireHours: getEnvAsInt("JWT_EXPIRE_HOURS", 24),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "debug"),
			File:  getEnv("LOG_FILE", "logs/app.log"),
		},
		Upload: UploadConfig{
			Dir:         getEnv("UPLOAD_DIR", "uploads"),
			MaxSize:     int64(getEnvAsInt("MAX_UPLOAD_SIZE", 10485760)),
			AllowedExts: []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx"},
		},
	}

	return nil
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode)
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func (j JWTConfig) ExpireDuration() time.Duration {
	return time.Duration(j.ExpireHours) * time.Hour
}

