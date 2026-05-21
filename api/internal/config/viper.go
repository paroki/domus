package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Port int          `mapstructure:"port"`
	Env  string       `mapstructure:"env"`
	Api  ApiConfig    `mapstructure:"api"`
	Log  LoggerConfig `mapstructure:"log"`
}

func GetConfig() (*Config, error) {
	// Root defaults
	viper.SetDefault("port", 8080)
	viper.SetDefault("env", "development")

	// Api defaults
	viper.SetDefault("api.app_name", "Domus API")
	viper.SetDefault("api.read_timeout", 5*time.Second)
	viper.SetDefault("api.write_timeout", 10*time.Second)
	viper.SetDefault("api.idle_timeout", 60*time.Second)
	viper.SetDefault("api.body_limit", 4*1024*1024) // 4MB
	viper.SetDefault("api.prefork", false)
	viper.SetDefault("api.cors", "")

	// Log defaults
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.adapter", "slog")

	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("DOMUS")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
