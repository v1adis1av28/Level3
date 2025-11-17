package validation

import (
	"fmt"

	"github.com/v1adis1av28/level3/eventbooker/internal/models"
)

func ValidateEvent(event *models.Event) (bool, error) {
	if event.Capacity < 0 {
		return false, fmt.Errorf("event capacity can`t be empty")
	}
	if len(event.Description) < 0 {
		return false, fmt.Errorf("description capacity can`t be empty")
	}
	if len(event.Name) < 1 {
		return false, fmt.Errorf("name capacity can`t be empty")
	}
	return true, nil
}
