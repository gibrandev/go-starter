package main

import (
	"log"

	"example.com/m/controllers"
	"example.com/m/database"
	"example.com/m/libs"
	"example.com/m/models"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Get value from .env
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error while reading config file %s", err)
	}
	router := gin.Default()
	// DB Connection
	dsn := "" + viper.Get("DB_USERNAME").(string) + ":" + viper.Get("DB_PASSWORD").(string) + "@tcp(" + viper.Get("DB_HOST").(string) + ":" + viper.Get("DB_PORT").(string) + ")/" + viper.Get("DB_DATABASE").(string) + "?charset=utf8mb4&parseTime=True&loc=Local"
	database.DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	// Migration
	errMigration := database.DB.AutoMigrate(&models.User{}, &models.Token{})
	if errMigration != nil {
		panic(errMigration)
	}
	// Cors
	router.Use(cors.Default())
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "hai",
		})
	})
	// V1
	v1 := router.Group("/v1")
	{
		v1.POST("/login", controllers.AuthLogin)
		v1.POST("/register", controllers.AuthRegister)
		v1.GET("/logout", libs.Authorization, controllers.AuthLogout)
		v1.GET("/users", libs.Authorization, controllers.UserIndex)
		v1.POST("/users", libs.Authorization, controllers.UserPost)
		v1.PATCH("/users/:id", libs.Authorization, controllers.UserPatch)
		v1.GET("/users/:id", libs.Authorization, controllers.UserShow)
		v1.DELETE("/users/:id", libs.Authorization, controllers.UserDelete)
	}
	err = router.Run(":" + viper.Get("APP_PORT").(string))
	if err != nil {
		panic(err)
	}
}
