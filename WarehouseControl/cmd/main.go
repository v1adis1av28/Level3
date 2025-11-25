package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/v1adis1av28/Level3/WarehouseControl/internal/config"
	"github.com/v1adis1av28/Level3/WarehouseControl/internal/server"
	"github.com/v1adis1av28/Level3/WarehouseControl/internal/storage"
	"github.com/wb-go/wbf/zlog"
)

func main() {
	zlog.InitConsole()
	zlog.SetLevel("debug")

	cfg, err := config.New("config/dev.yml")
	if err != nil {
		zlog.Logger.Err(err).Msg(err.Error())
		os.Exit(1)
	}
	db, err := storage.NewStorage(&cfg.DB)
	if err != nil {
		zlog.Logger.Err(err).Msg(err.Error())
		os.Exit(1)
	}
	server := server.New(&cfg.Server, db)

	go func() {
		zlog.Logger.Info().Msgf("Starting server on %s", cfg.Server.ListenAddr)
		if err := server.HttpServer.ListenAndServe(); err != nil {
			zlog.Logger.Err(err).Msg("Server stopped with error")
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
}
