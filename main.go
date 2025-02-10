package main

import (
	"fmt"
	"log"
	"net/http"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
    ID       uint   `gorm:"primaryKey" json:"id"`
    Username string `gorm:"unique" json:"username"`
    Name     string `json:"name"`
    Password string `json:"password"`
}

type Message struct {
    ID         uint            `gorm:"primaryKey" json:"id"`
    From       string          `json:"from"`
    To         json.RawMessage `json:"to"`
    Content    string          `json:"content"`
}
var db *gorm.DB

func initDB() {
	dsn := "postgresql://neondb_owner:npg_Hyjd6zFMt8fu@ep-soft-dust-aa6yflf5-pooler.westus3.azure.neon.tech/neondb?sslmode=require"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	db.AutoMigrate(&User{}, &Message{})
	fmt.Println("Database connected & migrated!")
}

func main() {
	initDB()

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Welcome to FPS Chat!"})
	})

	r.POST("/register", func(c *gin.Context) {
		var user User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists!"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User registered!"})
	})

	r.POST("/login", func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		var user User
		if err := db.Where("username = ? AND password = ?", req.Username, req.Password).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User logged in!"})
	})

	
	r.POST("/msg/:from/:to", AuthMiddleware(), func(c *gin.Context) {
		from := c.Param("from")
		to := c.Param("to")
	
		authenticatedUser, exists := c.Get("user")
		if !exists || authenticatedUser.(User).Username != from {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
	
		var msg Message
		if err := c.BindJSON(&msg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}
	
		msg.From = from
		msg.To = json.RawMessage([]byte(to))
	
		if err := db.Create(&msg).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
			return
		}
	
		c.JSON(http.StatusOK, gin.H{"message": "Message sent!"})
	})
	

	r.Run(":8080") 
}

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        username := c.GetHeader("X-User") 
        if username == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }

        var user User
        if err := db.Where("username = ?", username).First(&user).Error; err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
            c.Abort()
            return
        }

        c.Set("user", user)
        c.Next()
    }
}
