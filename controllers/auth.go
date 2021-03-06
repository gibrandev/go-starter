package controllers

import (
	"engine/database"
	"engine/libs"
	"engine/models"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type Login struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

type Register struct {
	Name     string `validate:"required"`
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

func AuthLogin(c *gin.Context) {
	validate := validator.New()
	var user models.User
	email := c.PostForm("email")
	password := c.PostForm("password")

	// Validation
	data := Login{
		Email:    email,
		Password: password,
	}
	err := validate.Struct(data)
	if err != nil {
		errorMessage := libs.FormattingErrors(err)
		c.JSON(400, errorMessage)
		return
	}

	// Query
	result := database.DB.Where("email = ?", email).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{
			"message": "User not found",
		})
		return
	}
	if result.Error != nil {
		c.JSON(400, gin.H{
			"message": "User invalid",
		})
		return
	}
	// Validate password
	checkPassword := libs.ComparePasswords(user.Password, []byte(password))
	if checkPassword {
		token := libs.GenerateToken(user.ID, "user", c)
		if token != nil {
			c.JSON(200, gin.H{
				"token": token,
			})
			return
		} else {
			c.JSON(400, gin.H{
				"message": "Failed generate token",
			})
			return
		}
	} else {
		c.JSON(400, gin.H{
			"message": "Password invalid",
		})
		return
	}
}

func AuthRegister(c *gin.Context) {
	validate := validator.New()
	name := c.PostForm("name")
	email := c.PostForm("email")
	password := c.PostForm("password")
	data := Register{
		Name:     name,
		Email:    email,
		Password: password,
	}
	// Validation
	err := validate.Struct(data)
	if err != nil {
		errorMessage := libs.FormattingErrors(err)
		c.JSON(400, errorMessage)
		return
	}
	// DB Transaction
	tx := database.DB.Begin()
	hashPassword := libs.HashAndSalt([]byte(password))
	user := models.User{
		Name:     name,
		Email:    email,
		Password: hashPassword,
	}

	result := tx.Create(&user)
	if result.RowsAffected > 0 {
		libs.SendEmail(email)
		tx.Commit()
		c.JSON(200, gin.H{
			"message": "User has registered",
		})
		return
	} else {
		tx.Rollback()
		c.JSON(400, gin.H{
			"message": "User failed to register",
		})
		return
	}
}

func AuthUser(c *gin.Context) {
	user, _ := c.Get("user")
	c.JSON(200, user)
}

func AuthLogout(c *gin.Context) {
	logout := libs.Logout(c)
	c.Set("user", nil)
	if logout {
		c.JSON(200, gin.H{
			"message": "User has logout",
		})
		return
	} else {
		c.JSON(401, gin.H{
			"message": "User failed logout",
		})
		return
	}
}
