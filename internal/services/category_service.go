// Package services 提供业务逻辑层的服务实现
package services

import (
	"enterprise-blog/internal/models"
	"enterprise-blog/internal/repository"

	"github.com/google/uuid"
)

// CategoryService 分类服务，提供分类相关的业务逻辑
type CategoryService struct {
	categoryRepo *repository.CategoryRepository
}

// NewCategoryService 创建新的分类服务实例
// categoryRepo: 分类数据访问层仓库
func NewCategoryService(categoryRepo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
	}
}

// Create 创建新分类
// req: 分类创建请求，包含名称、描述、父分类ID、排序等
// 返回: 创建成功的分类对象，如果创建失败则返回错误
func (s *CategoryService) Create(req *models.CategoryCreate) (*models.Category, error) {
	category := &models.Category{
		Name:        req.Name,
		Slug:        GenerateSlug(req.Name),
		Description: req.Description,
		ParentID:    req.ParentID,
		Order:       req.Order,
	}

	if err := s.categoryRepo.Create(category); err != nil {
		return nil, err
	}

	return s.categoryRepo.GetByID(category.ID)
}

// GetByID 根据ID获取分类详情
// id: 分类UUID
// 返回: 分类对象，如果不存在则返回错误
func (s *CategoryService) GetByID(id uuid.UUID) (*models.Category, error) {
	return s.categoryRepo.GetByID(id)
}

// List 获取所有分类列表
// 返回: 分类列表，如果查询失败则返回错误
func (s *CategoryService) List() ([]*models.Category, error) {
	return s.categoryRepo.List()
}

// Update 更新分类信息
// id: 分类UUID
// req: 分类更新请求，包含可选的名称、描述、父分类ID、排序等
// 返回: 更新后的分类对象，如果更新失败则返回错误
func (s *CategoryService) Update(id uuid.UUID, req *models.CategoryUpdate) (*models.Category, error) {
	category, err := s.categoryRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		category.Name = *req.Name
		category.Slug = GenerateSlug(*req.Name)
	}

	if req.Description != nil {
		category.Description = *req.Description
	}

	if req.ParentID != nil {
		category.ParentID = req.ParentID
	}

	if req.Order != nil {
		category.Order = *req.Order
	}

	if err := s.categoryRepo.Update(category); err != nil {
		return nil, err
	}

	return s.categoryRepo.GetByID(id)
}

// Delete 删除分类（软删除）
// id: 分类UUID
// 返回: 如果删除失败则返回错误
func (s *CategoryService) Delete(id uuid.UUID) error {
	return s.categoryRepo.Delete(id)
}

