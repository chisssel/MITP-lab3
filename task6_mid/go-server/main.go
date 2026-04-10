package main

import (
	"log"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	pythonService := "http://python-service:5000"
	rustService := "http://rust-service:4000"

	pythonURL, _ := url.Parse(pythonService)
	rustURL, _ := url.Parse(rustService)

	r.Any("/api/users/*path", func(c *gin.Context) {
		proxy := httputil.NewSingleHostReverseProxy(pythonURL)
		proxy.ServeHTTP(c.Writer, c.Request)
	})

	r.Any("/api/stats/*path", func(c *gin.Context) {
		proxy := httputil.NewSingleHostReverseProxy(rustURL)
		proxy.ServeHTTP(c.Writer, c.Request)
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	log.Println("Go API Gateway starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
