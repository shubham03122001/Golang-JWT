package helpers

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
)

func CheckUserType(c *gin.Context, role string) (err error) {
	UserType := c.GetString("user_type")
	fmt.Println("UserType after using token is", UserType)
	err = nil
	if UserType != role {
		err = errors.New("Unauthorised to access this resourse")
		return err
	}
	return err

}

func MatchUserTypeToUid(c *gin.Context, UserId string) (err error) {
	userType := c.GetString("user_type")
	uid := c.GetString("uid")

	if userType == "USER" && uid != UserId {
		err = errors.New("Unauthorized to acces this resource")
		return err
	}

	err = CheckUserType(c, userType)
	return err
}
