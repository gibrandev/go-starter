package controllers

import (
	"example.com/m/database"
	"example.com/m/libs"
	"example.com/m/models"
	"github.com/gin-gonic/gin"
)

func UserIndex(c *gin.Context) {
	var users []models.User
	// get existing auth user
	// data, _ := c.Get("user")
	database.DB.Order("created_at desc").Scopes(libs.Paginate(c)).Find(&users)
	c.JSON(200, users)
}

func UserPost(c *gin.Context) {
	name := c.PostForm("name")
	email := c.PostForm("email")
	password := c.PostForm("password")
	hashPassword := libs.HashAndSalt([]byte(password))
	user := models.User{
		Name:     name,
		Email:    email,
		Password: hashPassword,
	}

	result := database.DB.Create(&user)
	if result.RowsAffected > 0 {
		c.JSON(200, gin.H{
			"message": "User has created",
		})
	} else {
		c.JSON(400, gin.H{
			"message": "User failed to create",
		})
	}
}
func UserPatch(c *gin.Context) {
	id := c.Param("id")
	name := c.PostForm("name")
	email := c.PostForm("email")
	password := c.PostForm("password")
	hashPassword := libs.HashAndSalt([]byte(password))
	user := models.User{
		ID: id,
	}

	result := database.DB.First(&user)
	if result.RowsAffected > 0 {
		user.Name = name
		user.Email = email
		user.Password = hashPassword
		database.DB.Save(&user)
		c.JSON(200, gin.H{
			"message": "User has updated",
		})
	} else {
		c.JSON(400, gin.H{
			"message": "User failed to update",
		})
	}
}

func UserShow(c *gin.Context) {
	var user models.User
	id := c.Param("id")
	result := database.DB.Where("id = ?", id).First(&user)
	if result.Error == nil {
		c.JSON(200, user)
	} else {
		c.JSON(404, gin.H{
			"message": "User not found",
		})
	}
}

func UserDelete(c *gin.Context) {
	var user models.User
	id := c.Param("id")
	result := database.DB.Where("id = ?", id).Delete(&user)
	if result.RowsAffected > 0 {
		c.JSON(200, gin.H{
			"message": "User has deleted",
		})
	} else {
		c.JSON(400, gin.H{
			"message": "User failed to delete",
		})
	}
}
