package validation

import (
	"fmt"

	"github.com/v1adis1av28/level3/CommentTree/internal/models"
)

func IsRequestValid(req *models.CreateRequest) (bool, error) {
	fmt.Println("in validation")
	if len(req.Text) < 1 {
		return false, fmt.Errorf("comment text can`t be empty")
	}
	if len(req.Username) < 1 {
		return false, fmt.Errorf("username can`t be empty")
	}
	if req.ParentId < 0 {
		return false, fmt.Errorf("parrent id can`t be negative")
	}
	return true, nil
}
