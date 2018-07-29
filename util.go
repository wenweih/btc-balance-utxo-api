package main

import (
	"log"

	"github.com/gin-gonic/gin"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/olivere/elastic"
)

// HomeDir 获取服务器当前用户目录路径
func HomeDir() string {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err.Error())
	}
	return home
}

// NoRouteMiddleware 路由错误
func noRouteMiddleware(ginInstance *gin.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		ginInstance.NoRoute(func(c *gin.Context) {
			c.JSON(404, gin.H{"code": 404, "message": "Route Error"})
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
		"status": code,
		"error":  err.Error(),
	})
}

type elasticClientAlias struct {
	*elastic.Client
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
