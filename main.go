package main

import (
	"log"

	"engine/controllers"
	"engine/database"
	"engine/libs"
	"engine/models"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
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
	dsn := "host=" + viper.GetString("DB_HOST") + " user=" + viper.GetString("DB_USERNAME") + " password=" + viper.GetString("DB_PASSWORD") + " dbname=" + viper.GetString("DB_DATABASE") + " port=" + viper.GetString("DB_PORT") + " sslmode=disable TimeZone=Asia/Jakarta"
	database.DB, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})
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

	// Start router
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
		v1.GET("/me", libs.Authorization, controllers.AuthUser)

		v1.GET("/users", libs.Authorization, controllers.UserIndex)
		v1.POST("/users", libs.Authorization, controllers.UserPost)
		v1.PATCH("/users/:id", libs.Authorization, controllers.UserPatch)
		v1.GET("/users/:id", libs.Authorization, controllers.UserShow)
		v1.DELETE("/users/:id", libs.Authorization, controllers.UserDelete)
	}
	err = router.Run(":" + viper.GetString("APP_PORT"))
	if err != nil {
		panic(err)
	}
}
