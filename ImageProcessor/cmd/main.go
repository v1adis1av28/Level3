package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/v1adis1av28/level3/ImageProcessor/internal/config"
	"github.com/v1adis1av28/level3/ImageProcessor/internal/server"
	"github.com/v1adis1av28/level3/ImageProcessor/internal/storage"
	"github.com/wb-go/wbf/zlog"
)

func main() {
	zlog.InitConsole()
	zlog.SetLevel("debug")
	cfg, err := config.New("./config/dev.yml")
	if err != nil {
		fmt.Println("Error on reading config %v", err)
		os.Exit(1)
	}

	storage, err := storage.New(&cfg.DB)
	if err != nil {
		fmt.Println("error on setting storage %v", err)
		os.Exit(1)
	}
	server := server.New(&cfg.Server, storage)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		err := server.HttpServer.ListenAndServe()
		if err != nil {
			log.Fatal("error on serving http server")
		}
	}()
	<-done
}
