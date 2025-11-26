package services

import (
	"enterprise-blog/internal/models"
	"enterprise-blog/internal/repository"

	"github.com/google/uuid"
)

type TagService struct {
	tagRepo *repository.TagRepository
}

func NewTagService(tagRepo *repository.TagRepository) *TagService {
	return &TagService{
		tagRepo: tagRepo,
	}
}

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

func (s *TagService) GetByID(id uuid.UUID) (*models.Tag, error) {
	return s.tagRepo.GetByID(id)
}

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

func (s *TagService) Delete(id uuid.UUID) error {
	return s.tagRepo.Delete(id)
}

func (s *TagService) List() ([]*models.Tag, error) {
	return s.tagRepo.List()
}

