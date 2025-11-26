package repository

import (
	"database/sql"
	"errors"
	"time"

	"enterprise-blog/internal/database"
	"enterprise-blog/internal/models"

	"github.com/google/uuid"
)

type TagRepository struct{}

func NewTagRepository() *TagRepository {
	return &TagRepository{}
}

func (r *TagRepository) Create(tag *models.Tag) error {
	query := `
		INSERT INTO tags (id, name, slug, color, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	
	now := time.Now()
	tag.ID = uuid.New()
	tag.CreatedAt = now
	tag.UpdatedAt = now

	row := database.DB.Raw(
		query, tag.ID, tag.Name, tag.Slug, tag.Color, tag.CreatedAt, tag.UpdatedAt,
	).Row()
	return row.Scan(&tag.ID)
}

func (r *TagRepository) GetByID(id uuid.UUID) (*models.Tag, error) {
	tag := &models.Tag{}
	query := `SELECT id, name, slug, color, created_at, updated_at FROM tags WHERE id = $1`
	
	err := database.DB.Raw(query, id).Scan(tag).Error
	if err == sql.ErrNoRows {
		return nil, errors.New("tag not found")
	}
	return tag, err
}

func (r *TagRepository) GetBySlug(slug string) (*models.Tag, error) {
	tag := &models.Tag{}
	query := `SELECT id, name, slug, color, created_at, updated_at FROM tags WHERE slug = $1`
	
	err := database.DB.Raw(query, slug).Scan(tag).Error
	if err == sql.ErrNoRows {
		return nil, errors.New("tag not found")
	}
	return tag, err
}

func (r *TagRepository) Update(tag *models.Tag) error {
	query := `
		UPDATE tags 
		SET name = $2, slug = $3, color = $4, updated_at = $5
		WHERE id = $1
	`
	
	tag.UpdatedAt = time.Now()
	result := database.DB.Exec(query, tag.ID, tag.Name, tag.Slug, tag.Color, tag.UpdatedAt)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("tag not found")
	}
	return nil
}

func (r *TagRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM tags WHERE id = $1`
	result := database.DB.Exec(query, id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("tag not found")
	}
	return nil
}

func (r *TagRepository) List() ([]*models.Tag, error) {
	var tags []*models.Tag
	query := `SELECT id, name, slug, color, created_at, updated_at FROM tags ORDER BY name ASC`
	
	err := database.DB.Raw(query).Scan(&tags).Error
	return tags, err
}

