package models

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	ArticleID uuid.UUID  `json:"article_id" db:"article_id"`
	UserID    *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	User      *User      `json:"user,omitempty"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"`
	Content   string     `json:"content" db:"content"`
	Author    string     `json:"author" db:"author"`
	Email     string     `json:"email" db:"email"`
	Website   string     `json:"website" db:"website"`
	IP        string     `json:"ip" db:"ip"`
	Status    string     `json:"status" db:"status"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

type CommentCreate struct {
	ArticleID uuid.UUID  `json:"article_id" validate:"required"`
	ParentID  *uuid.UUID `json:"parent_id"`
	Content   string     `json:"content" validate:"required,min=1"`
	Author    string     `json:"author" validate:"required"`
	Email     string     `json:"email" validate:"required,email"`
	Website   string     `json:"website"`
}

type CommentUpdate struct {
	Content *string `json:"content,omitempty" validate:"omitempty,min=1"`
	Status  *string `json:"status,omitempty"`
}

