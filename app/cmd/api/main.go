package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/OscarISTUedu/MEDODS_Service/cmd/api/controllers"
	_ "github.com/OscarISTUedu/MEDODS_Service/cmd/api/docs"
	"github.com/OscarISTUedu/MEDODS_Service/cmd/api/middleware"
	database "github.com/OscarISTUedu/MEDODS_Service/internal/database"
	"github.com/OscarISTUedu/MEDODS_Service/internal/database/models"
	"github.com/OscarISTUedu/MEDODS_Service/internal/migrations"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// @title Gin Swagger API
// @version 1.0
// @description This is a MEDODS server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @host 127.0.0.1:8000
// @schemes http
func main() {
	db, err := database.Connect()
	if err != nil {
		return
	}
	database.DB = db
	migrations.AutoMigrateAll(db)
	///
	///
	///
	var model models.User
	result := db.First(&model)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		users := []models.User{
			{Login: "john_doe", Password: "SecurePass123!"},
			{Login: "alice_smith", Password: "Alice@2023"},
			{Login: "robert_j", Password: "Robert#456"},
			{Login: "emily_w", Password: "Em!lyW1lls"},
			{Login: "michael_b", Password: "M1ch@elB2023"},
			{Login: "sarah_k", Password: "S@rahK!99"},
			{Login: "david_miller", Password: "D@vidM1ller"},
			{Login: "lisa_jackson", Password: "L!saJ@cks0n"},
			{Login: "peter_parker", Password: "Sp1d3rM@n"},
			{Login: "natalie_wong", Password: "N@tW0ng2023"},
		}
		crypt_users := make([]models.User, 0, len(users))

		for _, u := range users {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), 12)
			if err != nil {
				log.Fatalf("Failed to hash password for user %s: %v", u.Login, err)
			}
			crypt_users = append(crypt_users, models.User{
				Login:    u.Login,
				Password: string(hashedPassword),
			})
		}
		db.Create(&crypt_users)
	} else {
		fmt.Println("В таблице есть пользователи")
	}
	///
	///
	///
	router := gin.Default()
	router.Use(middleware.TokenAutoRefresh())
	router.GET("/auth/tokens/:id", controllers.GetJwtTokens)
	router.GET("/update_tokens", controllers.UpdateTokens)
	router.GET("/get_id", controllers.GetUserId)
	router.GET("/log_out", controllers.LogOut)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.Run("0.0.0.0:8000")
}
