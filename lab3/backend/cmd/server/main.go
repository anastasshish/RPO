package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"transport-auth-server/backend/internal/db"
	"transport-auth-server/backend/internal/handlers"
	"transport-auth-server/backend/internal/middleware"
)

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func main() {
	database, err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}
	r := gin.Default()
	r.Use(cors())
	api := r.Group("/api/v1")
	h := handlers.H{DB: database}
	api.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	api.POST("/auth/login", h.Login)
	api.GET("/swagger", handlers.SwaggerPage)
	api.GET("/openapi.json", handlers.OpenAPI)
	handlers.MountCRUD(api, h, middleware.AdminOnly(), middleware.AuthRequired())
	log.Println("backend on :8080")
	log.Fatal(r.Run(":8080"))
}
