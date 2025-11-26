package handlers

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/v1adis1av28/Level3/WarehouseControl/internal/jwt"
	"github.com/v1adis1av28/Level3/WarehouseControl/internal/models"
	"github.com/v1adis1av28/Level3/WarehouseControl/internal/storage"
	"github.com/v1adis1av28/Level3/WarehouseControl/internal/validate"
	"github.com/wb-go/wbf/ginext"
)

type AuthHandler interface {
	LoginUser(req *models.LoginRequest) error
}

func Login(s AuthHandler, secret string) ginext.HandlerFunc {
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
		token, err := jwt.GenerateToken(secret, &req)
		if err != nil {
			c.JSON(500, ginext.H{"error": err.Error()})
			return
		}

		c.Writer.Header().Set("Authorization", "Bearer "+token)
		c.JSON(201, ginext.H{"result": "user succesfuly sign in", "user": req})
	}
}

func GetItemHistory(s *storage.Storage, secret string) ginext.HandlerFunc {
	return func(c *ginext.Context) {

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, ginext.H{"error": "invalid item id"})
			return
		}

		history, err := s.GetItemHistory(id)
		if err != nil {
			c.JSON(500, ginext.H{"error": err.Error()})
			return
		}

		c.JSON(200, ginext.H{"history": history})
	}
}
func CreateItem(s *storage.Storage, secret string) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		payload, err := jwt.ExtractPayloadFromClaims(c.GetHeader("Authorization")[7:], secret)
		if err != nil {
			c.JSON(500, ginext.H{"error": err.Error()})
			return
		}

		var item models.Item
		err = c.ShouldBindJSON(&item)
		if err != nil {
			c.JSON(400, ginext.H{"error": err.Error()})
			return
		}
		if !validate.IsValidItemRequest(&item) {
			c.JSON(400, ginext.H{"error": "invalid item request"})
			return
		}

		err = s.CreateItem(&item, payload.Username)
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

func UpdateItem(s *storage.Storage, secret string) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		payload, err := jwt.ExtractPayloadFromClaims(c.GetHeader("Authorization")[7:], secret)
		if err != nil {
			c.JSON(500, ginext.H{"error": err.Error()})
			return
		}
		if payload.Role != "admin" && payload.Role != "manager" {
			c.JSON(403, ginext.H{"error": "you don`t have permission for changing item"})
			return
		}

		var item models.Item
		err = c.ShouldBindJSON(&item)
		if err != nil {
			c.JSON(400, ginext.H{"error": err.Error()})
			return
		}
		if !validate.IsValidItemRequest(&item) {
			c.JSON(400, ginext.H{"error": "invalid item request"})
			return
		}

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, ginext.H{"error": "invalid item id"})
			return
		}
		err = s.UpdateItem(id, &item, payload.Username)
		if err != nil {
			c.JSON(500, ginext.H{"error": err.Error()})
			return
		}

		c.JSON(200, ginext.H{"result": "item succesfuly updated", "item": item})
	}
}

func DeleteItem(s *storage.Storage, secret string) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, ginext.H{"error": "invalid item id"})
			return
		}
		payload, err := jwt.ExtractPayloadFromClaims(c.GetHeader("Authorization")[7:], secret)
		if err != nil {
			c.JSON(500, ginext.H{"error": err.Error()})
			return
		}
		if payload.Role != "admin" {
			c.JSON(403, ginext.H{"error": "you don`t have permission for deleting item"})
			return
		}

		err = s.DeleteItem(id, payload.Username)
		if err != nil {
			c.JSON(500, ginext.H{"error": err.Error()})
			return
		}

		c.JSON(200, ginext.H{"result": "deleted item", "itemId": id})
	}
}
