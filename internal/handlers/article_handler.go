package handlers

import (
	"net/http"

	"enterprise-blog/internal/models"
	"enterprise-blog/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ArticleHandler struct {
	articleService *services.ArticleService
}

func NewArticleHandler(articleService *services.ArticleService) *ArticleHandler {
	return &ArticleHandler{
		articleService: articleService,
	}
}

func (h *ArticleHandler) Create(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.Error(401, "unauthorized"))
		return
	}

	var req models.ArticleCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, err.Error()))
		return
	}

	article, err := h.articleService.Create(userID.(uuid.UUID), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, models.Success(article))
}

func (h *ArticleHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, "invalid article id"))
		return
	}

	article, err := h.articleService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error(404, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(article))
}

func (h *ArticleHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	
	article, err := h.articleService.GetBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error(404, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(article))
}

func (h *ArticleHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, "invalid article id"))
		return
	}

	var req models.ArticleUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, err.Error()))
		return
	}

	article, err := h.articleService.Update(id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(article))
}

func (h *ArticleHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, "invalid article id"))
		return
	}

	if err := h.articleService.Delete(id); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(nil))
}

func (h *ArticleHandler) Like(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, "invalid article id"))
		return
	}

	if err := h.articleService.Like(id); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(nil))
}

func (h *ArticleHandler) List(c *gin.Context) {
	var query models.ArticleQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, err.Error()))
		return
	}

	// 支持通过 search_mode=es 切换到 Elasticsearch 搜索
	searchMode := c.Query("search_mode")
	if searchMode == "es" && query.Search != "" {
		articles, total, err := h.articleService.SearchWithElasticsearch(query.Search, query.Page, query.PageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.Error(500, err.Error()))
			return
		}
		c.JSON(http.StatusOK, models.Paginated(articles, query.Page, query.PageSize, total))
		return
	}

	articles, total, err := h.articleService.List(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Paginated(articles, query.Page, query.PageSize, total))
}

// AdminList 管理后台文章列表：包含所有状态、支持按作者/状态/搜索过滤
func (h *ArticleHandler) AdminList(c *gin.Context) {
	var query models.ArticleQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, err.Error()))
		return
	}

	articles, total, err := h.articleService.List(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Paginated(articles, query.Page, query.PageSize, total))
}

// AdminGetByID 管理后台查看文章详情（与公开详情相同，预留后续扩展）
func (h *ArticleHandler) AdminGetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, "invalid article id"))
		return
	}

	article, err := h.articleService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error(404, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(article))
}

// AdminUpdateStatus 管理后台修改文章状态（草稿/发布/归档）
func (h *ArticleHandler) AdminUpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, "invalid article id"))
		return
	}

	var payload struct {
		Status models.ArticleStatus `json:"status"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, err.Error()))
		return
	}

	req := models.ArticleUpdate{
		Status: &payload.Status,
	}

	article, err := h.articleService.Update(id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(article))
}

// AdminDelete 管理后台删除文章（复用已有删除逻辑）
func (h *ArticleHandler) AdminDelete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, "invalid article id"))
		return
	}

	if err := h.articleService.Delete(id); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(nil))
}

