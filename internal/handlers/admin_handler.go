package handlers

import (
	"net/http"

	"enterprise-blog/internal/config"
	"enterprise-blog/internal/database"
	"enterprise-blog/internal/models"

	"github.com/gin-gonic/gin"
)

// AdminHandler 提供仪表盘和系统配置等后台管理接口
type AdminHandler struct{}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

// AdminDashboardData 仪表盘统计数据
type AdminDashboardData struct {
	TotalUsers          int64 `json:"total_users"`
	TotalArticles       int64 `json:"total_articles"`
	PublishedArticles   int64 `json:"published_articles"`
	DraftArticles       int64 `json:"draft_articles"`
	ArchivedArticles    int64 `json:"archived_articles"`
	TotalComments       int64 `json:"total_comments"`
	TotalArticleViews   int64 `json:"total_article_views"`
	TotalArticleLikes   int64 `json:"total_article_likes"`
	TodayPublishedCount int64 `json:"today_published_count"`
}

// Dashboard 返回后台仪表盘核心统计
func (h *AdminHandler) Dashboard(c *gin.Context) {
	var data AdminDashboardData

	// 用户总数
	_ = database.DB.Raw("SELECT COUNT(*) FROM users WHERE deleted_at IS NULL").Scan(&data.TotalUsers).Error

	// 文章相关统计
	_ = database.DB.Raw("SELECT COUNT(*) FROM articles WHERE deleted_at IS NULL").Scan(&data.TotalArticles).Error
	_ = database.DB.Raw("SELECT COUNT(*) FROM articles WHERE status = 'published' AND deleted_at IS NULL").Scan(&data.PublishedArticles).Error
	_ = database.DB.Raw("SELECT COUNT(*) FROM articles WHERE status = 'draft' AND deleted_at IS NULL").Scan(&data.DraftArticles).Error
	_ = database.DB.Raw("SELECT COUNT(*) FROM articles WHERE status = 'archived' AND deleted_at IS NULL").Scan(&data.ArchivedArticles).Error
	_ = database.DB.Raw("SELECT COALESCE(SUM(view_count), 0) FROM articles WHERE deleted_at IS NULL").Scan(&data.TotalArticleViews).Error
	_ = database.DB.Raw("SELECT COALESCE(SUM(like_count), 0) FROM articles WHERE deleted_at IS NULL").Scan(&data.TotalArticleLikes).Error
	_ = database.DB.Raw("SELECT COUNT(*) FROM articles WHERE status = 'published' AND deleted_at IS NULL AND DATE(published_at) = CURRENT_DATE").Scan(&data.TodayPublishedCount).Error

	// 评论总数
	_ = database.DB.Raw("SELECT COUNT(*) FROM comments WHERE deleted_at IS NULL").Scan(&data.TotalComments).Error

	c.JSON(http.StatusOK, models.Success(data))
}

// SystemConfigInfo 对外暴露的系统配置（脱敏）
type SystemConfigInfo struct {
	Server struct {
		Host string `json:"host"`
		Port string `json:"port"`
		Mode string `json:"mode"`
	} `json:"server"`
	Database struct {
		Host               string `json:"host"`
		Port               string `json:"port"`
		User               string `json:"user"`
		Name               string `json:"name"`
		MaxOpenConns       int    `json:"max_open_conns"`
		MaxIdleConns       int    `json:"max_idle_conns"`
		ConnMaxLifetimeMin int    `json:"conn_max_lifetime_minutes"`
	} `json:"database"`
	Redis struct {
		Host string `json:"host"`
		Port string `json:"port"`
		DB   int    `json:"db"`
	} `json:"redis"`
	JWT struct {
		ExpireHours int `json:"expire_hours"`
	} `json:"jwt"`
	Log struct {
		Level string `json:"level"`
		File  string `json:"file"`
	} `json:"log"`
	Upload struct {
		Dir     string   `json:"dir"`
		MaxSize int64    `json:"max_size"`
		Exts    []string `json:"exts"`
	} `json:"upload"`
}

// SystemConfig 返回当前运行时的系统配置（只读，敏感字段已脱敏）
func (h *AdminHandler) SystemConfig(c *gin.Context) {
	cfg := config.AppConfig
	if cfg == nil {
		c.JSON(http.StatusInternalServerError, models.Error(500, "config not loaded"))
		return
	}

	var info SystemConfigInfo
	info.Server.Host = cfg.Server.Host
	info.Server.Port = cfg.Server.Port
	info.Server.Mode = cfg.Server.Mode

	info.Database.Host = cfg.Database.Host
	info.Database.Port = cfg.Database.Port
	info.Database.User = cfg.Database.User
	info.Database.Name = cfg.Database.Name
	info.Database.MaxOpenConns = cfg.Database.MaxOpenConns
	info.Database.MaxIdleConns = cfg.Database.MaxIdleConns
	info.Database.ConnMaxLifetimeMin = cfg.Database.ConnMaxLifetimeMinutes

	info.Redis.Host = cfg.Redis.Host
	info.Redis.Port = cfg.Redis.Port
	info.Redis.DB = cfg.Redis.DB

	info.JWT.ExpireHours = cfg.JWT.ExpireHours

	info.Log.Level = cfg.Log.Level
	info.Log.File = cfg.Log.File

	info.Upload.Dir = cfg.Upload.Dir
	info.Upload.MaxSize = cfg.Upload.MaxSize
	info.Upload.Exts = cfg.Upload.AllowedExts

	c.JSON(http.StatusOK, models.Success(info))
}


