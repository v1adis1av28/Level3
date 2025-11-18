package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/v1adis1av28/level3/eventbooker/internal/models"
	"github.com/v1adis1av28/level3/eventbooker/internal/storage"
	"github.com/v1adis1av28/level3/eventbooker/internal/validation"
	"github.com/wb-go/wbf/ginext"
)

type EventHandler interface {
	CreateEvent(event *models.Event) error
	BookSeat(eventId int) error
	ConfirmBook(bp *models.BookPayload) error
	GetEventById(eventId int) (*models.Event, error)
}

type ErrorResponse struct {
	StatusCode  int
	Description string
}

func CreateEvent(storage *storage.Storage) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		//получаем ивент из json парсим
		var event models.Event

		err := c.ShouldBindJSON(&event)
		if err != nil {
			c.JSON(http.StatusBadRequest, ginext.H{"error": ErrorResponse{StatusCode: 400, Description: err.Error()}})
			return
		}

		isOk, err := validation.ValidateEvent(&event)
		if !isOk {
			c.JSON(http.StatusBadRequest, ginext.H{"error": ErrorResponse{StatusCode: 400, Description: err.Error()}})
			return
		}
		//добавляем в бд запись
		err = storage.CreateEvent(&event)
		if err != nil {
			c.JSON(http.StatusBadRequest, ginext.H{"error": ErrorResponse{StatusCode: 500, Description: err.Error()}})
			return
		}
		event.CreatedAt = time.Now()
		c.JSON(http.StatusOK, ginext.H{"result": "succesfully created", "event": event})
	}
}

// POST /events/{id}/book — бронирование места;
func BookSeat(storage *storage.Storage) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		eventId, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, ginext.H{"error": ErrorResponse{StatusCode: 400, Description: err.Error()}})
			return
		}

		confirmationNeed, bookID, err := storage.BookSeat(eventId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusBadRequest, ginext.H{"error": ErrorResponse{StatusCode: 400, Description: "event not found"}})
				return
			}
			c.JSON(http.StatusBadRequest, ginext.H{"error": ErrorResponse{StatusCode: 500, Description: err.Error()}})
			return
		}

		c.JSON(http.StatusOK, ginext.H{
			"result":            "succesfully booked a seat",
			"eventId":           eventId,
			"bookId":            bookID,
			"confirmation_need": confirmationNeed,
		})
	}
}

// POST /events/{id}/confirm — оплата брони (если мероприятие требует этого);
func ConfirmBook(s *storage.Storage) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		var payload models.BookPayload
		eventId, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, ginext.H{"error": ErrorResponse{StatusCode: 400, Description: err.Error()}})
			return
		}
		err = c.ShouldBindJSON(&payload)
		if err != nil {
			c.JSON(http.StatusBadRequest, ginext.H{"error": ErrorResponse{StatusCode: 400, Description: err.Error()}})
			return
		}
		payload.EventId = eventId
		confirmation, err := s.IsConfirmationNeed(eventId)
		if !confirmation {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusBadRequest, ginext.H{"error": ErrorResponse{StatusCode: 400, Description: "event not found"}})
				return
			} else {
				c.JSON(http.StatusOK, ginext.H{"info": "This event doens`t need book confirmation"})
				return
			}
		}
		err = s.ConfirmBook(&payload)
		if err != nil {

			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusBadRequest, ginext.H{"error": ErrorResponse{StatusCode: 400, Description: "book not found"}})
				return
			} else {
				c.JSON(http.StatusInternalServerError, ginext.H{"error": ErrorResponse{StatusCode: 500, Description: err.Error()}})
				return
			}
		}
		//todo добавить обработку чтобы из очереди убиралась этот бук
		c.JSON(http.StatusOK, ginext.H{"result": "Book succesfully confirmed"})
	}
}

// GET /events/{id} — получение информации о мероприятии и свободных местах.
func GetEvent(s *storage.Storage) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		eventId, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, ginext.H{"error": ErrorResponse{StatusCode: 400, Description: err.Error()}})
			return
		}

		event, err := s.GetEventById(eventId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusBadRequest, ginext.H{"error": ErrorResponse{StatusCode: 400, Description: "event not found"}})
				return
			} else {
				c.JSON(http.StatusInternalServerError, ginext.H{"error": ErrorResponse{StatusCode: 500, Description: err.Error()}})
				return
			}
		}

		c.JSON(http.StatusOK, ginext.H{"result": "succesfully",
			"event": event})
	}
}
