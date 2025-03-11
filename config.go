package main

import (
	"github.com/spf13/viper"
)

type config struct {
	Port    int
	APIPath string
}

func loadConfig() (*config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetDefault("port", 18080)
	viper.SetDefault("api_path", "./api")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	return &config{
		Port:    viper.GetInt("port"),
		APIPath: viper.GetString("api_path"),
	}, nil
}
