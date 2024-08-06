package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/shubham03122001/golang-jwt-project/database"
	helper "github.com/shubham03122001/golang-jwt-project/helpers"
	"github.com/shubham03122001/golang-jwt-project/models"
)

var userCollection *mongo.Collection = database.OpenCollection("user")
var validate = validator.New()

func HashPassword(password string) string {
	hashPasswordInBytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	//return string([]byte(password))

	return string(hashPasswordInBytes)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))

	check := true
	msg := ""
	if err != nil {
		msg = fmt.Sprintf("Email or Password is incorrect")
		check = false
	}

	return check, msg

}

func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 180*time.Second)
		defer cancel()
		var User models.User
		if err := c.BindJSON(&User); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		validationErr := validate.Struct(User)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": validationErr.Error()})
			return
		}

		countOfEmail, err := userCollection.CountDocuments(ctx, bson.M{"email": User.Email})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "error occured while checking for the count of email "})
			return
		}
		fmt.Println("Count of email is:", countOfEmail)
		password := HashPassword(*User.Password)
		User.Password = &password
		countOfPhoneNumber, err := userCollection.CountDocuments(ctx, bson.M{"phone": User.Phone})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "error occured while checking for the count of mobileNumber"})
			return
		}
		fmt.Println("Count of PhoneNumber is :", countOfPhoneNumber)

		if countOfEmail > 0 || countOfPhoneNumber > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "This email or phoneNumber already exists!"})
			return
		}
		User.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		User.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		User.Id = primitive.NewObjectID()
		User.User_id = User.Id.Hex()
		token, refreshToken, _ := helper.GenerateAllTokens(*User.Email, *User.First_name, *User.Last_name, *User.User_type, *&User.User_id)
		User.Token = &token
		User.Refresh_token = &refreshToken

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, User)
		if insertErr != nil {
			msg := fmt.Sprintf("USer item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"Error": msg})
			return
		}

		c.JSON(http.StatusOK, resultInsertionNumber)

	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		var User models.User
		var foundUser models.User

		if err := c.BindJSON(&User); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}

		// var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
		err := userCollection.FindOne(ctx, bson.M{"email": User.Email}).Decode(&foundUser)

		defer cancel()

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Email or password is incorrect"})
			return
		}
		passwordIsValid, msg := VerifyPassword(*User.Password, *foundUser.Password)
		defer cancel()

		if passwordIsValid != true {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": msg})
			return
		}

		if foundUser.Email == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "User not found"})

		}

		token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, *foundUser.User_type, *&foundUser.User_id)
		fmt.Println("Token in Login Function is:", token)
		fmt.Println("Refresh token in Login Function is:", refreshToken)
		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)

		err = userCollection.FindOne(ctx, bson.M{"user_id": foundUser.User_id}).Decode(&foundUser)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, foundUser)

	}
}

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))

		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err1 := strconv.Atoi(c.Query("page"))

		if err1 != nil || page < 1 {
			page = 1
		}
		startIndex := (page - 1) * recordPerPage
		fmt.Println("StartIndex after applying pagination logic:", startIndex)

		startIndex, err = strconv.Atoi(c.Query("startIndex"))
		matchStage := bson.D{{"$match", bson.D{{}}}}
		groupStage := bson.D{{
			"$group", bson.D{
				{"_id", bson.D{{"_id", "null"}}},
				{"total_count", bson.D{{"$sum", 1}}},
				{"data", bson.D{{"$push", "$$ROOT"}}},
			},
		}}

		projectStage := bson.D{
			{"$project", bson.D{
				{"_id", 0},
				{"total_count", 1},
				{"user_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},
			}},
		}
		result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage,
		})
		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Error occured while listing user items"})
		}

		var allUsers []bson.M
		if err = result.All(ctx, &allUsers); err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, allUsers[0])

	}
}
func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")
		if err := helper.MatchUserTypeToUid(c, userId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var User models.User

		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&User)
		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, User)

	}
}
