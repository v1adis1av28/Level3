package config

import (
	"fmt"

	"github.com/wb-go/wbf/config"
)

type Config struct {
	Server    ServerConfig `yaml:"server"`
	Kafka     KafkaConfig  `yaml:"kafka"`
	DB        DBConfig     `yaml:"db"`
	UploadDir string       `yaml:"upload_dir"`
}

type ServerConfig struct {
	ListenAddr string `yaml:"listenAddr"`
}

type KafkaConfig struct {
	Addr  string `yaml:"addr"`
	Topic string `yaml:"topic"`
}

type DBConfig struct {
	Path        string `yaml:"path"`
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
	err = cfg.Unmarshal(conf)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %v", err)
	}

	conf.DB.MasterDbUrl = conf.DB.Path
	conf.DB.SlaverUrl = []string{conf.DB.Path}

	if conf.UploadDir == "" {
		conf.UploadDir = "./uploads"
	}

	return conf, nil
}
