package models

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Slug        string     `json:"slug" db:"slug"`
	Description string     `json:"description" db:"description"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"`
	Order       int        `json:"order" db:"order"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type CategoryCreate struct {
	Name        string     `json:"name" validate:"required,min=1,max=100"`
	Description string     `json:"description"`
	ParentID    *uuid.UUID `json:"parent_id"`
	Order       int        `json:"order"`
}

type CategoryUpdate struct {
	Name        *string    `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string    `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	Order       *int       `json:"order,omitempty"`
}

