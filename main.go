package main

import (
	"fmt"
	"htmx-go/components"
	"htmx-go/database"
	"htmx-go/gintemplrenderer"
	"htmx-go/templates"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	if os.Getenv("GO_ENV") != "prod" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env")
		}
	}

	router := gin.Default()
	router.HTMLRender = gintemplrenderer.Default

	// Disable trusted proxy warning.
	router.SetTrustedProxies(nil)

	// Adjust so this only runs in development
	router.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
		c.Header("Pragma", "no-cache")                                   // HTTP 1.0.
		c.Header("Expires", "0")                                         // Proxies.
	})

	router.Static("/public", "./public")

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require", host, port, user, password, dbname)

	router.GET("/", func(c *gin.Context) {
		db, db_err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

		if db_err != nil {
			c.String(500, "could not connect to database")
			return
		}

		var todos []database.Todo
		result := db.Order("created_at desc").Find(&todos)
		if result.Error != nil {
			log.Println(result.Error)
		}

		c.HTML(http.StatusOK, "", templates.Home(todos))
	})

	router.POST("/todos", func(c *gin.Context) {
		db, db_err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

		if db_err != nil {
			c.JSON(500, gin.H{"error": "could not connect to database"})
			return
		}
		var todo database.Todo
		if err := c.ShouldBind(&todo); err != nil {
			c.JSON(400, gin.H{"error": "could not connect to database"})
			return
		}

		id, _ := gonanoid.New()

		todo.ID = id
		todo.CreatedAt = time.Now()

		db.Create(&todo)

		c.HTML(http.StatusOK, "", components.Todo(todo))
	})

	router.DELETE("/todos/:id", func(c *gin.Context) {
		db, db_err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

		if db_err != nil {
			c.JSON(500, gin.H{"error": "could not connect to database"})
			return
		}
		id := c.Param("id")

		var todo database.Todo

		todo.ID = id

		db.Delete(&todo)
		c.Status(http.StatusOK)
	})

	router.GET("/other", func(c *gin.Context) {
		c.HTML(http.StatusOK, "", templates.Other())
	})

	if os.Getenv("GO_ENV") == "prod" {
		router.Run("0.0.0.0:443")
	} else {
		router.Run("0.0.0.0:8080")
	}

}
