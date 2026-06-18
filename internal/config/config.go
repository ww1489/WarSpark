package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

const (
	// AppName, AppDisplayName, and EnvPrefix are template identity constants
	// rewritten by init-template scripts. Runtime values are still resolved in Load.
	AppName        = "warspark"
	AppDisplayName = "WarSpark"
	AppVersion     = "dev"
	EnvPrefix      = "WARSPARK"

	defaultConfigPath = "configs/config.yaml"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	CORS   CORSConfig   `yaml:"cors"`
	MySQL  MySQLConfig  `yaml:"mysql"`
	Redis  RedisConfig  `yaml:"redis"`
	JWT    JWTConfig    `yaml:"jwt"`
	CoC    CoCConfig    `yaml:"coc"`
	Log    LogConfig    `yaml:"log"`
	Docs   DocsConfig   `yaml:"docs"`
}

type ServerConfig struct {
	Port              int           `yaml:"port"`
	Mode              string        `yaml:"mode"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
	ReadTimeout       time.Duration `yaml:"read_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"`
	StartupTimeout    time.Duration `yaml:"startup_timeout"`
	ShutdownTimeout   time.Duration `yaml:"shutdown_timeout"`
}

type CORSConfig struct {
	AllowedOrigins   []string      `yaml:"allowed_origins"`
	AllowedMethods   []string      `yaml:"allowed_methods"`
	AllowedHeaders   []string      `yaml:"allowed_headers"`
	ExposeHeaders    []string      `yaml:"expose_headers"`
	AllowCredentials bool          `yaml:"allow_credentials"`
	MaxAge           time.Duration `yaml:"max_age"`
}

type MySQLConfig struct {
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	Host            string        `yaml:"host"`
	Port            string        `yaml:"port"`
	DB              string        `yaml:"db"`
	Charset         string        `yaml:"charset"`
	ParseTime       bool          `yaml:"parse_time"`
	Loc             string        `yaml:"loc"`
	MaxOpen         int           `yaml:"max_open"`
	MaxIdle         int           `yaml:"max_idle"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
}

type RedisConfig struct {
	Addr            string        `yaml:"addr"`
	Port            string        `yaml:"port"`
	Password        string        `yaml:"password"`
	DB              int           `yaml:"db"`
	PoolSize        int           `yaml:"pool_size"`
	MinIdleConns    int           `yaml:"min_idle_conns"`
	DialTimeout     time.Duration `yaml:"dial_timeout"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

type JWTConfig struct {
	Secret        string        `yaml:"secret"`
	Issuer        string        `yaml:"issuer"`
	AccessExpire  time.Duration `yaml:"access_expire"`
	RefreshExpire time.Duration `yaml:"refresh_expire"`
}

type CoCConfig struct {
	BaseURL            string        `yaml:"base_url"`
	APIToken           string        `yaml:"api_token"`
	Timeout            time.Duration `yaml:"timeout"`
	CurrentWarCacheTTL time.Duration `yaml:"current_war_cache_ttl"`
}

type LogConfig struct {
	Level   string           `yaml:"level"`
	Console ConsoleLogConfig `yaml:"console"`
	File    FileLogConfig    `yaml:"file"`
}

type ConsoleLogConfig struct {
	Enabled bool `yaml:"enabled"`
}

type FileLogConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Path       string `yaml:"path"`
	MaxSizeMB  int    `yaml:"max_size_mb"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAgeDays int    `yaml:"max_age_days"`
	Compress   bool   `yaml:"compress"`
}

type DocsConfig struct {
	SwaggerEnabled bool `yaml:"swagger_enabled"`
}

func Load(path string, portOverride int) (Config, error) {
	v := newViper(Defaults())

	if path != "" {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			if !isMissingDefaultConfig(path, err) {
				return Config{}, fmt.Errorf("read config %s: %w", path, err)
			}
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg, decodeOptions); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	if portOverride > 0 {
		cfg.Server.Port = portOverride
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Defaults() Config {
	// Defaults are local fallbacks, not deployment policy. YAML, environment
	// variables, and CLI overrides are layered on top by Load.
	return Config{
		Server: ServerConfig{
			Port:              8080,
			Mode:              "debug",
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       120 * time.Second,
			StartupTimeout:    10 * time.Second,
			ShutdownTimeout:   10 * time.Second,
		},
		CORS: CORSConfig{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
			ExposeHeaders:  []string{"X-Request-ID"},
			MaxAge:         24 * time.Hour,
		},
		MySQL: MySQLConfig{
			User:            "root",
			Host:            "127.0.0.1",
			Port:            "3306",
			DB:              AppName,
			Charset:         "utf8mb4",
			ParseTime:       true,
			Loc:             "Local",
			MaxOpen:         100,
			MaxIdle:         10,
			ConnMaxLifetime: time.Hour,
			ConnMaxIdleTime: 10 * time.Minute,
		},
		Redis: RedisConfig{
			Addr:            "127.0.0.1",
			Port:            "6379",
			DB:              0,
			PoolSize:        10,
			MinIdleConns:    5,
			DialTimeout:     5 * time.Second,
			ReadTimeout:     3 * time.Second,
			WriteTimeout:    3 * time.Second,
			ConnMaxLifetime: 30 * time.Minute,
		},
		JWT: JWTConfig{
			Secret:        "change-me-in-production",
			Issuer:        AppName,
			AccessExpire:  30 * time.Minute,
			RefreshExpire: 168 * time.Hour,
		},
		CoC: CoCConfig{
			BaseURL:            "https://api.clashofclans.com/v1",
			Timeout:            10 * time.Second,
			CurrentWarCacheTTL: 2 * time.Minute,
		},
		Log: LogConfig{
			Level: "info",
			Console: ConsoleLogConfig{
				Enabled: true,
			},
			File: FileLogConfig{
				Enabled:    true,
				Path:       "logs/" + AppName + ".log",
				MaxSizeMB:  100,
				MaxBackups: 30,
				MaxAgeDays: 30,
				Compress:   true,
			},
		},
		Docs: DocsConfig{
			SwaggerEnabled: true,
		},
	}
}

func newViper(defaults Config) *viper.Viper {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetEnvPrefix(EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v, defaults)
	return v
}

func decodeOptions(decoderConfig *mapstructure.DecoderConfig) {
	decoderConfig.TagName = "yaml"
	decoderConfig.DecodeHook = mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	)
}

func setDefaults(v *viper.Viper, cfg Config) {
	v.SetDefault("server.port", cfg.Server.Port)
	v.SetDefault("server.mode", cfg.Server.Mode)
	v.SetDefault("server.read_header_timeout", cfg.Server.ReadHeaderTimeout)
	v.SetDefault("server.read_timeout", cfg.Server.ReadTimeout)
	v.SetDefault("server.write_timeout", cfg.Server.WriteTimeout)
	v.SetDefault("server.idle_timeout", cfg.Server.IdleTimeout)
	v.SetDefault("server.startup_timeout", cfg.Server.StartupTimeout)
	v.SetDefault("server.shutdown_timeout", cfg.Server.ShutdownTimeout)

	v.SetDefault("cors.allowed_origins", cfg.CORS.AllowedOrigins)
	v.SetDefault("cors.allowed_methods", cfg.CORS.AllowedMethods)
	v.SetDefault("cors.allowed_headers", cfg.CORS.AllowedHeaders)
	v.SetDefault("cors.expose_headers", cfg.CORS.ExposeHeaders)
	v.SetDefault("cors.allow_credentials", cfg.CORS.AllowCredentials)
	v.SetDefault("cors.max_age", cfg.CORS.MaxAge)

	v.SetDefault("mysql.user", cfg.MySQL.User)
	v.SetDefault("mysql.password", cfg.MySQL.Password)
	v.SetDefault("mysql.host", cfg.MySQL.Host)
	v.SetDefault("mysql.port", cfg.MySQL.Port)
	v.SetDefault("mysql.db", cfg.MySQL.DB)
	v.SetDefault("mysql.charset", cfg.MySQL.Charset)
	v.SetDefault("mysql.parse_time", cfg.MySQL.ParseTime)
	v.SetDefault("mysql.loc", cfg.MySQL.Loc)
	v.SetDefault("mysql.max_open", cfg.MySQL.MaxOpen)
	v.SetDefault("mysql.max_idle", cfg.MySQL.MaxIdle)
	v.SetDefault("mysql.conn_max_lifetime", cfg.MySQL.ConnMaxLifetime)
	v.SetDefault("mysql.conn_max_idle_time", cfg.MySQL.ConnMaxIdleTime)

	v.SetDefault("redis.addr", cfg.Redis.Addr)
	v.SetDefault("redis.port", cfg.Redis.Port)
	v.SetDefault("redis.password", cfg.Redis.Password)
	v.SetDefault("redis.db", cfg.Redis.DB)
	v.SetDefault("redis.pool_size", cfg.Redis.PoolSize)
	v.SetDefault("redis.min_idle_conns", cfg.Redis.MinIdleConns)
	v.SetDefault("redis.dial_timeout", cfg.Redis.DialTimeout)
	v.SetDefault("redis.read_timeout", cfg.Redis.ReadTimeout)
	v.SetDefault("redis.write_timeout", cfg.Redis.WriteTimeout)
	v.SetDefault("redis.conn_max_lifetime", cfg.Redis.ConnMaxLifetime)

	v.SetDefault("jwt.secret", cfg.JWT.Secret)
	v.SetDefault("jwt.issuer", cfg.JWT.Issuer)
	v.SetDefault("jwt.access_expire", cfg.JWT.AccessExpire)
	v.SetDefault("jwt.refresh_expire", cfg.JWT.RefreshExpire)

	v.SetDefault("coc.base_url", cfg.CoC.BaseURL)
	v.SetDefault("coc.api_token", cfg.CoC.APIToken)
	v.SetDefault("coc.timeout", cfg.CoC.Timeout)
	v.SetDefault("coc.current_war_cache_ttl", cfg.CoC.CurrentWarCacheTTL)

	v.SetDefault("log.level", cfg.Log.Level)
	v.SetDefault("log.console.enabled", cfg.Log.Console.Enabled)
	v.SetDefault("log.file.enabled", cfg.Log.File.Enabled)
	v.SetDefault("log.file.path", cfg.Log.File.Path)
	v.SetDefault("log.file.max_size_mb", cfg.Log.File.MaxSizeMB)
	v.SetDefault("log.file.max_backups", cfg.Log.File.MaxBackups)
	v.SetDefault("log.file.max_age_days", cfg.Log.File.MaxAgeDays)
	v.SetDefault("log.file.compress", cfg.Log.File.Compress)

	v.SetDefault("docs.swagger_enabled", cfg.Docs.SwaggerEnabled)
}

func isMissingDefaultConfig(path string, err error) bool {
	if filepath.Clean(path) != filepath.Clean(defaultConfigPath) {
		return false
	}

	var configFileNotFoundError viper.ConfigFileNotFoundError
	if errors.As(err, &configFileNotFoundError) {
		return true
	}
	return errors.Is(err, os.ErrNotExist)
}

func (c Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535")
	}
	if c.Server.Mode == "" {
		return fmt.Errorf("server.mode is required")
	}
	switch c.Server.Mode {
	case "debug", "release", "test":
	default:
		return fmt.Errorf("server.mode must be debug, release, or test")
	}
	if c.MySQL.User == "" {
		return fmt.Errorf("mysql.user is required")
	}
	if c.MySQL.Host == "" || c.MySQL.Port == "" || c.MySQL.DB == "" {
		return fmt.Errorf("mysql.host, mysql.port, and mysql.db are required")
	}
	if c.Redis.Addr == "" || c.Redis.Port == "" {
		return fmt.Errorf("redis.addr and redis.port are required")
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("jwt.secret is required")
	}
	if c.JWT.Issuer == "" {
		return fmt.Errorf("jwt.issuer is required")
	}
	if c.JWT.AccessExpire <= 0 || c.JWT.RefreshExpire <= 0 {
		return fmt.Errorf("jwt access_expire and refresh_expire must be positive")
	}
	if c.CoC.BaseURL == "" {
		return fmt.Errorf("coc.base_url is required")
	}
	if c.CoC.Timeout <= 0 {
		return fmt.Errorf("coc.timeout must be positive")
	}
	if c.CoC.CurrentWarCacheTTL <= 0 {
		return fmt.Errorf("coc.current_war_cache_ttl must be positive")
	}
	if !c.Log.Console.Enabled && !c.Log.File.Enabled {
		return fmt.Errorf("at least one log output must be enabled")
	}
	if c.Log.File.Enabled && c.Log.File.Path == "" {
		return fmt.Errorf("log.file.path is required when file logging is enabled")
	}
	return nil
}
