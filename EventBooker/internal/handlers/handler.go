package handlers

import (
	"net/http"
	"time"

	"github.com/v1adis1av28/level3/eventbooker/internal/models"
	"github.com/v1adis1av28/level3/eventbooker/internal/storage"
	"github.com/v1adis1av28/level3/eventbooker/internal/validation"
	"github.com/wb-go/wbf/ginext"
)

type EventHandler interface {
	CreateEvent(event *models.Event) error
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
