package repository

import (
	"database/sql"
	"errors"
	"time"

	"enterprise-blog/internal/database"
	"enterprise-blog/internal/models"

	"github.com/google/uuid"
)

type CommentRepository struct{}

func NewCommentRepository() *CommentRepository {
	return &CommentRepository{}
}

func (r *CommentRepository) Create(comment *models.Comment) error {
	query := `
		INSERT INTO comments (id, article_id, user_id, parent_id, content, author, email, website, ip, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`
	
	now := time.Now()
	comment.ID = uuid.New()
	comment.CreatedAt = now
	comment.UpdatedAt = now
	if comment.Status == "" {
		comment.Status = "pending"
	}

	row := database.DB.Raw(
		query,
		comment.ID, comment.ArticleID, comment.UserID, comment.ParentID,
		comment.Content, comment.Author, comment.Email, comment.Website,
		comment.IP, comment.Status, comment.CreatedAt, comment.UpdatedAt,
	).Row()
	return row.Scan(&comment.ID)
}

func (r *CommentRepository) GetByID(id uuid.UUID) (*models.Comment, error) {
	comment := &models.Comment{}
	query := `
		SELECT id, article_id, user_id, parent_id, content, author, email, website, ip, status, created_at, updated_at
		FROM comments WHERE id = $1
	`
	
	err := database.DB.Raw(query, id).Scan(comment).Error
	if err == sql.ErrNoRows {
		return nil, errors.New("comment not found")
	}
	if err != nil {
		return nil, err
	}

	// 加载用户信息
	if comment.UserID != nil {
		var user models.User
		err = database.DB.Raw("SELECT id, username, email, avatar FROM users WHERE id = $1", *comment.UserID).Scan(&user).Error
		if err == nil {
			comment.User = &user
		}
	}

	return comment, nil
}

func (r *CommentRepository) GetByArticleID(articleID uuid.UUID, page, pageSize int) ([]*models.Comment, int64, error) {
	var comments []*models.Comment
	var total int64

	offset := (page - 1) * pageSize

	// 获取总数
	countQuery := `SELECT COUNT(*) FROM comments WHERE article_id = $1 AND parent_id IS NULL`
	err := database.DB.Raw(countQuery, articleID).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取父评论
	query := `
		SELECT id, article_id, user_id, parent_id, content, author, email, website, ip, status, created_at, updated_at
		FROM comments
		WHERE article_id = $1 AND parent_id IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	err = database.DB.Raw(query, articleID, pageSize, offset).Scan(&comments).Error
	if err != nil {
		return nil, 0, err
	}

	// 加载用户信息和子评论
	for i := range comments {
		if comments[i].UserID != nil {
			var user models.User
			err = database.DB.Raw("SELECT id, username, email, avatar FROM users WHERE id = $1", *comments[i].UserID).Scan(&user).Error
			if err == nil {
				comments[i].User = &user
			}
		}
	}

	return comments, total, nil
}

func (r *CommentRepository) Update(comment *models.Comment) error {
	query := `
		UPDATE comments 
		SET content = $2, status = $3, updated_at = $4
		WHERE id = $1
	`
	
	comment.UpdatedAt = time.Now()
	result := database.DB.Exec(query, comment.ID, comment.Content, comment.Status, comment.UpdatedAt)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("comment not found")
	}
	return nil
}

func (r *CommentRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM comments WHERE id = $1`
	result := database.DB.Exec(query, id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("comment not found")
	}
	return nil
}

func (r *CommentRepository) IncrementCommentCount(articleID uuid.UUID) error {
	query := `UPDATE articles SET comment_count = comment_count + 1 WHERE id = $1`
	return database.DB.Exec(query, articleID).Error
}

