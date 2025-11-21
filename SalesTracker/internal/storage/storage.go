package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

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

	if item.Date.IsZero() {
		item.Date = time.Now()
	}

	query := `INSERT INTO SALES (PRICE, NAME, TYPE,CREATED_AT) VALUES ($1, $2, $3, $4) RETURNING ID;`
	stmt, err := s.DB.Master.Prepare(query)
	if err != nil {
		return fmt.Errorf("error preparing statement: %v", err)
	}
	var id int
	err = stmt.QueryRow(item.Price, item.Name, item.Type, item.Date).Scan(&id)
	if err != nil {
		return fmt.Errorf("error inserting item: %v", err)
	}

	zlog.Logger.Debug().Msgf("Item created with ID: %d", id)

	return nil
}

func (s *Storage) GetItems() ([]models.Item, error) {
	arr := make([]models.Item, 0)

	query := "SELECT S.NAME,S.PRICE,S.TYPE, S.CREATED_AT FROM SALES AS S;"
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
		err := rows.Scan(&item.Name, &item.Price, &item.Type, &item.Date)
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

func (s *Storage) UpdateItemByID(id int, item *models.Item) error {
	query := "UPDATE SALES SET PRICE = $1, NAME = $2, TYPE = $3 WHERE ID = $4;"
	stmt, err := s.DB.Master.Prepare(query)
	if err != nil {
		return fmt.Errorf("error on preparing %v", err)
	}
	result, err := stmt.Exec(item.Price, item.Name, item.Type, id)
	if err != nil {
		return fmt.Errorf("error on exec %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error on getting rows affected %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no item found with ID %d to update", id)
	}

	return nil
}

func (s *Storage) DeleteItemByID(id int) error {
	query := "DELETE FROM SALES WHERE ID = $1;"
	stmt, err := s.DB.Master.Prepare(query)
	if err != nil {
		return fmt.Errorf("error on preparing %v", err)
	}
	result, err := stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("error on exec %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error on getting rows affected %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no item found with ID %d to delete", id)
	}

	return nil
}

func (s *Storage) GetAnalytics(req *models.AnalyticsRequest) (*models.AnalyticsResponse, error) {
	exists, err := s.isTypeExist(req.Type)
	if err != nil && !exists {
		return nil, fmt.Errorf("type %s does not exist", req.Type)
	}

	var respone models.AnalyticsResponse
	respone.Type = req.Type

	var sum, count int
	var avg, median, percentile90 float64 //,
	query := "SELECT COALESCE(SUM(PRICE),0), COUNT(*), COALESCE(AVG(PRICE),0) FROM SALES WHERE TYPE = $1"
	args := []interface{}{req.Type}
	median, err = s.medianCount(req)
	if err != nil {
		return nil, fmt.Errorf("error on getting median %v", err)
	}
	percentile90, err = s.percentile90Count(req)
	if err != nil {
		return nil, fmt.Errorf("error on getting percentile90 %v", err)
	}
	if req.Date != nil {
		query += " AND DATE(CREATED_AT) = $2"
		args = append(args, req.Date.Format("2006-01-02"))
	} else {
		if req.From != nil {
			query += " AND DATE(CREATED_AT) >= $2"
			args = append(args, req.From.Format("2006-01-02"))
		}
		if req.To != nil {
			paramIndex := len(args) + 1
			query += fmt.Sprintf(" AND DATE(CREATED_AT) <= $%d", paramIndex)
			args = append(args, req.To.Format("2006-01-02"))
		}
	}

	stmt, err := s.DB.Master.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("error on preparing %v", err)
	}
	err = stmt.QueryRow(args...).Scan(&sum, &count, &avg)
	if err != nil {
		return nil, fmt.Errorf("error on querying %v", err)
	}
	respone.Sum = sum
	respone.Count = count
	respone.Avg = avg
	respone.Median = median
	respone.Percentile90 = percentile90
	return &respone, nil

}

func (s *Storage) isTypeExist(opType string) (bool, error) {
	query := "SELECT type FROM SALES WHERE TYPE = $1;"
	stmt, err := s.DB.Master.Prepare(query)
	if err != nil {
		return false, fmt.Errorf("error on preparing %v", err)
	}
	var typ string
	err = stmt.QueryRow(opType).Scan(&typ)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("type %s not found", opType)
		}
		return false, fmt.Errorf("error on querying %v", err)
	}
	return len(typ) == 0, nil
}

func (s *Storage) medianCount(req *models.AnalyticsRequest) (float64, error) {
	var median float64
	query := "SELECT percentile_cont(0.5) WITHIN GROUP(ORDER BY PRICE) FROM SALES WHERE TYPE = $1"
	args := []interface{}{req.Type}

	if req.Date != nil {
		query += " AND DATE(CREATED_AT) = $2"
		args = append(args, req.Date.Format("2006-01-02"))
	} else {
		if req.From != nil {
			query += " AND DATE(CREATED_AT) >= $2"
			args = append(args, req.From.Format("2006-01-02"))
		}
		if req.To != nil {
			paramIndex := len(args) + 1
			query += fmt.Sprintf(" AND DATE(CREATED_AT) <= $%d", paramIndex)
			args = append(args, req.To.Format("2006-01-02"))
		}
	}

	stmt, err := s.DB.Master.Prepare(query)
	if err != nil {
		return -1.0, fmt.Errorf("error on preparing %v", err)
	}
	err = stmt.QueryRow(args...).Scan(&median)
	if err != nil {
		return -1.0, fmt.Errorf("error on querying %v", err)
	}
	return median, nil
}

func (s *Storage) percentile90Count(req *models.AnalyticsRequest) (float64, error) {
	var percentile90 float64
	query := "SELECT percentile_cont(0.9) WITHIN GROUP(ORDER BY PRICE) FROM SALES WHERE TYPE = $1"
	args := []interface{}{req.Type}

	if req.Date != nil {
		query += " AND DATE(CREATED_AT) = $2"
		args = append(args, req.Date.Format("2006-01-02"))
	} else {
		if req.From != nil {
			query += " AND DATE(CREATED_AT) >= $2"
			args = append(args, req.From.Format("2006-01-02"))
		}
		if req.To != nil {
			paramIndex := len(args) + 1
			query += fmt.Sprintf(" AND DATE(CREATED_AT) <= $%d", paramIndex)
			args = append(args, req.To.Format("2006-01-02"))
		}
	}

	stmt, err := s.DB.Master.Prepare(query)
	if err != nil {
		return -1.0, fmt.Errorf("error on preparing %v", err)
	}
	err = stmt.QueryRow(args...).Scan(&percentile90)
	if err != nil {
		return -1.0, fmt.Errorf("error on querying %v", err)
	}
	return percentile90, nil
}
