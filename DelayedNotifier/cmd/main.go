package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/v1adis1av28/level3/DelayedNotifier/internal/app"
	"github.com/v1adis1av28/level3/DelayedNotifier/internal/config"
	"github.com/v1adis1av28/level3/DelayedNotifier/internal/consumer"
	"github.com/v1adis1av28/level3/DelayedNotifier/internal/handlers"
	"github.com/v1adis1av28/level3/DelayedNotifier/internal/service"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/zlog"
)

func main() {
	config, err := config.NewAppConfig()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
		return
	}

	zlog.Init()
	dbOpt := &dbpg.Options{
		MaxOpenConns:    50,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
	}
	db, err := dbpg.New(config.DBConfig.DB, config.DBConfig.Slaves, dbOpt)
	if err != nil {
		fmt.Errorf("eror on creating db %w", err)
		os.Exit(1)
	}

	notificationService := service.NewNotificationService(db)
	defer notificationService.Close()

	notificationHandler := handlers.NewHandler(notificationService)
	app := app.NewApp(db, config, notificationHandler)
	msgConsumer := rabbitmq.NewConsumer(notificationService.Producer, config.ConsumerConfig)
	processor := consumer.NewNotificationProcessor(notificationService)

	go func() {
		app.MustStart()
	}()

	go func() {
		consumer.ConsumeWithShutdown(msgConsumer, processor)
	}()

	zlog.Logger.Info().Msg("service started")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	app.Stop()
	zlog.Logger.Debug().Msg("Server gracefully stopped")
}
