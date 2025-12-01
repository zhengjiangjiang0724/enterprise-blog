package models

import (
	"database/sql/driver"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleEditor UserRole = "editor"
	RoleAuthor UserRole = "author"
	RoleReader UserRole = "reader"
)

type User struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	Username  string     `json:"username" db:"username"`
	Email     string     `json:"email" db:"email"`
	Password  string     `json:"-" db:"password"`
	Role      UserRole   `json:"role" db:"role"`
	Avatar    string     `json:"avatar" db:"avatar"`
	Bio       string     `json:"bio" db:"bio"`
	Status    string     `json:"status" db:"status"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type UserCreate struct {
	Username string   `json:"username" validate:"required,min=3,max=50"`
	Email    string   `json:"email" validate:"required,email"`
	Password string   `json:"password" validate:"required,min=6"`
	Role     UserRole `json:"role"`
}

type UserUpdate struct {
	Username *string   `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Email    *string   `json:"email,omitempty" validate:"omitempty,email"`
	Role     *UserRole `json:"role,omitempty"`
	Avatar   *string   `json:"avatar,omitempty"`
	Bio      *string   `json:"bio,omitempty"`
	Status   *string   `json:"status,omitempty"`
}

type UserLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (r UserRole) Value() (driver.Value, error) {
	return string(r), nil
}

func (r *UserRole) Scan(value interface{}) error {
	if value == nil {
		*r = RoleReader
		return nil
	}
	if str, ok := value.(string); ok {
		*r = UserRole(str)
		return nil
	}
	return errors.New("cannot scan UserRole")
}
