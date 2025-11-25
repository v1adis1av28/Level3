package handlers

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/v1adis1av28/Level3/WarehouseControl/internal/models"
	"github.com/v1adis1av28/Level3/WarehouseControl/internal/storage"
	"github.com/v1adis1av28/Level3/WarehouseControl/internal/validate"
	"github.com/wb-go/wbf/ginext"
)

type ControlHandler interface {
}

type AuthHandler interface {
	LoginUser(req *models.LoginRequest) error
}

func Login(s *storage.Storage, secret string) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		var req models.LoginRequest
		err := c.ShouldBindJSON(&req)
		if err != nil {
			c.JSON(400, ginext.H{"error": err.Error()})
			return
		}
		if !validate.IsValidLoginRequest(&req) {
			c.JSON(400, ginext.H{"error": "invalid login request"})
			return
		}

		err = s.LoginUser(&req)
		if err != nil {
			if errors.Is(err, fmt.Errorf("username already registred")) {
				c.JSON(400, ginext.H{"error": err.Error()})
				return
			} else {
				c.JSON(500, ginext.H{"error": err.Error()})
				return
			}
		}
		token, err := generateToken(secret, &req)
		if err != nil {
			c.JSON(500, ginext.H{"error": err.Error()})
			return
		}

		c.Writer.Header().Set("Authorization", "Bearer "+token)
		//TODO добавить добавление jwt вместо того что сверху
		c.JSON(201, ginext.H{"result": "user succesfuly sign in", "user": req})
	}
}

func CreateItem(s *storage.Storage) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		var item models.Item
		err := c.ShouldBindJSON(&item)
		if err != nil {
			c.JSON(400, ginext.H{"error": err.Error()})
			return
		}
		if !validate.IsValidItemRequest(&item) {
			c.JSON(400, ginext.H{"error": "invalid item request"})
			return
		}

		err = s.CreateItem(&item)
		if err != nil {
			c.JSON(500, ginext.H{"error": err.Error()})
			return
		}

		c.JSON(201, ginext.H{"result": "item succesfuly created", "item": item})
	}
}

func GetItems(s *storage.Storage) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		items, err := s.GetItems()
		if err != nil {
			c.JSON(500, ginext.H{"error": err.Error()})
			return
		}

		if len(items) == 0 {
			c.JSON(200, ginext.H{"info": "there is nothing in the ites"})
			return
		}
		c.JSON(200, ginext.H{"items": items})
	}
}

// todo мб перекинуть в отдельный пакет
func generateToken(secret string, req *models.LoginRequest) (string, error) {
	claims := jwt.MapClaims{
		"role":     strings.ToLower(req.Role),
		"username": strings.ToLower(req.Username),
		"exp":      time.Now().Add(time.Hour * 4).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
