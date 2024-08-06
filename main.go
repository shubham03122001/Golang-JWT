package main

import (
	"os"

	"github.com/gin-gonic/gin"
	routes "github.com/shubham03122001/golang-jwt-project/routes"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	router := gin.New()
	router.Use(gin.Logger())

	routes.Authroutes(router)
	routes.Userroutes(router)

	router.GET("/api-1", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"Success": "Acces granted for api-1"})

	})

	router.GET("/api-2", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"Success": "Access granted for api-2"})
	})
	router.Run(":" + port)
}
