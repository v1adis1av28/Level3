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
