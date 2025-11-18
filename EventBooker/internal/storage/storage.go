package storage

import (
	"database/sql"
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
	CONFIRMATION_NEED BOOLEAN NOT NULL DEFAULT TRUE,
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
	_, err := s.DB.Master.Exec("INSERT INTO EVENTS (NAME,DESCRIPTION,CONFIRMATION_NEED,CAPACITY,CREATED_AT) values($1,$2,$3,$4,$5);", event.Name, event.Description, event.ConfirmationNeed, event.Capacity, time.Now())
	if err != nil {
		return fmt.Errorf("error on prepare statemen inserting event: %v", err)
	}
	return nil
}

func (s *Storage) BookSeat(eventId int) (bool, int, error) {
	tx, err := s.DB.Master.Begin()
	if err != nil {
		return false, 0, fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	confirmation, err := s.IsConfirmationNeed(eventId)
	if err != nil {
		return false, 0, err
	}
	stmt := `INSERT INTO BOOK (CREATED_AT, CONFIRMED, EVENT_ID) 
             VALUES ($1, $2, $3) 
             RETURNING ID`
	var bookID int

	if confirmation {
		err = tx.QueryRow(stmt, time.Now(), false, eventId).Scan(&bookID)
	} else {
		err = tx.QueryRow(stmt, time.Now(), true, eventId).Scan(&bookID)

	}
	if err != nil {
		return false, 0, fmt.Errorf("error inserting booking: %v", err)
	}
	err = tx.Commit()
	if err != nil {
		return false, 0, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return confirmation, bookID, nil
}

func (s *Storage) IsConfirmationNeed(eventId int) (bool, error) {
	var confirmationNeed bool
	query := "SELECT e.CONFIRMATION_NEED FROM events as e WHERE id = $1;"
	err := s.DB.Master.QueryRow(query, eventId).Scan(&confirmationNeed)
	if err != nil {

		if err == sql.ErrNoRows {
			return false, err
		}
		return false, fmt.Errorf("error on serching event: %w", err)
	}
	return confirmationNeed, nil
}

func (s *Storage) ConfirmBook(bookPayload *models.BookPayload) error {
	err := s.IsBookExist(bookPayload.BookId)
	if err != nil {
		return err
	}
	stmt, err := s.DB.Master.Prepare("UPDATE BOOK SET CONFIRMED = TRUE where EVENT_ID = $1 AND ID = $2;")
	if err != nil {
		return fmt.Errorf("Error on prepare statment confirming book, err: %v", err)
	}
	_, err = stmt.Exec(bookPayload.EventId, bookPayload.BookId)
	if err != nil {
		return fmt.Errorf("error on executing confirming book with id %v", bookPayload.BookId)
	}
	zlog.Logger.Debug().Msgf("Confirm book with id %v", bookPayload.BookId)
	return nil
}

func (s *Storage) IsBookExist(bookId int) error {
	var exist bool
	query := "SELECT 1 FROM book WHERE id = $1;"
	err := s.DB.Master.QueryRow(query, bookId).Scan(&exist)
	if err != nil {
		if err == sql.ErrNoRows {
			return err
		}
		return fmt.Errorf("error on serching book: %w", err)
	}
	return nil
}

func (s *Storage) GetEventById(eventId int) (*models.Event, error) {
	var Event models.Event
	query := "SELECT E.NAME, E.DESCRIPTION, E.CAPACITY,E.CREATED_AT, E.CONFIRMATION_NEED FROM EVENTS AS E WHERE ID = $1;"
	stmt, err := s.DB.Master.Prepare(query)
	if err != nil {
		return nil, err
	}
	err = stmt.QueryRow(eventId).Scan(&Event.Name, &Event.Description, &Event.Capacity, &Event.CreatedAt, &Event.ConfirmationNeed)
	if err != nil {
		return nil, err
	}

	return &Event, nil
}
