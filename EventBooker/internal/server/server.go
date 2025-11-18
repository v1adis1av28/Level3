package server

import (
	"net/http"

	"github.com/v1adis1av28/level3/eventbooker/internal/config"
	"github.com/v1adis1av28/level3/eventbooker/internal/handlers"
	"github.com/v1adis1av28/level3/eventbooker/internal/storage"
	"github.com/wb-go/wbf/ginext"
)

type Server struct {
	Router     *ginext.Engine
	HttpServer *http.Server
	Storage    *storage.Storage
}

func New(serverConfig *config.ServerConfig, storage *storage.Storage) *Server {
	server := &Server{Router: ginext.New(""), Storage: storage}

	server.Router.Use(func(c *ginext.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
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
	s.Router.Static("/static", "./internal/web/static")

	s.Router.POST("/events", handlers.CreateEvent(s.Storage))
	s.Router.POST("/events/:id/book", handlers.BookSeat(s.Storage))
	s.Router.POST("/events/:id/confirm", handlers.ConfirmBook(s.Storage))
	s.Router.GET("/events/:id", handlers.GetEvent(s.Storage))

}
