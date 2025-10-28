package handlers

import (
	"net/http"

	"github.com/v1adis1av28/level3/CommentTree/internal/models"
	"github.com/v1adis1av28/level3/CommentTree/internal/validation"
	"github.com/wb-go/wbf/ginext"
)

// s.Router.GET("/comments", handlers.GetComments(c))
// 	s.Router.POST("/comments", handlers.CreateComment(c))
// 	s.Router.DELETE("/comments/:id", handlers.DeleteComment(c))

type CommentsHandler interface {
	CreateComment(req *models.CreateRequest) error
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

		c.JSON(http.StatusOK, ginext.H{"result": "Succefully create comment"})
	}
}

func GetComments(c *ginext.Context, storage CommentsHandler) ginext.HandlerFunc {
	return func(c *ginext.Context) {

	}
}

func DeleteComment(c *ginext.Context, storage CommentsHandler) ginext.HandlerFunc {
	return func(c *ginext.Context) {

	}
}
