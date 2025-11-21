package handlers

import (
	"fmt"
	"time"

	"github.com/v1adis1av28/level3/SalesTracker/internal/models"
	"github.com/v1adis1av28/level3/SalesTracker/internal/storage"
	"github.com/wb-go/wbf/ginext"
)

type StorageInterface interface {
	CreateItem(item *models.Item) error
	GetItems() ([]models.Item, error)
	GetItemByID(id int) (*models.Item, error)
	UpdateItemByID(id int, item *models.Item) error
	DeleteItemByID(id int) error
	GetAnalytics(req *models.AnalyticsRequest) (*models.AnalyticsResponse, error)
}

func CreateItem(storage *storage.Storage) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		var reqPayload models.Item
		err := c.ShouldBindJSON(&reqPayload)
		if err != nil {
			c.JSON(400, ginext.H{"error": "Invalid request payload"})
			return
		}

		isValid := reqPayload.Price > 0 && reqPayload.Name != "" && reqPayload.Type != ""
		if !isValid {
			c.JSON(400, ginext.H{"error": "Missing or invalid fields in request payload"})
			return
		}

		err = storage.CreateItem(&reqPayload)
		if err != nil {
			c.JSON(500, ginext.H{"error": "Failed to create item"})
			return
		}

		c.JSON(201, ginext.H{"message": "Item created successfully", "item": reqPayload})
	}
}

func GetItems(storage *storage.Storage) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		arr, err := storage.GetItems()
		if err != nil {
			c.JSON(500, ginext.H{"error": "Failed to retrieve items"})
			return
		}

		c.JSON(200, ginext.H{"items": arr})
	}
}

func GetItemByID(storage *storage.Storage) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		var idParam struct {
			ID int `uri:"id" binding:"required"`
		}
		if err := c.ShouldBindUri(&idParam); err != nil {
			c.JSON(400, ginext.H{"error": "Invalid ID parameter"})
			return
		}

		item, err := storage.GetItemByID(idParam.ID)
		if err != nil {
			c.JSON(500, ginext.H{"error": err.Error()})
			return
		}

		c.JSON(200, ginext.H{"item": item})
	}
}

func UpdateItemByID(storage *storage.Storage) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		var updatePayload models.Item
		if err := c.ShouldBindJSON(&updatePayload); err != nil {
			c.JSON(400, ginext.H{"error": "Invalid request payload"})
			return
		}

		var idParam struct {
			ID int `uri:"id" binding:"required"`
		}
		if err := c.ShouldBindUri(&idParam); err != nil {
			c.JSON(400, ginext.H{"error": "Invalid ID parameter"})
			return
		}
		err := storage.UpdateItemByID(idParam.ID, &updatePayload)
		if err != nil {
			c.JSON(500, ginext.H{"error": err.Error()})
			return
		}
		c.JSON(200, ginext.H{"message": "Item updated successfully"})
	}
}

func DeleteItemByID(storage *storage.Storage) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		var idParam struct {
			ID int `uri:"id" binding:"required"`
		}
		if err := c.ShouldBindUri(&idParam); err != nil {
			c.JSON(400, ginext.H{"error": "Invalid ID parameter"})
			return
		}

		err := storage.DeleteItemByID(idParam.ID)
		if err != nil {
			c.JSON(500, ginext.H{"error": "Failed to delete item, " + err.Error()})
			return
		}

		c.JSON(200, ginext.H{"message": "Item deleted successfully"})
	}
}

func GetAnalytics(storage *storage.Storage) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		req, err := parseAnalyticsParam(c)
		if err != nil {
			c.JSON(400, ginext.H{"error": "Invalid query parameters"})
			return
		}

		data, err := storage.GetAnalytics(req)
		if err != nil {
			c.JSON(500, ginext.H{"error": "Failed to retrieve analytics data, err: " + err.Error()})
			return
		}

		c.JSON(200, ginext.H{"analytics": data})
	}
}

func parseAnalyticsParam(c *ginext.Context) (*models.AnalyticsRequest, error) {
	var req models.AnalyticsRequest
	req.Type = c.Query("type")
	if req.Type == "" {
		return nil, fmt.Errorf("type parameter is required")
	}
	if dateStr := c.Query("date"); dateStr != "" {
		if date, err := time.Parse("2006-01-02", dateStr); err == nil {
			req.Date = &date
		} else {
			return &req, err
		}
	}

	if fromStr := c.Query("from"); fromStr != "" {
		if from, err := time.Parse("2006-01-02", fromStr); err == nil {
			req.From = &from
		} else {
			return &req, err
		}
	}

	if toStr := c.Query("to"); toStr != "" {
		if to, err := time.Parse("2006-01-02", toStr); err == nil {
			req.To = &to
		} else {
			return &req, err
		}
	}

	return &req, nil
}
