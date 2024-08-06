package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/shubham03122001/golang-jwt-project/controllers"
	"github.com/shubham03122001/golang-jwt-project/middleware"
)

func Userroutes(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.GET("/users", controller.GetUsers())
	incomingRoutes.GET("/users/:user_id", controller.GetUser())
}
