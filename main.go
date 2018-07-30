package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	config   *configure
	esClient *elasticClientAlias
	sugar    *zap.SugaredLogger
)

type balanceParams struct {
	Address string `form:"address" json:"address"`
}

func balanceEndPoint() {
	r := ginEngine()
	r.GET("/balance/:address", balanceHandler)
	if err := r.Run(":3000"); err != nil {
		sugar.Fatal("Balance Error:", err.Error())
	}
}

func balanceHandler(c *gin.Context) {
	address := c.Param("address")
	_, err := btcutil.DecodeAddress(address, &chaincfg.MainNetParams)
	if err != nil {
		ginResponseException(c, http.StatusBadRequest, err)
		return
	}

	searchResult, err := esClient.Search().Index("balance").Type("balance").Query(elastic.NewTermQuery("address", address)).Do(context.TODO())
	if err != nil {
		ginResponseException(c, http.StatusNotFound, err)
		return
	}

	if len(searchResult.Hits.Hits) != 1 {
		ginResponseException(c, http.StatusNotFound, errors.New(strings.Join([]string{"Fail to query balance for", address}, " ")))
		return
	}
	b := new(Balance)
	if err := json.Unmarshal(*(searchResult.Hits.Hits[0].Source), b); err != nil {
		ginResponseException(c, http.StatusBadRequest, errors.New("unmarshal result error"))
		return
	}

	c.JSON(200, gin.H{
		"status":  http.StatusOK,
		"address": address,
		"balance": b.Amount,
	})
}

func main() {
	balanceEndPoint()
}

func init() {
	config = new(configure)
	sugar = zap.NewExample().Sugar()
	defer sugar.Sync()
	config.InitConfig()
	c, err := config.elasticClient()
	if err != nil {
		sugar.Fatal("init es client error:", err.Error())
	}
	esClient = c

}

func (conf *configure) InitConfig() {
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
