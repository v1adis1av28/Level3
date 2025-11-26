package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/v1adis1av28/Level3/WarehouseControl/internal/config"
	"github.com/v1adis1av28/Level3/WarehouseControl/internal/models"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

var roles = []string{"admin", "viewer", "manager"}

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
	//TODO сделать триггер
	// 	Создать триггеры на items, которые при каждом INSERT, UPDATE, DELETE будут записывать изменения в item_history.
	// Пример: CREATE TRIGGER log_item_changes AFTER UPDATE ON items FOR EACH ROW EXECUTE FUNCTION log_item_change();
	return &Storage{DB: db, Mutex: &sync.Mutex{}}, nil
}

// для логин эндпоинта, будем возвращать ошибку,а в middleware передавать из handler loginRequest(username,role)
func (s *Storage) LoginUser(req *models.LoginRequest) error {
	exist, err := isValidRole(req.Role)
	if err != nil || !exist {
		return err
	}
	userExist, err := s.isUserExist(req.Username)
	if err != nil || userExist {
		return err //возвращаем ошибку что пользователь с таким ником уже существует
	}
	query := "INSERT INTO USERS(USERNAME, ROLE) VALUES($1,$2);"
	stmt, err := s.DB.Master.Prepare(query)
	if err != nil {
		zlog.Logger.Err(err)
		return fmt.Errorf("error on prepare statment insert users, err: %v", err)
	}
	_, err = stmt.Exec(strings.ToLower(req.Username), strings.ToLower(req.Role))
	if err != nil {
		zlog.Logger.Err(err).Msgf("error on executing insert operation loginReq: %v", req)
		return fmt.Errorf("error on execuring insert op, error: %v", err)
	}

	zlog.Logger.Info().Msgf("Succesfully insert new user, requst: %v", req)
	return nil
}

func isValidRole(role string) (bool, error) {
	valid := slices.Contains(roles, strings.ToLower(role))
	if !valid {
		return false, fmt.Errorf("invalid role, allowed only: admin, viewer, manager")
	}
	return true, nil
}

func (s *Storage) isUserExist(username string) (bool, error) {
	var exist bool
	stmt, err := s.DB.Master.Prepare("SELECT EXISTS (SELECT 1 FROM USERS WHERE USERNAME = $1);")
	if err != nil {
		return true, err
	}

	err = stmt.QueryRow(username).Scan(&exist)
	if err != nil || exist {
		return true, fmt.Errorf("username already registred")
	}
	return false, nil
}

func (s *Storage) CreateItem(item *models.Item) error {
	stmt, err := s.DB.Master.Prepare("INSERT INTO ITEMS(QUANTITY, NAME, DESCRIPTION) VALUES($1,$2,$3);")
	if err != nil {
		zlog.Logger.Err(err)
		return fmt.Errorf("error on prepare statment insert item, err: %v", err)
	}
	_, err = stmt.Exec(item.Quantity, item.Name, item.Description)
	if err != nil {
		zlog.Logger.Err(err).Msgf("error on executing insert operation item: %v", item)
		return fmt.Errorf("error on execuring insert op, error: %v", err)
	}

	zlog.Logger.Info().Msgf("Succesfully insert new item, item: %v", item)
	return nil
}

func (s *Storage) GetItems() ([]models.Item, error) {
	items := []models.Item{}
	rows, err := s.DB.Master.Query("SELECT I.ID, I.NAME, I.DESCRIPTION, I.QUANTITY FROM ITEMS AS I;")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("Items doesn`t exsts")
		} else {
			return nil, err
		}
	}
	for rows.Next() {
		var item models.Item
		err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Quantity)
		if err != nil {
			return nil, fmt.Errorf("error while scanning items! err : %v", err)
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *Storage) UpdateItem(itemId int, updateReq *models.Item) error {
	query := "UPDATE ITEMS SET NAME=$1, DESCRIPTION=$2, QUANTITY=$3, UPDATED_AT=CURRENT_TIMESTAMP WHERE ID=$4;"
	stmt, err := s.DB.Master.Prepare(query)
	if err != nil {
		zlog.Logger.Err(err)
		return fmt.Errorf("error on prepare statment update item, err: %v", err)
	}
	_, err = stmt.Exec(updateReq.Name, updateReq.Description, updateReq.Quantity, itemId)
	if err != nil {
		zlog.Logger.Err(err).Msgf("error on executing update operation item: %v", updateReq)
		return fmt.Errorf("error on execuring update op, error: %v", err)
	}

	zlog.Logger.Info().Msgf("Succesfully update item, item: %v", updateReq)
	return nil
}

func (s *Storage) DeleteItem(itemId int) error {
	//TODO implement me
	return nil
}
