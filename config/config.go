package config

import (
	"errors"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	DbDSN           string `mapstructure:"DB_DSN"`
	ApiKey          string `mapstructure:"API_KEY"`
	ImportUrl       string `mapstructure:"IMPORT_URL"`
	GameSheetAPIKey string `mapstructure:"GAMESHEET_API_KEY"`
}

func Init(name string, path ...string) {
	viper.SetConfigName(name)

	for _, p := range path {
		viper.AddConfigPath(p)
	}
}

func ReadConfig() (Config, error) {
	cfg := Config{}

	err := viper.ReadInConfig()
	if err != nil {
		return cfg, err
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, err
	}

	if cfg.DbDSN == "" {
		return cfg, errors.New("DB_DSN is required in config")
	}

	return cfg, err
}

func MustReadConfig() Config {
	cfg, err := ReadConfig()
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}
