package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type configure struct {
	ElasticURL   string
	ElasticSniff bool
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bitcoin-service-external-api",
	Short: "Bitcoin middleware for application",
}

// Execute 命令行入口
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf(err.Error())
	}
}

func init() {
	config = new(configure)
	config.InitConfig()
}

func (conf *configure) InitConfig() {
	viper.SetConfigType("yaml")
	viper.AddConfigPath(HomeDir())
	viper.SetConfigName("bitcoin-service-external-api")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err == nil {
		fmt.Println("Using Configure file:", viper.ConfigFileUsed())
	} else {
		log.Fatal("Error: bitcoin-service-external-api not found in: ", HomeDir())
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
