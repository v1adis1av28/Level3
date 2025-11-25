package validate

import (
	"github.com/v1adis1av28/Level3/WarehouseControl/internal/models"
)

func IsValidLoginRequest(req *models.LoginRequest) bool {
	if len(req.Role) < 1 || len(req.Username) < 1 {
		return false
	}
	return true
}

func IsValidItemRequest(item *models.Item) bool {
	if item.Quantity < 0 || len(item.Name) < 1 || len(item.Description) < 1 {
		return false
	}
	return true
}
