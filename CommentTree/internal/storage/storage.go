package storage

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/v1adis1av28/level3/CommentTree/internal/config"
	"github.com/v1adis1av28/level3/CommentTree/internal/models"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
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
			CREATE TABLE IF NOT EXISTS COMMENTS(
				ID SERIAL PRIMARY KEY,
				PARRENT_ID INT DEFAULT 0,
				TEXT VARCHAR(256),
				USERNAME VARCHAR(128) NOT NULL
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

func (s *Storage) CreateComment(req *models.CreateRequest) error {
	stmt, err := s.DB.Master.Prepare("INSERT INTO COMMENTS (PARRENT_ID,TEXT,USERNAME) VALUES($1,$2,$3);")
	if err != nil {
		return fmt.Errorf("error on processing prepare statment %v", err)
	}
	_, err = stmt.Exec(req.ParrentId, req.Text, req.Username)
	if err != nil {
		return fmt.Errorf("error on executing insert query %v", err)
	}
	zlog.Logger.Info().Msg("Create comment succesfully")
	return nil
}
