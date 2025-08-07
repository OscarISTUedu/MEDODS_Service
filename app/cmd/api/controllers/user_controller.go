package controllers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/OscarISTUedu/MEDODS_Service/internal"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type UserIDResponse struct {
	UserID int `json:"user_id" example:"0"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// GetUserId godoc
// @Summary Get ID of current user
// @Description Get ID of auth-ed user
// @Tags users
// @Produce  json
// @Failure 400 {object} ErrorResponse "Bad Request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Success 200 {object} UserIDResponse "Успешный ответ"
// @example {"user_id": 0}
// @Router /get_id [get]
func GetUserId(c *gin.Context) {
	access_token, err := c.Cookie("access_token")
	if err == nil {
		parsedToken, err := jwt.Parse(access_token, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "unexpected signing method"})
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			internal.LoadEnv()
			SecretKey := os.Getenv("SECRET_KEY")
			return []byte(SecretKey), nil
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "failed to parse access token"})
			return
		}
		if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
			userID := uint(claims["user_id"].(float64))
			c.JSON(http.StatusOK, gin.H{"user_id": userID})
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid access token"})
		}
	} else {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "access token is not provided"})
	}
}
