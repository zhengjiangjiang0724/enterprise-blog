/**
 * @package integration
 * @description 集成测试包，测试完整的 API 端点，包括用户认证、文章管理等核心功能。
 *
 * @remarks
 * - 使用真实的数据库和 Redis 连接（测试环境）
 * - 测试完整的请求-响应流程
 * - 验证 API 端点的正确性和业务逻辑
 * - 使用时间戳生成唯一测试数据，避免数据冲突
 * - 接受 200 OK 或 201 Created 状态码（符合 RESTful 规范）
 *
 * @test_strategy
 * - 每个测试使用唯一的邮箱和用户名（基于时间戳），避免重复注册错误
 * - 测试覆盖：用户注册、登录、文章列表、文章创建（授权/未授权）
 * - 验证 HTTP 状态码、响应格式、业务逻辑正确性
 *
 * @interview_points
 * - 如何避免测试数据冲突？（使用时间戳生成唯一标识）
 * - 为什么接受 200 和 201 状态码？（RESTful 规范：创建资源返回 201）
 * - 集成测试与单元测试的区别？（集成测试使用真实依赖，单元测试使用 mock）
 * - 如何确保测试的独立性？（每个测试使用唯一数据，不依赖其他测试）
 */
package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"enterprise-blog/internal/config"
	"enterprise-blog/internal/database"
	"enterprise-blog/internal/handlers"
	"enterprise-blog/internal/middleware"
	"enterprise-blog/internal/models"
	"enterprise-blog/internal/repository"
	"enterprise-blog/internal/services"
	"enterprise-blog/pkg/jwt"
	"enterprise-blog/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testRouter *gin.Engine     // 测试用的 Gin 路由实例
	testJWT    *jwt.JWTManager // 测试用的 JWT 管理器
)

/**
 * @function setupTestRouter
 * @description 初始化测试路由，设置测试环境。
 *
 * @remarks
 * - 设置 Gin 为测试模式
 * - 初始化数据库连接（使用测试数据库配置）
 * - 创建所有必要的 Repository、Service、Handler 实例
 * - 设置 API 路由（公开路由和需要认证的路由）
 * - 在 TestMain 中调用，确保所有测试共享同一个路由实例
 *
 * @interview_points
 * - 为什么在 TestMain 中初始化路由？（避免重复初始化，提高测试效率）
 * - 测试数据库如何配置？（通过环境变量或配置文件）
 * - 如何模拟认证中间件？（使用真实的 JWT 管理器，生成有效的 token）
 */
func setupTestRouter() {
	gin.SetMode(gin.TestMode)
	logger.Init("error", "")

	// 初始化数据库（使用测试数据库）
	config.Load()
	if err := database.Init(); err != nil {
		panic(err)
	}

	// 初始化JWT
	testJWT = jwt.NewJWTManager("test-secret-key-for-integration-tests", 3600*1000000000)

	// 初始化Repository
	userRepo := repository.NewUserRepository()
	articleRepo := repository.NewArticleRepository()
	categoryRepo := repository.NewCategoryRepository()
	tagRepo := repository.NewTagRepository()
	commentRepo := repository.NewCommentRepository()
	smsRepo := repository.NewSMSRepository()

	// 初始化Service
	userService := services.NewUserService(userRepo, testJWT)
	smsService := services.NewSMSService(smsRepo, userRepo)
	articleService := services.NewArticleService(articleRepo, categoryRepo, tagRepo)
	categoryService := services.NewCategoryService(categoryRepo)
	tagService := services.NewTagService(tagRepo)
	commentService := services.NewCommentService(commentRepo, articleRepo)

	// 初始化Handler
	userHandler := handlers.NewUserHandler(userService, smsService, testJWT)
	articleHandler := handlers.NewArticleHandler(articleService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	tagHandler := handlers.NewTagHandler(tagService)
	commentHandler := handlers.NewCommentHandler(commentService)

	// 创建路由
	testRouter = gin.New()
	testRouter.Use(middleware.LoggerMiddleware())
	testRouter.Use(middleware.CORSMiddleware())

	api := testRouter.Group("/api/v1")
	{
		// 公开路由
		public := api.Group("")
		{
			public.POST("/auth/register", userHandler.Register)
			public.POST("/auth/login", userHandler.Login)
			public.GET("/articles", articleHandler.List)
			public.GET("/categories", categoryHandler.List)
			public.GET("/tags", tagHandler.List)
		}

		// 需要认证的路由
		authenticated := api.Group("")
		authenticated.Use(middleware.AuthMiddleware(testJWT))
		{
			authenticated.GET("/users/profile", userHandler.GetProfile)
			authenticated.PUT("/users/profile", userHandler.UpdateProfile)
			authenticated.POST("/articles", articleHandler.Create)
			authenticated.GET("/articles/:id", articleHandler.GetByID)
			authenticated.PUT("/articles/:id", articleHandler.Update)
			authenticated.DELETE("/articles/:id", articleHandler.Delete)
			authenticated.POST("/articles/:id/like", articleHandler.Like)
			authenticated.POST("/articles/:id/comments", commentHandler.Create)
		}
	}
}

func TestMain(m *testing.M) {
	setupTestRouter()
	m.Run()
}

/**
 * @function TestUserRegister
 * @description 测试用户注册功能。
 *
 * @test_cases
 * - 成功注册：使用唯一邮箱和用户名，验证返回 200/201 状态码
 * - 数据唯一性：使用时间戳生成唯一标识，避免多次运行测试时的数据冲突
 *
 * @remarks
 * - 使用时间戳生成唯一的邮箱和用户名，确保测试可以多次运行
 * - 接受 200 OK 或 201 Created 状态码（RESTful 规范：创建资源返回 201）
 * - 验证响应格式和业务逻辑正确性
 *
 * @interview_points
 * - 如何避免测试数据冲突？（使用时间戳或 UUID 生成唯一标识）
 * - 为什么接受 200 和 201？（RESTful 规范允许两种状态码）
 * - 如何验证注册成功？（检查状态码、响应格式、返回的用户数据）
 */
func TestUserRegister(t *testing.T) {
	// 使用唯一邮箱避免重复注册错误（多次运行测试时不会因为邮箱已存在而失败）
	timestamp := time.Now().UnixNano()
	reqBody := models.UserCreate{
		Username: fmt.Sprintf("testuser_%d", timestamp),
		Email:    fmt.Sprintf("test_%d@example.com", timestamp),
		Password: "password123",
		Role:     models.RoleReader,
	}

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	// 如果失败，打印响应体以便调试
	if w.Code != http.StatusOK && w.Code != http.StatusCreated {
		t.Logf("Response body: %s", w.Body.String())
	}

	// 注册操作返回 201 Created 是符合 RESTful 规范的（创建新资源）
	// 但某些实现可能返回 200 OK，两种都是可接受的
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated, "Expected 200 or 201, got %d", w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])
}

/**
 * @function TestUserLogin
 * @description 测试用户登录功能。
 *
 * @test_cases
 * - 先注册一个用户，然后测试登录
 * - 验证登录成功：返回 200 状态码和有效的 JWT token
 * - 验证 token 不为空
 *
 * @remarks
 * - 使用唯一邮箱注册测试用户，避免数据冲突
 * - 验证登录响应包含 token 和用户信息
 * - 接受 200 OK 或 201 Created 状态码（注册操作）
 *
 * @interview_points
 * - 如何测试需要前置条件的场景？（先注册用户，再测试登录）
 * - 如何验证 JWT token 的有效性？（检查 token 不为空，实际验证需要解析 token）
 */
func TestUserLogin(t *testing.T) {
	// 先注册一个用户（使用唯一邮箱，避免多次运行测试时的数据冲突）
	timestamp := time.Now().UnixNano()
	email := fmt.Sprintf("login_%d@example.com", timestamp)
	registerReq := models.UserCreate{
		Username: fmt.Sprintf("loginuser_%d", timestamp),
		Email:    email,
		Password: "password123",
		Role:     models.RoleReader,
	}
	registerData, _ := json.Marshal(registerReq)
	registerHTTPReq, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(registerData))
	registerHTTPReq.Header.Set("Content-Type", "application/json")
	registerW := httptest.NewRecorder()
	testRouter.ServeHTTP(registerW, registerHTTPReq)
	if registerW.Code != http.StatusOK && registerW.Code != http.StatusCreated {
		t.Logf("Register response body: %s", registerW.Body.String())
	}
	// 注册操作返回 201 Created 是符合 RESTful 规范的
	require.True(t, registerW.Code == http.StatusOK || registerW.Code == http.StatusCreated, "Expected 200 or 201, got %d", registerW.Code)

	// 测试登录
	loginReq := models.UserLogin{
		Email:    email,
		Password: "password123",
	}
	loginData, _ := json.Marshal(loginReq)
	loginHTTPReq, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginData))
	loginHTTPReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, loginHTTPReq)

	if w.Code != http.StatusOK {
		t.Logf("Login response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])

	// 验证返回了token
	data, ok := response["data"].(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, data["token"])
}

func TestGetArticlesList(t *testing.T) {
	req, _ := http.NewRequest("GET", "/api/v1/articles?page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])
}

func TestCreateArticle_Unauthorized(t *testing.T) {
	reqBody := models.ArticleCreate{
		Title:   "Test Article",
		Content: "Test content",
		Status:  models.StatusDraft,
	}

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/articles", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	// 应该返回401未授权
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

/**
 * @function TestCreateArticle_Authorized
 * @description 测试已授权用户创建文章功能。
 *
 * @test_cases
 * - 注册并登录获取 JWT token
 * - 使用 token 创建文章
 * - 验证文章创建成功：返回 200/201 状态码和文章数据
 *
 * @remarks
 * - 测试完整的认证流程：注册 -> 登录 -> 获取 token -> 使用 token 创建资源
 * - 验证 Authorization header 的正确使用
 * - 接受 200 OK 或 201 Created 状态码（创建资源）
 *
 * @interview_points
 * - 如何测试需要认证的 API？（先获取 token，然后在请求头中携带）
 * - 如何模拟认证流程？（注册 -> 登录 -> 使用 token）
 * - 如何验证创建操作成功？（检查状态码、响应格式、返回的数据）
 */
func TestCreateArticle_Authorized(t *testing.T) {
	// 先注册并登录获取token（使用唯一邮箱，避免数据冲突）
	timestamp := time.Now().UnixNano()
	email := fmt.Sprintf("article_%d@example.com", timestamp)
	registerReq := models.UserCreate{
		Username: fmt.Sprintf("articleuser_%d", timestamp),
		Email:    email,
		Password: "password123",
		Role:     models.RoleAuthor,
	}
	registerData, _ := json.Marshal(registerReq)
	registerHTTPReq, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(registerData))
	registerHTTPReq.Header.Set("Content-Type", "application/json")
	registerW := httptest.NewRecorder()
	testRouter.ServeHTTP(registerW, registerHTTPReq)
	if registerW.Code != http.StatusOK && registerW.Code != http.StatusCreated {
		t.Logf("Register response body: %s", registerW.Body.String())
	}
	// 注册操作返回 201 Created 是符合 RESTful 规范的
	require.True(t, registerW.Code == http.StatusOK || registerW.Code == http.StatusCreated, "Expected 200 or 201, got %d", registerW.Code)

	// 登录获取token
	loginReq := models.UserLogin{
		Email:    email,
		Password: "password123",
	}
	loginData, _ := json.Marshal(loginReq)
	loginHTTPReq, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginData))
	loginHTTPReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	testRouter.ServeHTTP(loginW, loginHTTPReq)
	if loginW.Code != http.StatusOK {
		t.Logf("Login response body: %s", loginW.Body.String())
	}
	require.Equal(t, http.StatusOK, loginW.Code)

	var loginResponse map[string]interface{}
	json.Unmarshal(loginW.Body.Bytes(), &loginResponse)
	loginDataMap := loginResponse["data"].(map[string]interface{})
	token := loginDataMap["token"].(string)

	// 使用token创建文章
	articleReq := models.ArticleCreate{
		Title:   "Test Article",
		Content: "Test content",
		Status:  models.StatusDraft,
	}
	articleData, _ := json.Marshal(articleReq)
	articleHTTPReq, _ := http.NewRequest("POST", "/api/v1/articles", bytes.NewBuffer(articleData))
	articleHTTPReq.Header.Set("Content-Type", "application/json")
	articleHTTPReq.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, articleHTTPReq)

	if w.Code != http.StatusOK && w.Code != http.StatusCreated {
		t.Logf("Create article response body: %s", w.Body.String())
	}
	// 创建文章操作返回 201 Created 是符合 RESTful 规范的
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated, "Expected 200 or 201, got %d", w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])
}
