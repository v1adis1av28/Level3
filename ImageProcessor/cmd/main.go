package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/v1adis1av28/level3/ImageProcessor/internal/config"
	"github.com/v1adis1av28/level3/ImageProcessor/internal/kafka"
	"github.com/v1adis1av28/level3/ImageProcessor/internal/processor"
	"github.com/v1adis1av28/level3/ImageProcessor/internal/server"
	"github.com/v1adis1av28/level3/ImageProcessor/internal/storage"
	"github.com/wb-go/wbf/zlog"
)

func waitForDB(connectionString string, maxAttempts int) error {
	for i := 0; i < maxAttempts; i++ {
		db, err := sql.Open("postgres", connectionString)
		if err != nil {
			log.Printf("Attempt %d: Failed to open DB connection: %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}
		defer db.Close()

		err = db.Ping()
		if err != nil {
			log.Printf("Attempt %d: DB not ready yet: %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}

		return nil
	}
	return fmt.Errorf("failed to connect to database after %d attempts", maxAttempts)
}

func main() {
	zlog.InitConsole()
	zlog.SetLevel("debug")

	configPath := "./config/dev.yml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = "/app/config/dev.yml"
	}

	cfg, err := config.New(configPath)
	if err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		os.Exit(1)
	}

	if kafkaAddr := os.Getenv("KAFKA_ADDR"); kafkaAddr != "" {
		cfg.Kafka.Addr = kafkaAddr
	}
	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		cfg.DB.Path = dbPath
		cfg.DB.MasterDbUrl = dbPath
		cfg.DB.SlaverUrl = []string{dbPath}
	}
	err = waitForDB(cfg.DB.Path, 10)
	if err != nil {
		log.Fatalf("fail to connect to database: %v", err)
	}

	storage, err := storage.New(&cfg.DB, cfg.UploadDir)
	if err != nil {
		fmt.Printf("error setting  storage: %v\n", err)
		os.Exit(1)
	}

	producer, err := kafka.NewProducer([]string{cfg.Kafka.Addr})
	if err != nil {
		fmt.Printf("error creating Kafka producer: %v\n", err)
		os.Exit(1)
	}
	defer producer.Close()

	imgProcessor, err := processor.NewImageProcessor(storage, cfg.UploadDir)
	if err != nil {
		fmt.Printf("error creating image processor: %v\n", err)
		os.Exit(1)
	}

	consumer, err := kafka.NewConsumer([]string{cfg.Kafka.Addr}, imgProcessor)
	if err != nil {
		fmt.Printf("error creating Kafka consumer: %v\n", err)
		os.Exit(1)
	}
	defer consumer.Close()

	err = consumer.StartConsuming(cfg.Kafka.Topic)
	if err != nil {
		fmt.Printf("error starting consumer: %v\n", err)
		os.Exit(1)
	}

	server := server.New(&cfg.Server, storage, producer, cfg.UploadDir)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Printf("server start %s", cfg.Server.ListenAddr)
		err := server.HttpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("error http server: %v", err)
		}
	}()

	<-done
}
