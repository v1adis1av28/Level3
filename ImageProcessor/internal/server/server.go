package server

import (
	"fmt"
	"net/http"
	"os"

	"github.com/v1adis1av28/level3/ImageProcessor/internal/config"
	"github.com/v1adis1av28/level3/ImageProcessor/internal/handlers"
	"github.com/v1adis1av28/level3/ImageProcessor/internal/storage"
	"github.com/wb-go/wbf/ginext"
)

type Server struct {
	Router     *ginext.Engine
	HttpServer *http.Server
	Storage    *storage.Storage
}

func New(serverConfig *config.ServerConfig, storage *storage.Storage) *Server {
	server := &Server{Router: ginext.New(""), Storage: storage}

	// CORS middleware - исправленная версия
	server.Router.Use(func(c *ginext.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	})

	server.HttpServer = &http.Server{
		Addr:    serverConfig.Addr,
		Handler: server.Router,
	}

	server.setupRoutes()

	return server
}

func (s *Server) setupRoutes() {
	c := &ginext.Context{}
	s.Router.GET("/image/:id", handlers.GetPicture(c, s.Storage))
	s.Router.POST("/upload", handlers.UploadPicture(c, s.Storage))
	s.Router.DELETE("/image/:id", handlers.DeletePicture(c, s.Storage))

	staticPath := "./frontend/static"
	if _, err := os.Stat(staticPath); os.IsNotExist(err) {
		staticPath = "/app/frontend/static"
	}
	fmt.Println("Path")
	fmt.Println(staticPath)
	s.Router.Static("/static", staticPath)

	s.Router.GET("/", func(c *ginext.Context) {
		indexPath := staticPath + "/index.html"
		c.File(indexPath)
	})
}
