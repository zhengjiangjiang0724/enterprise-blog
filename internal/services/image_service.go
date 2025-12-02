// Package services 提供业务逻辑层的服务实现
//
// 设计原则：
// 1. 单一职责：每个服务只负责一个业务领域
// 2. 依赖注入：通过构造函数注入依赖，便于测试和扩展
// 3. 错误处理：统一返回错误，由上层处理
// 4. 事务管理：在 Service 层协调多个 Repository 操作
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
//
// 职责：
// - 图片上传（文件验证、存储、元数据管理）
// - 图片查询（列表、详情、搜索）
// - 图片更新（描述、标签）
// - 图片删除（软删除 + 文件删除）
//
// 设计考虑：
// - 文件存储：本地文件系统（生产环境可改为对象存储如 S3、OSS）
// - 元数据存储：PostgreSQL 数据库
// - 文件命名：使用 UUID 避免文件名冲突
// - 错误处理：文件操作失败时清理已创建的文件
type ImageService struct {
	imageRepo *repository.ImageRepository // 图片数据访问层
	uploadDir string                      // 图片上传目录路径
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
//
// 参数说明：
// - ctx: 上下文，用于控制请求超时和取消
// - uploaderID: 上传者用户UUID，用于记录图片归属
// - file: 上传的文件（multipart.FileHeader，来自 HTTP 请求）
// - description: 图片描述（可选），用于搜索和管理
// - tags: 图片标签列表（可选），用于分类和搜索
//
// 返回值：
// - *models.Image: 上传成功的图片对象（包含完整信息）
// - error: 如果上传失败则返回错误
//
// 功能流程：
// 1. 验证文件类型（MIME类型白名单）
// 2. 验证文件大小（防止过大文件）
// 3. 生成唯一文件名（UUID + 原始扩展名）
// 4. 保存文件到本地文件系统
// 5. 读取图片尺寸（宽度、高度）
// 6. 构建访问URL（相对路径）
// 7. 保存图片元数据到数据库
// 8. 返回完整图片对象
//
// 安全考虑：
// - 文件类型验证：只允许图片格式，防止上传恶意文件
// - 文件大小限制：防止DoS攻击
// - 文件名生成：使用UUID避免文件名冲突和路径遍历攻击
//
// 错误处理：
// - 如果文件保存失败，删除已创建的文件
// - 如果数据库保存失败，删除已上传的文件（保证数据一致性）
//
// 面试要点：
// - 为什么使用UUID作为文件名？避免文件名冲突，提高安全性
// - 如何处理并发上传？UUID保证唯一性，文件系统操作是原子的
// - 如何保证数据一致性？使用事务或失败时清理已创建的文件
func (s *ImageService) Upload(ctx context.Context, uploaderID uuid.UUID, file *multipart.FileHeader, description string, tags []string) (*models.Image, error) {
	// 步骤1：验证文件类型（MIME类型白名单）
	// 只允许常见的图片格式，防止上传恶意文件（如可执行文件）
	// 生产环境应该从配置文件读取允许的类型
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

	// 步骤2：验证文件大小
	// 限制文件大小，防止DoS攻击和存储空间浪费
	// TODO: 应该从 config.AppConfig.Upload.MaxSize 读取，这里使用常量简化
	const maxSize = 10 * 1024 * 1024 // 10MB
	if file.Size > maxSize {
		return nil, errors.New("image size exceeds 10MB limit")
	}

	// 步骤3：打开上传的文件
	// multipart.FileHeader 需要调用 Open() 方法获取文件内容
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close() // 确保文件句柄关闭，释放资源

	// 步骤4：生成唯一文件名
	// 使用 UUID 作为文件名前缀，避免文件名冲突
	// 保留原始文件的扩展名，便于识别文件类型
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(s.uploadDir, filename) // 构建完整文件路径

	// 步骤5：创建目标文件并保存
	// 使用 os.Create 创建新文件，如果文件已存在会被覆盖（UUID保证唯一性，不会发生）
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close() // 确保文件句柄关闭

	// 使用 io.Copy 将上传的文件内容复制到目标文件
	// 这是高效的文件复制方式，支持大文件
	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(filePath) // 如果复制失败，删除已创建的文件（清理资源）
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// 步骤6：读取图片尺寸
	// 需要重新打开文件，因为之前的文件句柄已经关闭
	// 使用 image.DecodeConfig 只解码图片配置（尺寸），不加载完整图片到内存，性能更好
	imgFile, err := os.Open(filePath)
	if err != nil {
		os.Remove(filePath) // 如果打开失败，删除已保存的文件
		return nil, fmt.Errorf("failed to read image: %w", err)
	}
	defer imgFile.Close()

	img, _, err := image.DecodeConfig(imgFile)
	if err != nil {
		os.Remove(filePath) // 如果解码失败，说明不是有效的图片文件，删除文件
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// TODO: 生成缩略图（可选功能）
	// 实际生产环境可以使用第三方库如 github.com/nfnt/resize 或 imagemagick
	// 缩略图可以用于列表展示，减少带宽和加载时间

	// 步骤7：构建访问URL
	// 使用相对路径，前端会根据基础URL拼接完整地址
	// 生产环境应该配置完整的CDN地址或对象存储URL
	url := fmt.Sprintf("/uploads/images/%s", filename)

	// 步骤8：创建图片记录并保存到数据库
	// 保存图片的元数据：文件名、原始文件名、路径、URL、MIME类型、尺寸、上传者等
	image := &models.Image{
		Filename:     filename,     // 存储的文件名（UUID）
		OriginalName: file.Filename, // 原始文件名（用户上传时的文件名）
		Path:         filePath,      // 文件系统路径（用于删除文件）
		URL:          url,           // 访问URL（用于前端显示）
		MimeType:     mimeType,      // MIME类型（用于HTTP响应头）
		Size:         file.Size,     // 文件大小（字节）
		Width:        img.Width,     // 图片宽度（像素）
		Height:       img.Height,    // 图片高度（像素）
		UploaderID:   uploaderID,    // 上传者ID（用于权限控制）
		Description:  description,   // 图片描述（用于搜索）
		Tags:         tags,          // 图片标签（用于分类和搜索）
	}

	// 保存到数据库
	if err := s.imageRepo.Create(ctx, image); err != nil {
		os.Remove(filePath) // 如果数据库保存失败，删除已上传的文件（保证数据一致性）
		return nil, fmt.Errorf("failed to save image record: %w", err)
	}

	// 重新从数据库获取完整数据（包含关联信息，如上传者信息）
	// 这样可以确保返回的数据是最新的，包含数据库自动生成的字段（如ID、创建时间等）
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
//
// 参数说明：
// - ctx: 上下文，用于控制请求超时和取消
// - id: 要删除的图片UUID
//
// 返回值：
// - error: 如果删除失败则返回错误
//
// 功能流程：
// 1. 从数据库获取图片信息（获取文件路径）
// 2. 删除文件系统中的文件
// 3. 软删除数据库记录（设置 deleted_at 字段）
//
// 设计考虑：
// - 软删除：数据库记录不真正删除，只标记为已删除，可以恢复
// - 文件删除：物理删除文件，释放存储空间
// - 错误处理：即使文件删除失败，也继续执行数据库删除（避免数据不一致）
//
// 面试要点：
// - 为什么使用软删除？可以恢复误删的数据，保留审计记录
// - 如何处理文件删除失败？记录警告日志，继续执行数据库删除，避免数据不一致
func (s *ImageService) Delete(ctx context.Context, id uuid.UUID) error {
	// 先获取图片信息，获取文件路径
	image, err := s.imageRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 删除文件系统中的文件
	// 如果文件删除失败，记录警告日志但不中断流程
	// 这样可以避免文件删除失败导致数据库记录无法删除的情况
	if err := os.Remove(image.Path); err != nil {
		l := logger.GetLogger()
		l.Warn().Err(err).Str("path", image.Path).Msg("failed to delete image file")
		// 继续执行数据库删除，即使文件删除失败
	}

	// 软删除数据库记录（设置 deleted_at 字段）
	// 软删除的好处：可以恢复数据，保留审计记录
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

