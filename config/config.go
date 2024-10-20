package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	DbDSN     string `mapstructure:"db_dsn"`
	ApiKey    string `mapstructure:"api_key"`
	ImportUrl string `mapstructure:"import_url"`
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

	return cfg, err
}

func MustReadConfig() Config {
	cfg, err := ReadConfig()
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}
