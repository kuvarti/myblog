package main

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Header struct {
	Name    string `json"Name`
	Caption string `json:Caption`
	Path    string `json:Path`
}

func main() {
	server := gin.Default()

	server.Use(cors.New(cors.Config{
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
		{Name: "contents", Caption: "Ana Sayfa", Path: "/"},
		{Name: "lists", Caption: "Blog Yazilari", Path: "/lists"},
		{Name: "myProjects", Caption: "Projelerim", Path: "/"},
		{Name: "communication", Caption: "Iletisim", Path: "/"},
	}
	server.GET("/MenuList", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"menu": headers,
		})
	})
	server.Run()
}
