package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Database struct {
	Host         string `mapstructure:"host"`
	Port         string `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DatabaseName string `mapstructure:"databaseName"`
}

type Config struct {
	Database      Database `mapstructure:"database"`
	JWTKey        string   `mapstructure:"jwt_key"`
	TokenLifeSpan int      `mapstructure:"TokenLifeSpan"`
}

var C Config

func Init() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()

	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Errorf("config file not found")
		} else {
			return err
		}
	}

	if err := viper.Unmarshal(&C); err != nil {
		return err
	}

	return nil
}
