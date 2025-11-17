package storage

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/v1adis1av28/level3/eventbooker/internal/config"
	"github.com/v1adis1av28/level3/eventbooker/internal/models"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

type Storage struct {
	DB    *dbpg.DB
	Mutex *sync.Mutex
}

func NewStorage(dbConf *config.DBConfig) (*Storage, error) {
	db, err := dbpg.New(dbConf.MasterDbUrl, dbConf.SlaverUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("error in initializing storge :%v", err)
	}

	stmt, err := db.Master.Prepare(`
	CREATE TABLE IF NOT EXISTS EVENTS(
	ID SERIAL PRIMARY KEY,
	CAPACITY INTEGER NOT NULL,
	NAME VARCHAR(255) NOT NULL,
	DESCRIPTION VARCHAR(512) NOT NULL,
	CREATED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP);`)
	if err != nil {
		log.Fatal("error on initializing event tabel %v", err)
		os.Exit(1)
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("error on exec %v", err)
	}

	stmt, err = db.Master.Prepare(`
	CREATE TABLE IF NOT EXISTS BOOK(
	ID SERIAL PRIMARY KEY,
	CREATED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	CONFIRMED BOOLEAN NOT NULL DEFAULT FALSE,
	EVENT_ID INTEGER REFERENCES EVENTS(ID));`)
	if err != nil {
		return nil, fmt.Errorf("error prepare book table %v", err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("error on exec book table %v", err)
	}

	zlog.Logger.Debug().Msg("Db succesfully created and table initializeed")

	return &Storage{DB: db, Mutex: &sync.Mutex{}}, nil
}

func (s *Storage) CreateEvent(event *models.Event) error {
	_, err := s.DB.Master.Exec("INSERT INTO EVENTS (NAME,DESCRIPTION,CAPACITY,CREATED_AT) values($1,$2,$3,$4);", event.Name, event.Description, event.Capacity, time.Now())
	if err != nil {
		return fmt.Errorf("error on prepare statemen inserting event: %v", err)
	}
	return nil
}
