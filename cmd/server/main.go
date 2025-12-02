package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"enterprise-blog/internal/config"
	"enterprise-blog/internal/database"
	"enterprise-blog/internal/handlers"
	"enterprise-blog/internal/middleware"
	"enterprise-blog/internal/repository"
	"enterprise-blog/internal/search"
	"enterprise-blog/internal/services"
	"enterprise-blog/pkg/jwt"
	"enterprise-blog/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	if err := config.Load(); err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 初始化日志
	if err := logger.Init(config.AppConfig.Log.Level, config.AppConfig.Log.File); err != nil {
		panic(fmt.Sprintf("Failed to init logger: %v", err))
	}

	// 初始化数据库
	if err := database.Init(); err != nil {
		l := logger.GetLogger()
		l.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer database.Close()

	// 初始化Redis
	if err := database.InitRedis(); err != nil {
		l := logger.GetLogger()
		l.Warn().Err(err).Msg("Failed to connect to redis, continuing without cache")
	} else {
		defer database.CloseRedis()
	}

	// 初始化 Elasticsearch（可选，失败仅记录日志）
	search.InitElasticsearch()

	// 初始化JWT管理器
	jwtMgr := jwt.NewJWTManager(config.AppConfig.JWT.Secret, config.AppConfig.JWT.ExpireDuration())

	// 初始化Repository
	userRepo := repository.NewUserRepository()
	articleRepo := repository.NewArticleRepository()
	categoryRepo := repository.NewCategoryRepository()
	tagRepo := repository.NewTagRepository()
	commentRepo := repository.NewCommentRepository()
	smsRepo := repository.NewSMSRepository()

	// 初始化Service
	userService := services.NewUserService(userRepo, jwtMgr)
	smsService := services.NewSMSService(smsRepo, userRepo)
	articleService := services.NewArticleService(articleRepo, categoryRepo, tagRepo)
	categoryService := services.NewCategoryService(categoryRepo)
	tagService := services.NewTagService(tagRepo)
	commentService := services.NewCommentService(commentRepo, articleRepo)

	// 初始化Handler
	userHandler := handlers.NewUserHandler(userService, smsService, jwtMgr)
	articleHandler := handlers.NewArticleHandler(articleService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	tagHandler := handlers.NewTagHandler(tagService)
	commentHandler := handlers.NewCommentHandler(commentService)
	adminHandler := handlers.NewAdminHandler()

	// 设置Gin模式
	gin.SetMode(config.AppConfig.Server.Mode)

	// 创建路由
	router := gin.New()

	// 中间件
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(gin.Recovery())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API路由组
	api := router.Group("/api/v1")
	{
		// 公开路由
		public := api.Group("")
		{
			// 用户认证
			public.POST("/auth/register", userHandler.Register)
			public.POST("/auth/login", userHandler.Login)
			public.POST("/auth/send-sms-code", userHandler.SendSMSCode)
			public.POST("/auth/login-phone", userHandler.LoginWithPhone)

			// 文章（公开访问）
			public.GET("/articles", articleHandler.List)
			public.GET("/articles/:id", articleHandler.GetByID)
			public.GET("/articles/slug/:slug", articleHandler.GetBySlug)
			public.POST("/articles/:id/like", articleHandler.Like)

			// 分类和标签
			public.GET("/categories", categoryHandler.List)
			public.GET("/tags", tagHandler.List)

			// 评论（使用文章 ID 路径参数 id，与 /articles/:id 保持一致）
			public.GET("/articles/:id/comments", commentHandler.GetByArticleID)
			public.POST("/articles/:id/comments", commentHandler.Create)
		}

		// 需要认证的路由
		authenticated := api.Group("")
		authenticated.Use(middleware.AuthMiddleware(jwtMgr))
		authenticated.Use(middleware.RateLimitMiddleware(100, time.Minute))
		{
			// 用户
			authenticated.GET("/users/profile", userHandler.GetProfile)
			authenticated.PUT("/users/profile", userHandler.UpdateProfile)
			authenticated.PUT("/users/password", userHandler.ChangePassword)

			// 文章（需要认证）
			authenticated.POST("/articles", articleHandler.Create)
			authenticated.PUT("/articles/:id", articleHandler.Update)
			authenticated.DELETE("/articles/:id", articleHandler.Delete)
		}

		// 管理员路由
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware(jwtMgr))
		admin.Use(middleware.RoleMiddleware("admin"))
		{
			// 仪表盘 & 系统配置
			admin.GET("/dashboard", adminHandler.Dashboard)
			admin.GET("/system/config", adminHandler.SystemConfig)

			admin.GET("/users", userHandler.ListUsers)
			admin.GET("/users/:id", userHandler.GetUser)
			admin.PUT("/users/:id", userHandler.AdminUpdateUser)
			// 管理后台文章管理
			admin.GET("/articles", articleHandler.AdminList)
			admin.GET("/articles/:id", articleHandler.AdminGetByID)
			admin.PUT("/articles/:id/status", articleHandler.AdminUpdateStatus)
			admin.DELETE("/articles/:id", articleHandler.AdminDelete)

			// 管理后台分类与标签管理
			admin.GET("/categories", categoryHandler.List)
			admin.POST("/categories", categoryHandler.Create)
			admin.GET("/categories/:id", categoryHandler.GetByID)
			admin.PUT("/categories/:id", categoryHandler.Update)
			admin.DELETE("/categories/:id", categoryHandler.Delete)

			admin.GET("/tags", tagHandler.List)
			admin.POST("/tags", tagHandler.Create)
			admin.GET("/tags/:id", tagHandler.GetByID)
			admin.PUT("/tags/:id", tagHandler.Update)
			admin.DELETE("/tags/:id", tagHandler.Delete)
		}
	}

	// 启动服务器
	addr := fmt.Sprintf("%s:%s", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// 启动 Redis 计数回刷 goroutine（view_count / like_count）
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_ = services.FlushArticleCountersFromRedis(ctx)
			cancel()
		}
	}()

	// 优雅关闭
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l := logger.GetLogger()
			l.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	l := logger.GetLogger()
	l.Info().Str("address", addr).Msg("Server started")

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	l2 := logger.GetLogger()
	l2.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		l3 := logger.GetLogger()
		l3.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	l4 := logger.GetLogger()
	l4.Info().Msg("Server exited")
}
