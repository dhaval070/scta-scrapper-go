package config

import (
	"errors"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	DbDSN                  string     `mapstructure:"DB_DSN"`
	ApiKey                 string     `mapstructure:"API_KEY"`
	ImportUrl              string     `mapstructure:"IMPORT_URL"`
	GameSheetAPIKey        string     `mapstructure:"GAMESHEET_API_KEY"`
	GamesheetStartDate     string     `mapstructure:"GAMESHEET_START_DATE"`
	MaxRequestsPerHost     int        `mapstructure:"MAX_REQUESTS_PER_HOST"`
	ExternalAddressFetcher bool       `mapstructure:"EXTERNAL_ADDRESS_FETCHER"`
	SmtpConfig             SmtpConfig `mapstructure:"SMTP_CONFIG"`
}

type SmtpConfig struct {
	Host string `mapstructure:"HOST"`
	Port int    `mapstructure:"PORT"`
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

	if cfg.GamesheetStartDate == "" {
		return cfg, errors.New("GAMESHEET_START_DATE is required in config")
	}

	if cfg.SmtpConfig.Host == "" {
		return cfg, errors.New("SMTP Host is required in config")

	}
	if cfg.SmtpConfig.Port == 0 {
		return cfg, errors.New("SMTP Port is required in config")

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
