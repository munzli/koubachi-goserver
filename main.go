package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"koubachi-goserver/pkg/api"
	"time"

	"koubachi-goserver/pkg/config"
)

func main() {
	configuration := config.Init()
	configuration.LastConfigChange = time.Now()

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowMethods:    []string{"GET", "POST", "PUT", "HEAD", "DELETE", "PUT"},
		AllowHeaders:    []string{"Authorization", "Origin", "Content-Length", "Content-Type"},
		AllowAllOrigins: true,
	}))

	a := api.NewAPI(configuration)

	apiRouting := router.Group("/")
	a.AttachRoutes(apiRouting)

	router.Run(":8005")
}