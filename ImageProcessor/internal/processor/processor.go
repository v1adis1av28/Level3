package processor

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/v1adis1av28/level3/ImageProcessor/internal/models"
	"github.com/v1adis1av28/level3/ImageProcessor/internal/storage"
	"golang.org/x/image/draw"
)

type ImageProcessor struct {
	storage   *storage.Storage
	basePath  string
	watermark image.Image
}

func NewImageProcessor(storage *storage.Storage, basePath string) (*ImageProcessor, error) {
	processor := &ImageProcessor{
		storage:  storage,
		basePath: basePath,
	}

	processor.createTextWatermark()

	return processor, nil
}

func (p *ImageProcessor) createTextWatermark() {
	width, height := 200, 50
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if x < 2 || x >= width-2 || y < 2 || y >= height-2 {
				img.Set(x, y, image.White)
			} else {
				img.Set(x, y, image.Transparent)
			}
		}
	}

	p.watermark = img
}

func (p *ImageProcessor) ProcessImage(message []byte) error {
	var task models.ImageTask
	err := json.Unmarshal(message, &task)
	if err != nil {
		return fmt.Errorf("error unmarshaling task: %v", err)
	}

	p.storage.UpdateImageStatus(task.ID, "processing")

	originalPath := filepath.Join(p.basePath, task.OriginalPath)

	src, err := p.loadImage(originalPath)
	if err != nil {
		p.storage.UpdateImageStatus(task.ID, "failed")
		return fmt.Errorf("error loading image: %v", err)
	}

	versions := map[string]string{
		"thumbnail":   p.generateThumbnail(task.ID, src),
		"resized":     p.generateResized(task.ID, src),
		"watermarked": p.generateWatermarked(task.ID, src),
	}

	for name, path := range versions {
		if path != "" {
			p.storage.AddProcessedVersion(task.ID, name, path)
		}
	}

	p.storage.UpdateImageStatus(task.ID, "completed")

	return nil
}

func (p *ImageProcessor) loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var img image.Image
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
	case ".png":
		img, err = png.Decode(file)
	default:
		return nil, fmt.Errorf("unsupported image format: %s", ext)
	}

	return img, err
}

func (p *ImageProcessor) saveImage(img image.Image, path string) error {
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
	case ".png":
		return png.Encode(file, img)
	default:
		return fmt.Errorf("unsupported output format: %s", ext)
	}
}

func (p *ImageProcessor) generateThumbnail(id string, src image.Image) string {
	thumbnail := imaging.Thumbnail(src, 150, 150, imaging.Lanczos)

	outputPath := filepath.Join(p.basePath, id, "thumbnail.jpg")
	err := p.saveImage(thumbnail, outputPath)
	if err != nil {
		fmt.Printf("Error saving thumbnail: %v\n", err)
		return ""
	}

	return fmt.Sprintf("%s/thumbnail.jpg", id)
}

func (p *ImageProcessor) generateResized(id string, src image.Image) string {
	resized := imaging.Resize(src, 800, 0, imaging.Lanczos)

	outputPath := filepath.Join(p.basePath, id, "resized.jpg")
	err := p.saveImage(resized, outputPath)
	if err != nil {
		fmt.Printf("Error saving resized: %v\n", err)
		return ""
	}

	return fmt.Sprintf("%s/resized.jpg", id)
}

func (p *ImageProcessor) generateWatermarked(id string, src image.Image) string {
	dst := image.NewRGBA(src.Bounds())
	draw.Draw(dst, dst.Bounds(), src, image.Point{}, draw.Src)

	watermarkPos := image.Pt(
		src.Bounds().Dx()-p.watermark.Bounds().Dx()-10,
		src.Bounds().Dy()-p.watermark.Bounds().Dy()-10,
	)

	draw.Draw(dst, p.watermark.Bounds().Add(watermarkPos), p.watermark, image.Point{}, draw.Over)

	outputPath := filepath.Join(p.basePath, id, "watermarked.jpg")
	err := p.saveImage(dst, outputPath)
	if err != nil {
		fmt.Printf("Error saving watermarked: %v\n", err)
		return ""
	}

	return fmt.Sprintf("%s/watermarked.jpg", id)
}
