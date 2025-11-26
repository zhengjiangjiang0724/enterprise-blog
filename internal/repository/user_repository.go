package repository

import (
	"errors"
	"time"

	"enterprise-blog/internal/database"
	"enterprise-blog/internal/models"

	"github.com/google/uuid"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (id, username, email, password, role, avatar, bio, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`
	
	now := time.Now()
	user.ID = uuid.New()
	user.CreatedAt = now
	user.UpdatedAt = now
	
	if user.Status == "" {
		user.Status = "active"
	}
	if user.Role == "" {
		user.Role = models.RoleReader
	}

	row := database.DB.Raw(
		query,
		user.ID, user.Username, user.Email, user.Password, user.Role,
		user.Avatar, user.Bio, user.Status, user.CreatedAt, user.UpdatedAt,
	).Row()
	return row.Scan(&user.ID)
}

func (r *UserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, email, password, role, avatar, bio, status, created_at, updated_at, deleted_at
			  FROM users WHERE id = $1 AND deleted_at IS NULL`
	
	result := database.DB.Raw(query, id).Scan(user)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, email, password, role, avatar, bio, status, created_at, updated_at, deleted_at
			  FROM users WHERE email = $1 AND deleted_at IS NULL`
	
	result := database.DB.Raw(query, email).Scan(user)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, email, password, role, avatar, bio, status, created_at, updated_at, deleted_at
			  FROM users WHERE username = $1 AND deleted_at IS NULL`
	
	result := database.DB.Raw(query, username).Scan(user)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *UserRepository) Update(user *models.User) error {
	query := `
		UPDATE users 
		SET username = $2, email = $3, role = $4, avatar = $5, bio = $6, status = $7, updated_at = $8
		WHERE id = $1 AND deleted_at IS NULL
	`
	
	user.UpdatedAt = time.Now()
	result := database.DB.Exec(query, user.ID, user.Username, user.Email, user.Role,
		user.Avatar, user.Bio, user.Status, user.UpdatedAt)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *UserRepository) Delete(id uuid.UUID) error {
	query := `UPDATE users SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	result := database.DB.Exec(query, time.Now(), id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *UserRepository) List(page, pageSize int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	offset := (page - 1) * pageSize

	// 获取总数
	countQuery := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`
	err := database.DB.Raw(countQuery).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	query := `SELECT id, username, email, role, avatar, bio, status, created_at, updated_at
			  FROM users WHERE deleted_at IS NULL
			  ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	
	err = database.DB.Raw(query, pageSize, offset).Scan(&users).Error
	return users, total, err
}

