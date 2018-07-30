package main

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type configure struct {
	ElasticURL   string
	ElasticSniff bool
}

func init() {
	config = new(configure)
	config.InitConfig()
}

func (conf *configure) InitConfig() {
	sugar = zap.NewExample().Sugar()
	defer sugar.Sync()
	viper.SetConfigType("yaml")
	viper.AddConfigPath(HomeDir())
	viper.SetConfigName("bitcoin-service-external-api")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err == nil {
		sugar.Info("Using Configure file:", viper.ConfigFileUsed())
	} else {
		sugar.Fatal("Error: configure bitcoin-service-external-api.yml not found in:", HomeDir())
	}

	for key, value := range viper.AllSettings() {
		switch key {
		case "elastic_url":
			conf.ElasticURL = value.(string)
		case "elastic_sniff":
			conf.ElasticSniff = value.(bool)
		}
	}
}
