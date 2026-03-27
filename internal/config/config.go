package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	RateLimit RateLimitConfig `mapstructure:"ratelimit"`
	APIKeys   APIKeysConfig   `mapstructure:"apikeys"`
}

type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	Host         string        `mapstructure:"host"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
	TLS          TLSConfig     `mapstructure:"tls"`
}

type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxConnections  int           `mapstructure:"max_connections"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type RateLimitConfig struct {
	Enabled              bool `mapstructure:"enabled"`
	DefaultRequests      int  `mapstructure:"default_requests"`
	DefaultBurst         int  `mapstructure:"default_burst"`
	FreeTierRequests     int  `mapstructure:"free_tier_requests"`
	FreeTierBurst        int  `mapstructure:"free_tier_burst"`
	StandardTierRequests int  `mapstructure:"standard_tier_requests"`
	StandardTierBurst    int  `mapstructure:"standard_tier_burst"`
	PremiumTierRequests  int  `mapstructure:"premium_tier_requests"`
	PremiumTierBurst     int  `mapstructure:"premium_tier_burst"`
}

type APIKeysConfig struct {
	Keys []APIKeyEntry `mapstructure:"keys"`
}

type APIKeyEntry struct {
	Key       string `mapstructure:"key"`
	Tier      string `mapstructure:"tier"`
	IsActive  bool   `mapstructure:"is_active"`
	RateLimit int    `mapstructure:"rate_limit"`
}

func Load() (*Config, error) {
	return LoadWithPath(".")
}

func LoadWithPath(configPath string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	viper.AddConfigPath("/etc/identity-validation-mx")
	viper.AddConfigPath("$HOME/.config/identity-validation-mx")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("IDVAL")

	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

func setDefaults() {
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.read_timeout", "10s")
	viper.SetDefault("server.write_timeout", "10s")
	viper.SetDefault("server.idle_timeout", "60s")
	viper.SetDefault("server.tls.enabled", false)

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.database", "identity_validation")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_connections", 100)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.conn_max_lifetime", "1h")

	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	viper.SetDefault("ratelimit.enabled", true)
	viper.SetDefault("ratelimit.default_requests", 60)
	viper.SetDefault("ratelimit.default_burst", 10)
	viper.SetDefault("ratelimit.free_tier_requests", 100)
	viper.SetDefault("ratelimit.free_tier_burst", 20)
	viper.SetDefault("ratelimit.standard_tier_requests", 500)
	viper.SetDefault("ratelimit.standard_tier_burst", 50)
	viper.SetDefault("ratelimit.premium_tier_requests", 2000)
	viper.SetDefault("ratelimit.premium_tier_burst", 200)
}

func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}