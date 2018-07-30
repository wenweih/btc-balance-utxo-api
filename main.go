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

func balanceHandler(c *gin.Context) {
	address := c.Param("address")
	// https://github.com/btcsuite/btcutil/blob/master/address_test.go
	_, err := btcutil.DecodeAddress(address, &chaincfg.MainNetParams)
	if err != nil {
		ginResponseException(c, http.StatusBadRequest, errors.New(strings.Join([]string{"Address format error:", err.Error()}, " ")))
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

func utxosHandle(c *gin.Context) {
	address := c.Param("address")
	// https://github.com/btcsuite/btcutil/blob/master/address_test.go
	_, err := btcutil.DecodeAddress(address, &chaincfg.MainNetParams)
	if err != nil {
		ginResponseException(c, http.StatusBadRequest, errors.New(strings.Join([]string{"Address format error:", err.Error()}, " ")))
		return
	}

	// addresses is an array datatype
	// https://www.elastic.co/guide/en/elasticsearch/reference/current/array.html

	// exists Query
	// https://www.elastic.co/guide/en/elasticsearch/reference/current/query-dsl-exists-query.html
	//
	// curl -H 'Content-Type: application/json' -XPOST http://47.75.159.52:9200/vout/vout/_search\?pretty\=true -d '
	// {
	//		"query": {"bool":{"must":{"term":{"addresses":"12cbQLTFMXRnSzktFkuoG3eHoMeFtpTu3S"}},"must_not":{"exists":{"field":"used"}}}}
	// }'
	q := elastic.NewBoolQuery().Must(elastic.NewTermQuery("addresses", address)).MustNot(elastic.NewExistsQuery("used"))
	searchResult, err := esClient.Search().Index("vout").Type("vout").Query(q).SortBy(elastic.NewFieldSort("value").Asc()).Do(context.TODO())
	if err != nil {
		ginResponseException(c, http.StatusBadRequest, errors.New(strings.Join([]string{"Query utxo error:", err.Error()}, " ")))
		return
	}

	var utxos []*esVout
	for _, vout := range searchResult.Hits.Hits {
		newVout := new(esVout)
		if err := json.Unmarshal(*vout.Source, newVout); err != nil {
			ginResponseException(c, http.StatusBadRequest, errors.New(strings.Join([]string{"fail to unmarshal esvout", err.Error()}, " ")))
			return
		}
		utxos = append(utxos, newVout)
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"address": address,
		"utxos":   utxos,
	})
}

func main() {
	r := ginEngine()
	r.GET("/balance/:address", balanceHandler)
	r.GET("/utxo/:address", utxosHandle)
	if err := r.Run(":3000"); err != nil {
		sugar.Fatal("New Router Error:", err.Error())
	}
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
