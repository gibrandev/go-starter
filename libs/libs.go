package libs

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"engine/database"
	"engine/models"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
)

// Middleware
func Authorization(c *gin.Context) {
	// Check header exist or not
	tokenString := c.Request.Header.Get("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Invalid header authorization",
		})
		c.Abort()
		return
	}
	// Validate token
	token := ValidateToken(tokenString)
	if token == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "not authorized",
		})
		c.Abort()
		return
	}

	// Check user exist or not
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		var user models.User
		var token models.Token
		sub := claims["sub"]
		jti := claims["jti"]
		// Check token exist or not
		resultToken := database.DB.Where("id = ?", jti).First(&token)
		if resultToken.Error == nil && token.Sub == sub {
			result := database.DB.Where("id = ?", sub).First(&user)
			if result.Error == nil {
				now := time.Now()
				token.LastAccessAt = &now
				token.Ip = c.ClientIP()
				database.DB.Save(&token)
				c.Set("user", user)
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"message": "Invalid user",
				})
				c.Abort()
				return
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Invalid token",
			})
			c.Abort()
			return
		}
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "not authorized",
		})
		c.Abort()
		return
	}
}

func GenerateToken(sub string, iss string, c *gin.Context) *string {
	token := jwt.New(jwt.GetSigningMethod("HS256"))
	ip := c.ClientIP()
	saveToken := models.Token{
		Sub: sub,
		Ip:  ip,
		Iss: iss,
	}
	if !viper.GetBool("MULTIPLE_LOGIN") {
		// Delete old token
		database.DB.Where("sub = ?", sub).Delete(&saveToken)
	}
	// Create token
	result := database.DB.Create(&saveToken)
	if result.RowsAffected > 0 {
		token.Claims.(jwt.MapClaims)["jti"] = saveToken.ID
		token.Claims.(jwt.MapClaims)["sub"] = sub
		token.Claims.(jwt.MapClaims)["iat"] = saveToken.CreatedAt
		token.Claims.(jwt.MapClaims)["iss"] = iss
		// Sign and get the complete encoded token as a string
		tokenString, err := token.SignedString([]byte(viper.Get("JWT_SECRET").(string)))
		if err == nil {
			return &tokenString
		} else {
			return nil
		}
	} else {
		return nil
	}
}

func ParseToken(tokenString string) interface{} {
	token := ValidateToken(tokenString)
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		jti := claims["jti"]
		return jti
	} else {
		return nil
	}
}

func ValidateToken(tokenString string) *jwt.Token {
	token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if jwt.GetSigningMethod("HS256") != token.Method {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(viper.GetString("JWT_SECRET")), nil
	})
	return token
}

// func CheckTokenDb(c *gin.Context, token *jwt.Token) bool {
// 	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
// 		var user models.User
// 		var token models.Token
// 		sub := claims["sub"]
// 		jti := claims["jti"]
// 		// Check token exist or not
// 		resultToken := database.DB.Where("id = ?", jti).First(&token)
// 		if resultToken.Error == nil && token.Sub == sub {
// 			result := database.DB.Where("id = ?", sub).First(&user)
// 			if result.Error == nil {
// 				now := time.Now()
// 				token.LastAccessAt = &now
// 				token.Ip = c.ClientIP()
// 				database.DB.Save(&token)
// 				return true
// 			} else {
// 				return false
// 			}
// 		} else {
// 			return false
// 		}
// 	} else {
// 		return false
// 	}
// }

func Logout(c *gin.Context) bool {
	var token models.Token
	tokenString := c.Request.Header.Get("Authorization")
	getSub := ParseToken(tokenString)
	if getSub != nil {
		if !viper.GetBool("MULTIPLE_LOGIN") {
			// Delete old token
			database.DB.Where("sub = ?", getSub).Delete(&token)
		} else {
			database.DB.Where("id = ?", getSub).Delete(&token)
		}
		return true
	} else {
		return false
	}
}

func FormatingErrors(err error) map[string]string {
	errors, _ := err.(validator.ValidationErrors)
	e := make(map[string]string)
	for _, err := range errors {
		e[err.Field()] = err.Tag()
	}
	return e
}

func HashAndSalt(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Fatal(err)
	}
	return string(hash)
}

func ComparePasswords(hashedPwd string, plainPwd []byte) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
}

func Paginate(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		page, _ := strconv.Atoi(c.Query("page"))
		if page == 0 {
			page = 1
		}

		pageSize, _ := strconv.Atoi(c.Query("page_size"))
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

func SendEmail(email string) bool {
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(viper.GetString("MAIL_FROM_ADDRESS"), viper.GetString("MAIL_FROM_NAME")))
	m.SetHeader("To", email)
	// m.SetAddressHeader("Cc", "dan@example.com", "Dan")
	m.SetHeader("Subject", "Hello!")
	m.SetBody("text/html", "Hello <b>Bob</b> and <i>Cora</i>!")
	// m.Attach("/home/Alex/lolcat.jpg")

	d := gomail.NewDialer(viper.GetString("MAIL_HOST"), viper.GetInt("MAIL_PORT"), viper.GetString("MAIL_USERNAME"), viper.GetString("MAIL_PASSWORD"))

	if err := d.DialAndSend(m); err != nil {
		// panic(err)
		return false
	}
	return true
}

func GetChatId(query string) string {
	get := url.URL{
		RawQuery: query,
	}
	value := get.Query()
	chatId := value["chatId"]
	str := strings.Join(chatId, ", ")
	return str
}
