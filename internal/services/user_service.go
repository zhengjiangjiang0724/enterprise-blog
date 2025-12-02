// Package services 提供业务逻辑层的服务实现
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

// UserService 用户服务，提供用户相关的业务逻辑
type UserService struct {
	userRepo *repository.UserRepository
	jwtMgr   *jwt.JWTManager
}

// NewUserService 创建新的用户服务实例
// userRepo: 用户数据访问层仓库
// jwtMgr: JWT管理器，用于生成和验证token
func NewUserService(userRepo *repository.UserRepository, jwtMgr *jwt.JWTManager) *UserService {
	return &UserService{
		userRepo: userRepo,
		jwtMgr:   jwtMgr,
	}
}

// Register 用户注册
// req: 用户注册请求，包含用户名、邮箱、密码等信息
// 返回: 注册成功的用户对象（密码已清除），如果注册失败则返回错误
// 注意: 会检查邮箱和用户名是否已存在，密码使用bcrypt加密存储
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

// Login 用户登录（邮箱密码方式）
// req: 用户登录请求，包含邮箱和密码
// 返回: JWT token、用户对象（密码已清除），如果登录失败则返回错误
// 注意: 会验证密码和用户状态，只有active状态的用户才能登录
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

// GetByID 根据ID获取用户详情
// id: 用户UUID
// 返回: 用户对象（密码已清除），如果不存在则返回错误
func (s *UserService) GetByID(id uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	user.Password = ""
	return user, nil
}

// Update 更新用户信息
// id: 用户UUID
// req: 用户更新请求，包含可选的用户名、邮箱、角色、头像、简介、状态等
// 返回: 更新后的用户对象（密码已清除），如果更新失败则返回错误
// 注意: 会检查用户名和邮箱是否已被其他用户使用
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

// Delete 删除用户（软删除）
// id: 用户UUID
// 返回: 如果删除失败则返回错误
func (s *UserService) Delete(id uuid.UUID) error {
	return s.userRepo.Delete(id)
}

// ChangePassword 修改用户密码（需提供旧密码进行验证）
// id: 用户UUID
// oldPassword: 当前密码，用于验证用户身份
// newPassword: 新密码，将使用bcrypt加密后存储
// 返回: 如果旧密码错误或更新失败则返回错误
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

// List 获取用户列表（分页）
// page: 页码，从1开始
// pageSize: 每页数量，最大100
// 返回: 用户列表、总数，如果查询失败则返回错误
// 注意: 返回的用户对象密码已清除
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

// GenerateSlug 生成URL友好的slug字符串
// text: 原始文本
// 返回: 转换后的slug（小写、空格和下划线替换为连字符）
// 注意: 这是简单的实现，生产环境建议使用更完善的slug生成库
func GenerateSlug(text string) string {
	slug := strings.ToLower(text)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	return slug
}
