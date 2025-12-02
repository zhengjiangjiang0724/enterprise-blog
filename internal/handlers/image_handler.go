// Package handlers 提供HTTP处理器
package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"enterprise-blog/internal/models"
	"enterprise-blog/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ImageHandler 图片处理器
type ImageHandler struct {
	imageService *services.ImageService
}

// NewImageHandler 创建新的图片处理器实例
func NewImageHandler(imageService *services.ImageService) *ImageHandler {
	return &ImageHandler{
		imageService: imageService,
	}
}

// Upload 上传图片
// POST /api/v1/images/upload
// 需要认证
// Content-Type: multipart/form-data
// 表单字段:
//   - file: 图片文件（必需）
//   - description: 图片描述（可选）
//   - tags: 图片标签，逗号分隔（可选）
func (h *ImageHandler) Upload(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.Error(401, "unauthorized"))
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, "file is required"))
		return
	}

	// 获取描述和标签
	description := c.PostForm("description")
	tagsStr := c.PostForm("tags")
	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}

	// 上传图片
	image, err := h.imageService.Upload(c.Request.Context(), userID.(uuid.UUID), file, description, tags)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, models.Success(image))
}

// GetByID 根据ID获取图片详情
// GET /api/v1/images/:id
func (h *ImageHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, "invalid image id"))
		return
	}

	image, err := h.imageService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error(404, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(image))
}

// List 获取图片列表
// GET /api/v1/images?page=1&page_size=20&uploader_id=xxx&search=keyword&tag=tag1
func (h *ImageHandler) List(c *gin.Context) {
	var query models.ImageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, err.Error()))
		return
	}

	images, total, err := h.imageService.List(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Paginated(images, query.Page, query.PageSize, total))
}

// Update 更新图片信息
// PUT /api/v1/images/:id
// 需要认证
func (h *ImageHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, "invalid image id"))
		return
	}

	// 检查权限：只能更新自己上传的图片
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.Error(401, "unauthorized"))
		return
	}

	// 验证图片所有权
	image, err := h.imageService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error(404, err.Error()))
		return
	}

	// 非管理员只能更新自己上传的图片
	if roleVal, ok := c.Get("role"); ok {
		roleStr, _ := roleVal.(string)
		if roleStr != string(models.RoleAdmin) && image.UploaderID != userID.(uuid.UUID) {
			c.JSON(http.StatusForbidden, models.Error(403, "forbidden: can only update your own images"))
			return
		}
	}

	var req models.ImageUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, err.Error()))
		return
	}

	updated, err := h.imageService.Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(updated))
}

// Delete 删除图片
// DELETE /api/v1/images/:id
// 需要认证
func (h *ImageHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, "invalid image id"))
		return
	}

	// 检查权限：只能删除自己上传的图片
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.Error(401, "unauthorized"))
		return
	}

	// 验证图片所有权
	image, err := h.imageService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error(404, err.Error()))
		return
	}

	// 非管理员只能删除自己上传的图片
	if roleVal, ok := c.Get("role"); ok {
		roleStr, _ := roleVal.(string)
		if roleStr != string(models.RoleAdmin) && image.UploaderID != userID.(uuid.UUID) {
			c.JSON(http.StatusForbidden, models.Error(403, "forbidden: can only delete your own images"))
			return
		}
	}

	if err := h.imageService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(nil))
}

// ServeImage 提供图片文件服务
// GET /uploads/images/:filename
// 用于直接访问上传的图片文件
func (h *ImageHandler) ServeImage(c *gin.Context) {
	filename := c.Param("filename")
	
	// 安全检查：防止路径遍历攻击
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		c.JSON(http.StatusBadRequest, models.Error(400, "invalid filename"))
		return
	}

	// 构建文件路径
	filePath := filepath.Join(h.imageService.GetUploadDir(), filename)
	
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, models.Error(404, "image not found"))
		return
	}

	// 返回文件
	c.File(filePath)
}

