package handlers

import (
	"github.com/v1adis1av28/level3/CommentTree/internal/storage"
	"github.com/wb-go/wbf/ginext"
)

// s.Router.GET("/comments", handlers.GetComments(c))
// 	s.Router.POST("/comments", handlers.CreateComment(c))
// 	s.Router.DELETE("/comments/:id", handlers.DeleteComment(c))

func GetComments(c *ginext.Context, storage *storage.Storage) ginext.HandlerFunc {
	return func(c *ginext.Context) {

	}
}

func CreateComment(c *ginext.Context, storage *storage.Storage) ginext.HandlerFunc {
	return func(c *ginext.Context) {

	}
}

func DeleteComment(c *ginext.Context, storage *storage.Storage) ginext.HandlerFunc {
	return func(c *ginext.Context) {

	}
}
