package server

import (
	"net/http"

	"github.com/v1adis1av28/level3/shortener/internal/config"
	"github.com/v1adis1av28/level3/shortener/internal/handlers/analytic"
	"github.com/v1adis1av28/level3/shortener/internal/handlers/shortener"
	"github.com/v1adis1av28/level3/shortener/internal/storage"
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

	// serverAdres := serverConfig.Addres + serverConfig.Port
	server.HttpServer = &http.Server{
		Addr:    serverConfig.Port,
		Handler: server.Router,
	}

	server.setupRoutes()

	return server
}

func (s *Server) setupRoutes() {
	c := &ginext.Context{}
	s.Router.GET("/s/:short_url", shortener.RedirectShortUrl(c, s.Storage))
	s.Router.POST("/shorten", shortener.ShortenURL(c, s.Storage))
	s.Router.GET("/analytics/:short_url", analytic.GetAnalytic)
}
