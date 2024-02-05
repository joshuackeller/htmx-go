package main

import (
	"fmt"
	"htmx-go/components"
	"htmx-go/database"
	"htmx-go/gintemplrenderer"
	"htmx-go/templates"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//  const (
//  	host     = "localhost"
//  	port     = 5432
//  	user     = "root"
//  	password = "rampart.mailbox"
//  	dbname   = "todo"
//  )

const (
	host     = "ep-hidden-scene-a4k7zue1.us-east-1.aws.neon.tech"
	port     = 5432
	user     = "joshuackeller"
	password = "EdxBOce41rKi"
	dbname   = "neondb"
)

func main() {
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

	// dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disabled", host, port, user, password, dbname)
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require", host, port, user, password, dbname)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	router.GET("/", func(c *gin.Context) {
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})

		}
		var todos []database.Todo
		result := db.Order("created_at desc").Find(&todos)
		if result.Error != nil {
			log.Println(result.Error)
		}

		c.HTML(http.StatusOK, "", templates.Home(todos))
	})

	router.POST("/todos", func(c *gin.Context) {
		var todo database.Todo
		if err := c.ShouldBind(&todo); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		id, _ := gonanoid.New()

		todo.ID = id
		todo.CreatedAt = time.Now()

		db.Create(&todo)

		c.HTML(http.StatusOK, "", components.Todo(todo))
	})

	router.DELETE("/todos/:id", func(c *gin.Context) {
		id := c.Param("id")

		var todo database.Todo

		todo.ID = id

		db.Delete(&todo)
		c.Status(http.StatusOK)
	})

	router.GET("/other", func(c *gin.Context) {
		c.HTML(http.StatusOK, "", templates.Other())
	})

	// router.Run("localhost:8080")
	router.Run("0.0.0.0:8080")
}
