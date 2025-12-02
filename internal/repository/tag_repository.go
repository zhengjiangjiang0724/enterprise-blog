// Package repository 提供数据访问层的实现
package repository

import (
	"database/sql"
	"errors"
	"time"

	"enterprise-blog/internal/database"
	"enterprise-blog/internal/models"

	"github.com/google/uuid"
)

// TagRepository 标签数据访问层，提供标签相关的数据库操作
type TagRepository struct{}

// NewTagRepository 创建新的标签仓库实例
func NewTagRepository() *TagRepository {
	return &TagRepository{}
}

// Create 创建新标签
// tag: 标签对象，会设置ID、创建时间、更新时间
// 返回: 如果创建失败则返回错误
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

// GetByID 根据ID获取标签
// id: 标签UUID
// 返回: 标签对象，如果不存在则返回错误
func (r *TagRepository) GetByID(id uuid.UUID) (*models.Tag, error) {
	tag := &models.Tag{}
	query := `SELECT id, name, slug, color, created_at, updated_at FROM tags WHERE id = $1`
	
	err := database.DB.Raw(query, id).Scan(tag).Error
	if err == sql.ErrNoRows {
		return nil, errors.New("tag not found")
	}
	return tag, err
}

// GetBySlug 根据slug获取标签
// slug: 标签URL友好的标识符
// 返回: 标签对象，如果不存在则返回错误
func (r *TagRepository) GetBySlug(slug string) (*models.Tag, error) {
	tag := &models.Tag{}
	query := `SELECT id, name, slug, color, created_at, updated_at FROM tags WHERE slug = $1`
	
	err := database.DB.Raw(query, slug).Scan(tag).Error
	if err == sql.ErrNoRows {
		return nil, errors.New("tag not found")
	}
	return tag, err
}

// Update 更新标签信息
// tag: 标签对象，会更新更新时间
// 返回: 如果更新失败或标签不存在则返回错误
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

// Delete 删除标签（硬删除）
// id: 标签UUID
// 返回: 如果删除失败或标签不存在则返回错误
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

