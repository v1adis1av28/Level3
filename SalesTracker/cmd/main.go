package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/v1adis1av28/level3/SalesTracker/internal/config"
	"github.com/v1adis1av28/level3/SalesTracker/internal/storage"
	"github.com/wb-go/wbf/zlog"
)

func main() {
	zlog.InitConsole()
	zlog.SetLevel("debug")

	cfg, err := config.New("config/dev.yml")
	db, err := storage.NewStorage(&cfg.DB)
	if err != nil {
		zlog.Logger.Err(err).Msg(err.Error())
		os.Exit(1)
	}
	_ = db

	_ = cfg
	fmt.Println(cfg)

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
}
