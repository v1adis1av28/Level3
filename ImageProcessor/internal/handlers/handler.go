package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/ginext"
)

type ImageProcessor interface {
}

func GetPicture(c *ginext.Context, storage ImageProcessor) ginext.HandlerFunc {
	return func(c *ginext.Context) {

	}
}

func UploadPicture(c *ginext.Context, storage ImageProcessor) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		fmt.Println("get")
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, ginext.H{"error": "failed to get file from form"})
			return
		}
		defer file.Close()
		allowedTypes := map[string]bool{
			"image/jpeg": true,
			"image/jpg":  true,
			"image/png":  true,
		}
		buffer := make([]byte, 512)
		_, err = file.Read(buffer)
		if err != nil {
			c.JSON(http.StatusBadRequest, ginext.H{"error": "Failed to read file"})
			return
		}

		fileType := http.DetectContentType(buffer)
		if !allowedTypes[fileType] {
			c.JSON(http.StatusBadRequest, ginext.H{"error": "Invalid file type. Only images are allowed"})
			return
		}

		_, err = file.Seek(0, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ginext.H{
				"error": "Failed to reset file pointer",
			})
			return
		}

		ext := ".jpg"
		switch fileType {
		case "image/jpeg", "image/jpg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		case "image/webp":
			ext = ".webp"
		}

		uploadDir := "./uploads"
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			err = os.MkdirAll(uploadDir, 0755)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to create upload directory",
				})
				return
			}
		}

		imageDir := filepath.Join(uploadDir, header.Filename)
		err = os.MkdirAll(imageDir, 0755)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ginext.H{"error": "Failed to create image directory"})
			return
		}

		originalPath := filepath.Join(imageDir, "original"+ext)
		dst, err := os.Create(originalPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ginext.H{"error": "Failed to create file on server"})
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ginext.H{"error": "Failed to save file"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"filename": header.Filename,
			"size":     header.Size,
			"type":     fileType,
			"url":      fmt.Sprintf("/images/%s/original%s", header.Filename, ext),
			"status":   "uploaded",
		})
	}
}

func DeletePicture(c *ginext.Context, storage ImageProcessor) ginext.HandlerFunc {
	return func(c *ginext.Context) {}
}
