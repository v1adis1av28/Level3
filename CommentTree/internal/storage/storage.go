package storage

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/v1adis1av28/level3/CommentTree/internal/config"
	"github.com/v1adis1av28/level3/CommentTree/internal/models"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

type Storage struct {
	DB    *dbpg.DB
	Mutex *sync.Mutex
}

func New(dbConf *config.DBConfig) (*Storage, error) {
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
		CREATE TABLE IF NOT EXISTS comments(
			id SERIAL PRIMARY KEY,
			parent_id INT DEFAULT 0,
			text VARCHAR(256),
			username VARCHAR(128) NOT NULL,
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

	return &Storage{DB: db, Mutex: &sync.Mutex{}}, nil
}

func (s *Storage) CreateComment(req *models.CreateRequest) error {
	stmt, err := s.DB.Master.Prepare("INSERT INTO comments (parent_id, text, username) VALUES($1, $2, $3);")
	if err != nil {
		return fmt.Errorf("error on processing prepare statment %v", err)
	}
	_, err = stmt.Exec(req.ParentId, req.Text, req.Username)
	if err != nil {
		return fmt.Errorf("error on executing insert query %v", err)
	}
	zlog.Logger.Info().Msg("Create comment succesfully")
	return nil
}

func (s *Storage) GetComments(parentId, page, limit int, search string) (*models.CommentsResponse, error) {
	offset := (page - 1) * limit

	var comments []*models.Comment
	var total int
	var err error

	if search != "" {
		comments, total, err = s.searchComments(search, page, limit, offset)
	} else {
		comments, total, err = s.getCommentsTree(parentId, page, limit, offset)
	}

	if err != nil {
		return nil, err
	}

	return &models.CommentsResponse{
		Comments: comments,
		Total:    total,
		Page:     page,
		Limit:    limit,
	}, nil
}

func (s *Storage) getCommentsTree(parentId, page, limit, offset int) ([]*models.Comment, int, error) {
	rows, err := s.DB.Master.Query(`
		SELECT id, parent_id, username, text, created_at 
		FROM comments 
		WHERE parent_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`, parentId, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(&comment.ID, &comment.ParentId, &comment.Username, &comment.Text, &comment.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		children, _, err := s.getCommentsTree(comment.ID, 1, 1000, 0)
		if err != nil {
			return nil, 0, err
		}
		comment.Children = children
		comments = append(comments, &comment)
	}

	var total int
	err = s.DB.Master.QueryRow("SELECT COUNT(*) FROM comments WHERE parent_id = $1", parentId).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

func (s *Storage) searchComments(query string, page, limit, offset int) ([]*models.Comment, int, error) {
	searchPattern := "%" + query + "%"

	rows, err := s.DB.Master.Query(`
		SELECT id, parent_id, username, text, created_at 
		FROM comments 
		WHERE text ILIKE $1 OR username ILIKE $2
		ORDER BY created_at DESC 
		LIMIT $3 OFFSET $4`, searchPattern, searchPattern, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(&comment.ID, &comment.ParentId, &comment.Username, &comment.Text, &comment.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		comments = append(comments, &comment)
	}

	var total int
	err = s.DB.Master.QueryRow("SELECT COUNT(*) FROM comments WHERE text ILIKE $1 OR username ILIKE $2",
		searchPattern, searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

func (s *Storage) DeleteComment(commentId int) error {
	stmt, err := s.DB.Master.Prepare(`
		WITH RECURSIVE comment_tree AS (
			SELECT id FROM comments WHERE id = $1
			UNION ALL
			SELECT c.id FROM comments c
			INNER JOIN comment_tree ct ON c.parent_id = ct.id
		)
		DELETE FROM comments WHERE id IN (SELECT id FROM comment_tree)
	`)
	if err != nil {
		return fmt.Errorf("error on preparing delete statement: %v", err)
	}
	_, err = stmt.Exec(commentId)
	if err != nil {
		return fmt.Errorf("error on executing delete: %v", err)
	}
	zlog.Logger.Info().Msgf("Deleted comment with ID %d and its children", commentId)
	return nil
}
