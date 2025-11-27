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

	articles, total, err := h.articleService.List(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Paginated(articles, query.Page, query.PageSize, total))
}

