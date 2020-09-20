package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"koubachi-goserver/pkg/api"
	"time"

	"koubachi-goserver/pkg/config"
)

func main() {
	configuration := config.New()
	configuration.LastConfigChange = time.Now()

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), cors.New(cors.Config{
		AllowMethods:    []string{"GET", "POST", "PUT", "HEAD", "DELETE", "PUT"},
		AllowHeaders:    []string{"Authorization", "Origin", "Content-Length", "Content-Type"},
		AllowAllOrigins: true,
	}))

	a := api.New(configuration)

	apiRouting := router.Group("/")
	a.AttachRoutes(apiRouting)

	router.Run(":8005")
}