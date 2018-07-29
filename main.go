package main

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	config *configure
)

func main() {
	Execute()

	sugar := zap.NewExample().Sugar()
	defer sugar.Sync()
	sugar.Info("Dump block")

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080
}
