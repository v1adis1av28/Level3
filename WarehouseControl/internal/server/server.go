package server

import (
	"net/http"

	"github.com/v1adis1av28/Level3/WarehouseControl/internal/config"
	"github.com/v1adis1av28/Level3/WarehouseControl/internal/handlers"
	"github.com/v1adis1av28/Level3/WarehouseControl/internal/middleware"
	"github.com/v1adis1av28/Level3/WarehouseControl/internal/storage"
	"github.com/wb-go/wbf/ginext"
)

type Server struct {
	Router     *ginext.Engine
	HttpServer *http.Server
	Storage    *storage.Storage
}

func New(Config *config.Config, storage *storage.Storage) *Server {
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
		Addr:    Config.Server.ListenAddr,
		Handler: server.Router,
	}

	server.setupRoutes(Config.JWTConfig.Secret)
	return server
}

func (s *Server) setupRoutes(jwtSecret string) {
	//s.Router.Static("/static", "./internal/web/static")
	s.Router.POST("/auth/login", handlers.Login(s.Storage, jwtSecret))
	//Для остальных роутов нужно использовать мидлварь на авторизацию
	s.Router.POST("/items", middleware.AuthMiddleware(jwtSecret), handlers.CreateItem(s.Storage, jwtSecret))
	s.Router.GET("/items", middleware.AuthMiddleware(jwtSecret), handlers.GetItems(s.Storage))
	s.Router.PUT("/items/:id", middleware.AuthMiddleware(jwtSecret), handlers.UpdateItem(s.Storage, jwtSecret))
	s.Router.DELETE("/items/:id", middleware.AuthMiddleware(jwtSecret), handlers.DeleteItem(s.Storage, jwtSecret))
}
