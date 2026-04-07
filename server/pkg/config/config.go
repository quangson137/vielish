package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	CORS     CORSConfig
	Tracing  TracingConfig
}

type AppConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	URL string
}

type RedisConfig struct {
	URL string
}

type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration `mapstructure:"access_ttl"`
	RefreshTTL time.Duration `mapstructure:"refresh_ttl"`
}

type CORSConfig struct {
	Origins []string
}

type TracingConfig struct {
	Enabled  bool
	Endpoint string
}

// Load reads config.yaml from the given search paths (or "." and "./cmd/api" by default)
// and applies environment variable overrides. Env vars use underscore-separated keys,
// e.g. APP_PORT overrides app.port, JWT_SECRET overrides jwt.secret.
func Load(searchPaths ...string) (Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	if len(searchPaths) == 0 {
		v.AddConfigPath(".")
		v.AddConfigPath("./cmd/api")
	} else {
		for _, p := range searchPaths {
			v.AddConfigPath(p)
		}
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		var notFoundErr viper.ConfigFileNotFoundError
		if !errors.As(err, &notFoundErr) {
			return Config{}, fmt.Errorf("reading config file: %w", err)
		}
	}

	var cfg Config
	decodeHook := func(dc *mapstructure.DecoderConfig) {
		dc.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			dc.DecodeHook,
		)
	}
	if err := v.Unmarshal(&cfg, decodeHook); err != nil {
		return Config{}, fmt.Errorf("unmarshalling config: %w", err)
	}
	return cfg, nil
}

// NewConfig is the fx provider — loads from default search paths.
func NewConfig() (Config, error) {
	return Load()
}

var Module = fx.Module("config",
	fx.Provide(NewConfig),
)
