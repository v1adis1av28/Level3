package storage

import (
	"context"
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
		return nil, fmt.Errorf("error in initializing storage: %v", err)
	}

	stmt, err := db.Master.Prepare(`
	CREATE TABLE IF NOT EXISTS USERS(
		ID SERIAL PRIMARY KEY,
		USERNAME VARCHAR(255) NOT NULL UNIQUE,
		ROLE VARCHAR(64) NOT NULL
	);`)
	if err != nil {
		log.Fatal("error on initializing users table %v", err)
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
		UPDATED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		log.Fatal("error on initializing items table %v", err)
		os.Exit(1)
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("error on exec %v", err)
	}

	stmt, err = db.Master.Prepare(`
	CREATE TABLE IF NOT EXISTS ITEM_HISTORY(
    ID SERIAL PRIMARY KEY,
    ITEM_ID INTEGER NOT NULL,
    ACTION VARCHAR(64) NOT NULL,
    OLD_VALUES INTEGER,
    NEW_VALUES INTEGER,
    CHANGED_BY VARCHAR(64) NOT NULL,
    CHANGED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (ITEM_ID) REFERENCES ITEMS(ID) ON DELETE CASCADE
);`)
	if err != nil {
		log.Fatal("error on initializing item_history table %v", err)
		os.Exit(1)
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("error on exec %v", err)
	}

	stmt, err = db.Master.Prepare(`
		CREATE OR REPLACE FUNCTION log_item_changes()
		RETURNS TRIGGER AS $$
		DECLARE
			current_user_name TEXT;
			current_user_role TEXT;
		BEGIN
			-- Получаем имя пользователя из сессии
			current_user_name := current_setting('app.current_user', true);

			-- Если не установлено, используем USER (пользователь БД)
			IF current_user_name IS NULL THEN
				current_user_name := USER;
			END IF;

			-- Получаем роль пользователя из таблицы USERS
			SELECT ROLE INTO current_user_role FROM USERS WHERE USERNAME = current_user_name;

			-- Если пользователя нет, бросаем ошибку
			IF current_user_role IS NULL THEN
				RAISE EXCEPTION 'User % does not exist in USERS table', current_user_name;
			END IF;

			IF TG_OP = 'INSERT' THEN
				INSERT INTO ITEM_HISTORY (ITEM_ID, ACTION, OLD_VALUES, NEW_VALUES, CHANGED_BY, CHANGED_AT)
				VALUES (NEW.ID, 'INSERT', NULL, NEW.QUANTITY, current_user_role, CURRENT_TIMESTAMP);
			ELSIF TG_OP = 'UPDATE' THEN
				INSERT INTO ITEM_HISTORY (ITEM_ID, ACTION, OLD_VALUES, NEW_VALUES, CHANGED_BY, CHANGED_AT)
				VALUES (NEW.ID, 'UPDATE', OLD.QUANTITY, NEW.QUANTITY, current_user_role, CURRENT_TIMESTAMP);
			ELSIF TG_OP = 'DELETE' THEN
				INSERT INTO ITEM_HISTORY (ITEM_ID, ACTION, OLD_VALUES, NEW_VALUES, CHANGED_BY, CHANGED_AT)
				VALUES (OLD.ID, 'DELETE', OLD.QUANTITY, NULL, current_user_role, CURRENT_TIMESTAMP);
			END IF;
			RETURN NULL;
		END;
		$$ LANGUAGE plpgsql;`)
	if err != nil {
		log.Fatal("error on initializing trigger function %v", err)
		os.Exit(1)
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("error on exec creating trigger function %v", err)
	}

	var triggerExists bool
	err = db.Master.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM pg_trigger
			WHERE tgname = 'trg_items_audit'
			AND tgrelid = 'items'::regclass
		);`).Scan(&triggerExists)
	if err != nil {
		return nil, fmt.Errorf("error checking trigger existence: %v", err)
	}

	if !triggerExists {
		stmt, err := db.Master.Prepare(`
			CREATE TRIGGER trg_items_audit
			AFTER INSERT OR UPDATE OR DELETE ON ITEMS
			FOR EACH ROW
			EXECUTE FUNCTION log_item_changes();`)
		if err != nil {
			log.Fatal("error on initializing trigger: %v", err)
			os.Exit(1)
		}
		_, err = stmt.Exec()
		if err != nil {
			return nil, fmt.Errorf("error on exec creating trigger: %v", err)
		}
	} else {
		zlog.Logger.Debug().Msg("Trigger trg_items_audit already exists, skipping creation.")
	}

	_, err = db.Master.Exec(`
		INSERT INTO USERS (USERNAME, ROLE) VALUES ('postgres', 'admin') ON CONFLICT (USERNAME) DO NOTHING;
	`)
	if err != nil {
		return nil, fmt.Errorf("error inserting default user: %v", err)
	}

	zlog.Logger.Debug().Msg("DB successfully created and tables initialized")

	return &Storage{DB: db, Mutex: &sync.Mutex{}}, nil
}

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

func (s *Storage) CreateItem(item *models.Item, username string) error {
	ctx := context.Background()
	tx, err := s.DB.Master.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error on starting transaction: %v", err)
	}
	defer tx.Rollback()

	var quotedUser string
	err = tx.QueryRowContext(ctx, "SELECT quote_literal($1)", username).Scan(&quotedUser)
	if err != nil {
		return fmt.Errorf("error quoting username: %v", err)
	}

	_, err = tx.ExecContext(ctx, fmt.Sprintf("SET LOCAL \"app.current_user\" = %s;", quotedUser))
	if err != nil {
		return fmt.Errorf("error on setting current user: %v", err)
	}

	stmt, err := s.DB.Master.Prepare("INSERT INTO ITEMS(QUANTITY, NAME, DESCRIPTION) VALUES($1,$2,$3);")
	if err != nil {
		zlog.Logger.Err(err)
		return fmt.Errorf("error on prepare statement insert item, err: %v", err)
	}
	_, err = stmt.ExecContext(ctx, item.Quantity, item.Name, item.Description)
	if err != nil {
		zlog.Logger.Err(err).Msgf("error on executing insert operation item: %v", item)
		return fmt.Errorf("error on executing insert op, error: %v", err)
	}

	zlog.Logger.Info().Msgf("Successfully insert new item, item: %v", item)
	return tx.Commit()
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

func (s *Storage) UpdateItem(itemId int, updateReq *models.Item, username string) error {
	ctx := context.Background()
	tx, err := s.DB.Master.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error on starting transaction: %v", err)
	}
	defer tx.Rollback()

	var quotedUser string
	err = tx.QueryRowContext(ctx, "SELECT quote_literal($1)", username).Scan(&quotedUser)
	if err != nil {
		return fmt.Errorf("error quoting username: %v", err)
	}

	_, err = tx.ExecContext(ctx, fmt.Sprintf("SET LOCAL \"app.current_user\" = %s;", quotedUser))
	if err != nil {
		return fmt.Errorf("error on setting current user: %v", err)
	}

	query := "UPDATE ITEMS SET NAME=$1, DESCRIPTION=$2, QUANTITY=$3, UPDATED_AT=CURRENT_TIMESTAMP WHERE ID=$4;"
	stmt, err := s.DB.Master.Prepare(query)
	if err != nil {
		zlog.Logger.Err(err)
		return fmt.Errorf("error on prepare statement update item, err: %v", err)
	}
	_, err = stmt.Exec(updateReq.Name, updateReq.Description, updateReq.Quantity, itemId)
	if err != nil {
		zlog.Logger.Err(err).Msgf("error on executing update operation item: %v", updateReq)
		return fmt.Errorf("error on executing update op, error: %v", err)
	}

	zlog.Logger.Info().Msgf("Successfully update item, item: %v", updateReq)
	return tx.Commit()
}

func (s *Storage) DeleteItem(itemId int, username string) error {
	ctx := context.Background()
	tx, err := s.DB.Master.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error on starting transaction: %v", err)
	}
	defer tx.Rollback()

	var quotedUser string
	err = tx.QueryRowContext(ctx, "SELECT quote_literal($1)", username).Scan(&quotedUser)
	if err != nil {
		return fmt.Errorf("error quoting username: %v", err)
	}

	_, err = tx.ExecContext(ctx, fmt.Sprintf("SET LOCAL \"app.current_user\" = %s;", quotedUser))
	if err != nil {
		return fmt.Errorf("error on setting current user: %v", err)
	}
	query := "DELETE FROM ITEMS WHERE ID = $1;"
	exist, err := s.isIdExist(itemId)
	if err != nil || !exist {
		return err
	}

	stmt, err := s.DB.Master.Prepare(query)
	if err != nil {
		zlog.Logger.Err(err)
		return fmt.Errorf("error on prepare statement deleting item, err: %v", err)
	}
	_, err = stmt.Exec(itemId)
	if err != nil {
		zlog.Logger.Err(err).Msgf("error on executing delete operation item: %v", itemId)
		return fmt.Errorf("error on executing delete op, error: %v", err)
	}

	zlog.Logger.Info().Msgf("Successfully delete item, itemId: %v", itemId)
	return tx.Commit()
}

func (s *Storage) isIdExist(id int) (bool, error) {
	var exist bool
	stmt, err := s.DB.Master.Prepare("SELECT EXISTS (SELECT 1 FROM items WHERE id = $1);")
	if err != nil {
		return false, err
	}

	err = stmt.QueryRow(id).Scan(&exist)
	if err != nil || !exist {
		return false, fmt.Errorf("id not found")
	}
	return true, nil
}
