package main

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Header struct {
	Name    string `json:"name"`
	Caption string `json:"Caption"`
}

func main() {
	r := gin.Default()

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "http://localhost:5173"
		},
		MaxAge: 12 * time.Hour,
	}))

	headers := []Header{
		{Name: "contents", Caption: "Ana Sayfa"},
		{Name: "lists", Caption: "Blog Yazilari"},
		{Name: "myProjects", Caption: "Projelerim"},
		{Name: "communication", Caption: "Iletisim"},
	}
	r.GET("/MenuList", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"menu": headers,
		})
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}
