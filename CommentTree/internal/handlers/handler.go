package handlers

import (
	"net/http"
	"strconv"

	"github.com/v1adis1av28/level3/CommentTree/internal/models"
	"github.com/v1adis1av28/level3/CommentTree/internal/validation"
	"github.com/wb-go/wbf/ginext"
)

type CommentsHandler interface {
	CreateComment(req *models.CreateRequest) error
	GetComments(parentId, page, limit int, search string) (*models.CommentsResponse, error)
	DeleteComment(commentId int) error
}

func CreateComment(c *ginext.Context, storage CommentsHandler) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		var req models.CreateRequest
		err := c.ShouldBindJSON(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
			return
		}

		isValid, err := validation.IsRequestValid(&req)
		if !isValid {
			c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
			return
		}

		err = storage.CreateComment(&req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, ginext.H{"result": "Successfully created comment"})
	}
}

func GetComments(c *ginext.Context, storage CommentsHandler) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		parentIdStr := c.DefaultQuery("parent", "0")
		search := c.Query("search")
		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "50")

		parentId, err := strconv.Atoi(parentIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ginext.H{"error": "Invalid parent ID"})
			return
		}

		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			limit = 50
		}

		response, err := storage.GetComments(parentId, page, limit, search)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

func DeleteComment(c *ginext.Context, storage CommentsHandler) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ginext.H{"error": "Invalid comment ID"})
			return
		}

		err = storage.DeleteComment(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, ginext.H{"result": "Successfully deleted comment and its children"})
	}
}
