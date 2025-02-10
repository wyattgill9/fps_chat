package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `gorm:"unique" json:"username"`
	Password string `json:"password"`
}

type Message struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	From    string `json:"from"`
	To      string `json:"to"`
	Content string `json:"content"`
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

	r.POST("/msg", func(c *gin.Context) {
		var msg Message
		if err := c.BindJSON(&msg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		db.Create(&msg)

		c.JSON(http.StatusOK, gin.H{"message": "Message sent!"})
	})

	r.GET("/messages", func(c *gin.Context) {
		var messages []Message
		db.Find(&messages)
		c.JSON(http.StatusOK, messages)
	})

	r.Run(":8080") 
}
