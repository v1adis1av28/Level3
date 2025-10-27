package main

import (
	"log"
	"os"

	"github.com/v1adis1av28/level3/CommentTree/internal/config"
	"github.com/v1adis1av28/level3/CommentTree/internal/server"
	"github.com/v1adis1av28/level3/CommentTree/internal/storage"
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
	server.HttpServer.ListenAndServe()

}
