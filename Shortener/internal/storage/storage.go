package storage

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/v1adis1av28/level3/shortener/internal/config"
	"github.com/wb-go/wbf/dbpg"
)

type Storage struct {
	DB *dbpg.DB
}

func New(dbConf *config.DBConfig) (*Storage, error) {
	dbOpt := &dbpg.Options{
		MaxOpenConns:    50,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
	}
	db, err := dbpg.New(dbConf.MasterDbUrl, dbConf.SlaverUrl, dbOpt)
	if err != nil {
		log.Fatal("error on creating new storage %v", err)
		os.Exit(1)
	}

	stmt, err := db.Master.Prepare(`
		CREATE TABLE IF NOT EXISTS URL(
			ID SERIAL PRIMARY KEY,
			URL VARCHAR(255) NOT NULL,
			ALIAS VARCHAR(128) UNIQUE
		);
`)
	if err != nil {
		log.Fatal("error on initializing url tables, err: %v", err)
		os.Exit(1)
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("error on exec %v", err)
	}

	stmt, err = db.Master.Prepare(`
		CREATE TABLE IF NOT EXISTS ANALYTIC(
			ID SERIAL PRIMARY KEY,
			ALIAS VARCHAR(128),
			USER_AGENT VARCHAR(255),
			REQUEST_TIME TIMESTAMP,
			CONSTRAINT ALIAS_NAME FOREIGN KEY(ALIAS) REFERENCES URL(ALIAS)
		);
	`)

	if err != nil {
		log.Fatal("error on initializing analytic tables, err: %v", err)
		os.Exit(1)
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("error on exec %v", err)
	}

	return &Storage{DB: db}, nil
}
