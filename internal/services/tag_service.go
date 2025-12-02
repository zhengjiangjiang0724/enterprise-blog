// Package services 提供业务逻辑层的服务实现
package services

import (
	"enterprise-blog/internal/models"
	"enterprise-blog/internal/repository"

	"github.com/google/uuid"
)

// TagService 标签服务，提供标签相关的业务逻辑
type TagService struct {
	tagRepo *repository.TagRepository
}

// NewTagService 创建新的标签服务实例
// tagRepo: 标签数据访问层仓库
func NewTagService(tagRepo *repository.TagRepository) *TagService {
	return &TagService{
		tagRepo: tagRepo,
	}
}

// Create 创建新标签
// req: 标签创建请求，包含名称和颜色
// 返回: 创建成功的标签对象，如果创建失败则返回错误
func (s *TagService) Create(req *models.TagCreate) (*models.Tag, error) {
	tag := &models.Tag{
		Name:  req.Name,
		Slug:  GenerateSlug(req.Name),
		Color: req.Color,
	}

	if err := s.tagRepo.Create(tag); err != nil {
		return nil, err
	}

	return s.tagRepo.GetByID(tag.ID)
}

// GetByID 根据ID获取标签详情
// id: 标签UUID
// 返回: 标签对象，如果不存在则返回错误
func (s *TagService) GetByID(id uuid.UUID) (*models.Tag, error) {
	return s.tagRepo.GetByID(id)
}

// Update 更新标签信息
// id: 标签UUID
// req: 标签更新请求，包含可选的名称和颜色
// 返回: 更新后的标签对象，如果更新失败则返回错误
func (s *TagService) Update(id uuid.UUID, req *models.TagUpdate) (*models.Tag, error) {
	tag, err := s.tagRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		tag.Name = *req.Name
		tag.Slug = GenerateSlug(*req.Name)
	}

	if req.Color != nil {
		tag.Color = *req.Color
	}

	if err := s.tagRepo.Update(tag); err != nil {
		return nil, err
	}

	return s.tagRepo.GetByID(id)
}

// Delete 删除标签（软删除）
// id: 标签UUID
// 返回: 如果删除失败则返回错误
func (s *TagService) Delete(id uuid.UUID) error {
	return s.tagRepo.Delete(id)
}

// List 获取所有标签列表
// 返回: 标签列表，如果查询失败则返回错误
func (s *TagService) List() ([]*models.Tag, error) {
	return s.tagRepo.List()
}
