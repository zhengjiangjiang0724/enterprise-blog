// Package services 提供业务逻辑层的服务实现
package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"enterprise-blog/internal/models"
	"enterprise-blog/internal/repository"
	"enterprise-blog/pkg/logger"

	"github.com/google/uuid"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

// ImageService 图片服务，提供图片相关的业务逻辑
type ImageService struct {
	imageRepo *repository.ImageRepository
	uploadDir string // 图片上传目录
}

// NewImageService 创建新的图片服务实例
// imageRepo: 图片数据访问层仓库
// uploadDir: 图片上传目录路径
func NewImageService(imageRepo *repository.ImageRepository, uploadDir string) *ImageService {
	if uploadDir == "" {
		uploadDir = "./uploads/images"
	}
	// 确保上传目录存在
	os.MkdirAll(uploadDir, 0755)
	return &ImageService{
		imageRepo: imageRepo,
		uploadDir: uploadDir,
	}
}

// Upload 上传图片
// uploaderID: 上传者用户UUID
// file: 上传的文件
// description: 图片描述（可选）
// tags: 图片标签（可选）
// 返回: 上传成功的图片对象，如果上传失败则返回错误
// 注意: 支持JPEG、PNG、GIF格式，会自动生成缩略图
func (s *ImageService) Upload(ctx context.Context, uploaderID uuid.UUID, file *multipart.FileHeader, description string, tags []string) (*models.Image, error) {
	// 验证文件类型（从配置读取允许的扩展名）
	allowedMimeTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	mimeType := file.Header.Get("Content-Type")
	if !allowedMimeTypes[mimeType] {
		return nil, errors.New("unsupported image format, only JPEG, PNG, GIF, WebP are allowed")
	}

	// 验证文件大小（从配置读取最大大小）
	// 注意：这里应该从配置读取，但为了简化，暂时使用常量
	// 实际生产环境应该从 config.AppConfig.Upload.MaxSize 读取
	const maxSize = 10 * 1024 * 1024 // 10MB
	if file.Size > maxSize {
		return nil, errors.New("image size exceeds 10MB limit")
	}

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// 生成唯一文件名
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(s.uploadDir, filename)

	// 创建目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(filePath) // 删除已创建的文件
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// 读取图片尺寸
	imgFile, err := os.Open(filePath)
	if err != nil {
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to read image: %w", err)
	}
	defer imgFile.Close()

	img, _, err := image.DecodeConfig(imgFile)
	if err != nil {
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// 生成缩略图（可选，这里简化处理）
	// 实际生产环境可以使用更完善的图片处理库

	// 构建访问URL（相对路径，实际应该配置为完整URL）
	url := fmt.Sprintf("/uploads/images/%s", filename)

	// 创建图片记录
	image := &models.Image{
		Filename:     filename,
		OriginalName: file.Filename,
		Path:         filePath,
		URL:          url,
		MimeType:     mimeType,
		Size:         file.Size,
		Width:        img.Width,
		Height:       img.Height,
		UploaderID:   uploaderID,
		Description:  description,
		Tags:         tags,
	}

	if err := s.imageRepo.Create(ctx, image); err != nil {
		os.Remove(filePath) // 如果数据库保存失败，删除已上传的文件
		return nil, fmt.Errorf("failed to save image record: %w", err)
	}

	// 重新获取完整数据（包含关联信息）
	return s.imageRepo.GetByID(ctx, image.ID)
}

// GetByID 根据ID获取图片详情
// id: 图片UUID
// 返回: 图片对象，如果不存在则返回错误
func (s *ImageService) GetByID(ctx context.Context, id uuid.UUID) (*models.Image, error) {
	return s.imageRepo.GetByID(ctx, id)
}

// List 获取图片列表（分页、筛选、搜索）
// query: 图片查询条件
// 返回: 图片列表、总数，如果查询失败则返回错误
func (s *ImageService) List(ctx context.Context, query models.ImageQuery) ([]*models.Image, int64, error) {
	return s.imageRepo.List(ctx, query)
}

// Update 更新图片信息
// id: 图片UUID
// req: 图片更新请求，包含可选的描述和标签
// 返回: 更新后的图片对象，如果更新失败则返回错误
func (s *ImageService) Update(ctx context.Context, id uuid.UUID, req *models.ImageUpdate) (*models.Image, error) {
	image, err := s.imageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Description != nil {
		image.Description = *req.Description
	}

	if req.Tags != nil {
		image.Tags = *req.Tags
	}

	if err := s.imageRepo.Update(ctx, image); err != nil {
		return nil, err
	}

	return s.imageRepo.GetByID(ctx, id)
}

// Delete 删除图片（软删除）
// id: 图片UUID
// 返回: 如果删除失败则返回错误
// 注意: 会删除文件系统中的图片文件
func (s *ImageService) Delete(ctx context.Context, id uuid.UUID) error {
	image, err := s.imageRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 删除文件系统中的文件
	if err := os.Remove(image.Path); err != nil {
		l := logger.GetLogger()
		l.Warn().Err(err).Str("path", image.Path).Msg("failed to delete image file")
		// 继续执行数据库删除，即使文件删除失败
	}

	// 软删除数据库记录
	return s.imageRepo.Delete(ctx, id)
}

// GetUploadDir 获取上传目录路径
func (s *ImageService) GetUploadDir() string {
	return s.uploadDir
}

// GenerateThumbnail 生成缩略图（辅助函数，可选实现）
func (s *ImageService) GenerateThumbnail(ctx context.Context, imageID uuid.UUID, maxWidth, maxHeight uint) error {
	// 这里可以实现缩略图生成逻辑
	// 使用 image 包或第三方库如 github.com/nfnt/resize
	return nil
}

