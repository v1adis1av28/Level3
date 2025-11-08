package storage

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/v1adis1av28/level3/ImageProcessor/internal/config"
	"github.com/wb-go/wbf/dbpg"
)

type Storage struct {
	DB    *dbpg.DB
	Mutex *sync.Mutex
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
		CREATE TABLE IF NOT EXISTS images(
			id SERIAL PRIMARY KEY,
			file_name varchar(256) not null,
			status VARCHAR(256) DEFAULT 'Uploaded',
			original_path VARCHAR(256) NOT NULL,
			processed_path VARCHAR(256) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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

	return &Storage{DB: db, Mutex: &sync.Mutex{}}, nil
}
