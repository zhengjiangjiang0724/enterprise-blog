// Package services 提供业务逻辑层的服务实现
package services

import (
	"enterprise-blog/internal/models"
	"enterprise-blog/internal/repository"

	"github.com/google/uuid"
)

// CommentService 评论服务，提供评论相关的业务逻辑
type CommentService struct {
	commentRepo  *repository.CommentRepository
	articleRepo  *repository.ArticleRepository
}

// NewCommentService 创建新的评论服务实例
// commentRepo: 评论数据访问层仓库
// articleRepo: 文章数据访问层仓库，用于验证文章是否存在
func NewCommentService(
	commentRepo *repository.CommentRepository,
	articleRepo *repository.ArticleRepository,
) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		articleRepo: articleRepo,
	}
}

// Create 创建新评论
// userID: 登录用户ID（可选，游客评论时为nil）
// ip: 评论者IP地址，用于记录
// req: 评论创建请求，包含文章ID、内容、作者信息等
// 返回: 创建成功的评论对象，如果创建失败则返回错误
// 注意: 新评论默认状态为pending（待审核），会异步更新文章评论数
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

// GetByArticleID 获取指定文章下的评论列表（分页）
// articleID: 文章UUID
// page: 页码，从1开始
// pageSize: 每页数量，默认20
// 返回: 评论列表、总数，如果查询失败则返回错误
// 注意: 只返回父评论（parent_id为NULL的评论）
func (s *CommentService) GetByArticleID(articleID uuid.UUID, page, pageSize int) ([]*models.Comment, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return s.commentRepo.GetByArticleID(articleID, page, pageSize)
}

// Update 更新评论信息
// id: 评论UUID
// req: 评论更新请求，包含可选的内容和状态
// 返回: 更新后的评论对象，如果更新失败则返回错误
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

// Delete 删除评论（硬删除）
// id: 评论UUID
// 返回: 如果删除失败则返回错误
func (s *CommentService) Delete(id uuid.UUID) error {
	return s.commentRepo.Delete(id)
}

