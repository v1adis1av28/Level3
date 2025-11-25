package storage

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/v1adis1av28/Level3/WarehouseControl/internal/config"
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
	CREATE TABLE IF NOT EXISTS USERS(
	ID SERIAL PRIMARY KEY,
	USERNAME VARCHAR(255) NOT NULL UNIQUE,
	ROLE VARCHAR(64) NOT NULL);`)
	if err != nil {
		log.Fatal("error on initializing user tabel %v", err)
		os.Exit(1)
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("error on exec %v", err)
	}

	stmt, err = db.Master.Prepare(`
	CREATE TABLE IF NOT EXISTS ITEMS(
	ID SERIAL PRIMARY KEY,
	QUANTITY INTEGER NOT NULL,
	NAME VARCHAR(255) NOT NULL,
	DESCRIPTION VARCHAR(64) NOT NULL,
	CREATED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	UPDATED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP);`)
	if err != nil {
		log.Fatal("error on initializing items tabel %v", err)
		os.Exit(1)
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("error on exec %v", err)
	}

	//table for triger
	stmt, err = db.Master.Prepare(`
	CREATE TABLE IF NOT EXISTS ITEM_HISTORY(
	ID SERIAL PRIMARY KEY,
	ITEM_ID INTEGER NOT NULL,
	ACTION VARCHAR(64) NOT NULL,
	OLD_VALUES INTEGER NOT NULL,
	NEW_VALUES INTEGER NOT NULL,
	CHANGED_BY VARCHAR(255) NOT NULL,
	CHANGED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (ITEM_ID) REFERENCES ITEMS(ID),
	FOREIGN KEY (CHANGED_BY) REFERENCES USERS(USERNAME));`)
	if err != nil {
		log.Fatal("error on initializing event tabel %v", err)
		os.Exit(1)
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("error on exec %v", err)
	}

	zlog.Logger.Debug().Msg("Db succesfully created and table initializeed")

	return &Storage{DB: db, Mutex: &sync.Mutex{}}, nil
}
