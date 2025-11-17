package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/v1adis1av28/level3/eventbooker/internal/config"
	"github.com/v1adis1av28/level3/eventbooker/internal/server"
	"github.com/v1adis1av28/level3/eventbooker/internal/storage"
	"github.com/wb-go/wbf/zlog"
)

func main() {
	zlog.InitConsole()
	zlog.SetLevel("debug")
	cfg, err := config.New("config/dev.yml")
	if err != nil {
		fmt.Println("error loading config:", err)
		return
	}
	db, err := storage.NewStorage(&cfg.DB)
	_ = db
	fmt.Println(cfg)
	server := server.New(&cfg.Server, db)
	go func() {
		err := server.HttpServer.ListenAndServe()
		if err != nil {
			log.Fatal("error on serving http server")
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done

	zlog.Logger.Debug().Msg("Graccefully shutdown")
}
