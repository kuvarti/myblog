package main

import (
	controllers "backend/Controllers"
	services "backend/Services"
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	server *gin.Engine
	ctx context.Context
	mongoClient *mongo.Client
)

func InitDB() {
	var err error
	// Connection string comes from MONGODB_URI (e.g. an Atlas mongodb+srv URI)
	// so the credential stays out of source control; falls back to the local
	// docker-compose Mongo when unset.
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://root:example@localhost:27017"
	}
	mongoconn := options.Client().ApplyURI(uri)
	mongoClient, err = mongo.Connect(ctx, mongoconn)
	if err != nil {
		log.Fatal("error while connecting with mongo", err)
	}
	err = mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal("error while trying to ping mongo", err)
	}
}

func InitServer() *gin.Engine {
	server := gin.Default()
	server.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "http://localhost:5173"
		},
		MaxAge: 12 * time.Hour,
	}))
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	server.Use(cors.New(config))
	ctx = context.TODO();
	InitDB()
	return server
}

func InitControllers(db *mongo.Database, apiGroup *gin.RouterGroup) {
	userService, tokenService := services.NewUserService(ctx, db.Collection("Users"))
	menuService := services.NewMenuService(ctx, db.Collection("Menus"))
	pageService := services.NewPageService(ctx, db.Collection("Pages"))
	cardService := services.NewCardService(pageService, menuService)

	controllers.InitUserController(userService, tokenService, apiGroup)
	controllers.InitMenuController(menuService, pageService, apiGroup)
	controllers.InitPageController(pageService, cardService, apiGroup)
	controllers.InitPanelController(pageService, menuService, cardService, tokenService, apiGroup)
}

func main() {
	server = InitServer()

	db := mongoClient.Database("KuvartiBlog");
	defer db.Client().Disconnect(ctx);
	apiGroup := server.Group("api")
	InitControllers(db, apiGroup)
	server.Run()
}
