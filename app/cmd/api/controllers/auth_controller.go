package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/OscarISTUedu/MEDODS_Service/cmd/api/utils"
	"github.com/OscarISTUedu/MEDODS_Service/internal/database"
	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"

	"encoding/base64"
	"strconv"

	"github.com/OscarISTUedu/MEDODS_Service/internal"
	"github.com/OscarISTUedu/MEDODS_Service/internal/database/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const AccessTokenExpire = 600
const WebHook = "https://127.0.0.1"
const TokenRefreshBefore = 300 * time.Second

// TokenPairResponse represents a pair of new tokens
type TokenPairResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"dGhpcyBpcyBhbiBleGFtcGxlIG9mIHJlZnJlc2ggdG9rZW4="`
}

type LogoutResponse struct {
	Status string `json:"Status" example:"success LogOut"`
}

// UpdateTokens godoc
// @Summary Update access and refresh tokens
// @Description Refreshes both access and refresh tokens. Validates current tokens, checks User-Agent/IP consistency, and issues new tokens.
// @Tags authentication
// @Produce json
// @Failure 400 {object} ErrorResponse "Bad Request - invalid token or parsing error"
// @Failure 401 {object} ErrorResponse "Unauthorized - access token not provided"
// @Failure 403 {object} ErrorResponse "Forbidden - User-Agent mismatch"
// @Failure 404 {object} ErrorResponse "Not Found - refresh token not found"
// @Failure 500 {object} ErrorResponse "Internal Server Error - database or env file error"
// @Success 200 {object} TokenPairResponse "New tokens pair"
// @example {"access_token": "new_jwt_token", "refresh_token": "base64_encoded_refresh_token"}
// @Router /update_tokens [post]
func UpdateTokens(c *gin.Context) {
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
			fmt.Println("Ошибка:", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "failed to parse access token"})
		}
		if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
			userId := uint(claims["user_id"].(float64))
			refreshTokenId := uint(claims["refresh_token_id"].(float64))
			var CurRefreshToken models.RefreshToken
			result := database.DB.Where("id = ? AND deleted_at IS NULL", refreshTokenId).First(&CurRefreshToken)
			if result.Error != nil {
				if errors.Is(result.Error, gorm.ErrRecordNotFound) {
					c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "access token has no refresh token"})
				} else {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				}
				return
			}
			if CurRefreshToken.UserAgent != c.GetHeader("User-Agent") {
				LogOut(c)
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "User-Agent has switched"})
				return
			}
			if CurRefreshToken.IpAdress != c.ClientIP() {
				utils.SendPost(userId, CurRefreshToken.IpAdress, c.ClientIP(), WebHook)
			}
			new_refresh_token_id, new_refresh_token := CreateRefreshToken(userId, c.ClientIP(), c.GetHeader("User-Agent"))
			refresh_token_base64 := base64.StdEncoding.EncodeToString([]byte(new_refresh_token))
			newAccessToken := CreateAccessToken(userId, new_refresh_token_id)
			c.SetCookie("access_token", newAccessToken, AccessTokenExpire, "/", "", true, true)
			c.JSON(http.StatusOK, gin.H{
				"access_token":  newAccessToken,
				"refresh_token": refresh_token_base64,
			})
		} else {
			fmt.Println("Ошибка:", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid access token"})
		}
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "access token is not provided"})
	}
}

// LogOut godoc
// @Summary Log out user
// @Description Invalidates the current refresh token and clears access token cookie. Requires valid access token.
// @Tags authentication
// @Produce json
// @Failure 400 {object} ErrorResponse "Bad Request - invalid token or parsing error"
// @Failure 401 {object} ErrorResponse "Unauthorized - access token not provided"
// @Failure 500 {object} ErrorResponse "Internal Server Error - database or env file error"
// @Success 200 {object} LogoutResponse "Successful logout"
// @example {"Status": "success LogOut"}
// @Router /logout [post]
func LogOut(c *gin.Context) {
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
			fmt.Println("Ошибка:", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "failed to parse access token"})
		}
		if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
			refreshTokenId := uint(claims["refresh_token_id"].(float64))
			token := &models.RefreshToken{Model: gorm.Model{ID: refreshTokenId}}
			result := database.DB.Delete(token)
			if result.Error != nil {
				log.Fatal(result.Error)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to LogOut"})
			}
			c.SetCookie(
				"access_token", // name
				"",             // value (пустое)
				-1,             // maxAge: -1 (удалить куку)
				"/",            // path
				"",             // domain
				true,           // secure (HTTPS)
				true,           // httpOnly
			)
			c.JSON(http.StatusOK, gin.H{"Status": "success LogOut"})

		} else {
			fmt.Println("Ошибка:", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid access token"})
		}

	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "access token is not provided"})
	}
}

func CreateAccessToken(user_id uint, refresh_token_id uint) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":          user_id,
		"exp":              time.Now().Add(time.Duration(AccessTokenExpire) * time.Second).Unix(),
		"refresh_token_id": refresh_token_id,
	})
	internal.LoadEnv()
	SecretKey := os.Getenv("SECRET_KEY")
	access_token, _ := token.SignedString([]byte(SecretKey))
	return access_token
}

func CreateRefreshToken(user_id uint, client_ip string, client_user_agent string) (uint, string) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})
	internal.LoadEnv()
	SecretKey := os.Getenv("SECRET_KEY")
	refresh_token, _ := token.SignedString([]byte(SecretKey))
	if len(refresh_token) > 72 {
		refresh_token = string([]byte(refresh_token)[:71])
	}
	hashedRefreshToken, _ := bcrypt.GenerateFromPassword([]byte(refresh_token), bcrypt.DefaultCost)
	refresh_token = string(hashedRefreshToken)
	new_refresh_token := models.RefreshToken{
		Token:     refresh_token,
		UserID:    user_id,
		UserAgent: client_user_agent,
		IpAdress:  client_ip,
	}
	database.DB.Create(&new_refresh_token)
	return new_refresh_token.ID, refresh_token
}

// GetJwtTokens godoc
// @Summary Get JWT tokens
// @Description Generates new access and refresh tokens for unauthenticated user. Requires user ID.
// @Tags authentication
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Failure 400 {object} ErrorResponse "Bad Request - missing user ID or already authorized"
// @Failure 404 {object} ErrorResponse "Not Found - user not found"
// @Failure 425 {object} ErrorResponse "Too Early - user already has valid access token"
// @Success 200 {object} TokenPairResponse "New tokens pair"
// @example {"access_token": "new_jwt_token", "refresh_token": "base64_encoded_refresh_token"}
// @Router /auth/tokens/{id} [get]
func GetJwtTokens(c *gin.Context) {
	_, err := c.Cookie("access_token")
	if err != nil {
		user_id := c.Param("id")
		if user_id == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
			return
		}
		userId, _ := strconv.Atoi(user_id)
		var found_user models.User
		err := database.DB.First(&found_user, userId).Error
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		clientIP := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")
		refresh_token_id, refresh_token := CreateRefreshToken(found_user.ID, clientIP, userAgent)
		refresh_token_base64 := base64.StdEncoding.EncodeToString([]byte(refresh_token))
		access_token := CreateAccessToken(found_user.ID, refresh_token_id)
		c.SetCookie("access_token", access_token, AccessTokenExpire, "/", "", true, true)
		c.JSON(http.StatusOK, gin.H{
			"access_token":  access_token,
			"refresh_token": refresh_token_base64,
		})
	} else {
		c.JSON(http.StatusTooEarly, gin.H{"error": "already authorized"})
	}
}
