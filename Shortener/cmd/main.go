package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/v1adis1av28/level3/shortener/internal/config"
	"github.com/v1adis1av28/level3/shortener/internal/server"
	"github.com/v1adis1av28/level3/shortener/internal/storage"
)

func main() {

	config, err := config.New("./config/local.yml")
	if err != nil {
		log.Fatal("Error on reading config err %v", err)
		os.Exit(1)
	}
	storage, err := storage.New(&config.DB)
	if err != nil {
		log.Fatal("error : %v", err)
		os.Exit(1)
	}

	server := server.New(&config.Server, storage)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		err := server.HttpServer.ListenAndServe()
		if err != nil {
			log.Fatal("error on serving http server")

		}
	}()

	<-done
}
