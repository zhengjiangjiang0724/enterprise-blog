// Package models 定义数据模型
package models

import (
	"time"

	"github.com/google/uuid"
)

// Image 图片模型
type Image struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Filename    string     `json:"filename" db:"filename"`
	OriginalName string    `json:"original_name" db:"original_name"`
	Path        string     `json:"path" db:"path"`           // 存储路径（相对路径或URL）
	URL         string     `json:"url" db:"url"`             // 访问URL
	MimeType    string     `json:"mime_type" db:"mime_type"`  // MIME类型
	Size        int64      `json:"size" db:"size"`            // 文件大小（字节）
	Width       int        `json:"width" db:"width"`           // 图片宽度（像素）
	Height      int        `json:"height" db:"height"`         // 图片高度（像素）
	UploaderID  uuid.UUID  `json:"uploader_id" db:"uploader_id"` // 上传者ID
	Uploader    *User      `json:"uploader,omitempty"`        // 上传者信息
	Description string     `json:"description" db:"description"` // 图片描述
	Tags        []string   `json:"tags" db:"tags"`            // 标签（JSON数组）
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ImageCreate 图片创建请求
type ImageCreate struct {
	Filename     string   `json:"filename" validate:"required"`
	OriginalName string   `json:"original_name" validate:"required"`
	Path         string   `json:"path" validate:"required"`
	URL          string   `json:"url" validate:"required"`
	MimeType     string   `json:"mime_type" validate:"required"`
	Size         int64    `json:"size" validate:"required,min=1"`
	Width        int      `json:"width" validate:"min=0"`
	Height       int      `json:"height" validate:"min=0"`
	Description  string   `json:"description"`
	Tags         []string `json:"tags"`
}

// ImageUpdate 图片更新请求
type ImageUpdate struct {
	Description *string   `json:"description,omitempty"`
	Tags        *[]string  `json:"tags,omitempty"`
}

// ImageQuery 图片查询条件
type ImageQuery struct {
	Page       int       `form:"page"`
	PageSize   int       `form:"page_size"`
	UploaderID *uuid.UUID `form:"uploader_id"`
	Search     string    `form:"search"` // 搜索文件名或描述
	Tag        string    `form:"tag"`    // 按标签筛选
	SortBy     string    `form:"sort_by"`
	Order      string    `form:"order"`
}

