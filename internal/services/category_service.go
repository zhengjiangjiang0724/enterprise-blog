package services

import (
	"enterprise-blog/internal/models"
	"enterprise-blog/internal/repository"

	"github.com/google/uuid"
)

type CategoryService struct {
	categoryRepo *repository.CategoryRepository
}

func NewCategoryService(categoryRepo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
	}
}

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

func (s *CategoryService) GetByID(id uuid.UUID) (*models.Category, error) {
	return s.categoryRepo.GetByID(id)
}

func (s *CategoryService) List() ([]*models.Category, error) {
	return s.categoryRepo.List()
}

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

func (s *CategoryService) Delete(id uuid.UUID) error {
	return s.categoryRepo.Delete(id)
}

