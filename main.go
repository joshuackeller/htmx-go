package main

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

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

	router.GET("/", func(c *gin.Context) {
		conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
		if err != nil {
			c.JSON(400, gin.H{"error": fmt.Sprintf("unable to conn: %v\n", err)})
			return
		}
		defer conn.Close(context.Background())

		rows, err := conn.Query(context.Background(), "SELECT id, name, created_at FROM todo ORDER BY created_at DESC")
		if err != nil {
			c.JSON(400, gin.H{"error": fmt.Sprintf("unable to query: %v\n", err)})
			return
		}
		defer rows.Close()

		todos := []database.Todo{}

		for rows.Next() {
			var todo database.Todo
			err := rows.Scan(&todo.ID, &todo.Name, &todo.CreatedAt)
			if err != nil {
				c.JSON(400, gin.H{"error": fmt.Sprintf("unable to scan: %v\n", err)})
				return
			}
			todos = append(todos, todo)
		}

		if rows.Err() != nil {
			c.JSON(400, gin.H{"error": fmt.Sprintf("unable to rows: %v\n", err)})
			return
		}
		c.HTML(http.StatusOK, "", templates.Home(todos))
	})

	router.POST("/todos", func(c *gin.Context) {

		var todo database.Todo
		if err := c.ShouldBind(&todo); err != nil {
			c.JSON(400, gin.H{"error": "could not parse form data"})
			return
		}

		conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
		if err != nil {
			c.JSON(400, gin.H{"error": fmt.Sprintf("Unable to connect: %v\n", err)})
			return
		}
		defer conn.Close(context.Background())

		id, _ := gonanoid.New()
		todo.ID = id
		todo.CreatedAt = time.Now()

		var newTodo database.Todo
		createErr := conn.QueryRow(
			context.Background(),
			`INSERT INTO todo(id, name, created_at)
            VALUES($1, $2, $3)
            RETURNING id, name, created_at`,
			todo.ID, todo.Name, todo.CreatedAt).Scan(&newTodo.ID, &newTodo.Name, &newTodo.CreatedAt)
		if createErr != nil {
			c.JSON(400, gin.H{"error": fmt.Sprintf("Unable to create: %v\n", createErr)})
			return
		}
		c.HTML(http.StatusOK, "", components.Todo(newTodo))
	})

	router.DELETE("/todos/:id", func(c *gin.Context) {
		id := c.Param("id")

		conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
		if err != nil {
			c.JSON(400, gin.H{"error": fmt.Sprintf("unable to conn: %v\n", err)})
			return
		}
		defer conn.Close(context.Background())

		commandTag, err := conn.Exec(context.Background(), "delete from todo where id=$1", id)
		if err != nil {
			c.JSON(400, gin.H{"error": fmt.Sprintf("unable to delete: %v\n", err)})
			return
		}
		if commandTag.RowsAffected() != 1 {
			c.JSON(400, gin.H{"error": "unable to delete"})
			return
		}
		c.Status(http.StatusOK)
	})

	router.GET("/other", func(c *gin.Context) {
		c.HTML(http.StatusOK, "", templates.Other())
	})

	router.GET("/chat", func(c *gin.Context) {
		c.HTML(http.StatusOK, "", templates.Chat())
	})

	type MessageObject struct {
		Message string `json:"message"`
	}

	// Create a global variable to hold all active connections
	var clients = make(map[*websocket.Conn]bool)

	router.GET("/chat/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		// defer conn.Close()

		// Add this connection to the list of clients
		clients[conn] = true

		go func() {
			defer func() {
				conn.Close()
				delete(clients, conn)
			}()

			for {
				_, msgBytes, err := conn.ReadMessage()
				if err != nil {
					return
				}

				var msg MessageObject
				jsonErr := json.Unmarshal([]byte(string(msgBytes)), &msg)
				if jsonErr != nil {
					log.Fatal(err)
				}

				html := new(bytes.Buffer)
				templateErr := components.ChatMessage(msg.Message).Render(c.Request.Context(), html)
				if templateErr != nil {
					return
				}

				for client := range clients {
					if writeMessageErr := client.WriteMessage(websocket.TextMessage, html.Bytes()); writeMessageErr != nil {
						fmt.Println("ERROR", writeMessageErr)
						client.Close()
						delete(clients, client)
					}
				}
			}
		}()
	})

	if os.Getenv("GO_ENV") == "prod" {
		router.Run("0.0.0.0:443")
	} else {
		router.Run("0.0.0.0:8080")
	}

}
