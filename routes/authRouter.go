package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/shubham03122001/golang-jwt-project/controllers"
)

func Authroutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("users/signup", controller.Signup())
	incomingRoutes.POST("users/login", controller.Login())

}
