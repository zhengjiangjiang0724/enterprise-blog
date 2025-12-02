package unit

import (
	"errors"
	"testing"

	"enterprise-blog/internal/models"
	"enterprise-blog/pkg/jwt"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository 模拟用户仓库
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByPhone(phone string) (*models.User, error) {
	args := m.Called(phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdatePassword(id uuid.UUID, hashedPassword string) error {
	args := m.Called(id, hashedPassword)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) List(page, pageSize int) ([]*models.User, int64, error) {
	args := m.Called(page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.User), args.Get(1).(int64), args.Error(2)
}

// MockJWTManager 模拟JWT管理器
type MockJWTManager struct {
	mock.Mock
}

func (m *MockJWTManager) GenerateToken(userID uuid.UUID, username, role string) (string, error) {
	args := m.Called(userID, username, role)
	return args.String(0), args.Error(1)
}

func (m *MockJWTManager) ValidateToken(token string) (*jwt.Claims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Claims), args.Error(1)
}

func TestUserService_Register(t *testing.T) {
	tests := []struct {
		name    string
		req     *models.UserCreate
		setup   func(*MockUserRepository, *MockJWTManager)
		wantErr bool
		errMsg  string
	}{
		{
			name: "成功注册",
			req: &models.UserCreate{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
				Role:     models.RoleReader,
			},
			setup: func(mockRepo *MockUserRepository, mockJWT *MockJWTManager) {
				// 邮箱不存在
				mockRepo.On("GetByEmail", "test@example.com").Return(nil, errors.New("not found"))
				// 用户名不存在
				mockRepo.On("GetByUsername", "testuser").Return(nil, errors.New("not found"))
				// 创建成功
				mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "邮箱已存在",
			req: &models.UserCreate{
				Username: "testuser",
				Email:    "existing@example.com",
				Password: "password123",
			},
			setup: func(mockRepo *MockUserRepository, mockJWT *MockJWTManager) {
				// 邮箱已存在
				mockRepo.On("GetByEmail", "existing@example.com").Return(&models.User{}, nil)
			},
			wantErr: true,
			errMsg:  "email already exists",
		},
		{
			name: "用户名已存在",
			req: &models.UserCreate{
				Username: "existinguser",
				Email:    "test@example.com",
				Password: "password123",
			},
			setup: func(mockRepo *MockUserRepository, mockJWT *MockJWTManager) {
				// 邮箱不存在
				mockRepo.On("GetByEmail", "test@example.com").Return(nil, errors.New("not found"))
				// 用户名已存在
				mockRepo.On("GetByUsername", "existinguser").Return(&models.User{}, nil)
			},
			wantErr: true,
			errMsg:  "username already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			mockJWT := new(MockJWTManager)
			tt.setup(mockRepo, mockJWT)

			// 创建服务（需要适配实际的构造函数）
			// 这里需要根据实际的UserService结构来调整
			// userService := services.NewUserService(mockRepo, mockJWT)

			// result, err := userService.Register(tt.req)

			// if tt.wantErr {
			// 	assert.Error(t, err)
			// 	assert.Contains(t, err.Error(), tt.errMsg)
			// 	assert.Nil(t, result)
			// } else {
			// 	assert.NoError(t, err)
			// 	assert.NotNil(t, result)
			// 	assert.Equal(t, tt.req.Username, result.Username)
			// 	assert.Equal(t, tt.req.Email, result.Email)
			// 	assert.Empty(t, result.Password) // 密码应该被清除
			// }

			mockRepo.AssertExpectations(t)
			mockJWT.AssertExpectations(t)
		})
	}
}

func TestUserService_Login(t *testing.T) {
	userID := uuid.New()
	testUser := &models.User{Password: "password123"}
	testUser.HashPassword()
	hashedPassword := testUser.Password
	
	tests := []struct {
		name    string
		req     *models.UserLogin
		setup   func(*MockUserRepository, *MockJWTManager)
		wantErr bool
		errMsg  string
	}{
		{
			name: "成功登录",
			req: &models.UserLogin{
				Email:    "test@example.com",
				Password: "password123",
			},
			setup: func(mockRepo *MockUserRepository, mockJWT *MockJWTManager) {
				user := &models.User{
					ID:       userID,
					Email:    "test@example.com",
					Password: hashedPassword,
					Status:   "active",
					Role:     models.RoleReader,
				}
				mockRepo.On("GetByEmail", "test@example.com").Return(user, nil)
				mockJWT.On("GenerateToken", userID, user.Username, string(user.Role)).Return("test-token", nil)
			},
			wantErr: false,
		},
		{
			name: "用户不存在",
			req: &models.UserLogin{
				Email:    "notfound@example.com",
				Password: "password123",
			},
			setup: func(mockRepo *MockUserRepository, mockJWT *MockJWTManager) {
				mockRepo.On("GetByEmail", "notfound@example.com").Return(nil, errors.New("not found"))
			},
			wantErr: true,
			errMsg:  "invalid email or password",
		},
		{
			name: "密码错误",
			req: &models.UserLogin{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			setup: func(mockRepo *MockUserRepository, mockJWT *MockJWTManager) {
				user := &models.User{
					ID:       userID,
					Email:    "test@example.com",
					Password: hashedPassword,
					Status:   "active",
				}
				mockRepo.On("GetByEmail", "test@example.com").Return(user, nil)
			},
			wantErr: true,
			errMsg:  "invalid email or password",
		},
		{
			name: "用户状态非active",
			req: &models.UserLogin{
				Email:    "test@example.com",
				Password: "password123",
			},
			setup: func(mockRepo *MockUserRepository, mockJWT *MockJWTManager) {
				user := &models.User{
					ID:       userID,
					Email:    "test@example.com",
					Password: hashedPassword,
					Status:   "inactive",
				}
				mockRepo.On("GetByEmail", "test@example.com").Return(user, nil)
			},
			wantErr: true,
			errMsg:  "user account is not active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			mockJWT := new(MockJWTManager)
			tt.setup(mockRepo, mockJWT)

			// 注意：这里需要根据实际的UserService结构来调整
			// 由于UserService使用了具体的repository类型，我们需要使用接口或调整测试方式

			mockRepo.AssertExpectations(t)
			mockJWT.AssertExpectations(t)
		})
	}
}

// 辅助函数：创建测试用户
func createTestUser() *models.User {
	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Role:     models.RoleReader,
		Status:   "active",
	}
	return user
}

// 辅助函数：验证用户密码
func TestUserPasswordHashing(t *testing.T) {
	user := &models.User{
		Password: "testpassword123",
	}

	err := user.HashPassword()
	assert.NoError(t, err)
	assert.NotEqual(t, "testpassword123", user.Password) // 密码应该被哈希
	assert.True(t, user.CheckPassword("testpassword123")) // 应该能验证正确密码
	assert.False(t, user.CheckPassword("wrongpassword"))  // 应该拒绝错误密码
}

