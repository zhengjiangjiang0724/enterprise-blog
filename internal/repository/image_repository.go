// Package repository 提供数据访问层的实现
package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"enterprise-blog/internal/database"
	"enterprise-blog/internal/models"

	"github.com/google/uuid"
)

// ImageRepository 图片数据访问层，提供图片相关的数据库操作
type ImageRepository struct{}

// NewImageRepository 创建新的图片仓库实例
func NewImageRepository() *ImageRepository {
	return &ImageRepository{}
}

// Create 创建新图片记录
// image: 图片对象，会设置ID、创建时间、更新时间
// 返回: 如果创建失败则返回错误
func (r *ImageRepository) Create(ctx context.Context, image *models.Image) error {
	query := `
		INSERT INTO images (id, filename, original_name, path, url, mime_type, size, width, height, uploader_id, description, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id
	`

	now := time.Now()
	image.ID = uuid.New()
	image.CreatedAt = now
	image.UpdatedAt = now

	// 将tags转换为JSONB格式
	tagsJSON := "[]"
	if len(image.Tags) > 0 {
		tagsJSON = `["` + strings.Join(image.Tags, `","`) + `"]`
	}

	row := database.DB.WithContext(ctx).Raw(
		query,
		image.ID, image.Filename, image.OriginalName, image.Path, image.URL,
		image.MimeType, image.Size, image.Width, image.Height, image.UploaderID,
		image.Description, tagsJSON, image.CreatedAt, image.UpdatedAt,
	).Row()
	return row.Scan(&image.ID)
}

// GetByID 根据ID获取图片
// id: 图片UUID
// 返回: 图片对象，如果不存在则返回错误
func (r *ImageRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Image, error) {
	image := &models.Image{}
	query := `
		SELECT id, filename, original_name, path, url, mime_type, size, width, height,
		       uploader_id, description, tags, created_at, updated_at, deleted_at
		FROM images
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := database.DB.WithContext(ctx).Raw(query, id).Scan(image).Error
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("image not found")
		}
		return nil, err
	}

	// 加载上传者信息
	if err := r.loadImageRelations(ctx, image); err != nil {
		return nil, err
	}

	return image, nil
}

// Update 更新图片信息
// image: 图片对象，会更新更新时间
// 返回: 如果更新失败或图片不存在则返回错误
func (r *ImageRepository) Update(ctx context.Context, image *models.Image) error {
	query := `
		UPDATE images
		SET description = $2, tags = $3, updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL
	`

	image.UpdatedAt = time.Now()

	// 将tags转换为JSONB格式
	tagsJSON := "[]"
	if len(image.Tags) > 0 {
		tagsJSON = `["` + strings.Join(image.Tags, `","`) + `"]`
	}

	result := database.DB.WithContext(ctx).Exec(query, image.ID, image.Description, tagsJSON, image.UpdatedAt)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("image not found")
	}
	return nil
}

// Delete 删除图片（软删除）
// id: 图片UUID
// 返回: 如果删除失败或图片不存在则返回错误
func (r *ImageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE images SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	result := database.DB.WithContext(ctx).Exec(query, time.Now(), id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("image not found")
	}
	return nil
}

// List 获取图片列表（分页、筛选、搜索）
// ctx: 上下文
// query: 图片查询条件
// 返回: 图片列表、总数，如果查询失败则返回错误
func (r *ImageRepository) List(ctx context.Context, query models.ImageQuery) ([]*models.Image, int64, error) {
	var images []*models.Image
	var total int64

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}
	offset := (query.Page - 1) * query.PageSize

	// 构建查询条件
	where := []string{"deleted_at IS NULL"}
	args := []interface{}{}

	if query.UploaderID != nil {
		where = append(where, "uploader_id = ?")
		args = append(args, *query.UploaderID)
	}

	if query.Search != "" {
		where = append(where, "to_tsvector('english', coalesce(filename, '') || ' ' || coalesce(description, '')) @@ to_tsquery('english', ?)")
		searchTerms := strings.Fields(query.Search)
		tsQuery := strings.Join(searchTerms, " & ")
		args = append(args, tsQuery)
	}

	if query.Tag != "" {
		where = append(where, "tags @> ?::jsonb")
		tagJSON := `["` + query.Tag + `"]`
		args = append(args, tagJSON)
	}

	whereClause := strings.Join(where, " AND ")

	// 获取总数
	countQuery := "SELECT COUNT(*) FROM images WHERE " + whereClause
	err := database.DB.WithContext(ctx).Raw(countQuery, args...).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 排序
	var orderBy string
	allowedSortFields := map[string]string{
		"id":          "id",
		"filename":    "filename",
		"created_at":  "created_at",
		"updated_at":  "updated_at",
		"size":        "size",
	}

	allowedOrders := map[string]bool{
		"asc":  true,
		"desc": true,
	}

	if query.SortBy != "" {
		sortField, ok := allowedSortFields[query.SortBy]
		if !ok {
			orderBy = "created_at DESC"
		} else {
			order := strings.ToLower(query.Order)
			if !allowedOrders[order] {
				order = "desc"
			}
			orderBy = sortField + " " + strings.ToUpper(order)
		}
	} else {
		orderBy = "created_at DESC"
	}

	// 获取列表
	listQuery := `
		SELECT id, filename, original_name, path, url, mime_type, size, width, height,
		       uploader_id, description, tags, created_at, updated_at
		FROM images
		WHERE ` + whereClause + `
		ORDER BY ` + orderBy + `
		LIMIT ? OFFSET ?
	`

	args = append(args, query.PageSize, offset)
	err = database.DB.WithContext(ctx).Raw(listQuery, args...).Scan(&images).Error
	if err != nil {
		return nil, 0, err
	}

	// 加载关联数据
	for _, image := range images {
		if err := r.loadImageRelations(ctx, image); err != nil {
			return nil, 0, err
		}
	}

	return images, total, nil
}

// loadImageRelations 加载图片关联数据（上传者信息）
func (r *ImageRepository) loadImageRelations(ctx context.Context, image *models.Image) error {
	var uploader models.User
	err := database.DB.WithContext(ctx).Raw(
		"SELECT id, username, email, avatar FROM users WHERE id = $1",
		image.UploaderID,
	).Scan(&uploader).Error
	if err == nil {
		image.Uploader = &uploader
	}
	return nil
}

