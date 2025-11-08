package config

import (
	"fmt"

	"github.com/wb-go/wbf/config"
)

type Config struct {
	Server ServerConfig
	Kafka  KafkaConfig
	DB     DBConfig
}

type ServerConfig struct {
	Addr string
}

type KafkaConfig struct {
	Addr  string
	Topic string
}

type DBConfig struct {
	MasterDbUrl string
	SlaverUrl   []string
}

func New(confPath string) (*Config, error) {
	conf := &Config{}
	cfg := config.New()
	err := cfg.LoadConfigFiles(confPath, confPath)
	if err != nil {
		return nil, fmt.Errorf("error loading config file: %v", err)
	}
	conf.Server.Addr = cfg.GetString("server.ListenAddr")
	conf.Kafka.Addr = cfg.GetString("kafka.addr")
	conf.Kafka.Topic = cfg.GetString("kafka.topic")
	conf.DB.MasterDbUrl = cfg.GetString("db.path")
	conf.DB.SlaverUrl = append(conf.DB.SlaverUrl, conf.DB.MasterDbUrl)

	return conf, nil
}
