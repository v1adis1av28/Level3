package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lib/pq"
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

func (s *Storage) GetURL(alias string) (string, error) {
	stmt, err := s.DB.Master.Prepare("Select U.URL from URL as U where alias = $1")
	if err != nil {
		return "", fmt.Errorf("error while getting url! error : %v", err)
	}
	var result string
	err = stmt.QueryRow(alias).Scan(&result)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("no url using this alias")
		}

		return "", err
	}

	return result, nil
}

func (s *Storage) ShortenURL(url, alias string) error {
	_, err := s.DB.Master.Exec("INSERT INTO URL (url, alias) values($1,$2);", url, alias)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { //unique constraint error handler
				return fmt.Errorf("this alias already present")
			}
		}
		return fmt.Errorf("error on prepare statemen inserting short url, err %v", err)
	}
	return nil
}

func (s *Storage) UpdateStats(alias, UA string) error {
	_, err := s.DB.Master.Exec("INSERT INTO ANALYTIC (ALIAS,USER_AGENT,REQUEST_TIME) VALUES($1,$2,$3);", alias, UA, time.Now())
	if err != nil {
		return fmt.Errorf("error on updating analytic")
	}
	return nil
}
