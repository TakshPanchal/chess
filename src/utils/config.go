package utils

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	SrvAddr string `mapstructure:"SERVER_ADDR"`

	// DB
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBName     string `mapstructure:"DB_NAME"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBConnStr  string `mapstructure:"DB_CONN_STR"`

	// JWT
	AccessTokenPrivateKeyPath  string `mapstructure:"ACCESS_TOKEN_PRIVATE_KEY_PATH"`
	AccessTokenPublicKeyPath   string `mapstructure:"ACCESS_TOKEN_PUBLIC_KEY_PATH"`
	RefreshTokenPrivateKeyPath string `mapstructure:"REFRESH_TOKEN_PRIVATE_KEY_PATH"`
	RefreshTokenPublicKeyPath  string `mapstructure:"REFRESH_TOKEN_PUBLIC_KEY_PATH"`
	JwtExpiration              int    `mapstructure:"JWT_EXPIRATION"` // in minutes
}

func NewConfig() *Config {
	conf := &Config{}

	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("Error reading config file %v", err)
	}

	return conf
}
