package services

import (
	"fmt"
	"strings"

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
	// 生成slug
	slug := GenerateSlug(req.Title)
	if slug == "" {
		slug = "article"
	}
	// 检查slug是否已存在，如果存在则添加数字后缀
	originalSlug := slug
	counter := 1

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
	}

	if article.Status == "" {
		article.Status = models.StatusDraft
	}

	// 创建时如果遇到 slug 唯一约束冲突，则自动追加数字后缀重试几次
	const maxSlugRetries = 5
	for retries := 0; retries < maxSlugRetries; retries++ {
		article.Slug = slug

		if err := s.articleRepo.Create(article); err != nil {
			// 唯一约束冲突：尝试下一个 slug
			if isSlugUniqueViolation(err) {
				slug = fmt.Sprintf("%s-%d", originalSlug, counter)
				counter++
				continue
			}
			return nil, fmt.Errorf("failed to create article: %w", err)
		}

		// 创建成功，重新从数据库获取完整数据（含作者等关联）
		return s.articleRepo.GetByID(article.ID)
	}

	return nil, fmt.Errorf("failed to create article: slug already exists after %d retries", maxSlugRetries)
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

	// 简化：不再更新分类和标签，文章仅保留基本信息

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

func (s *ArticleService) Like(id uuid.UUID) error {
	return s.articleRepo.IncrementLikeCount(id)
}

// isSlugUniqueViolation 判断是否为 articles.slug 唯一约束冲突
func isSlugUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	// 目前使用字符串包含判断，兼容 pq/pgx 等驱动的错误文案
	msg := err.Error()
	return strings.Contains(msg, "duplicate key value violates unique constraint") &&
		strings.Contains(msg, "articles_slug_key")
}

