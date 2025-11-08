package handlers

import "github.com/wb-go/wbf/ginext"

type ImageProcessor interface {
}

func GetPicture(c *ginext.Context, storage ImageProcessor) ginext.HandlerFunc {

	return func(c *ginext.Context) {}
}

func UploadPicture(c *ginext.Context, storage ImageProcessor) ginext.HandlerFunc {
	return func(c *ginext.Context) {}
}

func DeletePicture(c *ginext.Context, storage ImageProcessor) ginext.HandlerFunc {
	return func(c *ginext.Context) {}
}
