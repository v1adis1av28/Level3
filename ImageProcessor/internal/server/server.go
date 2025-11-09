package server

import (
	"fmt"
	"net/http"
	"os"

	"github.com/v1adis1av28/level3/ImageProcessor/internal/config"
	"github.com/v1adis1av28/level3/ImageProcessor/internal/handlers"
	"github.com/v1adis1av28/level3/ImageProcessor/internal/kafka"
	"github.com/v1adis1av28/level3/ImageProcessor/internal/storage"
	"github.com/wb-go/wbf/ginext"
)

type Server struct {
	Router     *ginext.Engine
	HttpServer *http.Server
	Storage    *storage.Storage
	Producer   *kafka.Producer
	UploadDir  string
}

func New(serverConfig *config.ServerConfig, storage *storage.Storage, producer *kafka.Producer, uploadDir string) *Server {
	server := &Server{
		Router:    ginext.New(""),
		Storage:   storage,
		Producer:  producer,
		UploadDir: uploadDir,
	}

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
		Addr:    serverConfig.ListenAddr,
		Handler: server.Router,
	}

	server.setupRoutes()

	return server
}

func (s *Server) setupRoutes() {
	imageHandlers := handlers.NewImageHandlers(s.Storage, s.Producer, s.UploadDir)

	s.Router.GET("/image/:id", imageHandlers.GetImage)
	s.Router.POST("/upload", imageHandlers.UploadPicture)
	s.Router.DELETE("/image/:id", imageHandlers.DeleteImage)
	s.Router.GET("/files/*path", imageHandlers.ServeImage)

	staticPath := "./frontend/static"
	if _, err := os.Stat(staticPath); os.IsNotExist(err) {
		staticPath = "/app/frontend/static"
	}
	fmt.Printf("Serving static files from: %s\n", staticPath)
	s.Router.Static("/static", staticPath)

	s.Router.GET("/", func(c *ginext.Context) {
		indexPath := staticPath + "/index.html"
		c.File(indexPath)
	})
}
