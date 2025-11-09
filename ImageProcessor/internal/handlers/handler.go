package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/v1adis1av28/level3/ImageProcessor/internal/kafka"
	"github.com/v1adis1av28/level3/ImageProcessor/internal/models"
	"github.com/v1adis1av28/level3/ImageProcessor/internal/storage"
	"github.com/wb-go/wbf/ginext"
)

type ImageHandlers struct {
	storage   *storage.Storage
	producer  *kafka.Producer
	uploadDir string
}

func NewImageHandlers(storage *storage.Storage, producer *kafka.Producer, uploadDir string) *ImageHandlers {
	return &ImageHandlers{
		storage:   storage,
		producer:  producer,
		uploadDir: uploadDir,
	}
}

func (h *ImageHandlers) UploadPicture(c *ginext.Context) {
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
		"image/gif":  true,
	}

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "Failed to read file"})
		return
	}

	fileType := http.DetectContentType(buffer)
	if !allowedTypes[fileType] {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "Invalid file type. Only JPEG, PNG, GIF images are allowed"})
		return
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "Failed to reset file pointer"})
		return
	}

	id := uuid.New().String()

	ext := ".jpg"
	switch fileType {
	case "image/jpeg", "image/jpg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/gif":
		ext = ".gif"
	}

	if _, err := os.Stat(h.uploadDir); os.IsNotExist(err) {
		err = os.MkdirAll(h.uploadDir, 0755)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ginext.H{"error": "Failed to create upload directory"})
			return
		}
	}

	imageDir := filepath.Join(h.uploadDir, id)
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

	task := &models.ImageTask{
		ID:           id,
		FileName:     header.Filename,
		Status:       "uploaded",
		OriginalPath: fmt.Sprintf("%s/original%s", id, ext),
		Versions:     make(map[string]string),
	}

	err = h.storage.SaveImage(task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "Failed to save image metadata"})
		return
	}

	taskData, err := json.Marshal(task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "Failed to prepare task"})
		return
	}

	err = h.producer.SendMessage("images", taskData)
	if err != nil {
		h.storage.UpdateImageStatus(id, "failed")
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "Failed to queue task"})
		return
	}

	h.storage.UpdateImageStatus(id, "processing")

	c.JSON(http.StatusOK, ginext.H{
		"id":       id,
		"filename": header.Filename,
		"size":     header.Size,
		"type":     fileType,
		"status":   "processing",
	})
}

func (h *ImageHandlers) GetImage(c *ginext.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "Image ID is required"})
		return
	}

	task, err := h.storage.GetImage(id)
	if err != nil {
		c.JSON(http.StatusNotFound, ginext.H{"error": "Image not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *ImageHandlers) DeleteImage(c *ginext.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "Image ID is required"})
		return
	}

	err := h.storage.DeleteImage(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "Failed to delete image"})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"message": "Image deleted successfully"})
}

func (h *ImageHandlers) ServeImage(c *ginext.Context) {
	path := c.Param("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "Path is required"})
		return
	}

	if strings.Contains(path, "..") {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "Invalid path"})
		return
	}

	filePath := filepath.Join(h.uploadDir, path)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, ginext.H{"error": "File not found"})
		return
	}

	c.File(filePath)
}
