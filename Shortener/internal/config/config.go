package config

import "github.com/wb-go/wbf/config"

type Config struct {
	Server ServerConfig
	DB     DBConfig
}

type ServerConfig struct {
	Addres string
	Port   string
}

type DBConfig struct {
	MasterDbUrl string
	SlaverUrl   []string
}

func New(configPath string) (*Config, error) {
	conf := &Config{}
	cfg := config.New()
	err := cfg.Load(configPath, "", "")
	if err != nil {
		return nil, err
	}

	conf.Server.Addres = cfg.GetString("server.addr")
	conf.Server.Port = cfg.GetString("server.port")
	conf.DB.MasterDbUrl = cfg.GetString("db.MasterDB")
	conf.DB.SlaverUrl = append(conf.DB.SlaverUrl, conf.DB.MasterDbUrl)

	return conf, nil
}
