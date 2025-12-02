package services

import (
	"errors"
	"fmt"
	"strings"

	"enterprise-blog/internal/models"
	"enterprise-blog/internal/repository"
	"enterprise-blog/pkg/jwt"

	"github.com/google/uuid"
)

type UserService struct {
	userRepo *repository.UserRepository
	jwtMgr   *jwt.JWTManager
}

func NewUserService(userRepo *repository.UserRepository, jwtMgr *jwt.JWTManager) *UserService {
	return &UserService{
		userRepo: userRepo,
		jwtMgr:   jwtMgr,
	}
}

func (s *UserService) Register(req *models.UserCreate) (*models.User, error) {
	// 检查邮箱是否已存在
	_, err := s.userRepo.GetByEmail(req.Email)
	if err == nil {
		return nil, errors.New("email already exists")
	}

	// 检查用户名是否已存在
	_, err = s.userRepo.GetByUsername(req.Username)
	if err == nil {
		return nil, errors.New("username already exists")
	}

	// 创建用户
	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	}

	if user.Role == "" {
		user.Role = models.RoleReader
	}

	// 加密密码
	if err := user.HashPassword(); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 保存用户
	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 清除密码
	user.Password = ""
	return user, nil
}

func (s *UserService) Login(req *models.UserLogin) (string, *models.User, error) {
	// 获取用户
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return "", nil, errors.New("invalid email or password")
	}

	// 验证密码
	if !user.CheckPassword(req.Password) {
		return "", nil, errors.New("invalid email or password")
	}

	// 检查用户状态
	if user.Status != "active" {
		return "", nil, errors.New("user account is not active")
	}

	// 生成JWT token
	token, err := s.jwtMgr.GenerateToken(user.ID, user.Username, string(user.Role))
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// 清除密码
	user.Password = ""
	return token, user, nil
}

func (s *UserService) GetByID(id uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	user.Password = ""
	return user, nil
}

func (s *UserService) Update(id uuid.UUID, req *models.UserUpdate) (*models.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Username != nil {
		// 检查用户名是否已被其他用户使用
		existing, err := s.userRepo.GetByUsername(*req.Username)
		if err == nil && existing.ID != id {
			return nil, errors.New("username already exists")
		}
		user.Username = *req.Username
	}

	if req.Email != nil {
		// 检查邮箱是否已被其他用户使用
		existing, err := s.userRepo.GetByEmail(*req.Email)
		if err == nil && existing.ID != id {
			return nil, errors.New("email already exists")
		}
		user.Email = *req.Email
	}

	if req.Role != nil {
		user.Role = *req.Role
	}

	if req.Avatar != nil {
		user.Avatar = *req.Avatar
	}

	if req.Bio != nil {
		user.Bio = *req.Bio
	}

	if req.Status != nil {
		user.Status = *req.Status
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	user.Password = ""
	return user, nil
}

func (s *UserService) Delete(id uuid.UUID) error {
	return s.userRepo.Delete(id)
}

// ChangePassword 修改用户密码（需提供旧密码）
func (s *UserService) ChangePassword(id uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return err
	}
	// 校验旧密码
	if !user.CheckPassword(oldPassword) {
		return errors.New("invalid old password")
	}
	// 设置新密码并哈希
	user.Password = newPassword
	if err := user.HashPassword(); err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	return s.userRepo.UpdatePassword(id, user.Password)
}

func (s *UserService) List(page, pageSize int) ([]*models.User, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	users, total, err := s.userRepo.List(page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 清除所有用户的密码
	for _, user := range users {
		user.Password = ""
	}

	return users, total, nil
}

func GenerateSlug(text string) string {
	// 简单的slug生成，生产环境应使用更完善的实现
	slug := strings.ToLower(text)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	return slug
}
