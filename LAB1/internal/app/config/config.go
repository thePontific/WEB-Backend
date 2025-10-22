package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceHost string
	ServicePort int
	JWTSecret   string
	Redis       RedisConfig
}

type RedisConfig struct {
	Host        string
	Password    string
	Port        int
	User        string
	DialTimeout time.Duration
	ReadTimeout time.Duration
}

func NewConfig() (*Config, error) {
	_ = godotenv.Load() // загружаем .env

	configName := "config"
	if os.Getenv("CONFIG_NAME") != "" {
		configName = os.Getenv("CONFIG_NAME")
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")
	viper.WatchConfig()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	// JWT Secret
	if os.Getenv("JWT_SECRET") != "" {
		cfg.JWTSecret = os.Getenv("JWT_SECRET")
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "test" // дефолт
	}

	// Redis
	if os.Getenv("REDIS_HOST") != "" {
		cfg.Redis.Host = os.Getenv("REDIS_HOST")
	}
	if os.Getenv("REDIS_PORT") != "" {
		port, err := strconv.Atoi(os.Getenv("REDIS_PORT"))
		if err != nil {
			return nil, err
		}
		cfg.Redis.Port = port
	}
	if os.Getenv("REDIS_USER") != "" {
		cfg.Redis.User = os.Getenv("REDIS_USER")
	}
	if os.Getenv("REDIS_PASSWORD") != "" {
		cfg.Redis.Password = os.Getenv("REDIS_PASSWORD")
	}
	// таймауты
	cfg.Redis.DialTimeout = 10 * time.Second
	cfg.Redis.ReadTimeout = 10 * time.Second

	log.Info("config parsed")
	return cfg, nil
}
