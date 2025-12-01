package models

import (
	"database/sql/driver"
	"errors"
	"time"

	"github.com/google/uuid"
)

type ArticleStatus string

const (
	StatusDraft        ArticleStatus = "draft"
	StatusReview       ArticleStatus = "review"
	StatusPublished    ArticleStatus = "published"
	StatusArchived     ArticleStatus = "archived"
)

type Article struct {
	ID           uuid.UUID     `json:"id" db:"id"`
	Title        string        `json:"title" db:"title"`
	Slug         string        `json:"slug" db:"slug"`
	Content      string        `json:"content" db:"content"`
	Excerpt      string        `json:"excerpt" db:"excerpt"`
	CoverImage   string        `json:"cover_image" db:"cover_image"`
	Status       ArticleStatus `json:"status" db:"status"`
	AuthorID     uuid.UUID     `json:"author_id" db:"author_id"`
	Author       *User         `json:"author,omitempty"`
	CategoryID   *uuid.UUID    `json:"category_id,omitempty" db:"category_id"`
	Category     *Category     `json:"category,omitempty"`
	// Tags 由单独的查询加载，不通过 GORM 关系映射
	Tags         []Tag         `json:"tags,omitempty" gorm:"-"`
	ViewCount    int           `json:"view_count" db:"view_count"`
	LikeCount    int           `json:"like_count" db:"like_count"`
	CommentCount int           `json:"comment_count" db:"comment_count"`
	PublishedAt  *time.Time    `json:"published_at,omitempty" db:"published_at"`
	CreatedAt    time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time    `json:"deleted_at,omitempty" db:"deleted_at"`
}

type ArticleCreate struct {
	Title      string        `json:"title" validate:"required,min=1,max=200"`
	Content    string        `json:"content" validate:"required"`
	Excerpt    string        `json:"excerpt"`
	CoverImage string        `json:"cover_image"`
	Status     ArticleStatus `json:"status"`
	CategoryID *uuid.UUID    `json:"category_id"`
	TagIDs     []uuid.UUID   `json:"tag_ids"`
}

type ArticleUpdate struct {
	Title      *string        `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Content    *string        `json:"content,omitempty"`
	Excerpt    *string        `json:"excerpt,omitempty"`
	CoverImage *string        `json:"cover_image,omitempty"`
	Status     *ArticleStatus `json:"status,omitempty"`
	CategoryID *uuid.UUID     `json:"category_id,omitempty"`
	TagIDs     []uuid.UUID    `json:"tag_ids,omitempty"`
}

type ArticleQuery struct {
	Page       int           `form:"page"`
	PageSize   int           `form:"page_size"`
	Status     ArticleStatus `form:"status"`
	CategoryID *uuid.UUID    `form:"category_id"`
	TagID      *uuid.UUID    `form:"tag_id"`
	AuthorID   *uuid.UUID    `form:"author_id"`
	Search     string        `form:"search"`
	SortBy     string        `form:"sort_by"`
	Order      string        `form:"order"`
}

func (s ArticleStatus) Value() (driver.Value, error) {
	return string(s), nil
}

func (s *ArticleStatus) Scan(value interface{}) error {
	if value == nil {
		*s = StatusDraft
		return nil
	}
	if str, ok := value.(string); ok {
		*s = ArticleStatus(str)
		return nil
	}
	return errors.New("cannot scan ArticleStatus")
}
