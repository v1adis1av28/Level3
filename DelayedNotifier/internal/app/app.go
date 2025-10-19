package app

import (
	"context"
	"net/http"
	"time"

	"github.com/v1adis1av28/level3/DelayedNotifier/internal/config"
	"github.com/v1adis1av28/level3/DelayedNotifier/internal/handlers"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

type App struct {
	DB      *dbpg.DB
	Router  *ginext.Engine
	Server  *http.Server
	Config  *config.Config
	Handler *handlers.Handler
}

func NewApp(db *dbpg.DB, cfg *config.Config, handler *handlers.Handler) *App {
	router := ginext.New("")

	router.Use(func(c *ginext.Context) {
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

	server := &http.Server{
		Addr:    cfg.ServerConfig.Addr,
		Handler: router,
	}

	app := &App{
		DB:      db,
		Router:  router,
		Server:  server,
		Config:  cfg,
		Handler: handler,
	}

	err := app.SetupRoutes()
	if err != nil {
		panic(err)
	}

	return app
}

func (a *App) SetupRoutes() error {
	a.Router.GET("/notify/:id", a.Handler.GetNotificationHandler)
	a.Router.POST("/notify", a.Handler.CreateNotificationHandler)
	a.Router.DELETE("/notify/:id", a.Handler.DeleteNotificationHandler)
	return nil
}

func (a *App) MustStart() {

	if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		zlog.Logger.Err(err)
	}
}

func (app *App) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Server.Shutdown(ctx); err != nil {
		zlog.Logger.Err(err)
	} else {
		zlog.Logger.Debug().Msg("server stoped gracefully")
	}
}
