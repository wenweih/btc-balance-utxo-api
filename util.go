package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/olivere/elastic"
)

type elasticClientAlias struct {
	*elastic.Client
}

// Balance type struct
type Balance struct {
	Address string  `json:"address"`
	Amount  float64 `json:"amount"`
}

type configure struct {
	ElasticURL   string
	ElasticSniff bool
}

// esVout type struct
type esVout struct {
	TxIDBelongTo string      `json:"txidbelongto"`
	Value        float64     `json:"value"`
	Voutindex    uint32      `json:"voutindex"`
	Coinbase     bool        `json:"coinbase"`
	Addresses    []string    `json:"addresses"`
	Used         interface{} `json:"used"`
}

type utxo struct {
	Txid      string  `json:"txid"`
	Amount    float64 `json:"amount"`
	VoutIndex uint32  `json:"voutindex"`
}

// HomeDir 获取服务器当前用户目录路径
func HomeDir() string {
	home, err := homedir.Dir()
	if err != nil {
		sugar.Fatal("Home Dir error:", err.Error())
	}
	return home
}

// NoRouteMiddleware 路由错误
func noRouteMiddleware(ginInstance *gin.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		ginInstance.NoRoute(func(c *gin.Context) {
			c.AbortWithStatusJSON(http.StatusNotFound, map[string]string{"message": "Route Error"})
		})
	}
}

func ginEngine() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(noRouteMiddleware(r))

	return r
}

func ginResponseException(c *gin.Context, code int, err error) {
	c.JSON(code, gin.H{
		"status":  code,
		"message": err.Error(),
	})
}

func (conf configure) elasticClient() (*elasticClientAlias, error) {
	client, err := elastic.NewClient(
		elastic.SetURL(conf.ElasticURL),
		elastic.SetSniff(conf.ElasticSniff))
	if err != nil {
		return nil, err
	}
	elasticClient := elasticClientAlias{client}
	return &elasticClient, nil
}
