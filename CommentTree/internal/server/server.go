package server

import (
	"net/http"

	"github.com/v1adis1av28/level3/CommentTree/internal/config"
	"github.com/v1adis1av28/level3/CommentTree/internal/handlers"
	"github.com/v1adis1av28/level3/CommentTree/internal/middleware"
	"github.com/v1adis1av28/level3/CommentTree/internal/storage"
	"github.com/wb-go/wbf/ginext"
)

type Server struct {
	Router     *ginext.Engine
	HttpServer *http.Server
	Storage    *storage.Storage
}

func New(serverConfig *config.ServerConfig, storage *storage.Storage) *Server {
	server := &Server{Router: ginext.New(""), Storage: storage}

	server.Router.Use(middleware.LoggingMiddleware())
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
		Addr:    serverConfig.Port,
		Handler: server.Router,
	}

	server.setupRoutes()

	return server
}

func (s *Server) setupRoutes() {
	c := &ginext.Context{}
	s.Router.GET("/comments", handlers.GetComments(c, s.Storage))
	s.Router.POST("/comments", handlers.CreateComment(c, s.Storage))
	s.Router.DELETE("/comments/:id", handlers.DeleteComment(c, s.Storage))
}
