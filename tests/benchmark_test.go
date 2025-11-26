package tests

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
)

var (
	router      *gin.Engine
	dbAvailable bool
)

func init() {
	// 初始化测试环境
	config.Load()
	logger.Init("error", "")

	// 初始化数据库（如果失败，则在基准测试中跳过相关用例）
	if err := database.Init(); err == nil {
		dbAvailable = true
	}

	gin.SetMode(gin.TestMode)
	router = gin.New()
	router.Use(middleware.LoggerMiddleware())

	// 初始化依赖（使用模拟数据，实际测试需要真实数据库）
	jwtMgr := jwt.NewJWTManager("test-secret", 24*60*60*1000000000)
	userRepo := repository.NewUserRepository()
	articleRepo := repository.NewArticleRepository()
	categoryRepo := repository.NewCategoryRepository()
	tagRepo := repository.NewTagRepository()
	// 如需在基准测试中覆盖评论接口，再初始化 commentRepo / commentService / commentHandler

	userService := services.NewUserService(userRepo, jwtMgr)
	articleService := services.NewArticleService(articleRepo, categoryRepo, tagRepo)
	categoryService := services.NewCategoryService(categoryRepo)
	tagService := services.NewTagService(tagRepo)
	// commentService := services.NewCommentService(commentRepo, articleRepo)

	userHandler := handlers.NewUserHandler(userService, jwtMgr)
	articleHandler := handlers.NewArticleHandler(articleService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	tagHandler := handlers.NewTagHandler(tagService)
	// commentHandler := handlers.NewCommentHandler(commentService)

	api := router.Group("/api/v1")
	{
		api.POST("/auth/register", userHandler.Register)
		api.POST("/auth/login", userHandler.Login)
		api.GET("/articles", articleHandler.List)
		api.GET("/categories", categoryHandler.List)
		api.GET("/tags", tagHandler.List)
	}
}

func BenchmarkUserRegister(b *testing.B) {
	if !dbAvailable {
		b.Skip("Database not available")
	}
	reqBody := models.UserCreate{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	jsonData, _ := json.Marshal(reqBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
	}
}

func BenchmarkArticleList(b *testing.B) {
	if !dbAvailable {
		b.Skip("Database not available")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/articles?page=1&page_size=10", nil)
		router.ServeHTTP(w, req)
	}
}

func BenchmarkCategoryList(b *testing.B) {
	if !dbAvailable {
		b.Skip("Database not available")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/categories", nil)
		router.ServeHTTP(w, req)
	}
}

func BenchmarkTagList(b *testing.B) {
	if !dbAvailable {
		b.Skip("Database not available")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/tags", nil)
		router.ServeHTTP(w, req)
	}
}

// 并发测试
func BenchmarkConcurrentArticleList(b *testing.B) {
	if !dbAvailable {
		b.Skip("Database not available")
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/articles?page=1&page_size=10", nil)
			router.ServeHTTP(w, req)
		}
	})
}
