package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
	testRouter *gin.Engine
	testJWT    *jwt.JWTManager
)

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

func TestUserRegister(t *testing.T) {
	reqBody := models.UserCreate{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Role:     models.RoleReader,
	}

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])
}

func TestUserLogin(t *testing.T) {
	// 先注册一个用户
	registerReq := models.UserCreate{
		Username: "loginuser",
		Email:    "login@example.com",
		Password: "password123",
		Role:     models.RoleReader,
	}
	registerData, _ := json.Marshal(registerReq)
	registerHTTPReq, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(registerData))
	registerHTTPReq.Header.Set("Content-Type", "application/json")
	registerW := httptest.NewRecorder()
	testRouter.ServeHTTP(registerW, registerHTTPReq)
	require.Equal(t, http.StatusOK, registerW.Code)

	// 测试登录
	loginReq := models.UserLogin{
		Email:    "login@example.com",
		Password: "password123",
	}
	loginData, _ := json.Marshal(loginReq)
	loginHTTPReq, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginData))
	loginHTTPReq.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, loginHTTPReq)

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

func TestCreateArticle_Authorized(t *testing.T) {
	// 先注册并登录获取token
	registerReq := models.UserCreate{
		Username: "articleuser",
		Email:    "article@example.com",
		Password: "password123",
		Role:     models.RoleAuthor,
	}
	registerData, _ := json.Marshal(registerReq)
	registerHTTPReq, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(registerData))
	registerHTTPReq.Header.Set("Content-Type", "application/json")
	registerW := httptest.NewRecorder()
	testRouter.ServeHTTP(registerW, registerHTTPReq)
	require.Equal(t, http.StatusOK, registerW.Code)

	// 登录获取token
	loginReq := models.UserLogin{
		Email:    "article@example.com",
		Password: "password123",
	}
	loginData, _ := json.Marshal(loginReq)
	loginHTTPReq, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginData))
	loginHTTPReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	testRouter.ServeHTTP(loginW, loginHTTPReq)
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

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])
}

