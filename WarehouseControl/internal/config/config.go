package config

import (
	"fmt"

	"github.com/wb-go/wbf/config"
)

type Config struct {
	Server    ServerConfig
	DB        DBConfig
	UploadDir string
	JWTConfig JWTConfig
}

type JWTConfig struct {
	Secret string
}

type ServerConfig struct {
	ListenAddr string
}

type DBConfig struct {
	Path        string
	MasterDbUrl string
	SlaverUrl   []string
}

func New(confPath string) (*Config, error) {
	conf := &Config{}
	cfg := config.New()
	err := cfg.LoadConfigFiles(confPath)
	if err != nil {
		return nil, fmt.Errorf("error loading config file: %v", err)
	}
	err = cfg.Unmarshal(conf)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %v", err)
	}
	conf.Server.ListenAddr = cfg.GetString("app.port")
	conf.DB.MasterDbUrl = cfg.GetString("db.url")
	conf.DB.SlaverUrl = []string{conf.DB.MasterDbUrl}
	conf.JWTConfig.Secret = cfg.GetString("jwt.secret")

	return conf, nil
}
