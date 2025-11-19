package handlers

import (
	"github.com/v1adis1av28/level3/SalesTracker/internal/models"
	"github.com/v1adis1av28/level3/SalesTracker/internal/storage"
	"github.com/wb-go/wbf/ginext"
)

type StorageInterface interface {
	CreateItem(item *models.Item) error
	// Other storage methods can be defined here
}

func CreateItem(storage *storage.Storage) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		// Handler logic to create an item
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
		// Handler logic to get all items
	}
}

func GetItemByID(storage *storage.Storage) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		// Handler logic to get an item by ID
	}
}

func UpdateItemByID(storage *storage.Storage) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		// Handler logic to update an item by ID
	}
}

func DeleteItemByID(storage *storage.Storage) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		// Handler logic to delete an item by ID
	}
}
