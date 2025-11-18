package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/v1adis1av28/level3/eventbooker/internal/config"
	"github.com/v1adis1av28/level3/eventbooker/internal/server"
	"github.com/v1adis1av28/level3/eventbooker/internal/storage"
	"github.com/v1adis1av28/level3/eventbooker/internal/worker"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

const (
	PaymentDeadlineMinutes = 1
	WorkerInterval         = 30 * time.Second
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
	strategy := retry.Strategy{
		Attempts: 3,
		Delay:    time.Second,
		Backoff:  2,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go worker.ExpiredBookingsWorker(ctx, strategy, WorkerInterval, PaymentDeadlineMinutes, db)

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done

	zlog.Logger.Debug().Msg("Graccefully shutdown")
}
