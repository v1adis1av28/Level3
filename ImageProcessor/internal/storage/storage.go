package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/v1adis1av28/level3/ImageProcessor/internal/config"
	"github.com/v1adis1av28/level3/ImageProcessor/internal/models"
	"github.com/wb-go/wbf/dbpg"
)

type Storage struct {
	DB       *dbpg.DB
	Mutex    *sync.Mutex
	BasePath string
}

func New(dbConf *config.DBConfig, basePath string) (*Storage, error) {
	dbOpt := &dbpg.Options{
		MaxOpenConns:    50,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
	}
	db, err := dbpg.New(dbConf.MasterDbUrl, dbConf.SlaverUrl, dbOpt)
	if err != nil {
		log.Fatal("error on creating new storage %v", err)
		os.Exit(1)
	}

	stmt, err := db.Master.Prepare(`
		CREATE TABLE IF NOT EXISTS images(
			id VARCHAR(36) PRIMARY KEY,
			file_name VARCHAR(256) NOT NULL,
			status VARCHAR(50) DEFAULT 'uploaded',
			original_path VARCHAR(256) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		log.Fatal("error on initializing url tables, err: %v", err)
		os.Exit(1)
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("error on exec %v", err)
	}

	stmt, err = db.Master.Prepare(`
		CREATE TABLE IF NOT EXISTS image_versions(
			id SERIAL PRIMARY KEY,
			image_id VARCHAR(36) REFERENCES images(id),
			version_name VARCHAR(100) NOT NULL,
			file_path VARCHAR(256) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		log.Fatal("error on initializing image_versions table, err: %v", err)
		os.Exit(1)
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("error on exec %v", err)
	}

	return &Storage{DB: db, Mutex: &sync.Mutex{}, BasePath: basePath}, nil
}

func (s *Storage) SaveImage(task *models.ImageTask) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	stmt, err := s.DB.Master.Prepare(`
		INSERT INTO images (id, file_name, status, original_path) 
		VALUES ($1, $2, $3, $4)
	`)
	if err != nil {
		return fmt.Errorf("error preparing insert statement: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(task.ID, task.FileName, task.Status, task.OriginalPath)
	if err != nil {
		return fmt.Errorf("error inserting image: %v", err)
	}

	return nil
}

func (s *Storage) GetImage(id string) (*models.ImageTask, error) {
	row := s.DB.Master.QueryRow(`
		SELECT id, file_name, status, original_path, created_at 
		FROM images WHERE id = $1
	`, id)

	task := &models.ImageTask{
		Versions: make(map[string]string),
	}
	err := row.Scan(&task.ID, &task.FileName, &task.Status, &task.OriginalPath, &task.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("image not found")
		}
		return nil, fmt.Errorf("error querying image: %v", err)
	}

	rows, err := s.DB.Master.Query(`
		SELECT version_name, file_path 
		FROM image_versions 
		WHERE image_id = $1
	`, id)
	if err != nil {
		return nil, fmt.Errorf("error querying versions: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var versionName, filePath string
		err := rows.Scan(&versionName, &filePath)
		if err != nil {
			continue
		}
		task.Versions[versionName] = filePath
	}

	return task, nil
}

func (s *Storage) DeleteImage(id string) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	tx, err := s.DB.Master.Begin()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %v", err)
	}
	defer tx.Rollback()

	// Delete versions
	_, err = tx.Exec("DELETE FROM image_versions WHERE image_id = $1", id)
	if err != nil {
		return fmt.Errorf("error deleting versions: %v", err)
	}

	// Delete image
	_, err = tx.Exec("DELETE FROM images WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("error deleting image: %v", err)
	}

	return tx.Commit()
}

func (s *Storage) UpdateImageStatus(id, status string) error {
	_, err := s.DB.Master.Exec(`
		UPDATE images SET status = $1 WHERE id = $2
	`, status, id)
	return err
}

func (s *Storage) AddProcessedVersion(id, versionName, filePath string) error {
	_, err := s.DB.Master.Exec(`
		INSERT INTO image_versions (image_id, version_name, file_path) 
		VALUES ($1, $2, $3)
	`, id, versionName, filePath)
	return err
}
