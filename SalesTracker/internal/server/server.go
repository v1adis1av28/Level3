package server

import (
	"net/http"

	"github.com/v1adis1av28/level3/SalesTracker/internal/config"
	"github.com/v1adis1av28/level3/SalesTracker/internal/handlers"
	"github.com/v1adis1av28/level3/SalesTracker/internal/storage"

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
	// – POST /items;
	s.Router.POST("/items", handlers.CreateItem(s.Storage))
	// – GET /items;
	s.Router.GET("/items", handlers.GetItems(s.Storage))
	// – GET /items/{id};
	s.Router.GET("/items/:id", handlers.GetItemByID(s.Storage))
	// – PUT /items/{id};
	s.Router.PUT("/items/:id", handlers.UpdateItemByID(s.Storage))
	// – DELETE /items/{id}
	s.Router.DELETE("/items/:id", handlers.DeleteItemByID(s.Storage))

}
