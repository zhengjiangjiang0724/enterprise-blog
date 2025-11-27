package repository

import (
	"database/sql"
	"errors"
	"time"

	"enterprise-blog/internal/database"
	"enterprise-blog/internal/models"

	"github.com/google/uuid"
)

type CategoryRepository struct{}

func NewCategoryRepository() *CategoryRepository {
	return &CategoryRepository{}
}

func (r *CategoryRepository) Create(category *models.Category) error {
	query := `
		INSERT INTO categories (id, name, slug, description, parent_id, "order", created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	
	now := time.Now()
	category.ID = uuid.New()
	category.CreatedAt = now
	category.UpdatedAt = now

	row := database.DB.Raw(
		query,
		category.ID, category.Name, category.Slug, category.Description,
		category.ParentID, category.Order, category.CreatedAt, category.UpdatedAt,
	).Row()
	return row.Scan(&category.ID)
}

func (r *CategoryRepository) GetByID(id uuid.UUID) (*models.Category, error) {
	category := &models.Category{}
	query := `SELECT id, name, slug, description, parent_id, "order", created_at, updated_at
			  FROM categories WHERE id = $1`
	
	err := database.DB.Raw(query, id).Scan(category).Error
	if err == sql.ErrNoRows {
		return nil, errors.New("category not found")
	}
	return category, err
}

func (r *CategoryRepository) Update(category *models.Category) error {
	query := `
		UPDATE categories 
		SET name = $2, slug = $3, description = $4, parent_id = $5, "order" = $6, updated_at = $7
		WHERE id = $1
	`
	
	category.UpdatedAt = time.Now()
	result := database.DB.Exec(query, category.ID, category.Name, category.Slug,
		category.Description, category.ParentID, category.Order, category.UpdatedAt)
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return errors.New("category not found")
	}
	return nil
}

func (r *CategoryRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM categories WHERE id = $1`
	result := database.DB.Exec(query, id)
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return errors.New("category not found")
	}
	return nil
}

func (r *CategoryRepository) List() ([]*models.Category, error) {
	var categories []*models.Category
	query := `SELECT id, name, slug, description, parent_id, "order", created_at, updated_at
			  FROM categories ORDER BY "order" ASC, created_at DESC`
	
	err := database.DB.Raw(query).Scan(&categories).Error
	return categories, err
}

