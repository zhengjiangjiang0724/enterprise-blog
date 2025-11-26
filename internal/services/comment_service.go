package services

import (
	"enterprise-blog/internal/models"
	"enterprise-blog/internal/repository"

	"github.com/google/uuid"
)

type CommentService struct {
	commentRepo  *repository.CommentRepository
	articleRepo  *repository.ArticleRepository
}

func NewCommentService(
	commentRepo *repository.CommentRepository,
	articleRepo *repository.ArticleRepository,
) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		articleRepo: articleRepo,
	}
}

func (s *CommentService) Create(userID *uuid.UUID, ip string, req *models.CommentCreate) (*models.Comment, error) {
	// 验证文章是否存在
	_, err := s.articleRepo.GetByID(req.ArticleID)
	if err != nil {
		return nil, err
	}

	comment := &models.Comment{
		ArticleID: req.ArticleID,
		UserID:    userID,
		ParentID:  req.ParentID,
		Content:   req.Content,
		Author:    req.Author,
		Email:     req.Email,
		Website:   req.Website,
		IP:        ip,
		Status:    "pending", // 默认待审核
	}

	if err := s.commentRepo.Create(comment); err != nil {
		return nil, err
	}

	// 更新文章评论数
	go s.commentRepo.IncrementCommentCount(req.ArticleID)

	return s.commentRepo.GetByID(comment.ID)
}

func (s *CommentService) GetByArticleID(articleID uuid.UUID, page, pageSize int) ([]*models.Comment, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return s.commentRepo.GetByArticleID(articleID, page, pageSize)
}

func (s *CommentService) Update(id uuid.UUID, req *models.CommentUpdate) (*models.Comment, error) {
	comment, err := s.commentRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Content != nil {
		comment.Content = *req.Content
	}

	if req.Status != nil {
		comment.Status = *req.Status
	}

	if err := s.commentRepo.Update(comment); err != nil {
		return nil, err
	}

	return s.commentRepo.GetByID(id)
}

func (s *CommentService) Delete(id uuid.UUID) error {
	return s.commentRepo.Delete(id)
}

