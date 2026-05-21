package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port int `mapstructure:"port"`
}

func Load() (*Config, error) {
	viper.SetDefault("port", 8080)

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
