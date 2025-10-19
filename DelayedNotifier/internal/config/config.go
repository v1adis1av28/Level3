package config

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/rabbitmq"
)

type serverConfig struct {
	Addr string
}

type rabbiMQConfig struct {
	Host    string
	Retries int
	Port    string
}

type DbConfig struct {
	DB     string
	Slaves []string
}

type Config struct {
	ServerConfig   serverConfig
	RabbiMQConfig  rabbiMQConfig
	DBConfig       DbConfig
	ConsumerConfig *rabbitmq.ConsumerConfig
}

func NewAppConfig() (*Config, error) {
	appConfig := &Config{}
	cfg := config.New()
	err := cfg.Load("./config/dev.yml", "", "")
	if err != nil {
		return appConfig, fmt.Errorf("failed to load config: %w", err)
	}
	appConfig.ServerConfig.Addr = cfg.GetString("server.addr")
	appConfig.RabbiMQConfig.Host = cfg.GetString("rabbitMQ.host")
	appConfig.RabbiMQConfig.Port = cfg.GetString("rabbitMQ.port")
	appConfig.RabbiMQConfig.Retries = cfg.GetInt("rabbitMQ.retries")
	appConfig.DBConfig.DB = cfg.GetString("db.masterDB")

	appConfig.DBConfig.Slaves = append(appConfig.DBConfig.Slaves, appConfig.DBConfig.DB)

	consConf := &rabbitmq.ConsumerConfig{
		Queue:     "notifications",
		Consumer:  "notification-processor",
		AutoAck:   false,
		Exclusive: false,
		NoLocal:   false,
		NoWait:    false,
		Args: amqp091.Table{
			"x-priority": 5,
		},
	}
	appConfig.ConsumerConfig = consConf
	return appConfig, nil
}
