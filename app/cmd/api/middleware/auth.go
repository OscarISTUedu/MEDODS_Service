package middleware

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/OscarISTUedu/MEDODS_Service/cmd/api/controllers"
	"github.com/OscarISTUedu/MEDODS_Service/cmd/api/utils"
	"github.com/OscarISTUedu/MEDODS_Service/internal"
	database "github.com/OscarISTUedu/MEDODS_Service/internal/database"
	"github.com/OscarISTUedu/MEDODS_Service/internal/database/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"
)

const (
	WebHook            = "https://127.0.0.1"
	AccessTokenCookie  = "access_token"
	RefreshTokenCookie = "refresh_token"
	TokenRefreshBefore = 300 * time.Second // Обновлять, если осталось меньше n секунд
	AccessTokenExpire  = 600
)

func TokenAutoRefresh() gin.HandlerFunc {
	return func(c *gin.Context) {
		// fmt.Printf("%#v\n", utils.SendPost) // посмотреть структуру пакета
		accessToken, err := c.Cookie("access_token")
		if err != nil {
			c.Next()
			return
		}
		parsedToken, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			internal.LoadEnv()
			return []byte(os.Getenv("SECRET_KEY")), nil
		})
		if err != nil {
			if ve, ok := err.(*jwt.ValidationError); ok {
				if ve.Errors&(jwt.ValidationErrorMalformed|jwt.ValidationErrorSignatureInvalid) != 0 {
					log.Println("Invalid token:", err)
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
					return
				}
			} else {
				log.Println("Token parsing error:", err)
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Failed to parse token"})
				return
			}
		}
		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			log.Println("Invalid token claims")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}
		userId := uint(claims["user_id"].(float64))
		refreshTokenId := uint(claims["refresh_token_id"].(float64))
		exp := int64(claims["exp"].(float64))
		expTime := time.Unix(exp, 0)
		var CurRefreshToken models.RefreshToken
		result := database.DB.Where("id = ? AND deleted_at IS NULL", refreshTokenId).First(&CurRefreshToken)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Access token has no refresh token"})
			} else {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			}
			return
		}
		if CurRefreshToken.UserAgent != c.GetHeader("User-Agent") {
			controllers.LogOut(c)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "User-Agent has switched"})
			return
		}
		if CurRefreshToken.IpAdress != c.ClientIP() {
			utils.SendPost(userId, CurRefreshToken.IpAdress, c.ClientIP(), WebHook)
		}
		needRefresh := time.Until(expTime) <= TokenRefreshBefore
		if needRefresh {
			newAccessToken := controllers.CreateAccessToken(userId, refreshTokenId)
			c.SetCookie("access_token", newAccessToken, AccessTokenExpire, "/", "", true, true)
			log.Println("Access token refreshed for user", userId)
		}
		c.Next()
	}
}
