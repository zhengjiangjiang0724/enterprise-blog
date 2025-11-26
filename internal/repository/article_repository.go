package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"enterprise-blog/internal/database"
	"enterprise-blog/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ArticleRepository struct{}

func NewArticleRepository() *ArticleRepository {
	return &ArticleRepository{}
}

func (r *ArticleRepository) Create(article *models.Article) error {
	tx := database.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	query := `
		INSERT INTO articles (id, title, slug, content, excerpt, cover_image, status, author_id, category_id, view_count, like_count, comment_count, published_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id
	`
	
	now := time.Now()
	article.ID = uuid.New()
	article.CreatedAt = now
	article.UpdatedAt = now
	if article.Status == models.StatusPublished && article.PublishedAt == nil {
		article.PublishedAt = &now
	}

	row := tx.Raw(
		query,
		article.ID, article.Title, article.Slug, article.Content, article.Excerpt,
		article.CoverImage, article.Status, article.AuthorID, article.CategoryID,
		article.ViewCount, article.LikeCount, article.CommentCount,
		article.PublishedAt, article.CreatedAt, article.UpdatedAt,
	).Row()
	if err := row.Scan(&article.ID); err != nil {
		return err
	}

	// 处理标签关联
	if len(article.Tags) > 0 {
		if err := r.setArticleTags(tx, article.ID, article.Tags); err != nil {
			return err
		}
	}

	return tx.Commit().Error
}

func (r *ArticleRepository) GetByID(id uuid.UUID) (*models.Article, error) {
	article := &models.Article{}
	query := `
		SELECT a.id, a.title, a.slug, a.content, a.excerpt, a.cover_image, a.status,
			   a.author_id, a.category_id, a.view_count, a.like_count, a.comment_count,
			   a.published_at, a.created_at, a.updated_at, a.deleted_at
		FROM articles a
		WHERE a.id = $1 AND a.deleted_at IS NULL
	`
	
	err := database.DB.Raw(query, id).Scan(article).Error
	if err == sql.ErrNoRows {
		return nil, errors.New("article not found")
	}
	if err != nil {
		return nil, err
	}

	// 加载作者信息
	if err := r.loadArticleRelations(article); err != nil {
		return nil, err
	}

	return article, nil
}

func (r *ArticleRepository) GetBySlug(slug string) (*models.Article, error) {
	article := &models.Article{}
	query := `
		SELECT a.id, a.title, a.slug, a.content, a.excerpt, a.cover_image, a.status,
			   a.author_id, a.category_id, a.view_count, a.like_count, a.comment_count,
			   a.published_at, a.created_at, a.updated_at, a.deleted_at
		FROM articles a
		WHERE a.slug = $1 AND a.deleted_at IS NULL
	`
	
	err := database.DB.Raw(query, slug).Scan(article).Error
	if err == sql.ErrNoRows {
		return nil, errors.New("article not found")
	}
	if err != nil {
		return nil, err
	}

	if err := r.loadArticleRelations(article); err != nil {
		return nil, err
	}

	return article, nil
}

func (r *ArticleRepository) Update(article *models.Article) error {
	tx := database.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	query := `
		UPDATE articles 
		SET title = $2, slug = $3, content = $4, excerpt = $5, cover_image = $6,
			status = $7, category_id = $8, updated_at = $9, published_at = $10
		WHERE id = $1 AND deleted_at IS NULL
	`
	
	article.UpdatedAt = time.Now()
	if article.Status == models.StatusPublished && article.PublishedAt == nil {
		now := time.Now()
		article.PublishedAt = &now
	}

	result := tx.Exec(query, article.ID, article.Title, article.Slug, article.Content,
		article.Excerpt, article.CoverImage, article.Status, article.CategoryID,
		article.UpdatedAt, article.PublishedAt)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("article not found")
	}

	// 更新标签关联
	if len(article.Tags) > 0 {
		// 删除旧关联
		if err := tx.Exec("DELETE FROM article_tags WHERE article_id = $1", article.ID).Error; err != nil {
			return err
		}
		// 创建新关联
		if err := r.setArticleTags(tx, article.ID, article.Tags); err != nil {
			return err
		}
	}

	return tx.Commit().Error
}

func (r *ArticleRepository) Delete(id uuid.UUID) error {
	query := `UPDATE articles SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	result := database.DB.Exec(query, time.Now(), id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("article not found")
	}
	return nil
}

func (r *ArticleRepository) List(query models.ArticleQuery) ([]*models.Article, int64, error) {
	var articles []*models.Article
	var total int64

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	offset := (query.Page - 1) * query.PageSize

	// 构建查询条件
	where := []string{"a.deleted_at IS NULL"}
	args := []interface{}{}
	argIndex := 1

	if query.Status != "" {
		where = append(where, fmt.Sprintf("a.status = $%d", argIndex))
		args = append(args, query.Status)
		argIndex++
	}

	if query.CategoryID != nil {
		where = append(where, fmt.Sprintf("a.category_id = $%d", argIndex))
		args = append(args, *query.CategoryID)
		argIndex++
	}

	if query.AuthorID != nil {
		where = append(where, fmt.Sprintf("a.author_id = $%d", argIndex))
		args = append(args, *query.AuthorID)
		argIndex++
	}

	if query.Search != "" {
		where = append(where, fmt.Sprintf("(a.title ILIKE $%d OR a.content ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+query.Search+"%")
		argIndex++
	}

	if query.TagID != nil {
		where = append(where, fmt.Sprintf("EXISTS (SELECT 1 FROM article_tags WHERE article_id = a.id AND tag_id = $%d)", argIndex))
		args = append(args, *query.TagID)
		argIndex++
	}

	whereClause := strings.Join(where, " AND ")

	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM articles a WHERE %s", whereClause)
	err := database.DB.Raw(countQuery, args...).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 排序
	orderBy := "a.created_at DESC"
	if query.SortBy != "" {
		if query.Order == "asc" {
			orderBy = fmt.Sprintf("a.%s ASC", query.SortBy)
		} else {
			orderBy = fmt.Sprintf("a.%s DESC", query.SortBy)
		}
	}

	// 获取列表
	listQuery := fmt.Sprintf(`
		SELECT a.id, a.title, a.slug, a.content, a.excerpt, a.cover_image, a.status,
			   a.author_id, a.category_id, a.view_count, a.like_count, a.comment_count,
			   a.published_at, a.created_at, a.updated_at
		FROM articles a
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIndex, argIndex+1)
	
	args = append(args, query.PageSize, offset)
	err = database.DB.Raw(listQuery, args...).Scan(&articles).Error
	if err != nil {
		return nil, 0, err
	}

	// 加载关联数据
	for _, article := range articles {
		if err := r.loadArticleRelations(article); err != nil {
			return nil, 0, err
		}
	}

	return articles, total, nil
}

func (r *ArticleRepository) IncrementViewCount(id uuid.UUID) error {
	query := `UPDATE articles SET view_count = view_count + 1 WHERE id = $1`
	return database.DB.Exec(query, id).Error
}

func (r *ArticleRepository) setArticleTags(tx *gorm.DB, articleID uuid.UUID, tags []models.Tag) error {
	if len(tags) == 0 {
		return nil
	}

	tagIDs := make([]uuid.UUID, len(tags))
	for i, tag := range tags {
		tagIDs[i] = tag.ID
	}

	query := `INSERT INTO article_tags (article_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	for _, tagID := range tagIDs {
		if err := tx.Exec(query, articleID, tagID).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *ArticleRepository) loadArticleRelations(article *models.Article) error {
	// 加载作者
	var author models.User
	err := database.DB.Raw("SELECT id, username, email, avatar FROM users WHERE id = $1", article.AuthorID).Scan(&author).Error
	if err == nil {
		article.Author = &author
	}

	// 加载分类
	if article.CategoryID != nil {
		var category models.Category
		err := database.DB.Raw("SELECT id, name, slug FROM categories WHERE id = $1", *article.CategoryID).Scan(&category).Error
		if err == nil {
			article.Category = &category
		}
	}

	// 加载标签
	var tags []models.Tag
	err = database.DB.Raw(`
		SELECT t.id, t.name, t.slug, t.color
		FROM tags t
		INNER JOIN article_tags at ON t.id = at.tag_id
		WHERE at.article_id = $1
	`, article.ID).Scan(&tags).Error
	if err == nil {
		article.Tags = tags
	}

	return nil
}

