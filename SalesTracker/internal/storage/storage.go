package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/v1adis1av28/level3/SalesTracker/internal/config"
	"github.com/v1adis1av28/level3/SalesTracker/internal/models"
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
	CREATE TABLE IF NOT EXISTS SALES(
	ID SERIAL PRIMARY KEY,
	PRICE INTEGER NOT NULL,
	NAME VARCHAR(255) NOT NULL,
	TYPE VARCHAR(64) NOT NULL,
	CREATED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP);`)
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

func (s *Storage) CreateItem(item *models.Item) error {

	query := `INSERT INTO SALES (PRICE, NAME, TYPE) VALUES ($1, $2, $3) RETURNING ID;`
	stmt, err := s.DB.Master.Prepare(query)
	if err != nil {
		return fmt.Errorf("error preparing statement: %v", err)
	}
	var id int
	err = stmt.QueryRow(item.Price, item.Name, item.Type).Scan(&id)
	if err != nil {
		return fmt.Errorf("error inserting item: %v", err)
	}

	zlog.Logger.Debug().Msgf("Item created with ID: %d", id)

	return nil
}

func (s *Storage) GetItems() ([]models.Item, error) {
	arr := make([]models.Item, 0)

	query := "SELECT S.NAME,S.PRICE,S.TYPE FROM SALES AS S;"
	stmt, err := s.DB.Master.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("error on preparing %v", err)
	}
	rows, err := stmt.Query()
	if err != nil {
		return nil, fmt.Errorf("error on querying %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Item
		err := rows.Scan(&item.Name, &item.Price, &item.Type)
		if err != nil {
			return nil, fmt.Errorf("error on scanning %v", err)
		}
		arr = append(arr, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error on rows iteration %v", err)
	}

	return arr, nil
}

func (s *Storage) GetItemByID(id int) (*models.Item, error) {
	query := "SELECT S.ID, S.NAME, S.PRICE, S.TYPE FROM SALES AS S WHERE S.ID = $1;"
	stmt, err := s.DB.Master.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("error on preparing %v", err)
	}
	row := stmt.QueryRow(id)

	var item models.Item
	err = row.Scan(&item.ID, &item.Name, &item.Price, &item.Type)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("item with ID %d not found", id)
		}
		return nil, fmt.Errorf("error on scanning %v", err)
	}

	return &item, nil
}
