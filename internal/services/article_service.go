package services

import (
	"errors"
	"fmt"

	"enterprise-blog/internal/models"
	"enterprise-blog/internal/repository"

	"github.com/google/uuid"
)

type ArticleService struct {
	articleRepo  *repository.ArticleRepository
	categoryRepo *repository.CategoryRepository
	tagRepo      *repository.TagRepository
}

func NewArticleService(
	articleRepo *repository.ArticleRepository,
	categoryRepo *repository.CategoryRepository,
	tagRepo *repository.TagRepository,
) *ArticleService {
	return &ArticleService{
		articleRepo:  articleRepo,
		categoryRepo: categoryRepo,
		tagRepo:      tagRepo,
	}
}

func (s *ArticleService) Create(authorID uuid.UUID, req *models.ArticleCreate) (*models.Article, error) {
	// 验证分类
	if req.CategoryID != nil {
		_, err := s.categoryRepo.GetByID(*req.CategoryID)
		if err != nil {
			return nil, errors.New("category not found")
		}
	}

	// 验证标签
	var tags []models.Tag
	if len(req.TagIDs) > 0 {
		for _, tagID := range req.TagIDs {
			tag, err := s.tagRepo.GetByID(tagID)
			if err != nil {
				return nil, fmt.Errorf("tag %s not found", tagID)
			}
			tags = append(tags, *tag)
		}
	}

	// 生成slug
	slug := GenerateSlug(req.Title)
	// 检查slug是否已存在，如果存在则添加数字后缀
	originalSlug := slug
	counter := 1
	for {
		_, err := s.articleRepo.GetBySlug(slug)
		if err != nil {
			break // slug不存在，可以使用
		}
		slug = fmt.Sprintf("%s-%d", originalSlug, counter)
		counter++
	}

	// 生成摘要
	excerpt := req.Excerpt
	if excerpt == "" && len(req.Content) > 200 {
		excerpt = req.Content[:200] + "..."
	} else if excerpt == "" {
		excerpt = req.Content
	}

	article := &models.Article{
		Title:      req.Title,
		Slug:       slug,
		Content:    req.Content,
		Excerpt:    excerpt,
		CoverImage: req.CoverImage,
		Status:     req.Status,
		AuthorID:   authorID,
		CategoryID: req.CategoryID,
		Tags:       tags,
	}

	if article.Status == "" {
		article.Status = models.StatusDraft
	}

	if err := s.articleRepo.Create(article); err != nil {
		return nil, fmt.Errorf("failed to create article: %w", err)
	}

	return s.articleRepo.GetByID(article.ID)
}

func (s *ArticleService) GetByID(id uuid.UUID) (*models.Article, error) {
	return s.articleRepo.GetByID(id)
}

func (s *ArticleService) GetBySlug(slug string) (*models.Article, error) {
	article, err := s.articleRepo.GetBySlug(slug)
	if err != nil {
		return nil, err
	}

	// 增加浏览次数
	go s.articleRepo.IncrementViewCount(article.ID)

	return article, nil
}

func (s *ArticleService) Update(id uuid.UUID, req *models.ArticleUpdate) (*models.Article, error) {
	article, err := s.articleRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		article.Title = *req.Title
		// 如果标题改变，更新slug
		article.Slug = GenerateSlug(*req.Title)
	}

	if req.Content != nil {
		article.Content = *req.Content
		// 如果内容改变但没有摘要，自动生成摘要
		if req.Excerpt == nil {
			if len(*req.Content) > 200 {
				article.Excerpt = (*req.Content)[:200] + "..."
			} else {
				article.Excerpt = *req.Content
			}
		}
	}

	if req.Excerpt != nil {
		article.Excerpt = *req.Excerpt
	}

	if req.CoverImage != nil {
		article.CoverImage = *req.CoverImage
	}

	if req.Status != nil {
		article.Status = *req.Status
	}

	if req.CategoryID != nil {
		if *req.CategoryID != uuid.Nil {
			_, err := s.categoryRepo.GetByID(*req.CategoryID)
			if err != nil {
				return nil, errors.New("category not found")
			}
		}
		article.CategoryID = req.CategoryID
	}

	if len(req.TagIDs) > 0 {
		var tags []models.Tag
		for _, tagID := range req.TagIDs {
			tag, err := s.tagRepo.GetByID(tagID)
			if err != nil {
				return nil, fmt.Errorf("tag %s not found", tagID)
			}
			tags = append(tags, *tag)
		}
		article.Tags = tags
	}

	if err := s.articleRepo.Update(article); err != nil {
		return nil, err
	}

	return s.articleRepo.GetByID(id)
}

func (s *ArticleService) Delete(id uuid.UUID) error {
	return s.articleRepo.Delete(id)
}

func (s *ArticleService) List(query models.ArticleQuery) ([]*models.Article, int64, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	return s.articleRepo.List(query)
}

