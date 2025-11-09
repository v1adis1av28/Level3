package models

import (
	"time"
)

type ImageTask struct {
	ID           string            `json:"id"`
	FileName     string            `json:"file_name"`
	OriginalPath string            `json:"original_path"`
	Status       string            `json:"status"`
	Versions     map[string]string `json:"versions,omitempty"`
	CreatedAt    time.Time         `json:"created_at,omitempty"`
}

type ImageProcessor interface {
	SaveImage(task *ImageTask) error
	GetImage(id string) (*ImageTask, error)
	DeleteImage(id string) error
	UpdateImageStatus(id, status string) error
	AddProcessedVersion(id, versionName, filePath string) error
}
